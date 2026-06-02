package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/infra/queue"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------- sentinel errors ----------

var ErrLedgerNotFound = errors.New("ledger not found")

// ---------- response types ----------

// LedgerSummary 账本汇总数据。
type LedgerSummary struct {
	TotalIncome       float64           `json:"total_income"`
	TotalExpense      float64           `json:"total_expense"`
	Balance           float64           `json:"balance"`
	BaseCurrency      string            `json:"base_currency"`
	ExpenseByCategory []LedgerCategorySummary `json:"expense_by_category"`
}

// LedgerCategorySummary 分类汇总单项。
type LedgerCategorySummary struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	CategoryIcon string    `json:"category_icon"`
	Total        float64   `json:"total"`
	Count        int64     `json:"count"`
}

// MonthlyTrendItem 月度收支趋势单项。
type MonthlyTrendItem struct {
	Month   string  `json:"month" example:"2026-01"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

// CategoryBreakdownItem 分类分布单项。
type CategoryBreakdownItem struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	CategoryIcon string    `json:"category_icon"`
	Type         string    `json:"type"`
	Total        float64   `json:"total"`
	Percentage   float64   `json:"percentage"`
}

// TagStatsItem 标签统计单项。
type TagStatsItem struct {
	Tag              string  `json:"tag"`
	TotalExpense     float64 `json:"total_expense"`
	TotalIncome      float64 `json:"total_income"`
	TransactionCount int64   `json:"transaction_count"`
	Percentage       float64 `json:"percentage"`
}

// DailyTransactionItem 每日交易汇总单项。
type DailyTransactionItem struct {
	Date    string  `json:"date" example:"2026-05-01"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Count   int64   `json:"count"`
}

// ---------- LedgerService methods ----------

// CreateLedger 创建账本并自动添加创建者为 owner。。
func (s *LedgerService) CreateLedger(userID uuid.UUID, name string, currency string, description, icon, color *string) (*models.Ledger, error) {
	if currency == "" {
		currency = "CNY"
	}

	ledger := models.Ledger{
		ID:           uuid.New(),
		UserID:       userID,
		Name:         name,
		Description:  description,
		BaseCurrency: currency,
		Icon:         icon,
		Color:        color,
	}

	if err := s.DB.Create(&ledger).Error; err != nil {
		return nil, fmt.Errorf("failed to create ledger: %w", err)
	}

	// Auto-add creator as owner member
	member := models.LedgerMember{
		LedgerID: ledger.ID,
		UserID:   userID,
		Role:     models.RoleOwner,
		JoinedAt: ledger.CreatedAt,
	}
	if err := s.DB.Create(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to add owner member: %w", err)
	}

	return &ledger, nil
}

// GetLedger 查询单个账本，返回 (ledger, nil) 或 (nil, ErrLedgerNotFound)。
// 通过 member 表校验访问权限。
func (s *LedgerService) GetLedger(id, userID uuid.UUID) (*models.Ledger, error) {
	// Verify membership first
	var memberCount int64
	s.DB.Model(&models.LedgerMember{}).Where("ledger_id = ? AND user_id = ?", id, userID).Count(&memberCount)
	if memberCount == 0 {
		return nil, ErrLedgerNotFound
	}

	var ledger models.Ledger
	if err := s.DB.Where("id = ?", id).First(&ledger).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLedgerNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &ledger, nil
}

// ListLedgers 查询用户的所有账本（通过 member 表）。
func (s *LedgerService) ListLedgers(userID uuid.UUID) ([]models.Ledger, error) {
	var ledgers []models.Ledger
	if err := s.DB.
		Joins("JOIN ledger_members ON ledger_members.ledger_id = ledgers.id").
		Where("ledger_members.user_id = ?", userID).
		Preload("User").
		Order("ledgers.sort_order asc, ledgers.created_at desc").
		Find(&ledgers).Error; err != nil {
		return nil, fmt.Errorf("failed to list ledgers: %w", err)
	}
	return ledgers, nil
}

// UpdateLedger 更新账本字段。
// updates 可以包含：name, description, base_currency, icon, color, is_archived, sort_order。
// 通过 member 表校验访问权限，仅 owner/admin 可更新。
func (s *LedgerService) UpdateLedger(id, userID uuid.UUID, updates map[string]interface{}) (*models.Ledger, error) {
	// Check membership and role
	var member models.LedgerMember
	if err := s.DB.Where("ledger_id = ? AND user_id = ?", id, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLedgerNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	if member.Role != models.RoleOwner && member.Role != models.RoleAdmin {
		return nil, errors.New("only owner or admin can update the ledger")
	}

	var ledger models.Ledger
	if err := s.DB.Where("id = ?", id).First(&ledger).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLedgerNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if err := s.DB.Model(&ledger).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update ledger: %w", err)
	}

	// Reload to get the latest state
	s.DB.First(&ledger, ledger.ID)
	return &ledger, nil
}

