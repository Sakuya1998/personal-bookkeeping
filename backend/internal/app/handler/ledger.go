package handlers

import (
	"net/http"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"

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
	id := c.Param("id")

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
	id := c.Param("id")

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
	id := c.Param("id")

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
	id := c.Param("id")

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
