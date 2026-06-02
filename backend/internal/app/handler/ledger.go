package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LedgerHandler struct {
	svc *service.LedgerService
}

func NewLedgerHandler(svc *service.LedgerService) *LedgerHandler {
	return &LedgerHandler{svc: svc}
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

	ledgers, err := h.svc.ListLedgers(user.ID)
	if err != nil {
		InternalError(c, "failed to list ledgers")
		return
	}
	if ledgers == nil {
		ledgers = []models.Ledger{}
	}

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

	ledger, err := h.svc.CreateLedger(user.ID, input.Name, input.BaseCurrency, input.Description, input.Icon, input.Color)
	if err != nil {
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
	idStr := c.Param("ledger_id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	ledger, err := h.svc.GetLedger(id, user.ID)
	if err != nil {
		if errors.Is(err, service.ErrLedgerNotFound) {
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
	idStr := c.Param("ledger_id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
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

	ledger, err := h.svc.UpdateLedger(id, user.ID, updates)
	if err != nil {
		if errors.Is(err, service.ErrLedgerNotFound) {
			NotFound(c, "ledger not found")
			return
		}
		InternalError(c, "failed to update ledger")
		return
	}

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
	idStr := c.Param("ledger_id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	err = h.svc.DeleteLedger(id, user.ID)
	if err != nil {
		if errors.Is(err, service.ErrLedgerNotFound) {
			NotFound(c, "ledger not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "failed to delete ledger")
		return
	}

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
	idStr := c.Param("ledger_id")

	ledgerID, err := uuid.Parse(idStr)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	summary, err := h.svc.Summary(ledgerID, user.ID)
	if err != nil {
		if errors.Is(err, service.ErrLedgerNotFound) {
			NotFound(c, "ledger not found")
			return
		}
		InternalError(c, "failed to get summary")
		return
	}

	RespondJSON(c, http.StatusOK, ledgerSummary{
		TotalIncome:       summary.TotalIncome,
		TotalExpense:      summary.TotalExpense,
		Balance:           summary.Balance,
		BaseCurrency:      summary.BaseCurrency,
		ExpenseByCategory: summary.ExpenseByCategory,
	})
}

type ledgerSummary struct {
	TotalIncome       float64                         `json:"total_income"`
	TotalExpense      float64                         `json:"total_expense"`
	Balance           float64                         `json:"balance"`
	BaseCurrency      string                          `json:"base_currency" example:"CNY"`
	ExpenseByCategory []service.LedgerCategorySummary `json:"expense_by_category"`
}

// LedgerCategorySummary 分类汇总单项（handler 响应类型）。
type LedgerCategorySummary = service.LedgerCategorySummary

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

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	// Verify ledger ownership via service
	if _, err := h.svc.GetLedger(lid, user.ID); err != nil {
		if errors.Is(err, service.ErrLedgerNotFound) {
			NotFound(c, "ledger not found")
			return
		}
		InternalError(c, "database error")
		return
	}

	format := c.DefaultQuery("format", "csv")
	if format != "csv" && format != "json" {
		BadRequest(c, "format must be csv or json")
		return
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Get count and data via service
	transactions, total, err := h.svc.ExportTransactions(lid, user.ID, startDate, endDate)
	if err != nil {
		InternalError(c, "failed to export transactions")
		return
	}

	const syncLimit = 5000
	if total >= syncLimit {
		// Redirect to async task
		taskID, err := h.svc.SubmitExportTask(user.ID, lid, startDate, endDate, format)
		if err != nil {
			InternalError(c, err.Error())
			return
		}
		RespondJSON(c, http.StatusOK, map[string]interface{}{
			"message": "export submitted as async task",
			"task_id": taskID,
			"total":   total,
		})
		return
	}

	// Synchronous export (< 5000 records)
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

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	tags, err := h.svc.Tags(lid, user.ID)
	if err != nil {
		if errors.Is(err, service.ErrLedgerNotFound) {
			NotFound(c, "ledger not found")
			return
		}
		InternalError(c, "database error")
		return
	}

	RespondJSON(c, http.StatusOK, tags)
}

// ---------- helpers ----------

var csvHeader = []string{"id", "date", "type", "amount", "currency", "base_amount", "description", "category_id"}

func writeCSVStream(c *gin.Context, transactions []models.Transaction) {
	w := csv.NewWriter(c.Writer)
	w.Write(csvHeader)
	for _, t := range transactions {
		desc := ""
		if t.Description != nil {
			desc = *t.Description
		}
		w.Write([]string{
			t.ID.String(),
			t.TransactionDate,
			t.Type,
			fmt.Sprintf("%.2f", t.Amount),
			t.Currency,
			fmt.Sprintf("%.2f", t.BaseAmount),
			desc,
			t.CategoryID.String(),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		slog.Error("csv export flush error", "error", err)
	}
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
