package handler

import (
	"net/http"

	"personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExchangeRateHandler struct {
	svc *service.ExchangeRateService
}

func NewExchangeRateHandler(svc *service.ExchangeRateService) *ExchangeRateHandler {
	return &ExchangeRateHandler{svc: svc}
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

	rate, wasUpdated, err := h.svc.Create(input.FromCurrency, input.ToCurrency, input.Rate, input.Date, input.Source)
	if err != nil {
		InternalError(c, "failed to create exchange rate")
		return
	}

	if wasUpdated {
		RespondJSON(c, http.StatusOK, rate)
	} else {
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
