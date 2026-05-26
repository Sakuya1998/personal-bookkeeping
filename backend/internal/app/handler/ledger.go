package handlers

import (
	"fmt"
	"net/http"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"
	"personal-bookkeeping/internal/infra/queue"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LedgerHandler struct{}

func NewLedgerHandler() *LedgerHandler {
	return &LedgerHandler{}
}

type CreateLedgerInput struct {
	Name         string  `json:"name" binding:"required,max=100" example:"日常账本"`
	Description  *string `json:"description" example:"日常收支记录"`
	BaseCurrency string  `json:"base_currency" example:"CNY"`
	Icon         *string `json:"icon" example:"📒"`
	Color        *string `json:"color" example:"#1890ff"`
}

type UpdateLedgerInput struct {
	Name         *string `json:"name" example:"新账本名"`
	Description  *string `json:"description"`
	BaseCurrency *string `json:"base_currency"`
	Icon         *string `json:"icon"`
	Color        *string `json:"color"`
	IsArchived   *bool   `json:"is_archived"`
	SortOrder    *int    `json:"sort_order"`
}

// List  godoc
// @Summary      账本列表
// @Tags         ledgers
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} Response
// @Router       /ledgers [get]
func (h *LedgerHandler) List(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	var ledgers []models.Ledger
	database.GetDB().Where("user_id = ?", user.ID).Order("sort_order asc, created_at desc").Find(&ledgers)
	RespondJSON(c, http.StatusOK, ledgers)
}

// Create  godoc
// @Summary      创建账本
// @Tags         ledgers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body CreateLedgerInput true "账本信息"
// @Success      201 {object} Response
// @Router       /ledgers [post]
func (h *LedgerHandler) Create(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	var input CreateLedgerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}
	if input.BaseCurrency == "" {
		input.BaseCurrency = "CNY"
	}

	ledger := models.Ledger{
		ID:           uuid.New(),
		UserID:       user.ID,
		Name:         input.Name,
		Description:  input.Description,
		BaseCurrency: input.BaseCurrency,
		Icon:         input.Icon,
		Color:        input.Color,
	}

	if err := database.GetDB().Create(&ledger).Error; err != nil {
		InternalError(c, "failed to create ledger")
		return
	}

	RespondJSON(c, http.StatusCreated, ledger)
}

// Get  godoc
// @Summary      账本详情
// @Tags         ledgers
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "账本 ID"
// @Success      200 {object} Response
// @Failure      404 {object} Response
// @Router       /ledgers/{id} [get]
func (h *LedgerHandler) Get(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("ledger_id")

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).First(&ledger).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "ledger not found")
			return
		}
		InternalError(c, "database error")
		return
	}

	RespondJSON(c, http.StatusOK, ledger)
}

// Update  godoc
// @Summary      更新账本
// @Tags         ledgers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path string             true "账本 ID"
// @Param        input body UpdateLedgerInput  true "更新内容"
// @Success      200 {object} Response
// @Router       /ledgers/{id} [put]
func (h *LedgerHandler) Update(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("ledger_id")

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).First(&ledger).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "ledger not found")
			return
		}
		InternalError(c, "database error")
		return
	}

	var input UpdateLedgerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.BaseCurrency != nil {
		updates["base_currency"] = *input.BaseCurrency
	}
	if input.Icon != nil {
		updates["icon"] = *input.Icon
	}
	if input.Color != nil {
		updates["color"] = *input.Color
	}
	if input.IsArchived != nil {
		updates["is_archived"] = *input.IsArchived
	}
	if input.SortOrder != nil {
		updates["sort_order"] = *input.SortOrder
	}

	database.GetDB().Model(&ledger).Updates(updates)
	database.GetDB().First(&ledger, ledger.ID)

	RespondJSON(c, http.StatusOK, ledger)
}

// Delete  godoc
// @Summary      删除账本
// @Tags         ledgers
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "账本 ID"
// @Success      200 {object} Response
// @Router       /ledgers/{id} [delete]
func (h *LedgerHandler) Delete(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("ledger_id")

	result := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.Ledger{})
	if result.RowsAffected == 0 {
		NotFound(c, "ledger not found")
		return
	}

	database.GetDB().Where("ledger_id = ?", id).Delete(&models.Transaction{})
	database.GetDB().Where("ledger_id = ?", id).Delete(&models.Category{})

	RespondJSON(c, http.StatusOK, nil)
}

// Summary  godoc
// @Summary      账本汇总
// @Tags         ledgers
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "账本 ID"
// @Success      200 {object} Response{data=ledgerSummary}
// @Router       /ledgers/{id}/summary [get]
func (h *LedgerHandler) Summary(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("ledger_id")

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	var totalIncome, totalExpense float64
	database.GetDB().Model(&models.Transaction{}).
		Where("ledger_id = ? AND type = ?", id, "income").
		Select("COALESCE(SUM(base_amount), 0)").Scan(&totalIncome)

		database.GetDB().Model(&models.Transaction{}).
		Where("ledger_id = ? AND type = ?", id, "expense").
		Select("COALESCE(SUM(base_amount), 0)").Scan(&totalExpense)

	var expenseByCategory []CategorySummary
	database.GetDB().Raw(`
		SELECT t.category_id, c.name as category_name, COALESCE(c.icon,'') as category_icon,
		       COALESCE(SUM(t.base_amount),0) as total, COUNT(*) as count
		FROM transactions t
		JOIN categories c ON c.id = t.category_id
		WHERE t.ledger_id = ? AND t.type = 'expense'
		GROUP BY t.category_id, c.name, c.icon
		ORDER BY total DESC
	`, id).Scan(&expenseByCategory)

	RespondJSON(c, http.StatusOK, ledgerSummary{
		TotalIncome:       totalIncome,
		TotalExpense:      totalExpense,
		Balance:           totalIncome - totalExpense,
		BaseCurrency:      ledger.BaseCurrency,
		ExpenseByCategory: expenseByCategory,
	})
}