// DeleteLedger 级联删除账本及其关联数据（transactions, categories, budgets, recurring_rules）。
func (s *LedgerService) DeleteLedger(id, userID uuid.UUID) error {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// 先删除关联表记录，再删账本（外键约束）
		// Delete member records first
		if err := tx.Where("ledger_id = ?", id).Delete(&models.LedgerMember{}).Error; err != nil {
			return err
		}
		if err := tx.Where("ledger_id = ?", id).Delete(&models.Transaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("ledger_id = ?", id).Delete(&models.Category{}).Error; err != nil {
			return err
		}
		if err := tx.Where("ledger_id = ?", id).Delete(&models.Budget{}).Error; err != nil {
			return err
		}
		if err := tx.Where("ledger_id = ?", id).Delete(&models.RecurringRule{}).Error; err != nil {
			return err
		}

		result := tx.Where("id = ?", id).Delete(&models.Ledger{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrLedgerNotFound
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Summary 获取账本汇总统计。
func (s *LedgerService) Summary(ledgerID, userID uuid.UUID) (*LedgerSummary, error) {
	// Verify ledger ownership
	ledger, err := s.GetLedger(ledgerID, userID)
	if err != nil {
		return nil, err
	}

	var totalIncome, totalExpense float64
	s.DB.Model(&models.Transaction{}).
		Where("ledger_id = ? AND type = ?", ledgerID, "income").
		Select("COALESCE(SUM(base_amount), 0)").Scan(&totalIncome)

	s.DB.Model(&models.Transaction{}).
		Where("ledger_id = ? AND type = ?", ledgerID, "expense").
		Select("COALESCE(SUM(base_amount), 0)").Scan(&totalExpense)

	var expenseByCategory []LedgerCategorySummary
	s.DB.Raw(`
		SELECT t.category_id, c.name as category_name, COALESCE(c.icon,'') as category_icon,
		       COALESCE(SUM(t.base_amount),0) as total, COUNT(*) as count
		FROM transactions t
		JOIN categories c ON c.id = t.category_id
		WHERE t.ledger_id = ? AND t.type = 'expense' AND t.deleted_at IS NULL
		GROUP BY t.category_id, c.name, c.icon
		ORDER BY total DESC
	`, ledgerID).Scan(&expenseByCategory)

	return &LedgerSummary{
		TotalIncome:       totalIncome,
		TotalExpense:      totalExpense,
		Balance:           totalIncome - totalExpense,
		BaseCurrency:      ledger.BaseCurrency,
		ExpenseByCategory: expenseByCategory,
	}, nil
}

// Tags 获取账本所有标签（去重）。
func (s *LedgerService) Tags(ledgerID, userID uuid.UUID) ([]string, error) {
	// Verify ledger ownership
	if _, err := s.GetLedger(ledgerID, userID); err != nil {
		return nil, err
	}

	var rawTags []string
	s.DB.Model(&models.Transaction{}).
		Where("ledger_id = ? AND user_id = ? AND tags IS NOT NULL AND tags <> ''", ledgerID, userID).
		Distinct("tags").
		Pluck("tags", &rawTags)

	// Split and deduplicate
	seen := make(map[string]struct{})
	var tags []string
	for _, raw := range rawTags {
		for _, t := range splitTags(raw) {
			t = trim(t)
			if t == "" {
				continue
			}
			if _, ok := seen[t]; !ok {
				seen[t] = struct{}{}
				tags = append(tags, t)
			}
		}
	}

	if tags == nil {
		tags = []string{}
	}
	return tags, nil
}

// ---------- Export ----------

// ExportTransactions 查询导出范围内的交易记录，返回 (交易列表, 总数, error)。
func (s *LedgerService) ExportTransactions(ledgerID, userID uuid.UUID, startDate, endDate string) ([]models.Transaction, int64, error) {
	// Count first
	var total int64
	countQuery := s.DB.Model(&models.Transaction{}).
		Where("ledger_id = ? AND user_id = ?", ledgerID, userID)
	if startDate != "" {
		countQuery = countQuery.Where("transaction_date >= ?", startDate)
	}
	if endDate != "" {
		countQuery = countQuery.Where("transaction_date <= ?", endDate)
	}
	countQuery.Count(&total)

	// Fetch records
	query := s.DB.Where("ledger_id = ? AND user_id = ?", ledgerID, userID)
	if startDate != "" {
		query = query.Where("transaction_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("transaction_date <= ?", endDate)
	}
	var transactions []models.Transaction
	if err := query.Order("transaction_date desc").Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query export transactions: %w", err)
	}

	return transactions, total, nil
}

// SubmitExportTask 提交异步导出任务。
func (s *LedgerService) SubmitExportTask(userID, ledgerID uuid.UUID, startDate, endDate, format string) (string, error) {
	if s.Queue == nil {
		return "", errors.New("task queue not available")
	}

	taskID := uuid.New().String()
	if err := s.Queue.Submit(nil, queue.Task{
		ID:   taskID,
		Type: "export_report",
		Payload: map[string]interface{}{
			"user_id":    userID.String(),
			"ledger_id":  ledgerID.String(),
			"start_date": startDate,
			"end_date":   endDate,
			"format":     format,
		},
	}); err != nil {
		return "", fmt.Errorf("failed to submit export task: %w", err)
	}
	return taskID, nil
}

// ---------- Analytics ----------

// MonthlyTrend 获取月度收支趋势。
func (s *LedgerService) MonthlyTrend(ledgerID, userID uuid.UUID, months int) ([]MonthlyTrendItem, error) {
	if months < 1 {
		months = 6
	}
	if months > 24 {
		months = 24
	}

	endDate := time.Now().AddDate(0, 1, 0) // include current month
	startDate := endDate.AddDate(0, -months, 0)

	type row struct {
		Month        string
		TotalIncome  float64
		TotalExpense float64
	}
	var rows []row
	s.DB.Raw(`
		SELECT
			to_char(transaction_date, 'YYYY-MM') AS month,
			COALESCE(SUM(CASE WHEN type = 'income'  THEN base_amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN base_amount ELSE 0 END), 0) AS total_expense
		FROM transactions
		WHERE ledger_id = ? AND user_id = ? AND transaction_date >= ? AND transaction_date < ? AND deleted_at IS NULL
		GROUP BY month
		ORDER BY month ASC
	`, ledgerID, userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).Scan(&rows)

	items := make([]MonthlyTrendItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, MonthlyTrendItem{
			Month:   r.Month,
			Income:  r.TotalIncome,
			Expense: r.TotalExpense,
		})
	}
	return items, nil
}

// CategoryBreakdown 获取分类支出/收入分布。
func (s *LedgerService) CategoryBreakdown(ledgerID, userID uuid.UUID, startDate, endDate, txnType string) ([]CategoryBreakdownItem, error) {
	type row struct {
		CategoryID   uuid.UUID
		CategoryName string
		CategoryIcon string
		Type         string
		Total        float64
	}

	query := s.DB.Table("transactions t").
		Select(`
			t.category_id,
			COALESCE(c.name, '')    AS category_name,
			COALESCE(c.icon, '')    AS category_icon,
			t.type,
			COALESCE(SUM(t.base_amount), 0) AS total
		`).
		Joins("LEFT JOIN categories c ON c.id = t.category_id").
		Where("t.ledger_id = ? AND t.user_id = ?", ledgerID, userID)

	if startDate != "" {
		query = query.Where("t.transaction_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("t.transaction_date <= ?", endDate)
	}
	if txnType != "" {
		query = query.Where("t.type = ?", txnType)
	}

	var rows []row
	query = query.Group("t.category_id, c.name, c.icon, t.type").Order("total DESC")
	if err := query.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to query category breakdown: %w", err)
	}

	// Calculate total for percentages
	var grandTotal float64
	for _, r := range rows {
		grandTotal += r.Total
	}

	items := make([]CategoryBreakdownItem, 0, len(rows))
	for _, r := range rows {
		pct := 0.0
		if grandTotal > 0 {
			pct = r.Total / grandTotal * 100
		}
		items = append(items, CategoryBreakdownItem{
			CategoryID:   r.CategoryID,
			CategoryName: r.CategoryName,
			CategoryIcon: r.CategoryIcon,
			Type:         r.Type,
			Total:        r.Total,
			Percentage:   pct,
		})
	}
	return items, nil
}

// TagStats 获取账本内各标签的收支统计。
// 支持可选的时间范围筛选。百分比基于总支出+总收入计算。
func (s *LedgerService) TagStats(ledgerID, userID uuid.UUID, startDate, endDate string) ([]TagStatsItem, error) {
	type row struct {
		Tag              string
		TotalExpense     float64
		TotalIncome      float64
		TransactionCount int64
	}

	query := s.DB.Table("transactions").
		Select(`
			TRIM(t.tag) AS tag,
			COALESCE(SUM(CASE WHEN transactions.type = 'expense' THEN transactions.base_amount ELSE 0 END), 0) AS total_expense,
			COALESCE(SUM(CASE WHEN transactions.type = 'income' THEN transactions.base_amount ELSE 0 END), 0) AS total_income,
			COUNT(*) AS transaction_count
		`).
		Joins("CROSS JOIN LATERAL unnest(string_to_array(transactions.tags, ',')) AS t(tag)").
		Where("transactions.ledger_id = ? AND transactions.user_id = ?", ledgerID, userID).
		Where("transactions.tags IS NOT NULL AND transactions.tags <> ''")

	if startDate != "" {
		query = query.Where("transactions.transaction_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("transactions.transaction_date <= ?", endDate)
	}

	var rows []row
	query = query.Group("TRIM(t.tag)").Order("total_expense DESC")
	if err := query.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to query tag stats: %w", err)
	}

	// Calculate grand total for percentages
	var grandTotal float64
	for _, r := range rows {
		grandTotal += r.TotalExpense + r.TotalIncome
	}

	items := make([]TagStatsItem, 0, len(rows))
	for _, r := range rows {
		pct := 0.0
		if grandTotal > 0 {
			pct = r.TotalExpense / grandTotal * 100
		}
		items = append(items, TagStatsItem{
			Tag:              r.Tag,
			TotalExpense:     r.TotalExpense,
			TotalIncome:      r.TotalIncome,
			TransactionCount: r.TransactionCount,
			Percentage:       pct,
		})
	}
	return items, nil
}

// DailyTransactions 获取指定年月的每日交易汇总。
func (s *LedgerService) DailyTransactions(ledgerID, userID uuid.UUID, year, month int) ([]DailyTransactionItem, error) {
	now := time.Now()
	if year == 0 {
		year = now.Year()
	}
	if month == 0 {
		month = int(now.Month())
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	endDate := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	type row struct {
		Date         string
		TotalIncome  float64
		TotalExpense float64
		TxnCount     int64
	}
	var rows []row
	s.DB.Raw(`
		SELECT
			transaction_date AS date,
			COALESCE(SUM(CASE WHEN type = 'income'  THEN base_amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN base_amount ELSE 0 END), 0) AS total_expense,
			COUNT(*) AS txn_count
		FROM transactions
		WHERE ledger_id = ? AND user_id = ? AND transaction_date >= ? AND transaction_date < ? AND deleted_at IS NULL
		GROUP BY transaction_date
		ORDER BY transaction_date ASC
	`, ledgerID, userID, startDate, endDate).Scan(&rows)

	items := make([]DailyTransactionItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, DailyTransactionItem{
			Date:    r.Date,
			Income:  r.TotalIncome,
			Expense: r.TotalExpense,
			Count:   r.TxnCount,
		})
	}
	return items, nil
}

// ---------- helpers ----------

// splitTags 将逗号分隔的标签字符串拆分为切片。
func splitTags(raw string) []string {
	var parts []string
	current := ""
	for _, ch := range raw {
		if ch == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// trim 去除字符串首尾的空格和制表符。
func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// CSVRow 将单条交易转换为 CSV 行（不含换行）。
func CSVRow(t models.Transaction) string {
	desc := ""
	if t.Description != nil {
		desc = *t.Description
	}
	return stringsJoin([]string{
		t.ID.String(),
		t.TransactionDate,
		t.Type,
		fmt.Sprintf("%.2f", t.Amount),
		t.Currency,
		fmt.Sprintf("%.2f", t.BaseAmount),
		desc,
		t.CategoryID.String(),
	}, ",")
}

// CSVHeader 返回 CSV 表头。
func CSVHeader() []string {
	return []string{"id", "date", "type", "amount", "currency", "base_amount", "description", "category_id"}
}

// FormatAmount 使用 strconv 格式化金额（保留两位小数）。
func FormatAmount(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// stringsJoin 使用指定分隔符合并字符串切片。
func stringsJoin(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += sep + parts[i]
	}
	return out
}
