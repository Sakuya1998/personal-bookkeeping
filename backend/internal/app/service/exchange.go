package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"personal-bookkeeping/internal/app/models"
	cch "personal-bookkeeping/internal/infra/cache"

	"gorm.io/gorm"
)

// --- RateProvider abstraction for testability ---

// RateProvider abstracts the database and cache dependencies needed by GetExchangeRate.
type RateProvider interface {
	GetCache() cch.Cache
	QueryForwardRate(fromCurrency, toCurrency, date string) (*models.ExchangeRate, error)
	QueryReverseRate(fromCurrency, toCurrency, date string) (*models.ExchangeRate, error)
	SetCacheFloat(key string, val float64, ttl time.Duration)
}

// defaultProvider is the production implementation backed by injected DB and cache.
type defaultProvider struct {
	db    *gorm.DB
	cache cch.Cache
}

func newDefaultProvider(db *gorm.DB, c cch.Cache) *defaultProvider {
	return &defaultProvider{db: db, cache: c}
}

func (p *defaultProvider) GetCache() cch.Cache {
	return p.cache
}

func (p *defaultProvider) QueryForwardRate(fromCurrency, toCurrency, date string) (*models.ExchangeRate, error) {
	var rate models.ExchangeRate
	err := p.db.Where("from_currency = ? AND to_currency = ? AND date <= ?",
		fromCurrency, toCurrency, date).
		Order("date DESC").
		First(&rate).Error
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

func (p *defaultProvider) QueryReverseRate(fromCurrency, toCurrency, date string) (*models.ExchangeRate, error) {
	var reverse models.ExchangeRate
	err := p.db.Where("from_currency = ? AND to_currency = ? AND date <= ?",
		toCurrency, fromCurrency, date).
		Order("date DESC").
		First(&reverse).Error
	if err != nil {
		return nil, err
	}
	return &reverse, nil
}

func (p *defaultProvider) SetCacheFloat(key string, val float64, ttl time.Duration) {
	if p.cache == nil {
		return
	}
	if err := p.cache.Set(context.Background(), key, fmt.Sprintf("%f", val), ttl); err != nil {
		slog.Warn("cache set failed", "key", key, "error", err)
	}
}

// provider is the package-level provider used by GetExchangeRate.
// It must be initialized via InitExchangeRateProvider before use.
var provider RateProvider

// InitExchangeRateProvider initializes the global exchange rate provider.
// Called during server startup in main.go with the infra layer's DB and cache.
func InitExchangeRateProvider(db *gorm.DB, c cch.Cache) {
	provider = newDefaultProvider(db, c)
}

// SetRateProvider allows tests to inject a mock RateProvider.
func SetRateProvider(p RateProvider) {
	provider = p
}

// GetExchangeRate returns the exchange rate from foreign_curr to base_curr on the given date.
// Returns 0 if not found. Results are cached for 1 hour.
// Panics if InitExchangeRateProvider was not called during startup.
func GetExchangeRate(fromCurrency, toCurrency, date string) (float64, error) {
	if provider == nil {
		return 0, fmt.Errorf("exchange rate provider not initialized")
	}
	return getExchangeRate(provider, fromCurrency, toCurrency, date)
}

func getExchangeRate(p RateProvider, fromCurrency, toCurrency, date string) (float64, error) {
	c := p.GetCache()
	if c != nil {
		key := cch.KeyExchangeRate(fromCurrency, toCurrency, date)
		if val, err := c.Get(context.Background(), key); err == nil {
			if r, parseErr := strconv.ParseFloat(val, 64); parseErr == nil {
				return r, nil
			}
		}
	}

	rate, err := p.QueryForwardRate(fromCurrency, toCurrency, date)
	if err != nil {
		reverse, err2 := p.QueryReverseRate(fromCurrency, toCurrency, date)
		if err2 != nil {
			return 0, err2
		}
		if reverse.Rate > 0 {
			r := 1.0 / reverse.Rate
			p.SetCacheFloat(cch.KeyExchangeRate(fromCurrency, toCurrency, date), r, time.Hour)
			return r, nil
		}
		return 0, err2
	}

	p.SetCacheFloat(cch.KeyExchangeRate(fromCurrency, toCurrency, date), rate.Rate, time.Hour)
	return rate.Rate, nil
}
