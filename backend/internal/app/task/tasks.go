package task

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"
	"personal-bookkeeping/internal/infra/queue"

	"github.com/google/uuid"
)

// Task type constants
const (
	TypeExportReport       = "export_report"
	TypeImportTransactions = "import_transactions"
)

// ExportReportPayload 导出报表任务参数
type ExportReportPayload struct {
	UserID    string `json:"user_id"`
	LedgerID  string `json:"ledger_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Format    string `json:"format"` // csv | json
}

// ImportTransactionsPayload 导入交易记录任务参数
type ImportTransactionsPayload struct {
	UserID   string `json:"user_id"`
	LedgerID string `json:"ledger_id"`
	Format   string `json:"format"`      // csv | json
	Content  string `json:"content"`     // raw file content
}

// RegisterAll 注册所有异步任务处理器
func RegisterAll(q queue.Queue) {
	q.Register(TypeExportReport, handleExportReport)
	q.Register(TypeImportTransactions, handleImportTransactions)
	slog.Info("task handlers registered", "types", []string{TypeExportReport, TypeImportTransactions})
}

// handleExportReport 异步导出报表：查询交易记录 → 生成文件 → 落盘
func handleExportReport(ctx context.Context, task queue.Task) error {
	payload, err := decodePayload[ExportReportPayload](task.Payload)
	if err != nil {
		return fmt.Errorf("export: decode payload: %w", err)
	}

	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("export: database not available")
	}

	var transactions []models.Transaction
	query := db.Where("ledger_id = ? AND user_id = ?", payload.LedgerID, payload.UserID)
	if payload.StartDate != "" {
		query = query.Where("transaction_date >= ?", payload.StartDate)
	}
	if payload.EndDate != "" {
		query = query.Where("transaction_date <= ?", payload.EndDate)
	}
	if err := query.Order("transaction_date DESC").Find(&transactions).Error; err != nil {
		return fmt.Errorf("export: query transactions: %w", err)
	}

	switch payload.Format {
	case "csv":
		return writeCSV(transactions)
	case "json":
		return writeJSON(transactions)
	default:
		return fmt.Errorf("export: unsupported format: %s", payload.Format)
	}
}

// handleImportTransactions 异步导入：解析 CSV/JSON → 批量写入
func handleImportTransactions(ctx context.Context, task queue.Task) error {
	payload, err := decodePayload[ImportTransactionsPayload](task.Payload)
	if err != nil {
		return fmt.Errorf("import: decode payload: %w", err)
	}

	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("import: database not available")
	}

	var transactions []models.Transaction

	switch payload.Format {
	case "csv":
		transactions, err = parseCSV(payload.Content, payload.UserID, payload.LedgerID)
	case "json":
		transactions, err = parseJSON(payload.Content, payload.UserID, payload.LedgerID)
	default:
		return fmt.Errorf("import: unsupported format: %s", payload.Format)
	}
	if err != nil {
		return fmt.Errorf("import: parse: %w", err)
	}

	if len(transactions) == 0 {
		slog.Warn("import: no transactions to import")
		return nil
	}

	// Batch insert in chunks of 100
	batchSize := 100
	for i := 0; i < len(transactions); i += batchSize {
		end := i + batchSize
		if end > len(transactions) {
			end = len(transactions)
		}
		batch := transactions[i:end]
		if err := db.Create(&batch).Error; err != nil {
			return fmt.Errorf("import: batch insert [%d:%d]: %w", i, end, err)
		}
	}
	slog.Info("import completed", "count", len(transactions), "ledger_id", payload.LedgerID)
	return nil
}

// --- helpers ---

func decodePayload[T any](payload any) (T, error) {
	var zero T
	data, err := json.Marshal(payload)
	if err != nil {
		return zero, fmt.Errorf("marshal payload: %w", err)
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return zero, fmt.Errorf("unmarshal payload: %w", err)
	}
	return result, nil
}

func writeCSV(transactions []models.Transaction) error {
	out := &strings.Builder{}
	w := csv.NewWriter(out)
	w.Write([]string{"id", "date", "type", "amount", "currency", "base_amount", "description", "category_id"})
	for _, t := range transactions {
		w.Write([]string{
			t.ID.String(),
			t.TransactionDate,
			t.Type,
			fmt.Sprintf("%.2f", t.Amount),
			t.Currency,
			fmt.Sprintf("%.2f", t.BaseAmount),
			nullableStr(t.Description),
			t.CategoryID.String(),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("csv write: %w", err)
	}
	slog.Info("export csv done", "rows", len(transactions))
	return nil
}

func writeJSON(transactions []models.Transaction) error {
	data, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	slog.Info("export json done", "bytes", len(data), "rows", len(transactions))
	return nil
}

func parseCSV(content, userID, ledgerID string) ([]models.Transaction, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv read: %w", err)
	}
	if len(records) < 2 {
		return nil, nil // header only or empty
	}

	uid := uuid.MustParse(userID)
	lid := uuid.MustParse(ledgerID)
	now := time.Now()
	var txns []models.Transaction

	for _, row := range records[1:] {
		if len(row) < 5 {
			continue
		}
		amount, _ := strconv.ParseFloat(row[3], 64)
		desc := row[4]
		txn := models.Transaction{
			LedgerID:        lid,
			UserID:          uid,
			TransactionDate: row[0],
			Type:            row[1],
			Amount:          amount,
			BaseAmount:      amount,
			Currency:        "CNY",
			ExchangeRate:    1.0,
			Description:     &desc,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if len(row) > 5 && row[5] != "" {
			cid := uuid.MustParse(row[5])
			txn.CategoryID = cid
		}
		txns = append(txns, txn)
	}
	return txns, nil
}

func parseJSON(content, userID, ledgerID string) ([]models.Transaction, error) {
	var txns []models.Transaction
	if err := json.Unmarshal([]byte(content), &txns); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}
	uid := uuid.MustParse(userID)
	lid := uuid.MustParse(ledgerID)
	for i := range txns {
		txns[i].UserID = uid
		txns[i].LedgerID = lid
		if txns[i].Currency == "" {
			txns[i].Currency = "CNY"
		}
		if txns[i].ExchangeRate == 0 {
			txns[i].ExchangeRate = 1.0
		}
		if txns[i].BaseAmount == 0 {
			txns[i].BaseAmount = txns[i].Amount
		}
	}
	return txns, nil
}

func nullableStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
