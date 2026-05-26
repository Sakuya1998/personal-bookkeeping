package handlers

import (
	"net/http"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BudgetHandler struct{}

func NewBudgetHandler() *BudgetHandler {
	return &BudgetHandler{}
}

type UpsertBudgetInput struct {
	LedgerID   string  `json:"ledger_id" binding:"required"`
	CategoryID *string `json:"category_id"` // null = global budget
	Month      string  `json:"month" binding:"required,len=7"` // "2026-05"
	Amount     float64 `json:"amount" binding:"required,gt=0"`
}

// Upsert  godoc
// @Summary      创建/更新预算（同一 ledger+category+month 覆盖）
// @Tags         budgets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body UpsertBudgetInput true "预算信息"
// @Success      201 {object} Response
// @Router       /budgets [post]
func (h *BudgetHandler) Upsert(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	var input UpsertBudgetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	ledgerUUID, err := uuid.Parse(input.LedgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}
	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", ledgerUUID, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	var catUUID *uuid.UUID
	if input.CategoryID != nil && *input.CategoryID != "" {
		parsed, err := uuid.Parse(*input.CategoryID)
		if err != nil {
			BadRequest(c, "invalid category_id")
			return
		}
		catUUID = &parsed
	}

	// Upsert: delete existing then create
	db := database.GetDB()
	delQuery := db.Where("user_id = ? AND ledger_id = ? AND month = ?", user.ID, ledgerUUID, input.Month)
	if catUUID != nil {
		delQuery = delQuery.Where("category_id = ?", catUUID)
	} else {
		delQuery = delQuery.Where("category_id IS NULL")
	}
	delQuery.Delete(&models.Budget{})

	budget := models.Budget{
		UserID:     user.ID,
		LedgerID:   ledgerUUID,
		CategoryID: catUUID,
		Month:      input.Month,
		Amount:     input.Amount,
	}

	if err := db.Create(&budget).Error; err != nil {
		InternalError(c, "failed to create budget")
		return
	}

	RespondJSON(c, http.StatusCreated, budget)
}

// List  godoc
// @Summary      查询预算列表
// @Tags         budgets
// @Produce      json
// @Security     BearerAuth
// @Param        month query string true "月份 YYYY-MM"
// @Success      200 {object} Response{data=[]models.Budget}
// @Router       /budgets [get]
func (h *BudgetHandler) List(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	month := c.Query("month")
	if month == "" {
		BadRequest(c, "month is required (YYYY-MM)")
		return
	}

	var budgets []models.Budget
	database.GetDB().Where("user_id = ? AND month = ?", user.ID, month).
		Order("created_at desc").Find(&budgets)

	RespondJSON(c, http.StatusOK, budgets)
}

// BudgetStatusItem returns actual spend vs budget
type BudgetStatusItem struct {
	BudgetID   string  `json:"budget_id,omitempty"`
	CategoryID string  `json:"category_id,omitempty"`
	Name       string  `json:"name,omitempty"`
	Icon       string  `json:"icon,omitempty"`
	Budget     float64 `json:"budget"`
	Spent      float64 `json:"spent"`
	Percentage float64 `json:"percentage"`
}

// Status  godoc
// @Summary      预算执行状态
// @Tags         budgets
// @Produce      json
// @Security     BearerAuth
// @Param        month     query string true  "月份 YYYY-MM"
// @Param        ledger_id query string true  "账本 ID"
// @Success      200 {object} Response{data=[]BudgetStatusItem}
// @Router       /budgets/status [get]
func (h *BudgetHandler) Status(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	month := c.Query("month")
	ledgerID := c.Query("ledger_id")
	if month == "" || ledgerID == "" {
		BadRequest(c, "month and ledger_id are required")
		return
	}

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", lid, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	db := database.GetDB()
	startDate := month + "-01"
	endDate := month + "-32" // safe overflow for querying

	var results []BudgetStatusItem

	// Get category-level budgets with actual spending
	type budgetRow struct {
		BudgetID   uuid.UUID
		CategoryID uuid.UUID
		Amount     float64
	}
	var budgets []budgetRow
	db.Raw(`SELECT b.id AS budget_id, b.category_id, b.amount
		FROM budgets b
		WHERE b.user_id = ? AND b.ledger_id = ? AND b.month = ? AND b.category_id IS NOT NULL`,
		user.ID, lid, month).Scan(&budgets)

	for _, b := range budgets {
		var spent float64
		db.Raw(`SELECT COALESCE(SUM(base_amount), 0) FROM transactions
			WHERE user_id = ? AND ledger_id = ? AND category_id = ? AND type = 'expense'
			AND transaction_date >= ? AND transaction_date < ?`,
			user.ID, lid, b.CategoryID, startDate, endDate).Scan(&spent)

		var cat models.Category
		db.First(&cat, b.CategoryID)

		pct := 0.0
		if b.Amount > 0 {
			pct = spent / b.Amount * 100
		}
		results = append(results, BudgetStatusItem{
			BudgetID:   b.BudgetID.String(),
			CategoryID: b.CategoryID.String(),
			Name:       cat.Name,
			Icon:       stringPtr(cat.Icon),
			Budget:     b.Amount,
			Spent:      spent,
			Percentage: pct,
		})
	}

	// Get global budget (category_id IS NULL) — compare against all expenses
	var globalBudget struct {
		BudgetID uuid.UUID
		Amount   float64
	}
	err = db.Raw(`SELECT id AS budget_id, amount FROM budgets
		WHERE user_id = ? AND ledger_id = ? AND month = ? AND category_id IS NULL`,
		user.ID, lid, month).Scan(&globalBudget).Error
	if err == nil && globalBudget.Amount > 0 {
		var totalSpent float64
		db.Raw(`SELECT COALESCE(SUM(base_amount), 0) FROM transactions
			WHERE user_id = ? AND ledger_id = ? AND type = 'expense'
			AND transaction_date >= ? AND transaction_date < ?`,
			user.ID, lid, startDate, endDate).Scan(&totalSpent)

		pct := 0.0
		if globalBudget.Amount > 0 {
			pct = totalSpent / globalBudget.Amount * 100
		}
		results = append(results, BudgetStatusItem{
			BudgetID:   globalBudget.BudgetID.String(),
			Name:       "全部支出",
			Budget:     globalBudget.Amount,
			Spent:      totalSpent,
			Percentage: pct,
		})
	}

	if results == nil {
		results = []BudgetStatusItem{}
	}

	RespondJSON(c, http.StatusOK, results)
}

// Delete  godoc
// @Summary      删除预算
// @Tags         budgets
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "预算 ID"
// @Success      200 {object} Response
// @Router       /budgets/{id} [delete]
func (h *BudgetHandler) Delete(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("id")

	result := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.Budget{})
	if result.RowsAffected == 0 {
		NotFound(c, "budget not found")
		return
	}

	RespondJSON(c, http.StatusOK, nil)
}

