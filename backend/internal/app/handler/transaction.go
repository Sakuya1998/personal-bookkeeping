package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	svc *service.TransactionService
}

func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

type CreateTransactionInput struct {
	LedgerID        string   `json:"ledger_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	CategoryID      string   `json:"category_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Type            string   `json:"type" binding:"required,oneof=income expense" example:"expense"`
	Amount          any      `json:"amount" binding:"required"` // 支持 string 或 number
	Currency        string   `json:"currency" example:"CNY"`
	Description     *string  `json:"description"`
	TransactionDate string   `json:"transaction_date" example:"2024-01-01"`
	Tags            []string `json:"tags" example:"[\"food\",\"lunch\"]"`
}

type UpdateTransactionInput struct {
	CategoryID      *string  `json:"category_id"`
	Type            *string  `json:"type" binding:"omitempty,oneof=income expense"`
	Amount          *float64 `json:"amount" binding:"omitempty,gt=0"`
	Currency        *string  `json:"currency"`
	Description     *string  `json:"description"`
	TransactionDate *string  `json:"transaction_date"`
	Tags            []string `json:"tags"`
	IsReconciled    *bool    `json:"is_reconciled"`
}

// List  godoc
// @Summary      交易记录列表
// @Tags         transactions
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id  path   string true  "账本 ID"
// @Param        type       query  string false "筛选：income/expense"
// @Param        category_id query string false "分类 ID"
// @Param        start_date query  string false "开始日期 2006-01-02"
// @Param        end_date   query  string false "结束日期 2006-01-02"
// @Param        keyword    query  string false "描述关键词搜索"
// @Param        page       query  int    false "页码" default(1)
// @Param        page_size  query  int    false "每页数量" default(20)
// @Success      200 {object} Response{data=transactionList}
// @Router       /ledgers/{ledger_id}/transactions [get]
func (h *TransactionHandler) List(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")

	ledgerUUID, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	// Build filters map
	filters := make(map[string]string)
	if t := c.Query("type"); t != "" {
		filters["type"] = t
	}
	if catID := c.Query("category_id"); catID != "" {
		filters["category_id"] = catID
	}
	if start := c.Query("start_date"); start != "" {
		filters["start_date"] = start
	}
	if end := c.Query("end_date"); end != "" {
		filters["end_date"] = end
	}
	if keyword := c.Query("keyword"); keyword != "" {
		filters["keyword"] = keyword
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	transactions, total, totalPages, err := h.svc.ListTransactions(ledgerUUID, user.ID, page, pageSize, filters)
	if err != nil {
		InternalError(c, "database error")
		return
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	RespondJSON(c, http.StatusOK, transactionList{
		Items:      transactions,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// Create  godoc
// @Summary      创建交易记录
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body CreateTransactionInput true "交易信息"
// @Success      201 {object} Response
// @Router       /transactions [post]
func (h *TransactionHandler) Create(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	var input CreateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	ledgerUUID, err := uuid.Parse(input.LedgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id format")
		return
	}
	categoryUUID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		BadRequest(c, "invalid category_id format")
		return
	}

	// Parse amount (accept string or number)
	amount, err := parseAmount(input.Amount)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	txn, overBudget, err := h.svc.CreateTransaction(
		ledgerUUID, user.ID, categoryUUID,
		input.Type, amount, input.Currency,
		input.Description, input.TransactionDate, input.Tags,
	)
	if err != nil {
		InternalError(c, "failed to create transaction")
		return
	}

	RespondJSON(c, http.StatusCreated, gin.H{
		"transaction": txn,
		"over_budget": overBudget,
	})
}

// Update  godoc
// @Summary      更新交易记录
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path string                  true "交易 ID"
// @Param        input body UpdateTransactionInput  true "更新内容"
// @Success      200 {object} Response
// @Router       /transactions/{id} [put]
func (h *TransactionHandler) Update(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		BadRequest(c, "invalid transaction id")
		return
	}

	var input UpdateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Convert category_id string pointer to uuid pointer
	var catPtr *uuid.UUID
	if input.CategoryID != nil {
		parsed, err := uuid.Parse(*input.CategoryID)
		if err != nil {
			BadRequest(c, "invalid category_id format")
			return
		}
		catPtr = &parsed
	}

	txn, overBudget, err := h.svc.UpdateTransaction(
		id, user.ID, catPtr, input.Type, input.Amount,
		input.Currency, input.Description, input.TransactionDate,
		input.Tags, input.IsReconciled,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, "transaction not found")
			return
		}
		InternalError(c, "failed to update transaction")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"transaction": txn,
		"over_budget": overBudget,
	})
}

