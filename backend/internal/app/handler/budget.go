package handler

import (
	"errors"
	"net/http"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BudgetHandler struct {
	svc *service.BudgetService
}

func NewBudgetHandler(svc *service.BudgetService) *BudgetHandler {
	return &BudgetHandler{svc: svc}
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

	var catUUID *uuid.UUID
	if input.CategoryID != nil && *input.CategoryID != "" {
		parsed, err := uuid.Parse(*input.CategoryID)
		if err != nil {
			BadRequest(c, "invalid category_id")
			return
		}
		catUUID = &parsed
	}

	budget, err := h.svc.UpsertBudget(user.ID, ledgerUUID, catUUID, input.Month, input.Amount)
	if err != nil {
		InternalError(c, "failed to upsert budget")
		return
	}

	RespondJSON(c, http.StatusCreated, budget)
}

// List  godoc
// @Summary      查询预算列表
// @Tags         budgets
// @Produce      json
// @Security     BearerAuth
// @Param        month     query string true  "月份 YYYY-MM"
// @Param        ledger_id query string true  "账本 ID"
// @Success      200 {object} Response{data=[]models.Budget}
// @Router       /budgets [get]
func (h *BudgetHandler) List(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	month := c.Query("month")
	if month == "" {
		BadRequest(c, "month is required (YYYY-MM)")
		return
	}

	ledgerID := c.Query("ledger_id")
	if ledgerID == "" {
		BadRequest(c, "ledger_id is required")
		return
	}

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	budgets, err := h.svc.ListBudgets(user.ID, lid, month)
	if err != nil {
		InternalError(c, "failed to query budgets")
		return
	}

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

	results, err := h.svc.Status(lid, user.ID, month)
	if err != nil {
		InternalError(c, "failed to query budget status")
		return
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

	uid, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, "invalid budget id")
		return
	}

	if err := h.svc.DeleteBudget(uid, user.ID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			NotFound(c, "budget not found")
			return
		}
		InternalError(c, "failed to delete budget")
		return
	}

	RespondJSON(c, http.StatusOK, nil)
}
