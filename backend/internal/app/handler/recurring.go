package handler

import (
	"net/http"
	"time"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RecurringHandler struct {
	svc *service.RecurringService
}

func NewRecurringHandler(svc *service.RecurringService) *RecurringHandler {
	return &RecurringHandler{svc: svc}
}

type CreateRecurringInput struct {
	LedgerID    string   `json:"ledger_id" binding:"required"`
	CategoryID  string   `json:"category_id" binding:"required"`
	Type        string   `json:"type" binding:"required,oneof=income expense"`
	Amount      float64  `json:"amount" binding:"required,gt=0"`
	Currency    string   `json:"currency"`
	Description *string  `json:"description"`
	Tags        []string `json:"tags"`
	Frequency   string   `json:"frequency" binding:"required,oneof=daily weekly monthly yearly"`
	Interval    int      `json:"interval"`
	DayOfMonth  *int     `json:"day_of_month"`
	Weekday     *int     `json:"weekday"`
	StartDate   string   `json:"start_date" binding:"required"`
	EndDate     *string  `json:"end_date"`
}

type UpdateRecurringInput struct {
	CategoryID  *string  `json:"category_id"`
	Type        *string  `json:"type" binding:"omitempty,oneof=income expense"`
	Amount      *float64 `json:"amount" binding:"omitempty,gt=0"`
	Currency    *string  `json:"currency"`
	Description *string  `json:"description"`
	Tags        []string `json:"tags"`
	Frequency   *string  `json:"frequency" binding:"omitempty,oneof=daily weekly monthly yearly"`
	Interval    *int     `json:"interval"`
	DayOfMonth  *int     `json:"day_of_month"`
	Weekday     *int     `json:"weekday"`
	EndDate     *string  `json:"end_date"`
	IsActive    *bool    `json:"is_active"`
}

// List  godoc
// @Summary      周期性规则列表
// @Tags         recurring
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} Response{data=[]models.RecurringRule}
// @Router       /recurring [get]
func (h *RecurringHandler) List(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	rules, err := h.svc.List(user.ID)
	if err != nil {
		InternalError(c, "failed to query recurring rules")
		return
	}
	RespondJSON(c, http.StatusOK, rules)
}

// Create  godoc
// @Summary      创建周期性规则
// @Tags         recurring
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body CreateRecurringInput true "规则信息"
// @Success      201 {object} Response
// @Router       /recurring [post]
func (h *RecurringHandler) Create(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	var input CreateRecurringInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	ledgerUUID, err := uuid.Parse(input.LedgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}
	catUUID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		BadRequest(c, "invalid category_id")
		return
	}

	rule, err := h.svc.Create(
		ledgerUUID, catUUID, user.ID,
		input.Type, input.Amount, input.Currency,
		input.Description, input.Tags,
		input.Frequency, input.Interval,
		input.DayOfMonth, input.Weekday,
		input.StartDate, input.EndDate,
	)
	if err != nil {
		if err == service.ErrLedgerNotFound {
			NotFound(c, "ledger not found")
			return
		}
		if err == service.ErrNotFound {
			NotFound(c, "category not found")
			return
		}
		InternalError(c, "failed to create recurring rule")
		return
	}

	RespondJSON(c, http.StatusCreated, rule)
}

// Update  godoc
// @Summary      更新周期性规则
// @Tags         recurring
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path string true "规则 ID"
// @Param        input body UpdateRecurringInput true "更新内容"
// @Success      200 {object} Response
// @Router       /recurring/{id} [put]
func (h *RecurringHandler) Update(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("id")

	var input UpdateRecurringInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	ruleID, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, "invalid id")
		return
	}

	updates := map[string]interface{}{}
	if input.CategoryID != nil {
		parsed, err := uuid.Parse(*input.CategoryID)
		if err != nil {
			BadRequest(c, "invalid category_id")
			return
		}
		updates["category_id"] = parsed
	}
	if input.Type != nil {
		updates["type"] = *input.Type
	}
	if input.Amount != nil {
		updates["amount"] = *input.Amount
	}
	if input.Currency != nil {
		updates["currency"] = *input.Currency
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Frequency != nil {
		updates["frequency"] = *input.Frequency
	}
	if input.Interval != nil {
		updates["interval"] = *input.Interval
	}
	if input.DayOfMonth != nil {
		updates["day_of_month"] = *input.DayOfMonth
	}
	if input.Weekday != nil {
		updates["weekday"] = *input.Weekday
	}
	if input.EndDate != nil {
		updates["end_date"] = *input.EndDate
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
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

	rule, err := h.svc.Update(ruleID, user.ID, updates)
	if err != nil {
		if err == service.ErrNotFound {
			NotFound(c, "recurring rule not found")
			return
		}
		InternalError(c, "failed to update recurring rule")
		return
	}

	RespondJSON(c, http.StatusOK, rule)
}

// Delete  godoc
// @Summary      删除周期性规则
// @Tags         recurring
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "规则 ID"
// @Success      200 {object} Response
// @Router       /recurring/{id} [delete]
func (h *RecurringHandler) Delete(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("id")

	ruleID, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, "invalid id")
		return
	}

	if err := h.svc.Delete(ruleID, user.ID); err != nil {
		if err == service.ErrNotFound {
			NotFound(c, "recurring rule not found")
			return
		}
		InternalError(c, "failed to delete recurring rule")
		return
	}

	RespondJSON(c, http.StatusOK, nil)
}

// Upcoming  godoc
// @Summary      即将到期（未来7天内）的周期性规则
// @Tags         recurring
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} Response{data=[]models.RecurringRule}
// @Router       /recurring/upcoming [get]
func (h *RecurringHandler) Upcoming(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	rules, err := h.svc.Upcoming(user.ID)
	if err != nil {
		InternalError(c, "failed to query upcoming rules")
		return
	}

	RespondJSON(c, http.StatusOK, rules)
}

// ComputeNextRunDate is exported for use in the cron task.
func ComputeNextRunDate(from time.Time, freq string, interval int, dayOfMonth, weekday *int) string {
	next := service.ComputeNextRunDate(from, freq, interval, dayOfMonth, weekday)
	return next.Format("2006-01-02")
}