// Delete  godoc
// @Summary      删除交易记录
// @Tags         transactions
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "交易 ID"
// @Success      200 {object} Response
// @Router       /transactions/{id} [delete]
func (h *TransactionHandler) Delete(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		BadRequest(c, "invalid transaction id")
		return
	}

	err = h.svc.DeleteTransaction(id, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, "transaction not found")
			return
		}
		InternalError(c, "failed to delete transaction")
		return
	}

	RespondJSON(c, http.StatusOK, nil)
}

type transactionList struct {
	Items      []models.Transaction `json:"items"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// ---------- batch operations ----------

// BatchDelete  godoc
// @Summary      批量删除交易
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body object{ids=[]string} true "交易 ID 列表"
// @Success      200 {object} Response{data=map[string]int}
// @Router       /transactions/batch-delete [post]
func (h *TransactionHandler) BatchDelete(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var input struct {
		IDs []string `json:"ids" binding:"required,min=1,max=500"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	uuids := make([]uuid.UUID, 0, len(input.IDs))
	for _, id := range input.IDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			BadRequest(c, "invalid id: "+id)
			return
		}
		uuids = append(uuids, parsed)
	}

	deleted, err := h.svc.BatchDelete(uuids, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			BadRequest(c, "some transactions not found or not owned by current user")
			return
		}
		InternalError(c, "failed to batch delete transactions")
		return
	}

	RespondJSON(c, http.StatusOK, map[string]interface{}{
		"deleted": deleted,
	})
}

// BatchUpdate  godoc
// @Summary      批量修改交易分类
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body object{ids=[]string,category_id=string} true "交易 ID 列表 + 目标分类 ID"
// @Success      200 {object} Response{data=map[string]int}
// @Router       /transactions/batch-update [put]
func (h *TransactionHandler) BatchUpdate(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var input struct {
		IDs        []string `json:"ids" binding:"required,min=1,max=500"`
		CategoryID string   `json:"category_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	catUUID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		BadRequest(c, "invalid category_id")
		return
	}

	uuids := make([]uuid.UUID, 0, len(input.IDs))
	for _, id := range input.IDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			BadRequest(c, "invalid id: "+id)
			return
		}
		uuids = append(uuids, parsed)
	}

	updated, err := h.svc.BatchUpdateCategory(uuids, catUUID, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			BadRequest(c, "some transactions not found or not owned by current user")
			return
		}
		InternalError(c, "failed to batch update transactions")
		return
	}

	RespondJSON(c, http.StatusOK, map[string]interface{}{
		"updated": updated,
	})
}

// parseAmount 将 any 类型（string、number、json.Number）解析为 float64。
func parseAmount(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		if val <= 0 {
			return 0, fmt.Errorf("amount must be positive")
		}
		return val, nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid amount: %w", err)
		}
		if f <= 0 {
			return 0, fmt.Errorf("amount must be positive")
		}
		return f, nil
	case json.Number:
		f, err := val.Float64()
		if err != nil {
			return 0, fmt.Errorf("invalid amount: %w", err)
		}
		if f <= 0 {
			return 0, fmt.Errorf("amount must be positive")
		}
		return f, nil
	default:
		return 0, fmt.Errorf("amount must be a number or numeric string")
	}
}
