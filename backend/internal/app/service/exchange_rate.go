package service

import (
	"context"
	"fmt"
	"time"

	"personal-bookkeeping/internal/app/models"
	cch "personal-bookkeeping/internal/infra/cache"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

// LatestRate 最新汇率 DTO
type LatestRate struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Rate         float64 `json:"rate"`
	Date         string  `json:"date"`
}

// List returns exchange rates filtered by optional date/from/to.
func (s *ExchangeRateService) List(date, from, to string) ([]models.ExchangeRate, error) {
	var rates []models.ExchangeRate
	query := s.DB.Order("date desc, from_currency asc")

	if date != "" {
		query = query.Where("date = ?", date)
	}
	if from != "" {
		query = query.Where("from_currency = ?", from)
	}
	if to != "" {
		query = query.Where("to_currency = ?", to)
	}

	if err := query.Find(&rates).Error; err != nil {
		return nil, fmt.Errorf("failed to query exchange rates: %w", err)
	}
	return rates, nil
}

// Create creates or updates an exchange rate.
// Uses upsert on (from_currency, to_currency) — only the latest rate is kept per pair.
// Returns (rate, wasUpdated, error).
func (s *ExchangeRateService) Create(fromCurrency, toCurrency string, rate float64, date string, source *string) (*models.ExchangeRate, bool, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	now := time.Now()
	sourceStr := source

	newRate := models.ExchangeRate{
		ID:           uuid.New(),
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Rate:         rate,
		Date:         date,
		Source:       sourceStr,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := s.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "from_currency"}, {Name: "to_currency"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"rate":       rate,
			"date":       date,
			"source":     sourceStr,
			"updated_at": now,
		}),
	}).Create(&newRate).Error
	if err != nil {
		return nil, false, fmt.Errorf("failed to upsert exchange rate: %w", err)
	}

	wasUpdated := s.DB.RowsAffected != 1
	s.invalidateCache(fromCurrency, toCurrency, date)
	return &newRate, wasUpdated, nil
}

	// Latest returns the latest exchange rate for each currency pair.
func (s *ExchangeRateService) Latest() ([]LatestRate, error) {
	var rates []LatestRate
	if err := s.DB.Raw(`
		SELECT DISTINCT ON (from_currency, to_currency)
			from_currency, to_currency, rate, date
		FROM exchange_rates
		WHERE deleted_at IS NULL
		ORDER BY from_currency, to_currency, date DESC
	`).Scan(&rates).Error; err != nil {
		return nil, fmt.Errorf("failed to query latest rates: %w", err)
	}
	return rates, nil
}

// Delete removes an exchange rate by ID.
func (s *ExchangeRateService) Delete(id uuid.UUID) error {
	result := s.DB.Delete(&models.ExchangeRate{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete exchange rate: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// invalidateCache removes the cached exchange rate for the given parameters.
func (s *ExchangeRateService) invalidateCache(from, to, date string) {
	if s.Cache == nil {
		return
	}
	_ = s.Cache.Delete(context.Background(), cch.KeyExchangeRate(from, to, date))
}
