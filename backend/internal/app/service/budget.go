package service

import (
	"errors"
	"fmt"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/pkg/strutil"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Sentinel errors for service layer.
var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("conflict error")
)

// BudgetStatusItem 预算执行状态项
type BudgetStatusItem struct {
	BudgetID   string  `json:"budget_id,omitempty"`
	CategoryID string  `json:"category_id,omitempty"`
	Name       string  `json:"name,omitempty"`
	Icon       string  `json:"icon,omitempty"`
	Budget     float64 `json:"budget"`
	Spent      float64 `json:"spent"`
	Percentage float64 `json:"percentage"`
}

// UpsertBudget 创建或更新预算（同一 ledger + category + month 覆盖）
func (s *BudgetService) UpsertBudget(userID, ledgerID uuid.UUID, categoryID *uuid.UUID, month string, amount float64) (*models.Budget, error) {
	var budget *models.Budget
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Delete existing matching budget
		delQuery := tx.Where("user_id = ? AND ledger_id = ? AND month = ?", userID, ledgerID, month)
		if categoryID != nil {
			delQuery = delQuery.Where("category_id = ?", categoryID)
		} else {
			delQuery = delQuery.Where("category_id IS NULL")
		}
		if err := delQuery.Delete(&models.Budget{}).Error; err != nil {
			return fmt.Errorf("failed to clear existing budget: %w", err)
		}

		b := models.Budget{
			UserID:     userID,
			LedgerID:   ledgerID,
			CategoryID: categoryID,
			Month:      month,
			Amount:     amount,
		}

		if err := tx.Create(&b).Error; err != nil {
			return fmt.Errorf("failed to create budget: %w", err)
		}
		budget = &b
		return nil
	})
	if err != nil {
		return nil, err
	}
	return budget, nil
}

// ListBudgets 查询预算列表
func (s *BudgetService) ListBudgets(userID, ledgerID uuid.UUID, month string) ([]models.Budget, error) {
	var budgets []models.Budget
	if err := s.DB.Where("user_id = ? AND ledger_id = ? AND month = ?", userID, ledgerID, month).
		Order("created_at desc").Find(&budgets).Error; err != nil {
		return nil, fmt.Errorf("failed to query budgets: %w", err)
	}
	return budgets, nil
}

// Status 查询预算执行状态（含实际支出）
func (s *BudgetService) Status(ledgerID, userID uuid.UUID, month string) ([]BudgetStatusItem, error) {
	startDate := month + "-01"
	endDate := month + "-32"

	var results []BudgetStatusItem

	// Category-level budgets with actual spending
	type budgetRow struct {
		BudgetID   uuid.UUID
		CategoryID uuid.UUID
		Amount     float64
	}
	var budgets []budgetRow
	s.DB.Raw(`SELECT b.id AS budget_id, b.category_id, b.amount
		FROM budgets b
		WHERE b.user_id = ? AND b.ledger_id = ? AND b.month = ? AND b.category_id IS NOT NULL AND b.deleted_at IS NULL`,
		userID, ledgerID, month).Scan(&budgets)

	for _, b := range budgets {
		var spent float64
		s.DB.Raw(`SELECT COALESCE(SUM(base_amount), 0) FROM transactions
			WHERE user_id = ? AND ledger_id = ? AND category_id = ? AND type = 'expense'
			AND transaction_date >= ? AND transaction_date < ? AND deleted_at IS NULL`,
			userID, ledgerID, b.CategoryID, startDate, endDate).Scan(&spent)

		var cat models.Category
		if err := s.DB.First(&cat, b.CategoryID).Error; err != nil {
			continue
		}

		pct := 0.0
		if b.Amount > 0 {
			pct = spent / b.Amount * 100
		}
		results = append(results, BudgetStatusItem{
			BudgetID:   b.BudgetID.String(),
			CategoryID: b.CategoryID.String(),
			Name:       cat.Name,
			Icon:       strutil.NullableStr(cat.Icon),
			Budget:     b.Amount,
			Spent:      spent,
			Percentage: pct,
		})
	}

	// Global budget (category_id IS NULL) — compare against all expenses
	var globalBudget struct {
		BudgetID uuid.UUID
		Amount   float64
	}
	err := s.DB.Raw(`SELECT id AS budget_id, amount FROM budgets
		WHERE user_id = ? AND ledger_id = ? AND month = ? AND category_id IS NULL AND deleted_at IS NULL`,
		userID, ledgerID, month).Scan(&globalBudget).Error
	if err == nil && globalBudget.Amount > 0 {
		var totalSpent float64
		s.DB.Raw(`SELECT COALESCE(SUM(base_amount), 0) FROM transactions
			WHERE user_id = ? AND ledger_id = ? AND type = 'expense'
			AND transaction_date >= ? AND transaction_date < ? AND deleted_at IS NULL`,
			userID, ledgerID, startDate, endDate).Scan(&totalSpent)

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

	return results, nil
}

// DeleteBudget 删除预算
func (s *BudgetService) DeleteBudget(id, userID uuid.UUID) error {
	result := s.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Budget{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete budget: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CheckBudgetOverrun 检查新交易是否会导致预算超限。
// 返回 true 表示超限。
func (s *BudgetService) CheckBudgetOverrun(userID, ledgerID, categoryID uuid.UUID, month string, amount float64) bool {
	startDate := month + "-01"
	endDate := month + "-32"

	// Check category-level budget
	var catBudget float64
	s.DB.Raw(`SELECT COALESCE(amount, 0) FROM budgets
		WHERE user_id = ? AND ledger_id = ? AND category_id = ? AND month = ? AND deleted_at IS NULL`,
		userID, ledgerID, categoryID, month).Scan(&catBudget)

	if catBudget > 0 {
		var currentSpent float64
		s.DB.Raw(`SELECT COALESCE(SUM(base_amount), 0) FROM transactions
			WHERE user_id = ? AND ledger_id = ? AND category_id = ? AND type = 'expense'
			AND transaction_date >= ? AND transaction_date < ? AND deleted_at IS NULL`,
			userID, ledgerID, categoryID, startDate, endDate).Scan(&currentSpent)

		if currentSpent+amount > catBudget {
			return true
		}
	}

	// Check global budget
	var globalBudget float64
	s.DB.Raw(`SELECT COALESCE(amount, 0) FROM budgets
		WHERE user_id = ? AND ledger_id = ? AND category_id IS NULL AND month = ? AND deleted_at IS NULL`,
		userID, ledgerID, month).Scan(&globalBudget)

	if globalBudget > 0 {
		var totalSpent float64
		s.DB.Raw(`SELECT COALESCE(SUM(base_amount), 0) FROM transactions
			WHERE user_id = ? AND ledger_id = ? AND type = 'expense'
			AND transaction_date >= ? AND transaction_date < ? AND deleted_at IS NULL`,
			userID, ledgerID, startDate, endDate).Scan(&totalSpent)

		if totalSpent+amount > globalBudget {
			return true
		}
	}

	return false
}
