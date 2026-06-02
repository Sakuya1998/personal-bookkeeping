package service

import (
	"math"
	"time"

	"personal-bookkeeping/internal/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateTransaction 创建交易记录：汇率转换 + 预算检查
func (s *TransactionService) CreateTransaction(
	ledgerID, userID, categoryID uuid.UUID,
	txnType string,
	amount float64,
	currency string,
	description *string,
	transactionDate string,
	tags []string,
) (*models.Transaction, bool, error) {
	if currency == "" {
		currency = "CNY"
	}

	if transactionDate == "" {
		transactionDate = time.Now().Format("2006-01-02")
	}

	// 查询账本获取 base_currency
	var ledger models.Ledger
	if err := s.DB.Where("id = ? AND user_id = ?", ledgerID, userID).First(&ledger).Error; err != nil {
		return nil, false, err
	}

	exchangeRate := 1.0
	baseAmount := amount
	if currency != ledger.BaseCurrency {
		rate, err := GetExchangeRate(currency, ledger.BaseCurrency, transactionDate)
		if err == nil && rate > 0 {
			exchangeRate = rate
			baseAmount = amount * rate
		}
	}
	baseAmount = math.Round(baseAmount*100) / 100

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

	txn := models.Transaction{
		ID:              uuid.New(),
		LedgerID:        ledgerID,
		UserID:          userID,
		CategoryID:      categoryID,
		Type:            txnType,
		Amount:          amount,
		Currency:        currency,
		ExchangeRate:    exchangeRate,
		BaseAmount:      baseAmount,
		Description:     description,
		TransactionDate: transactionDate,
		Tags:            tagsStr,
	}

	if err := s.DB.Create(&txn).Error; err != nil {
		return nil, false, err
	}

	// 重新加载预加载 Category
	if err := s.DB.Preload("Category").First(&txn, txn.ID).Error; err != nil {
		return nil, false, err
	}

	// Budget check
	overBudget := checkBudgetOverrun(s.DB, userID, ledgerID, categoryID, txn.Type, txn.BaseAmount, txn.TransactionDate)

	return &txn, overBudget, nil
}

// UpdateTransaction 更新交易记录：汇率转换 + 预算检查
func (s *TransactionService) UpdateTransaction(
	id, userID uuid.UUID,
	categoryID *uuid.UUID,
	txnType *string,
	amount *float64,
	currency *string,
	description *string,
	transactionDate *string,
	tags []string,
	isReconciled *bool,
) (*models.Transaction, bool, error) {
	var txn models.Transaction
	if err := s.DB.Where("id = ? AND user_id = ?", id, userID).First(&txn).Error; err != nil {
		return nil, false, err
	}

	updates := map[string]interface{}{}
	if amount != nil {
		updates["amount"] = *amount
	}
	if currency != nil {
		updates["currency"] = *currency
	}
	if txnType != nil {
		updates["type"] = *txnType
	}
	if description != nil {
		updates["description"] = *description
	}
	if transactionDate != nil {
		updates["transaction_date"] = *transactionDate
	}
	if isReconciled != nil {
		updates["is_reconciled"] = *isReconciled
	}
	if categoryID != nil {
		updates["category_id"] = *categoryID
	}
	if tags != nil {
		s := ""
		for i, tag := range tags {
			if i > 0 {
				s += ","
			}
			s += tag
		}
		updates["tags"] = s
	}

	newAmount, hasAmount := updates["amount"].(float64)
	newCurrency, hasCurrency := updates["currency"].(string)
	if hasAmount || hasCurrency {
		if !hasAmount {
			newAmount = txn.Amount
		}
		if !hasCurrency {
			newCurrency = txn.Currency
		}

		var ledger models.Ledger
		if err := s.DB.First(&ledger, "id = ?", txn.LedgerID).Error; err == nil {
			rate := 1.0
			if newCurrency != ledger.BaseCurrency {
				txnDate := txn.TransactionDate
				if transactionDate != nil {
					txnDate = *transactionDate
				}
				r, err := GetExchangeRate(newCurrency, ledger.BaseCurrency, txnDate)
				if err == nil && r > 0 {
					rate = r
				}
			}
			baseAmount := math.Round(newAmount*rate*100) / 100
			updates["exchange_rate"] = rate
			updates["base_amount"] = baseAmount
		}
	}

	if err := s.DB.Model(&txn).Updates(updates).Error; err != nil {
		return nil, false, err
	}

	// 重新加载
	if err := s.DB.Preload("Category").First(&txn, txn.ID).Error; err != nil {
		return nil, false, err
	}

	// Budget check
	overBudget := checkBudgetOverrun(s.DB, userID, txn.LedgerID, txn.CategoryID, txn.Type, txn.BaseAmount, txn.TransactionDate)

	return &txn, overBudget, nil
}

// DeleteTransaction 删除交易记录（验证所有权）
func (s *TransactionService) DeleteTransaction(id, userID uuid.UUID) error {
	result := s.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Transaction{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListTransactions 分页查询交易记录
func (s *TransactionService) ListTransactions(
	ledgerID, userID uuid.UUID,
	page, pageSize int,
	filters map[string]string,
) ([]models.Transaction, int64, int, error) {
	// 验证账本所有权
	var ledger models.Ledger
	if err := s.DB.Where("id = ? AND user_id = ?", ledgerID, userID).First(&ledger).Error; err != nil {
		return nil, 0, 0, err
	}

	query := s.DB.Where("ledger_id = ? AND user_id = ?", ledgerID, userID)
	if t, ok := filters["type"]; ok && t != "" {
		query = query.Where("type = ?", t)
	}
	if catID, ok := filters["category_id"]; ok && catID != "" {
		query = query.Where("category_id = ?", catID)
	}
	if start, ok := filters["start_date"]; ok && start != "" {
		query = query.Where("transaction_date >= ?", start)
	}
	if end, ok := filters["end_date"]; ok && end != "" {
		query = query.Where("transaction_date <= ?", end)
	}
	if keyword, ok := filters["keyword"]; ok && keyword != "" {
		query = query.Where("description ILIKE ?", "%"+keyword+"%")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := query.Model(&models.Transaction{}).Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	var transactions []models.Transaction
	if err := query.Preload("Category").
		Order("transaction_date desc, created_at desc").
		Offset(offset).Limit(pageSize).
		Find(&transactions).Error; err != nil {
		return nil, 0, 0, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return transactions, total, totalPages, nil
}

// BatchDelete 批量删除交易（验证所有权）
func (s *TransactionService) BatchDelete(ids []uuid.UUID, userID uuid.UUID) (int64, error) {
	var rowsAffected int64
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Transaction{}).
			Where("id IN ? AND user_id = ?", ids, userID).
			Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(ids)) {
			return gorm.ErrRecordNotFound
		}

		result := tx.Where("id IN ?", ids).Delete(&models.Transaction{})
		if result.Error != nil {
			return result.Error
		}
		rowsAffected = result.RowsAffected
		return nil
	})
	return rowsAffected, err
}

// BatchUpdateCategory 批量修改交易分类（验证所有权）
func (s *TransactionService) BatchUpdateCategory(ids []uuid.UUID, categoryID uuid.UUID, userID uuid.UUID) (int64, error) {
	var rowsAffected int64
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// 验证分类存在且属于用户
		var cat models.Category
		if err := tx.Where("id = ? AND user_id = ?", categoryID, userID).First(&cat).Error; err != nil {
			return err
		}

		// 验证交易所有权
		var count int64
		if err := tx.Model(&models.Transaction{}).
			Where("id IN ? AND user_id = ?", ids, userID).
			Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(ids)) {
			return gorm.ErrRecordNotFound
		}

		result := tx.Model(&models.Transaction{}).
			Where("id IN ?", ids).
			Update("category_id", categoryID)
		if result.Error != nil {
			return result.Error
		}
		rowsAffected = result.RowsAffected
		return nil
	})
	return rowsAffected, err
}

// checkBudgetOverrun 检查交易是否会导致预算超支
func checkBudgetOverrun(db *gorm.DB, userID, ledgerID, categoryID uuid.UUID, txnType string, txnAmount float64, txnDate string) bool {
	if txnType != "expense" {
		return false
	}
	month := txnDate[:7]
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
