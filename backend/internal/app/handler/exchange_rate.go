package handler

import (
	"log/slog"
	"net/http"

	"personal-bookkeeping/internal/app/service"
	"personal-bookkeeping/internal/infra/config"
	"personal-bookkeeping/internal/infra/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExchangeRateHandler struct {
	svc *service.ExchangeRateService
}

func NewExchangeRateHandler(svc *service.ExchangeRateService) *ExchangeRateHandler {
	return &ExchangeRateHandler{svc: svc}
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
	date := c.Query("date")
	from := c.Query("from")
	to := c.Query("to")

	rates, err := h.svc.List(date, from, to)
	if err != nil {
		InternalError(c, "failed to query exchange rates")
		return
	}
	RespondJSON(c, http.StatusOK, rates)
}

// Sync  godoc
// @Summary      手动同步最新汇率
// @Description  从外部 API 拉取最新汇率并更新数据库
// @Tags         exchange-rates
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} Response
// @Router       /exchange-rates/sync [post]
func (h *ExchangeRateHandler) Sync(c *gin.Context) {
	cfg := config.Load()
	if cfg == nil {
		InternalError(c, "config not available")
		return
	}

	if err := service.UpdateExchangeRates(database.GetDB(), &cfg.ExchangeRate); err != nil {
		slog.Error("exchange rate sync failed",
			"error", err,
			"provider", cfg.ExchangeRate.Provider,
			"base", cfg.ExchangeRate.Base,
			"has_api_key", cfg.ExchangeRate.APIKey != "",
		)
		InternalError(c, "sync failed: "+err.Error())
		return
	}
	RespondJSON(c, http.StatusOK, gin.H{"message": "exchange rates synced"})
}

// Latest  godoc
// @Summary      最新汇率（每种币对一条）
// @Tags         exchange-rates
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} Response
// @Router       /exchange-rates/latest [get]
func (h *ExchangeRateHandler) Latest(c *gin.Context) {
	rates, err := h.svc.Latest()
	if err != nil {
		InternalError(c, "failed to query latest rates")
		return
	}
	RespondJSON(c, http.StatusOK, rates)
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
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid exchange rate id")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if err == service.ErrNotFound {
			NotFound(c, "exchange rate not found")
			return
		}
		InternalError(c, "failed to delete exchange rate")
		return
	}
	RespondJSON(c, http.StatusOK, nil)
}
