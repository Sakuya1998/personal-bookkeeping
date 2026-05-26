package handlers

import (
	"context"
	"net/http"
	"time"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"
	cch "personal-bookkeeping/internal/infra/cache"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExchangeRateHandler struct{}

func NewExchangeRateHandler() *ExchangeRateHandler {
	return &ExchangeRateHandler{}
}

type CreateExchangeRateInput struct {
	FromCurrency string  `json:"from_currency" binding:"required" example:"USD"`
	ToCurrency   string  `json:"to_currency" binding:"required" example:"CNY"`
	Rate         float64 `json:"rate" binding:"required,gt=0" example:"7.24"`
	Date         string  `json:"date" example:"2024-01-15"`
	Source       *string `json:"source" example:"bank-of-china"`
}

// List  godoc
// @Summary      汇率列表
// @Tags         exchange-rates
// @Produce      json
// @Security     BearerAuth
// @Param        date  query string false "日期筛选 2006-01-02"
// @Param        from  query string false "源币种"
// @Param        to    query string false "目标币种"
// @Success      200 {object} Response
// @Router       /exchange-rates [get]
func (h *ExchangeRateHandler) List(c *gin.Context) {
	var rates []models.ExchangeRate
	query := database.GetDB().Order("date desc, from_currency asc")

	if date := c.Query("date"); date != "" {
		query = query.Where("date = ?", date)
	}
	if from := c.Query("from"); from != "" {
		query = query.Where("from_currency = ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("to_currency = ?", to)
	}

	query.Find(&rates)
	RespondJSON(c, http.StatusOK, rates)
}

// Create  godoc
// @Summary      创建汇率（同日期+币种自动覆盖）
// @Tags         exchange-rates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body CreateExchangeRateInput true "汇率信息"
// @Success      201 {object} Response
// @Router       /exchange-rates [post]
func (h *ExchangeRateHandler) Create(c *gin.Context) {
	var input CreateExchangeRateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if input.Date == "" {
		input.Date = time.Now().Format("2006-01-02")
	}

	rate := models.ExchangeRate{
		ID:           uuid.New(),
		FromCurrency: input.FromCurrency,
		ToCurrency:   input.ToCurrency,
		Rate:         input.Rate,
		Date:         input.Date,
		Source:       input.Source,
	}

	var existing models.ExchangeRate
	result := database.GetDB().Where("from_currency = ? AND to_currency = ? AND date = ?",
		input.FromCurrency, input.ToCurrency, input.Date).First(&existing)

	if result.Error == nil {
		database.GetDB().Model(&existing).Updates(map[string]interface{}{
			"rate":   input.Rate,
			"source": input.Source,
		})
		invalidateRateCache(input.FromCurrency, input.ToCurrency, input.Date)
		RespondJSON(c, http.StatusOK, existing)
	} else {
		if err := database.GetDB().Create(&rate).Error; err != nil {
			InternalError(c, "failed to create exchange rate")
			return
		}
		invalidateRateCache(input.FromCurrency, input.ToCurrency, input.Date)
		RespondJSON(c, http.StatusCreated, rate)
	}
}

// Latest  godoc
// @Summary      最新汇率（每种币对一条）
// @Tags         exchange-rates
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} Response
// @Router       /exchange-rates/latest [get]
func (h *ExchangeRateHandler) Latest(c *gin.Context) {
	type LatestRate struct {
		FromCurrency string  `json:"from_currency"`
		ToCurrency   string  `json:"to_currency"`
		Rate         float64 `json:"rate"`
		Date         string  `json:"date"`
	}

	var rates []LatestRate
	database.GetDB().Raw(`
		SELECT DISTINCT ON (from_currency, to_currency)
			from_currency, to_currency, rate, date
		FROM exchange_rates
		ORDER BY from_currency, to_currency, date DESC
	`).Scan(&rates)

	RespondJSON(c, http.StatusOK, rates)
}

type latestRate struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Rate         float64 `json:"rate"`
	Date         string  `json:"date"`
}

// Delete  godoc
// @Summary      删除汇率
// @Tags         exchange-rates
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "汇率 ID"
// @Success      200 {object} Response
// @Router       /exchange-rates/{id} [delete]
func (h *ExchangeRateHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	result := database.GetDB().Delete(&models.ExchangeRate{}, "id = ?", id)
	if result.RowsAffected == 0 {
		NotFound(c, "exchange rate not found")
		return
	}
	RespondJSON(c, http.StatusOK, nil)
}

func invalidateRateCache(from, to, date string) {
	c := database.GetCache()
	if c == nil {
		return
	}
	ctx := context.Background()
	_ = c.Delete(ctx, cch.KeyExchangeRate(from, to, date))
}
