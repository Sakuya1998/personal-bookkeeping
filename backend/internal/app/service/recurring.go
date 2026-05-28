package service

import (
	"errors"
	"fmt"
	"time"

	"personal-bookkeeping/internal/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------- RecurringService methods ----------

// Create 创建周期性规则。
func (s *RecurringService) Create(
	ledgerID, categoryID, userID uuid.UUID,
	txnType string,
	amount float64,
	currency string,
	description *string,
	tags []string,
	frequency string,
	interval int,
	dayOfMonth, weekday *int,
	startDate string,
	endDate *string,
) (*models.RecurringRule, error) {
	// 验证账本所有权
	var ledger models.Ledger
	if err := s.DB.Where("id = ? AND user_id = ?", ledgerID, userID).First(&ledger).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLedgerNotFound
		}
		return nil, fmt.Errorf("failed to query ledger: %w", err)
	}
	_ = ledger

	// 验证分类所有权
	var cat models.Category
	if err := s.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&cat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query category: %w", err)
	}
	_ = cat

	if currency == "" {
		currency = "CNY"
	}
	if interval <= 0 {
		interval = 1
	}

	startDateParsed, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format, use YYYY-MM-DD: %w", err)
	}
	nextRunDate := ComputeNextRunDate(startDateParsed, frequency, interval, dayOfMonth, weekday)

	var tagsStr *string
	if len(tags) > 0 {
		s := ""
		for i, tag := range tags {
			if i > 0 {
				s += ","
			}
			s += tag
		}
		tagsStr = &s
	}

	rule := models.RecurringRule{
		UserID:      userID,
		LedgerID:    ledgerID,
		CategoryID:  categoryID,
		Type:        txnType,
		Amount:      amount,
		Currency:    currency,
		Description: description,
		Tags:        tagsStr,
		Frequency:   frequency,
		Interval:    interval,
		DayOfMonth:  dayOfMonth,
		Weekday:     weekday,
		StartDate:   startDate,
		EndDate:     endDate,
		NextRunDate: nextRunDate.Format("2006-01-02"),
		IsActive:    true,
	}

	if err := s.DB.Create(&rule).Error; err != nil {
		return nil, fmt.Errorf("failed to create recurring rule: %w", err)
	}

	return &rule, nil
}

// Update 更新周期性规则。
func (s *RecurringService) Update(id, userID uuid.UUID, updates map[string]interface{}) (*models.RecurringRule, error) {
	var rule models.RecurringRule
	if err := s.DB.Where("id = ? AND user_id = ?", id, userID).First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query recurring rule: %w", err)
	}

	if err := s.DB.Model(&rule).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update recurring rule: %w", err)
	}

	// 重新加载最新状态
	if err := s.DB.First(&rule, rule.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload recurring rule: %w", err)
	}

	return &rule, nil
}

// Delete 删除周期性规则（验证所有权）。
func (s *RecurringService) Delete(id, userID uuid.UUID) error {
	result := s.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.RecurringRule{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete recurring rule: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List 查询用户的所有周期性规则。
func (s *RecurringService) List(userID uuid.UUID) ([]models.RecurringRule, error) {
	var rules []models.RecurringRule
	if err := s.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to query recurring rules: %w", err)
	}
	return rules, nil
}

// Upcoming 查询未来7天内到期的活跃周期性规则。
func (s *RecurringService) Upcoming(userID uuid.UUID) ([]models.RecurringRule, error) {
	today := time.Now().Format("2006-01-02")
	nextWeek := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	var rules []models.RecurringRule
	if err := s.DB.
		Where("user_id = ? AND is_active = true AND next_run_date >= ? AND next_run_date <= ?", userID, today, nextWeek).
		Order("next_run_date asc").
		Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to query upcoming rules: %w", err)
	}
	return rules, nil
}

// ---------- helpers ----------

// ComputeNextRunDate 计算周期性规则的下次执行日期，返回第一个 >= from 的有效日期。
func ComputeNextRunDate(from time.Time, freq string, interval int, dayOfMonth, weekday *int) time.Time {
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
			lastDay := DaysInMonth(next.Year(), int(next.Month()))
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

// DaysInMonth 返回指定年月的天数。
func DaysInMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}
