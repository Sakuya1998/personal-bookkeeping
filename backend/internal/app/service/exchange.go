package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"
	cch "personal-bookkeeping/internal/infra/cache"
)

// --- RateProvider abstraction for testability ---

// RateProvider abstracts the database and cache dependencies needed by GetExchangeRate.
type RateProvider interface {
	// GetCache returns the cache instance.
	GetCache() cch.Cache
	// QueryForwardRate queries the exchange rate from fromCurrency to toCurrency for the given date.
	QueryForwardRate(fromCurrency, toCurrency, date string) (*models.ExchangeRate, error)
	// QueryReverseRate queries the exchange rate from toCurrency to fromCurrency for the given date.
	QueryReverseRate(fromCurrency, toCurrency, date string) (*models.ExchangeRate, error)
	// SetCacheFloat stores a float64 value in cache with the given TTL.
	SetCacheFloat(key string, val float64, ttl time.Duration)
}

// defaultProvider is the production implementation that uses global DB and cache.
type defaultProvider struct{}

func (p *defaultProvider) GetCache() cch.Cache {
	return database.GetCache()
}

func (p *defaultProvider) QueryForwardRate(fromCurrency, toCurrency, date string) (*models.ExchangeRate, error) {
	var rate models.ExchangeRate
	err := database.GetDB().Where("from_currency = ? AND to_currency = ? AND date <= ?",
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
	err := database.GetDB().Where("from_currency = ? AND to_currency = ? AND date <= ?",
		toCurrency, fromCurrency, date).
		Order("date DESC").
		First(&reverse).Error
	if err != nil {
		return nil, err
	}
	return &reverse, nil
}

func (p *defaultProvider) SetCacheFloat(key string, val float64, ttl time.Duration) {
	c := database.GetCache()
	if c == nil {
		return
	}
	if err := c.Set(context.Background(), key, fmt.Sprintf("%f", val), ttl); err != nil {
		slog.Warn("cache set failed", "key", key, "error", err)
	}
}

// provider is the package-level provider used by GetExchangeRate.
// It can be replaced in tests via SetRateProvider.
var provider RateProvider = &defaultProvider{}

// SetRateProvider allows tests to inject a mock RateProvider.
func SetRateProvider(p RateProvider) {
	provider = p
}

// GetExchangeRate returns the exchange rate from foreign_curr to base_curr on the given date.
// Returns 0 if not found. Results are cached for 1 hour.
func GetExchangeRate(fromCurrency, toCurrency, date string) (float64, error) {
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

	// Try forward rate
	rate, err := p.QueryForwardRate(fromCurrency, toCurrency, date)
	if err != nil {
		// Try reverse rate
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

// setCacheFloat is a legacy helper kept for backward compatibility.
func setCacheFloat(key string, val float64, ttl time.Duration) {
	(&defaultProvider{}).SetCacheFloat(key, val, ttl)
}
