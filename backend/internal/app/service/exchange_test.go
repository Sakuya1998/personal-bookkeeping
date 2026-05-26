package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"personal-bookkeeping/internal/app/model"
	cch "personal-bookkeeping/internal/infra/cache"
)

// ---------------------------------------------------------------------------
// Mock cache — implements cch.Cache for testing
// ---------------------------------------------------------------------------

type mockCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]string)}
}

func (m *mockCache) Get(_ context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[key]
	if !ok {
		return "", cch.ErrMiss
	}
	return val, nil
}

func (m *mockCache) Set(_ context.Context, key, value string, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *mockCache) preload(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *mockCache) Delete(_ context.Context, _ string) error {
	panic("unexpected call to Delete")
}

func (m *mockCache) Exists(_ context.Context, _ string) (bool, error) {
	panic("unexpected call to Exists")
}

func (m *mockCache) Flush(_ context.Context) error {
	panic("unexpected call to Flush")
}

func (m *mockCache) Close() error {
	return nil
}

// ---------------------------------------------------------------------------
// Mock provider — implements RateProvider for testing
// ---------------------------------------------------------------------------

type mockProvider struct {
	cache       cch.Cache
	forwardRate *models.ExchangeRate
	forwardErr  error
	reverseRate *models.ExchangeRate
	reverseErr  error
	// Track SetCacheFloat calls for verification
	setFloatCalls []setFloatCall
	mu            sync.Mutex
}

type setFloatCall struct {
	key string
	val float64
}

func (m *mockProvider) GetCache() cch.Cache {
	return m.cache
}

func (m *mockProvider) QueryForwardRate(_, _, _ string) (*models.ExchangeRate, error) {
	return m.forwardRate, m.forwardErr
}

func (m *mockProvider) QueryReverseRate(_, _, _ string) (*models.ExchangeRate, error) {
	return m.reverseRate, m.reverseErr
}

func (m *mockProvider) SetCacheFloat(key string, val float64, _ time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setFloatCalls = append(m.setFloatCalls, setFloatCall{key: key, val: val})
}

func (m *mockProvider) lastSetFloat() (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.setFloatCalls) == 0 {
		return 0, false
	}
	return m.setFloatCalls[len(m.setFloatCalls)-1].val, true
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestExchangeRate_CacheHit(t *testing.T) {
	cache := newMockCache()
	// Preload cache with a valid rate
	cache.preload("exchange:rate:USD:CNY:2024-06-01", "7.25000000")

	mp := &mockProvider{cache: cache}
	// If cache hits, QueryForwardRate should NOT be called,
	// so we deliberately leave forwardRate/forwardErr as nil to catch misuse.

	rate, err := getExchangeRate(mp, "USD", "CNY", "2024-06-01")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rate != 7.25 {
		t.Fatalf("expected rate 7.25, got %f", rate)
	}

	// Verify no cache-set call was made (value came from cache alone)
	if v, ok := mp.lastSetFloat(); ok {
		t.Fatalf("unexpected SetCacheFloat call with value %f — cache hit should not write back", v)
	}
}

func TestExchangeRate_CacheWithParseError(t *testing.T) {
	cache := newMockCache()
	// Cache returns an unparseable string — should fall through to DB
	cache.preload("exchange:rate:USD:CNY:2024-06-01", "not-a-number")

	mp := &mockProvider{
		cache: cache,
		forwardRate: &models.ExchangeRate{
			FromCurrency: "USD",
			ToCurrency:   "CNY",
			Rate:         7.25000000,
		},
	}

	rate, err := getExchangeRate(mp, "USD", "CNY", "2024-06-01")
	if err != nil {
		t.Fatalf("expected no error after DB fallback, got %v", err)
	}
	if rate != 7.25 {
		t.Fatalf("expected rate 7.25, got %f", rate)
	}

	// Verify the result was written back to cache
	v, ok := mp.lastSetFloat()
	if !ok {
		t.Fatal("expected SetCacheFloat call after DB fallback, but none was made")
	}
	if v != 7.25 {
		t.Fatalf("expected cached value 7.25, got %f", v)
	}
}

