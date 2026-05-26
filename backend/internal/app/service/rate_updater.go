package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"
	"personal-bookkeeping/internal/infra/config"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExchangeRateAPI response from exchangerate-api.com
type exchangeRateAPIResponse struct {
	Result         string             `json:"result"`
	BaseCode       string             `json:"base_code"`
	ConversionRates map[string]float64 `json:"conversion_rates"`
	TimeLastUpdateUnix int64          `json:"time_last_update_unix"`
}

// Frankfurter response (frankfurter.app)
type frankfurterResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

// UpdateExchangeRates 从外部 API 拉取最新汇率并写入 DB。
// 只更新涉及配置中 base 货币的汇率对。
func UpdateExchangeRates(cfg *config.ExchangeRateConfig) error {
	if cfg.APIKey == "" && cfg.Provider == "exchangerate-api" {
		slog.Warn("exchange rate auto-update: no API key configured, skipping")
		return nil
	}

	base := strings.ToUpper(cfg.Base)
	if base == "" {
		base = "USD"
	}

	rates, source, err := fetchRates(cfg.Provider, cfg.APIKey, base)
	if err != nil {
		return fmt.Errorf("fetch rates: %w", err)
	}

	date := time.Now().Format("2006-01-02")
	today := date

	return storeRates(database.GetDB(), base, rates, today, source)
}

func fetchRates(provider, apiKey, base string) (map[string]float64, string, error) {
	switch provider {
	case "exchangerate-api":
		return fetchExchangeRateAPI(apiKey, base)
	case "frankfurter":
		return fetchFrankfurter(base)
	default:
		return nil, "", fmt.Errorf("unknown exchange rate provider: %q", provider)
	}
}

func fetchExchangeRateAPI(apiKey, base string) (map[string]float64, string, error) {
	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", apiKey, base)
	body, err := httpGet(url, 10*time.Second)
	if err != nil {
		return nil, "", err
	}

	var resp exchangeRateAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, "", fmt.Errorf("decode response: %w", err)
	}
	if resp.Result != "success" {
		return nil, "", fmt.Errorf("API error: result=%q", resp.Result)
	}
	return resp.ConversionRates, "exchangerate-api", nil
}

func fetchFrankfurter(base string) (map[string]float64, string, error) {
	url := fmt.Sprintf("https://api.frankfurter.app/latest?from=%s", base)
	body, err := httpGet(url, 10*time.Second)
	if err != nil {
		return nil, "", err
	}

	var resp frankfurterResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, "", fmt.Errorf("decode frankfurter response: %w", err)
	}
	return resp.Rates, "frankfurter", nil
}

func httpGet(url string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "personal-bookkeeping/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// storeRates 将汇率写入 DB。
// 对于 base→target 的汇率，直接写入。
// 对于 target→base 的汇率，计算倒数后写入。
func storeRates(db *gorm.DB, base string, rates map[string]float64, date, source string) error {
	if db == nil {
		return fmt.Errorf("database not available")
	}

	now := time.Now()
	created := 0
	sourceStr := source

	for currency, rate := range rates {
		if currency == base {
			continue // skip self
		}
		if rate <= 0 {
			continue
		}

		// Store both directions
		pairs := []struct {
			from, to string
			r        float64
		}{
			{from: base, to: currency, r: rate},
			{from: currency, to: base, r: 1.0 / rate},
		}

		for _, p := range pairs {
			// Check if this rate already exists for today
			var existing int64
			db.Model(&models.ExchangeRate{}).
				Where("from_currency = ? AND to_currency = ? AND date = ?", p.from, p.to, date).
				Count(&existing)
			if existing > 0 {
				continue
			}

			record := models.ExchangeRate{
				ID:           uuid.New(),
				FromCurrency: p.from,
				ToCurrency:   p.to,
				Rate:         p.r,
				Date:         date,
				Source:       &sourceStr,
				CreatedAt:    now,
			}
			if err := db.Create(&record).Error; err != nil {
				slog.Warn("store rate failed", "from", p.from, "to", p.to, "error", err)
				continue
			}
			created++
		}
	}

	slog.Info("exchange rates updated", "base", base, "currencies", len(rates), "created", created, "source", source)
	return nil
}
