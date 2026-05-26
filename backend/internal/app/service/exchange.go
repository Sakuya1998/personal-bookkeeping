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

// GetExchangeRate returns the exchange rate from foreign_curr to base_curr on the given date.
// Returns 0 if not found. Results are cached for 1 hour.
func GetExchangeRate(fromCurrency, toCurrency, date string) (float64, error) {
	c := database.GetCache()
	if c != nil {
		key := cch.KeyExchangeRate(fromCurrency, toCurrency, date)
		if val, err := c.Get(context.Background(), key); err == nil {
			if r, parseErr := strconv.ParseFloat(val, 64); parseErr == nil {
				return r, nil
			}
		}
	}

	var rate models.ExchangeRate
	err := database.GetDB().Where("from_currency = ? AND to_currency = ? AND date <= ?",
		fromCurrency, toCurrency, date).
		Order("date DESC").
		First(&rate).Error
	if err != nil {
		// Try reverse rate
		var reverse models.ExchangeRate
		err2 := database.GetDB().Where("from_currency = ? AND to_currency = ? AND date <= ?",
			toCurrency, fromCurrency, date).
			Order("date DESC").
			First(&reverse).Error
		if err2 != nil {
			return 0, err2
		}
		if reverse.Rate > 0 {
			r := 1.0 / reverse.Rate
			setCacheFloat(cch.KeyExchangeRate(fromCurrency, toCurrency, date), r, time.Hour)
			return r, nil
		}
		return 0, err2
	}

	setCacheFloat(cch.KeyExchangeRate(fromCurrency, toCurrency, date), rate.Rate, time.Hour)
	return rate.Rate, nil
}

func setCacheFloat(key string, val float64, ttl time.Duration) {
	c := database.GetCache()
	if c == nil {
		return
	}
	if err := c.Set(context.Background(), key, fmt.Sprintf("%f", val), ttl); err != nil {
		slog.Warn("cache set failed", "key", key, "error", err)
	}
}
