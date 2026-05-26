package handlers

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"
	"personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionHandler struct{}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}

type CreateTransactionInput struct {
	LedgerID        string   `json:"ledger_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	CategoryID      string   `json:"category_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Type            string   `json:"type" binding:"required,oneof=income expense" example:"expense"`
	Amount          float64  `json:"amount" binding:"required,gt=0" example:"29.90"`
	Currency        string   `json:"currency" example:"CNY"`
	Description     *string  `json:"description" example:"午餐"`
	TransactionDate string   `json:"transaction_date" example:"2024-01-15"`
	Tags            []string `json:"tags" example:"[\"午餐\",\"外卖\"]"`
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

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", ledgerID, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	query := database.GetDB().Where("ledger_id = ? AND user_id = ?", ledgerID, user.ID)
	if t := c.Query("type"); t != "" {
		query = query.Where("type = ?", t)
	}
	if catID := c.Query("category_id"); catID != "" {
		query = query.Where("category_id = ?", catID)
	}
	if start := c.Query("start_date"); start != "" {
		query = query.Where("transaction_date >= ?", start)
	}
	if end := c.Query("end_date"); end != "" {
		query = query.Where("transaction_date <= ?", end)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("description ILIKE ?", "%"+keyword+"%")
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	query.Model(&models.Transaction{}).Count(&total)

	var transactions []models.Transaction
	query.Preload("Category").Order("transaction_date desc, created_at desc").Offset(offset).Limit(pageSize).Find(&transactions)

	RespondJSON(c, http.StatusOK, transactionList{
		Items:      transactions,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
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

	ledgerUUID, _ := uuid.Parse(input.LedgerID)
	categoryUUID, _ := uuid.Parse(input.CategoryID)

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", ledgerUUID, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	if input.Currency == "" {
		input.Currency = "CNY"
	}

	txnDate := input.TransactionDate
	if txnDate == "" {
		txnDate = time.Now().Format("2006-01-02")
	}

	exchangeRate := 1.0
	baseAmount := input.Amount
	if input.Currency != ledger.BaseCurrency {
		rate, err := services.GetExchangeRate(input.Currency, ledger.BaseCurrency, txnDate)
		if err == nil && rate > 0 {
			exchangeRate = rate
			baseAmount = input.Amount * rate
		}
	}

	baseAmount = math.Round(baseAmount*100) / 100

	var tagsStr *string
	if len(input.Tags) > 0 {
		s := ""
		for i, tag := range input.Tags {
			if i > 0 {
				s += ","
			}
			s += tag
		}
		tagsStr = &s
	}

	txn := models.Transaction{
		ID:              uuid.New(),
		LedgerID:        ledgerUUID,
		UserID:          user.ID,
		CategoryID:      categoryUUID,
		Type:            input.Type,
		Amount:          input.Amount,
		Currency:        input.Currency,
		ExchangeRate:    exchangeRate,
		BaseAmount:      baseAmount,
		Description:     input.Description,
		TransactionDate: txnDate,
		Tags:            tagsStr,
	}

	if err := database.GetDB().Create(&txn).Error; err != nil {
		InternalError(c, "failed to create transaction")
		return
	}

	database.GetDB().Preload("Category").First(&txn, txn.ID)

	// Budget check
	overBudget := CheckBudgetOverrun(database.GetDB(), user.ID, ledgerUUID, categoryUUID, txn.Type, txn.BaseAmount, txn.TransactionDate)

	RespondJSON(c, http.StatusCreated, gin.H{
		"transaction":  txn,
		"over_budget":  overBudget,
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
	id := c.Param("id")

	var txn models.Transaction
	if err := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).First(&txn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "transaction not found")
			return
		}
		InternalError(c, "database error")
		return
	}

	var input UpdateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if input.Amount != nil {
		updates["amount"] = *input.Amount
	}
	if input.Currency != nil {
		updates["currency"] = *input.Currency
	}
	if input.Type != nil {
		updates["type"] = *input.Type
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.TransactionDate != nil {
		updates["transaction_date"] = *input.TransactionDate
	}
	if input.IsReconciled != nil {
		updates["is_reconciled"] = *input.IsReconciled
	}
	if input.CategoryID != nil {
		if parsed, err := uuid.Parse(*input.CategoryID); err == nil {
			updates["category_id"] = parsed
		}
	}
	if input.Tags != nil {
		s := ""
		for i, tag := range input.Tags {
			if i > 0 {
				s += ","
			}
			s += tag
		}
		updates["tags"] = s
	}

	amount, hasAmount := updates["amount"].(float64)
	currency, hasCurrency := updates["currency"].(string)
	if hasAmount || hasCurrency {
		if !hasAmount {
			amount = txn.Amount
		}
		if !hasCurrency {
			currency = txn.Currency
		}

		var ledger models.Ledger
		if err := database.GetDB().First(&ledger, "id = ?", txn.LedgerID).Error; err == nil {
			rate := 1.0
			if currency != ledger.BaseCurrency {
				txnDate := txn.TransactionDate
				if input.TransactionDate != nil {
					txnDate = *input.TransactionDate
				}
				r, err := services.GetExchangeRate(currency, ledger.BaseCurrency, txnDate)
				if err == nil && r > 0 {
					rate = r
				}
			}
			baseAmount := math.Round(amount*rate*100) / 100
			updates["exchange_rate"] = rate
			updates["base_amount"] = baseAmount
		}
	}

	database.GetDB().Model(&txn).Updates(updates)
	database.GetDB().Preload("Category").First(&txn, txn.ID)

	// Budget check
	overBudget := CheckBudgetOverrun(database.GetDB(), user.ID, txn.LedgerID, txn.CategoryID, txn.Type, txn.BaseAmount, txn.TransactionDate)

	RespondJSON(c, http.StatusOK, gin.H{
		"transaction":  txn,
		"over_budget":  overBudget,
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
	id := c.Param("id")

	result := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.Transaction{})
	if result.RowsAffected == 0 {
		NotFound(c, "transaction not found")
		return
	}

	RespondJSON(c, http.StatusOK, nil)
}

type transactionList struct {
	Items      []models.Transaction `json:"items"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
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

	// verify ownership and collect valid UUIDs
	uuids := make([]uuid.UUID, 0, len(input.IDs))
	for _, id := range input.IDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			BadRequest(c, "invalid id: "+id)
			return
		}
		uuids = append(uuids, parsed)
	}

	var count int64
	database.GetDB().Model(&models.Transaction{}).Where("id IN ? AND user_id = ?", uuids, user.ID).Count(&count)
	if count != int64(len(uuids)) {
		BadRequest(c, "some transactions not found or not owned by current user")
		return
	}

	result := database.GetDB().Where("id IN ?", uuids).Delete(&models.Transaction{})
	RespondJSON(c, http.StatusOK, map[string]interface{}{
		"deleted": result.RowsAffected,
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

	// verify category exists and belongs to user
	catUUID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		BadRequest(c, "invalid category_id")
		return
	}
	var cat models.Category
	if err := database.GetDB().Where("id = ? AND user_id = ?", catUUID, user.ID).First(&cat).Error; err != nil {
		NotFound(c, "category not found")
		return
	}

	// verify all transaction IDs
	uuids := make([]uuid.UUID, 0, len(input.IDs))
	for _, id := range input.IDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			BadRequest(c, "invalid id: "+id)
			return
		}
		uuids = append(uuids, parsed)
	}

	var count int64
	database.GetDB().Model(&models.Transaction{}).Where("id IN ? AND user_id = ?", uuids, user.ID).Count(&count)
	if count != int64(len(uuids)) {
		BadRequest(c, "some transactions not found or not owned by current user")
		return
	}

	result := database.GetDB().Model(&models.Transaction{}).Where("id IN ?", uuids).Update("category_id", catUUID)
	RespondJSON(c, http.StatusOK, map[string]interface{}{
		"updated": result.RowsAffected,
	})
}