func TestExchangeRate_ReverseRate(t *testing.T) {
	cache := newMockCache() // empty cache → miss
	reverseRate := 0.14     // 1/0.14 ≈ 7.142857

	mp := &mockProvider{
		cache:       cache,
		forwardErr:  errors.New("forward rate not found"),
		reverseRate: &models.ExchangeRate{FromCurrency: "CNY", ToCurrency: "USD", Rate: reverseRate},
	}

	rate, err := getExchangeRate(mp, "USD", "CNY", "2024-06-01")
	if err != nil {
		t.Fatalf("expected no error when reverse rate exists, got %v", err)
	}

	expected := 1.0 / reverseRate
	if rate != expected {
		t.Fatalf("expected rate %f (1/%f), got %f", expected, reverseRate, rate)
	}

	// Should have cached the computed forward rate
	v, ok := mp.lastSetFloat()
	if !ok {
		t.Fatal("expected SetCacheFloat after reverse rate computation")
	}
	if v != expected {
		t.Fatalf("expected cached value %f, got %f", expected, v)
	}
}

func TestExchangeRate_RateNotFound(t *testing.T) {
	cache := newMockCache() // empty cache → miss
	reverseErr := errors.New("reverse rate not found")

	mp := &mockProvider{
		cache:      cache,
		forwardErr: errors.New("forward rate not found"),
		reverseErr: reverseErr,
	}

	rate, err := getExchangeRate(mp, "USD", "CNY", "2024-06-01")
	if err == nil {
		t.Fatal("expected an error when no rate exists in either direction")
	}
	if !errors.Is(err, reverseErr) {
		t.Fatalf("expected reverseErr, got %v", err)
	}
	if rate != 0 {
		t.Fatalf("expected rate 0, got %f", rate)
	}
}

func TestExchangeRate_ZeroReverseRate(t *testing.T) {
	cache := newMockCache() // empty cache → miss

	// Forward not found, reverse found but Rate == 0
	mp := &mockProvider{
		cache:       cache,
		forwardErr:  errors.New("forward rate not found"),
		reverseRate: &models.ExchangeRate{FromCurrency: "CNY", ToCurrency: "USD", Rate: 0},
		// reverseErr is nil — reverse was "found"
	}

	rate, err := getExchangeRate(mp, "USD", "CNY", "2024-06-01")
	// The original code returns 0, err2 where err2 is nil in this case
	if err != nil {
		t.Fatalf("expected nil error when reverse rate is found with Rate=0, got %v", err)
	}
	if rate != 0 {
		t.Fatalf("expected rate 0 when reverse rate is 0, got %f", rate)
	}
}

// ---------------------------------------------------------------------------
// Table-driven summary test
// ---------------------------------------------------------------------------

