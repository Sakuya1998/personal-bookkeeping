package handlers

import (
	"net/http"
	"strconv"
	"time"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"

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

	// verify ledger ownership
	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", ledgerID, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	months := 6
	if m := c.Query("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 24 {
			months = parsed
		}
	}

	endDate := time.Now().AddDate(0, 1, 0) // include current month
	startDate := endDate.AddDate(0, -months, 0)

	var rows []struct {
		Month         string  `json:"month"`
		TotalIncome   float64 `json:"total_income"`
		TotalExpense  float64 `json:"total_expense"`
	}
	database.GetDB().Raw(`
		SELECT
			to_char(transaction_date, 'YYYY-MM') AS month,
			COALESCE(SUM(CASE WHEN type = 'income'  THEN base_amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN base_amount ELSE 0 END), 0) AS total_expense
		FROM transactions
		WHERE ledger_id = ? AND user_id = ? AND transaction_date >= ? AND transaction_date < ?
		GROUP BY month
		ORDER BY month ASC
	`, ledgerID, user.ID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).Scan(&rows)

	items := make([]MonthlyTrendItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, MonthlyTrendItem{
			Month:   r.Month,
			Income:  r.TotalIncome,
			Expense: r.TotalExpense,
		})
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

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", ledgerID, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	txnType := c.Query("type") // optional filter

	var rows []struct {
		CategoryID   uuid.UUID `json:"category_id"`
		CategoryName string    `json:"category_name"`
		CategoryIcon string    `json:"category_icon"`
		Type         string    `json:"type"`
		Total        float64   `json:"total"`
	}

	query := database.GetDB().Raw(`
		SELECT
			t.category_id,
			COALESCE(c.name, '')    AS category_name,
			COALESCE(c.icon, '')    AS category_icon,
			t.type,
			COALESCE(SUM(t.base_amount), 0) AS total
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.ledger_id = ? AND t.user_id = ?
	`, ledgerID, user.ID)

	if startDate != "" {
		query = query.Where("t.transaction_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("t.transaction_date <= ?", endDate)
	}
	if txnType != "" {
		query = query.Where("t.type = ?", txnType)
	}

	query = query.Group("t.category_id, c.name, c.icon, t.type").Order("total DESC")
	query.Scan(&rows)

	// calculate total for percentages
	var grandTotal float64
	for _, r := range rows {
		grandTotal += r.Total
	}

	items := make([]CategoryBreakdownItem, 0, len(rows))
	for _, r := range rows {
		pct := 0.0
		if grandTotal > 0 {
			pct = r.Total / grandTotal * 100
		}
		items = append(items, CategoryBreakdownItem{
			CategoryID:   r.CategoryID,
			CategoryName: r.CategoryName,
			CategoryIcon: r.CategoryIcon,
			Type:         r.Type,
			Total:        r.Total,
			Percentage:   pct,
		})
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

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", ledgerID, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
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

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	endDate := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	var rows []struct {
		Date         string  `json:"date"`
		TotalIncome  float64 `json:"total_income"`
		TotalExpense float64 `json:"total_expense"`
		TxnCount     int64   `json:"txn_count"`
	}
	database.GetDB().Raw(`
		SELECT
			transaction_date AS date,
			COALESCE(SUM(CASE WHEN type = 'income'  THEN base_amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN base_amount ELSE 0 END), 0) AS total_expense,
			COUNT(*) AS txn_count
		FROM transactions
		WHERE ledger_id = ? AND user_id = ? AND transaction_date >= ? AND transaction_date < ?
		GROUP BY transaction_date
		ORDER BY transaction_date ASC
	`, ledgerID, user.ID, startDate, endDate).Scan(&rows)

	items := make([]DailyTransactionItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, DailyTransactionItem{
			Date:    r.Date,
			Income:  r.TotalIncome,
			Expense: r.TotalExpense,
			Count:   r.TxnCount,
		})
	}

	RespondJSON(c, http.StatusOK, items)
}