type ledgerSummary struct {
	TotalIncome       float64           `json:"total_income"`
	TotalExpense      float64           `json:"total_expense"`
	Balance           float64           `json:"balance"`
	BaseCurrency      string            `json:"base_currency" example:"CNY"`
	ExpenseByCategory []CategorySummary `json:"expense_by_category"`
}

type CategorySummary struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	CategoryIcon string    `json:"category_icon"`
	Total        float64   `json:"total"`
	Count        int64     `json:"count"`
}

// Export  godoc
// @Summary      导出交易数据（同步 CSV/JSON，超 5000 条转异步）
// @Tags         ledgers
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id  path  string true "账本 ID"
// @Param        format     query string false "导出格式: csv|json" default(csv)
// @Param        start_date query string false "开始日期 2006-01-02"
// @Param        end_date   query string false "结束日期 2006-01-02"
// @Success      200 {file} file
// @Router       /ledgers/{ledger_id}/export [get]
func (h *LedgerHandler) Export(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", ledgerID, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	format := c.DefaultQuery("format", "csv")
	if format != "csv" && format != "json" {
		BadRequest(c, "format must be csv or json")
		return
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// count first
	var total int64
	countQuery := database.GetDB().Model(&models.Transaction{}).
		Where("ledger_id = ? AND user_id = ?", ledgerID, user.ID)
	if startDate != "" {
		countQuery = countQuery.Where("transaction_date >= ?", startDate)
	}
	if endDate != "" {
		countQuery = countQuery.Where("transaction_date <= ?", endDate)
	}
	countQuery.Count(&total)

	const syncLimit = 5000
	if total >= syncLimit {
		// redirect to async task
		q := database.GetQueue()
		if q == nil {
			InternalError(c, "task queue not available")
			return
		}
		taskID := uuid.New().String()
		if err := q.Submit(c.Request.Context(), queue.Task{
			ID:   taskID,
			Type: "export_report",
			Payload: map[string]interface{}{
				"user_id":    user.ID.String(),
				"ledger_id":  ledgerID,
				"start_date": startDate,
				"end_date":   endDate,
				"format":     format,
			},
		}); err != nil {
			InternalError(c, "failed to submit export task")
			return
		}
		RespondJSON(c, http.StatusOK, map[string]interface{}{
			"message": "export submitted as async task",
			"task_id": taskID,
			"total":   total,
		})
		return
	}

	// synchronous export (< 5000 records)
	query := database.GetDB().Where("ledger_id = ? AND user_id = ?", ledgerID, user.ID)
	if startDate != "" {
		query = query.Where("transaction_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("transaction_date <= ?", endDate)
	}
	var transactions []models.Transaction
	query.Order("transaction_date desc").Find(&transactions)

	switch format {
	case "csv":
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename=export.csv")
		writeCSVStream(c, transactions)
	case "json":
		RespondJSON(c, http.StatusOK, transactions)
	}
}

// Tags  godoc
// @Summary      获取账本所有标签
// @Tags         ledgers
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Success      200 {object} Response{data=[]string}
// @Router       /ledgers/{ledger_id}/tags [get]
func (h *LedgerHandler) Tags(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", ledgerID, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	var rawTags []string
	database.GetDB().Model(&models.Transaction{}).
		Where("ledger_id = ? AND user_id = ? AND tags IS NOT NULL AND tags <> ''", ledgerID, user.ID).
		Distinct("tags").
		Pluck("tags", &rawTags)

	// split and deduplicate
	seen := make(map[string]struct{})
	var tags []string
	for _, raw := range rawTags {
		for _, t := range splitTags(raw) {
			t = trim(t)
			if t == "" {
				continue
			}
			if _, ok := seen[t]; !ok {
				seen[t] = struct{}{}
				tags = append(tags, t)
			}
		}
	}

	if tags == nil {
		tags = []string{}
	}

	RespondJSON(c, http.StatusOK, tags)
}

// ---------- helpers ----------

var csvHeader = []string{"id", "date", "type", "amount", "currency", "base_amount", "description", "category_id"}

func writeCSVStream(c *gin.Context, transactions []models.Transaction) {
	c.Writer.WriteString(stringsJoin(csvHeader, ",") + "\n")
	for _, t := range transactions {
		c.Writer.WriteString(csvRow(t))
	}
}

func csvRow(t models.Transaction) string {
	desc := ""
	if t.Description != nil {
		desc = *t.Description
	}
	return stringsJoin([]string{
		t.ID.String(),
		t.TransactionDate,
		t.Type,
		fmt.Sprintf("%.2f", t.Amount),
		t.Currency,
		fmt.Sprintf("%.2f", t.BaseAmount),
		desc,
		t.CategoryID.String(),
	}, ",") + "\n"
}

func stringsJoin(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += sep + parts[i]
	}
	return out
}

func splitTags(raw string) []string {
	// tags are stored comma-separated
	var parts []string
	current := ""
	for _, ch := range raw {
		if ch == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