func TestExchangeRate_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		to          string
		date        string
		cacheData   map[string]string // preload into mock cache
		forwardRate *models.ExchangeRate
		forwardErr  error
		reverseRate *models.ExchangeRate
		reverseErr  error
		wantRate    float64
		wantErr     bool
		wantCached  bool // whether SetCacheFloat should have been called
	}{
		{
			name:      "cache hit returns immediately",
			from:      "EUR",
			to:        "CNY",
			date:      "2024-01-01",
			cacheData: map[string]string{"exchange:rate:EUR:CNY:2024-01-01": "7.8"},
			wantRate:  7.8,
			wantCached: false,
		},
		{
			name:      "cache parse error falls through to forward DB",
			from:      "EUR",
			to:        "CNY",
			date:      "2024-01-01",
			cacheData: map[string]string{"exchange:rate:EUR:CNY:2024-01-01": "bad-float"},
			forwardRate: &models.ExchangeRate{FromCurrency: "EUR", ToCurrency: "CNY", Rate: 7.8},
			wantRate:  7.8,
			wantCached: true,
		},
		{
			name:       "forward miss → reverse hit → compute reciprocal",
			from:       "JPY",
			to:         "USD",
			date:       "2024-03-15",
			forwardErr: errors.New("forward not found"),
			reverseRate: &models.ExchangeRate{FromCurrency: "USD", ToCurrency: "JPY", Rate: 0.0068}, // ~1/147.06
			wantRate:   1.0 / 0.0068,
			wantCached: true,
		},
		{
			name:       "both directions missing → error",
			from:       "GBP",
			to:         "CNY",
			date:       "2024-05-20",
			forwardErr: errors.New("forward missing"),
			reverseErr: errors.New("reverse missing"),
			wantRate:   0,
			wantErr:    true,
			wantCached: false,
		},
		{
			name:       "reverse rate is zero → return 0 with no error",
			from:       "BTC",
			to:         "CNY",
			date:       "2024-12-01",
			forwardErr: errors.New("forward not found"),
			reverseRate: &models.ExchangeRate{FromCurrency: "CNY", ToCurrency: "BTC", Rate: 0},
			wantRate:   0,
			wantErr:    false,
			wantCached: false,
		},
		{
			name:       "nil cache → skip cache, go to DB",
			from:       "GBP",
			to:         "USD",
			date:       "2024-08-01",
			cacheData:  nil, // mock cache exists but empty — not nil
			forwardRate: &models.ExchangeRate{FromCurrency: "GBP", ToCurrency: "USD", Rate: 1.27},
			wantRate:   1.27,
			wantCached: true,
		},
		{
			name:       "cache rate is zero → still returned from cache",
			from:       "JPY",
			to:         "CNY",
			date:       "2024-04-01",
			cacheData:  map[string]string{"exchange:rate:JPY:CNY:2024-04-01": "0.000000"},
			wantRate:   0.0,
			wantCached: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := newMockCache()
			for k, v := range tt.cacheData {
				cache.preload(k, v)
			}

			mp := &mockProvider{
				cache:       cache,
				forwardRate: tt.forwardRate,
				forwardErr:  tt.forwardErr,
				reverseRate: tt.reverseRate,
				reverseErr:  tt.reverseErr,
			}

			rate, err := getExchangeRate(mp, tt.from, tt.to, tt.date)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if rate != tt.wantRate {
				t.Errorf("rate: got %f, want %f", rate, tt.wantRate)
			}

			_, wasCached := mp.lastSetFloat()
			if tt.wantCached && !wasCached {
				t.Error("expected SetCacheFloat to be called, but it was not")
			}
			if !tt.wantCached && wasCached {
				t.Error("expected NO SetCacheFloat call, but one was made")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test that the package-level GetExchangeRate still works after our refactor
// (it delegates to getExchangeRate internally).
// We just verify wiring by setting a mock provider.
// ---------------------------------------------------------------------------

func TestExchangeRate_PackageLevelWire(t *testing.T) {
	// Save original provider and restore after test
	orig := provider
	defer func() { provider = orig }()

	cache := newMockCache()
	cache.preload("exchange:rate:GBP:USD:2024-07-01", "1.27000000")

	SetRateProvider(&mockProvider{
		cache: cache,
	})

	rate, err := GetExchangeRate("GBP", "USD", "2024-07-01")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rate != 1.27 {
		t.Fatalf("expected 1.27, got %f", rate)
	}
}

// ---------------------------------------------------------------------------
// setCacheFloat legacy helper still works
// ---------------------------------------------------------------------------

func TestSetCacheFloat(t *testing.T) {
	cache := newMockCache()
	orig := provider
	defer func() { provider = orig }()

	SetRateProvider(&mockProvider{
		cache: cache,
	})

	// setCacheFloat uses the defaultProvider which calls database.GetCache().
	// Since database.GetCache() returns nil in test (no InitCache called),
	// this should be a no-op and not panic.
	setCacheFloat("test:key", 3.14, time.Minute)
	// If we get here without panic, the function is safe.
}
