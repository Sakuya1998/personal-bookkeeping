package handler

import (
	"net/http"
	"strconv"
	"time"

	"personal-bookkeeping/internal/app/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------- response types ----------

// MonthlyTrendItem 月度收支趋势单项。
type MonthlyTrendItem struct {
	Month   string  `json:"month" example:"2026-01"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

// CategoryBreakdownItem 分类支出/收入分布单项。
type CategoryBreakdownItem struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	CategoryIcon string    `json:"category_icon"`
	Type         string    `json:"type"`
	Total        float64   `json:"total"`
	Percentage   float64   `json:"percentage"`
}

// DailyTransactionItem 日历视图每日汇总。
type DailyTransactionItem struct {
	Date    string  `json:"date" example:"2026-05-01"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Count   int64   `json:"count"`
}

// ---------- handler methods ----------

// MonthlyTrend  godoc
// @Summary      月度收支趋势
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Param        months    query int    false "统计月数" default(6)
// @Success      200 {object} Response{data=[]MonthlyTrendItem}
// @Router       /ledgers/{ledger_id}/monthly-trend [get]
func (h *LedgerHandler) MonthlyTrend(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	months := 6
	if m := c.Query("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 24 {
			months = parsed
		}
	}

	items, err := h.svc.MonthlyTrend(lid, user.ID, months)
	if err != nil {
		InternalError(c, "failed to query monthly trend")
		return
	}

	RespondJSON(c, http.StatusOK, items)
}

// CategoryBreakdown  godoc
// @Summary      分类支出/收入分布
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id  path string true "账本 ID"
// @Param        start_date query string false "开始日期 2006-01-02"
// @Param        end_date   query string false "结束日期 2006-01-02"
// @Param        type       query string false "筛选：income/expense（留空返回全部）"
// @Success      200 {object} Response{data=[]CategoryBreakdownItem}
// @Router       /ledgers/{ledger_id}/category-breakdown [get]
func (h *LedgerHandler) CategoryBreakdown(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	txnType := c.Query("type")

	items, err := h.svc.CategoryBreakdown(lid, user.ID, startDate, endDate, txnType)
	if err != nil {
		InternalError(c, "failed to query category breakdown")
		return
	}

	RespondJSON(c, http.StatusOK, items)
}

// TagStats  godoc
// @Summary      标签使用统计
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id  path string true "账本 ID"
// @Param        start_date query string false "开始日期 2006-01-02"
// @Param        end_date   query string false "结束日期 2006-01-02"
// @Success      200 {object} Response{data=[]service.TagStatsItem}
// @Router       /ledgers/{ledger_id}/tag-stats [get]
func (h *LedgerHandler) TagStats(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	items, err := h.svc.TagStats(lid, user.ID, startDate, endDate)
	if err != nil {
		InternalError(c, "failed to query tag stats")
		return
	}
	RespondJSON(c, http.StatusOK, items)
}

// DailyTransactions  godoc
// @Summary      日历视图 — 每日交易汇总
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Param        year      query int    true  "年份"
// @Param        month     query int    true  "月份 (1-12)"
// @Success      200 {object} Response{data=[]DailyTransactionItem}
// @Router       /ledgers/{ledger_id}/daily-transactions [get]
func (h *LedgerHandler) DailyTransactions(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := c.Query("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil && parsed >= 2000 && parsed <= 2099 {
			year = parsed
		}
	}
	if m := c.Query("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = parsed
		}
	}

	items, err := h.svc.DailyTransactions(lid, user.ID, year, month)
	if err != nil {
		InternalError(c, "failed to query daily transactions")
		return
	}

	RespondJSON(c, http.StatusOK, items)
}
