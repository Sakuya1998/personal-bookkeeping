package handlers

import (
	"net/http"
	"time"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RecurringHandler struct{}

func NewRecurringHandler() *RecurringHandler {
	return &RecurringHandler{}
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

	var rules []models.RecurringRule
	database.GetDB().Where("user_id = ?", user.ID).Order("created_at desc").Find(&rules)
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
	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", ledgerUUID, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

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

	if input.Currency == "" {
		input.Currency = "CNY"
	}
	if input.Interval <= 0 {
		input.Interval = 1
	}

	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		BadRequest(c, "invalid start_date format, use YYYY-MM-DD")
		return
	}
	nextRunDate := computeNextRunDate(startDate, input.Frequency, input.Interval, input.DayOfMonth, input.Weekday)

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

	rule := models.RecurringRule{
		UserID:      user.ID,
		LedgerID:    ledgerUUID,
		CategoryID:  catUUID,
		Type:        input.Type,
		Amount:      input.Amount,
		Currency:    input.Currency,
		Description: input.Description,
		Tags:        tagsStr,
		Frequency:   input.Frequency,
		Interval:    input.Interval,
		DayOfMonth:  input.DayOfMonth,
		Weekday:     input.Weekday,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		NextRunDate: nextRunDate.Format("2006-01-02"),
		IsActive:    true,
	}

	if err := database.GetDB().Create(&rule).Error; err != nil {
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

	var rule models.RecurringRule
	if err := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "recurring rule not found")
			return
		}
		InternalError(c, "database error")
		return
	}

	var input UpdateRecurringInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if input.CategoryID != nil {
		if parsed, err := uuid.Parse(*input.CategoryID); err == nil {
			updates["category_id"] = parsed
		}
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

	database.GetDB().Model(&rule).Updates(updates)
	database.GetDB().First(&rule, rule.ID)

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

	result := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.RecurringRule{})
	if result.RowsAffected == 0 {
		NotFound(c, "recurring rule not found")
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

	today := time.Now().Format("2006-01-02")
	nextWeek := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	var rules []models.RecurringRule
	database.GetDB().
		Where("user_id = ? AND is_active = true AND next_run_date >= ? AND next_run_date <= ?", user.ID, today, nextWeek).
		Order("next_run_date asc").
		Find(&rules)

	RespondJSON(c, http.StatusOK, rules)
}

// ---------- helpers ----------

// computeNextRunDate calculates the next run date for a recurring rule.
// It returns the first valid date >= startDate.
func computeNextRunDate(from time.Time, freq string, interval int, dayOfMonth, weekday *int) time.Time {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	if interval <= 0 {
		interval = 1
	}

	switch freq {
	case "daily":
		return from
	case "weekly":
		if weekday == nil {
			return from
		}
		target := int(from.Weekday())
		want := *weekday % 7
		offset := (want - target + 7) % 7
		return from.AddDate(0, 0, offset)
	case "monthly":
		dom := 1
		if dayOfMonth != nil && *dayOfMonth >= 1 && *dayOfMonth <= 31 {
			dom = *dayOfMonth
		}
		// Try the target day; if it exceeds days in month, clamp to last day.
		next := time.Date(from.Year(), from.Month(), dom, 0, 0, 0, 0, time.UTC)
		if next.Before(from) {
			// next month
			next = time.Date(from.Year(), from.Month()+1, 1, 0, 0, 0, 0, time.UTC)
			lastDay := daysInMonth(next.Year(), int(next.Month()))
			if dom > lastDay {
				dom = lastDay
			}
			next = time.Date(next.Year(), next.Month(), dom, 0, 0, 0, 0, time.UTC)
		}
		return next
	case "yearly":
		next := time.Date(from.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		if !next.After(from) {
			next = time.Date(from.Year()+1, time.January, 1, 0, 0, 0, 0, time.UTC)
		}
		return next
	default:
		return from
	}
}

// ComputeNextRunDate is exported for use in the cron task.
func ComputeNextRunDate(from time.Time, freq string, interval int, dayOfMonth, weekday *int) string {
	next := computeNextRunDate(from, freq, interval, dayOfMonth, weekday)
	return next.Format("2006-01-02")
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}