// CheckBudgetOverrun checks if a new transaction would exceed the budget.
// Returns (overrun, nil) — nil means under budget.
func CheckBudgetOverrun(db *gorm.DB, userID, ledgerID, categoryID uuid.UUID, txnType string, txnAmount float64, txnDate string) (overrunFlag bool) {
	if txnType != "expense" {
		return false // income never triggers budget
	}
	month := txnDate[:7] // "2026-05"

	startDate := month + "-01"
	endDate := month + "-32"

	// Check category budget
	var catBudget float64
	db.Raw(`SELECT COALESCE(amount, 0) FROM budgets
		WHERE user_id = ? AND ledger_id = ? AND category_id = ? AND month = ?`,
		userID, ledgerID, categoryID, month).Scan(&catBudget)

	if catBudget > 0 {
		var currentSpent float64
		db.Raw(`SELECT COALESCE(SUM(base_amount), 0) FROM transactions
			WHERE user_id = ? AND ledger_id = ? AND category_id = ? AND type = 'expense'
			AND transaction_date >= ? AND transaction_date < ?`,
			userID, ledgerID, categoryID, startDate, endDate).Scan(&currentSpent)

		if currentSpent+txnAmount > catBudget {
			return true
		}
	}

	// Check global budget
	var globalBudget float64
	db.Raw(`SELECT COALESCE(amount, 0) FROM budgets
		WHERE user_id = ? AND ledger_id = ? AND category_id IS NULL AND month = ?`,
		userID, ledgerID, month).Scan(&globalBudget)

	if globalBudget > 0 {
		var totalSpent float64
		db.Raw(`SELECT COALESCE(SUM(base_amount), 0) FROM transactions
			WHERE user_id = ? AND ledger_id = ? AND type = 'expense'
			AND transaction_date >= ? AND transaction_date < ?`,
			userID, ledgerID, startDate, endDate).Scan(&totalSpent)

		if totalSpent+txnAmount > globalBudget {
			return true
		}
	}

	return false
}

func stringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
