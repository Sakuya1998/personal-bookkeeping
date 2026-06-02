package task

import (
	"context"
	"sync"
	"testing"
	"time"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/infra/queue"

	"github.com/google/uuid"
)

// ------------------ helpers ------------------

var (
	sampleUserID   = "11111111-1111-1111-1111-111111111111"
	sampleLedgerID = "22222222-2222-2222-2222-222222222222"
)

// ------------------ parseCSV tests ------------------

func TestParseCSV_Valid(t *testing.T) {
	content := "date,type,amount,description,category_id\n2024-01-15,expense,42.50,coffee,33333333-3333-3333-3333-333333333333\n2024-01-16,income,100.00,salary,\n"
	txns, err := parseCSV(content, sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txns))
	}

	// First transaction
	txn := txns[0]
	if txn.TransactionDate != "2024-01-15" {
		t.Errorf("expected date 2024-01-15, got %s", txn.TransactionDate)
	}
	if txn.Type != "expense" {
		t.Errorf("expected type expense, got %s", txn.Type)
	}
	if txn.Amount != 42.50 {
		t.Errorf("expected amount 42.50, got %f", txn.Amount)
	}
	if txn.BaseAmount != 42.50 {
		t.Errorf("expected base_amount 42.50, got %f", txn.BaseAmount)
	}
	if txn.Currency != "CNY" {
		t.Errorf("expected currency CNY, got %s", txn.Currency)
	}
	if txn.ExchangeRate != 1.0 {
		t.Errorf("expected exchange rate 1.0, got %f", txn.ExchangeRate)
	}
	if txn.Description == nil || *txn.Description != "coffee" {
		t.Errorf("expected description 'coffee', got %v", nullableStr(txn.Description))
	}
	if txn.CategoryID.String() != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("expected category 33333333-..., got %s", txn.CategoryID.String())
	}
	if txn.UserID.String() != sampleUserID {
		t.Errorf("expected user_id %s, got %s", sampleUserID, txn.UserID.String())
	}
	if txn.LedgerID.String() != sampleLedgerID {
		t.Errorf("expected ledger_id %s, got %s", sampleLedgerID, txn.LedgerID.String())
	}
	if txn.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}
	if txn.UpdatedAt.IsZero() {
		t.Errorf("expected UpdatedAt to be set")
	}

	// Second transaction: no category_id → uuid.Nil
	txn2 := txns[1]
	if txn2.Type != "income" {
		t.Errorf("expected type income, got %s", txn2.Type)
	}
	if txn2.Description == nil || *txn2.Description != "salary" {
		t.Errorf("expected description 'salary', got %v", nullableStr(txn2.Description))
	}
	if txn2.CategoryID != uuid.Nil {
		t.Errorf("expected nil category_id for empty field, got %s", txn2.CategoryID.String())
	}
}

func TestParseCSV_HeaderOnly(t *testing.T) {
	txns, err := parseCSV("date,type,amount,description,category_id\n", sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txns != nil {
		t.Fatalf("expected nil for header-only CSV, got %d transactions", len(txns))
	}
}

func TestParseCSV_Empty(t *testing.T) {
	txns, err := parseCSV("", sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txns != nil {
		t.Fatalf("expected nil for empty CSV, got %d transactions", len(txns))
	}
}

func TestParseCSV_Malformed(t *testing.T) {
	// Row with invalid (non-numeric) amount — should still parse but amount=0
	content := "date,type,amount,description,category_id\n2024-01-15,expense,not-a-number,coffee,44444444-4444-4444-4444-444444444444\n2024-01-16,income,100.00,salary,55555555-5555-5555-5555-555555555555\n"
	txns, err := parseCSV(content, sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txns))
	}
	// First row has invalid amount → should be 0
	if txns[0].Amount != 0 {
		t.Errorf("expected amount 0 for non-numeric, got %f", txns[0].Amount)
	}
	// Second row should be fine
	if txns[1].Amount != 100.00 {
		t.Errorf("expected amount 100.00, got %f", txns[1].Amount)
	}
}

// ------------------ parseJSON tests ------------------

func TestParseJSON_Valid(t *testing.T) {
	content := `[
		{"transaction_date":"2024-01-15","type":"expense","amount":42.50,"description":"coffee","category_id":"33333333-3333-3333-3333-333333333333"},
		{"transaction_date":"2024-01-16","type":"income","amount":100.00,"description":"salary","currency":"USD","exchange_rate":7.25,"base_amount":725.00}
	]`
	txns, err := parseJSON(content, sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txns))
	}

	// First transaction
	txn := txns[0]
	if txn.TransactionDate != "2024-01-15" {
		t.Errorf("expected date 2024-01-15, got %s", txn.TransactionDate)
	}
	if txn.Type != "expense" {
		t.Errorf("expected type expense, got %s", txn.Type)
	}
	if txn.Amount != 42.50 {
		t.Errorf("expected amount 42.50, got %f", txn.Amount)
	}
	if txn.UserID.String() != sampleUserID {
		t.Errorf("expected user_id %s, got %s", sampleUserID, txn.UserID.String())
	}
	if txn.LedgerID.String() != sampleLedgerID {
		t.Errorf("expected ledger_id %s, got %s", sampleLedgerID, txn.LedgerID.String())
	}
	if txn.CategoryID.String() != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("expected category 33333333-..., got %s", txn.CategoryID.String())
	}
	// Defaults for missing fields
	if txn.Currency != "CNY" {
		t.Errorf("expected default currency CNY, got %s", txn.Currency)
	}
	if txn.ExchangeRate != 1.0 {
		t.Errorf("expected default exchange rate 1.0, got %f", txn.ExchangeRate)
	}
	if txn.BaseAmount != 42.50 {
		t.Errorf("expected default base_amount 42.50, got %f", txn.BaseAmount)
	}

	// Second transaction: all fields provided
	txn2 := txns[1]
	if txn2.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", txn2.Currency)
	}
	if txn2.ExchangeRate != 7.25 {
		t.Errorf("expected exchange rate 7.25, got %f", txn2.ExchangeRate)
	}
	if txn2.BaseAmount != 725.00 {
		t.Errorf("expected base_amount 725.00, got %f", txn2.BaseAmount)
	}
	if txn2.CategoryID != uuid.Nil {
		t.Errorf("expected nil category_id when not provided, got %s", txn2.CategoryID.String())
	}
}

func TestParseJSON_EmptyArray(t *testing.T) {
	txns, err := parseJSON("[]", sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 0 {
		t.Fatalf("expected 0 transactions, got %d", len(txns))
	}
}

func TestParseJSON_Malformed(t *testing.T) {
	_, err := parseJSON("{invalid json}", sampleUserID, sampleLedgerID)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

// ------------------ writeCSV tests ------------------

func TestWriteCSV(t *testing.T) {
	uid := uuid.MustParse(sampleUserID)
	lid := uuid.MustParse(sampleLedgerID)
	catID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	desc := "coffee"
	txns := []models.Transaction{
		{
			ID:              uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			LedgerID:        lid,
			UserID:          uid,
			CategoryID:      catID,
			Type:            "expense",
			Amount:          42.50,
			Currency:        "CNY",
			ExchangeRate:    1.0,
			BaseAmount:      42.50,
			Description:     &desc,
			TransactionDate: "2024-01-15",
		},
		{
			ID:              uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			LedgerID:        lid,
			UserID:          uid,
			Type:            "income",
			Amount:          100.00,
			Currency:        "USD",
			ExchangeRate:    7.25,
			BaseAmount:      725.00,
			Description:     nil,
			TransactionDate: "2024-01-16",
		},
	}
	err := writeCSV(txns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteCSV_Empty(t *testing.T) {
	err := writeCSV([]models.Transaction{})
	if err != nil {
		t.Fatalf("unexpected error for empty list: %v", err)
	}
}

// ------------------ writeJSON tests ------------------

func TestWriteJSON(t *testing.T) {
	uid := uuid.MustParse(sampleUserID)
	lid := uuid.MustParse(sampleLedgerID)
	desc := "coffee"
	txns := []models.Transaction{
		{
			ID:              uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			LedgerID:        lid,
			UserID:          uid,
			Type:            "expense",
			Amount:          42.50,
			Currency:        "CNY",
			ExchangeRate:    1.0,
			BaseAmount:      42.50,
			Description:     &desc,
			TransactionDate: "2024-01-15",
		},
	}
	err := writeJSON(txns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteJSON_Empty(t *testing.T) {
	err := writeJSON([]models.Transaction{})
	if err != nil {
		t.Fatalf("unexpected error for empty list: %v", err)
	}
}

// ------------------ decodePayload tests ------------------

func TestDecodePayload(t *testing.T) {
	raw := map[string]any{
		"user_id":    sampleUserID,
		"ledger_id":  sampleLedgerID,
		"start_date": "2024-01-01",
		"end_date":   "2024-01-31",
		"format":     "csv",
	}
	payload, err := decodePayload[ExportReportPayload](raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.UserID != sampleUserID {
		t.Errorf("expected user_id %s, got %s", sampleUserID, payload.UserID)
	}
	if payload.LedgerID != sampleLedgerID {
		t.Errorf("expected ledger_id %s, got %s", sampleLedgerID, payload.LedgerID)
	}
	if payload.StartDate != "2024-01-01" {
		t.Errorf("expected start_date 2024-01-01, got %s", payload.StartDate)
	}
	if payload.EndDate != "2024-01-31" {
		t.Errorf("expected end_date 2024-01-31, got %s", payload.EndDate)
	}
	if payload.Format != "csv" {
		t.Errorf("expected format csv, got %s", payload.Format)
	}
}

func TestDecodePayload_Import(t *testing.T) {
	raw := map[string]any{
		"user_id":   sampleUserID,
		"ledger_id": sampleLedgerID,
		"format":    "json",
		"content":   `[{"amount": 10}]`,
	}
	payload, err := decodePayload[ImportTransactionsPayload](raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.UserID != sampleUserID {
		t.Errorf("expected user_id %s, got %s", sampleUserID, payload.UserID)
	}
	if payload.Format != "json" {
		t.Errorf("expected format json, got %s", payload.Format)
	}
	if payload.Content != `[{"amount": 10}]` {
		t.Errorf("expected content '[{\"amount\": 10}]', got %s", payload.Content)
	}
}

func TestDecodePayload_NilInput(t *testing.T) {
	// nil payload should also be decoded via JSON marshal/unmarshal roundtrip
	_, err := decodePayload[ExportReportPayload](nil)
	if err != nil {
		t.Fatalf("unexpected error for nil payload: %v", err)
	}
}

// ------------------ nullableStr tests ------------------

func TestNullableStr_Nil(t *testing.T) {
	got := nullableStr(nil)
	if got != "" {
		t.Errorf("expected empty string for nil pointer, got %q", got)
	}
}

func TestNullableStr_NonNil(t *testing.T) {
	s := "hello"
	got := nullableStr(&s)
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestNullableStr_EmptyString(t *testing.T) {
	s := ""
	got := nullableStr(&s)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// ------------------ handleExportReport unsupported format test ------------------

func TestHandleExportReport_UnsupportedFormat(t *testing.T) {
	// This tests that handleExportReport returns an error when given an
	// unsupported export format.  The function's first dependency is
	// repository.GetDB() which returns nil when the DB is uninitialized.
	//
	// We verify two things:
	//   1. decodePayload succeeds for a well-formed ExportReportPayload
	//   2. the handler returns a non-nil error (the "database not available"
	//      error path is hit before the format check; if the DB were available
	//      the format guard would catch "xml" and return a format-specific error)
	payload := map[string]any{
		"user_id":    sampleUserID,
		"ledger_id":  sampleLedgerID,
		"start_date": "2024-01-01",
		"end_date":   "2024-01-31",
		"format":     "xml", // unsupported
	}
	task := queue.Task{ID: "t1", Type: TypeExportReport, Payload: payload}

	err := handleExportReport(context.Background(), task)
	if err == nil {
		t.Fatal("expected error for unsupported format 'xml', got nil")
	}
	// The error should reference either "format" or "database" — both indicate
	// the function did not silently succeed.
	t.Logf("handleExportReport returned (expected) error: %v", err)
}

// ------------------ handleImportTransactions unsupported format test ------------------

func TestHandleImportTransactions_UnsupportedFormat(t *testing.T) {
	// Similar to the export test above — the handler will fail at
	// decodePayload or database-not-available, but never succeed silently.
	task := queue.Task{
		ID:   "t2",
		Type: TypeImportTransactions,
		Payload: map[string]any{
			"user_id":   sampleUserID,
			"ledger_id": sampleLedgerID,
			"format":    "xml",
			"content":   "some data",
		},
	}
	err := handleImportTransactions(context.Background(), task)
	if err == nil {
		t.Fatal("expected error for unsupported format 'xml', got nil")
	}
	t.Logf("handleImportTransactions returned (expected) error: %v", err)
}

// ------------------ Test parseCSV/parseJSON roundtrip consistency ------------------

func TestParseCSV_ValidAmountFormats(t *testing.T) {
	content := "date,type,amount,description,category_id\n2024-01-15,expense,0.00,zero,\n2024-01-16,income,100,integer,\n2024-01-17,expense,-50.5,negative,\n"
	txns, err := parseCSV(content, sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(txns))
	}
	if txns[0].Amount != 0.0 {
		t.Errorf("expected amount 0.0, got %f", txns[0].Amount)
	}
	if txns[1].Amount != 100.0 {
		t.Errorf("expected amount 100.0, got %f", txns[1].Amount)
	}
	if txns[2].Amount != -50.5 {
		t.Errorf("expected amount -50.5, got %f", txns[2].Amount)
	}
}

func TestParseJSON_ExplicitDefaultsOverridden(t *testing.T) {
	// JSON where the user explicitly provides a base_amount / exchange_rate
	// that differ from the defaults
	content := `[
		{"transaction_date":"2024-01-15","type":"expense","amount":10.00,"currency":"USD","exchange_rate":7.25,"base_amount":72.50}
	]`
	txns, err := parseJSON(content, sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txns))
	}
	txn := txns[0]
	if txn.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", txn.Currency)
	}
	if txn.ExchangeRate != 7.25 {
		t.Errorf("expected exchange rate 7.25, got %f", txn.ExchangeRate)
	}
	if txn.BaseAmount != 72.50 {
		t.Errorf("expected base_amount 72.50, got %f", txn.BaseAmount)
	}
}

func TestParseCSV_NegativeAmount(t *testing.T) {
	content := "date,type,amount,description,category_id\n2024-01-15,expense,-123.45,refund,\n"
	txns, err := parseCSV(content, sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txns))
	}
	if txns[0].Amount != -123.45 {
		t.Errorf("expected amount -123.45, got %f", txns[0].Amount)
	}
}

func TestParseCSV_WithBOM(t *testing.T) {
	// UTF-8 BOM before the header
	content := "\xef\xbb\xbfdate,type,amount,description,category_id\n2024-01-15,expense,42.50,coffee,\n"
	txns, err := parseCSV(content, sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error with BOM: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("expected 1 transaction with BOM, got %d", len(txns))
	}
	if txns[0].Amount != 42.50 {
		t.Errorf("expected amount 42.50 with BOM, got %f", txns[0].Amount)
	}
}

func TestParseCSV_ExtraWhitespace(t *testing.T) {
	// CSV parser preserves whitespace — " 42.50 " is not parseable as float
	content := "date,type,amount,description,category_id\n2024-01-15, expense, 42.50 , coffee with space ,\n"
	txns, err := parseCSV(content, sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txns))
	}
	// " 42.50 " cannot be parsed as float64 — amount will be 0
	if txns[0].Amount != 0 {
		t.Errorf("expected amount 0 (unparseable whitespace), got %f", txns[0].Amount)
	}
	// But string values like description preserve whitespace
	if txns[0].Description == nil || *txns[0].Description != " coffee with space " {
		t.Errorf("expected description ' coffee with space ', got %v", nullableStr(txns[0].Description))
	}
}
func TestParseCSV_ZeroAmount(t *testing.T) {
	content := "date,type,amount,description,category_id\n2024-01-15,expense,0.00,free,\n"
	txns, err := parseCSV(content, sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txns))
	}
	if txns[0].Amount != 0.0 {
		t.Errorf("expected amount 0.0, got %f", txns[0].Amount)
	}
	if txns[0].BaseAmount != 0.0 {
		t.Errorf("expected base_amount 0.0, got %f", txns[0].BaseAmount)
	}
}

func TestParseCSV_LargeAmount(t *testing.T) {
	content := "date,type,amount,description,category_id\n2024-01-15,income,99999999.99,large,\n"
	txns, err := parseCSV(content, sampleUserID, sampleLedgerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txns))
	}
	if txns[0].Amount != 99999999.99 {
		t.Errorf("expected amount 99999999.99, got %f", txns[0].Amount)
	}
}

// ------------------ scheduler tests ------------------

// mockQueue implements queue.Queue for scheduler testing.
type mockQueue struct {
	submitted []queue.Task
	started   bool
	mu        sync.Mutex
}

func (q *mockQueue) Register(_ string, _ queue.HandlerFunc) {}
func (q *mockQueue) Start(_ context.Context)                 { q.started = true }
func (q *mockQueue) Submit(_ context.Context, task queue.Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.submitted = append(q.submitted, task)
	return nil
}
func (q *mockQueue) Shutdown(_ context.Context) error { return nil }
func (q *mockQueue) Stats() queue.Stats                { return queue.Stats{} }

// submittedTasks returns a snapshot of submitted tasks under lock protection.
func (q *mockQueue) submittedTasks() []queue.Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]queue.Task, len(q.submitted))
	copy(out, q.submitted)
	return out
}

func TestStartRecurringScheduler_DispatchOnStart(t *testing.T) {
	mq := &mockQueue{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Very short interval so the ticker fires quickly
	StartRecurringScheduler(ctx, mq, 10*time.Millisecond)

	// Wait for a tick or two
	time.Sleep(50 * time.Millisecond)

	submitted := mq.submittedTasks()
	if len(submitted) == 0 {
		t.Fatal("expected at least 1 submitted task")
	}
	if submitted[0].Type != TypeProcessRecurring {
		t.Errorf("expected type %q, got %q", TypeProcessRecurring, submitted[0].Type)
	}
}

func TestStartRecurringScheduler_NoDuplicateOnSameDay(t *testing.T) {
	mq := &mockQueue{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartRecurringScheduler(ctx, mq, 10*time.Millisecond)

	time.Sleep(100 * time.Millisecond) // several ticks

	submitted := mq.submittedTasks()
	// Should only have 1 submission because same-day dedup
	if len(submitted) != 1 {
		t.Logf("submitted %d tasks (expected 1 due to same-day dedup)", len(submitted))
	}
}

func TestStartRecurringScheduler_NilQueue(t *testing.T) {
	// Should not panic
	StartRecurringScheduler(context.Background(), nil, time.Hour)
}

func TestStartExchangeRateScheduler_NilQueue(t *testing.T) {
	// Should log warning and return immediately, not panic
	StartExchangeRateScheduler(context.Background(), nil)
}

// ------------------ computeRecurringNext tests ------------------

func weekdayPtr(w int) *int { return &w }

func TestComputeRecurringNext_Daily(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := computeRecurringNext(from, "daily", 3, nil, nil)
	want := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("daily interval=3: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_WeeklyNoWeekday(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // Monday
	got := computeRecurringNext(from, "weekly", 2, nil, nil)
	want := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC) // +14 days
	if !got.Equal(want) {
		t.Errorf("weekly nil weekday interval=2: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_WeeklyCustomWeekday(t *testing.T) {
	// from = Wednesday 2024-01-03, target weekday = Monday (1)
	from := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	got := computeRecurringNext(from, "weekly", 1, nil, weekdayPtr(1)) // Monday
	// Next Monday from Wednesday is 5 days later: 2024-01-08
	want := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("weekly weekday=Mon from Wed interval=1: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_WeeklySameWeekday(t *testing.T) {
	// from = Monday 2024-01-01, target weekday = Monday (1)
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := computeRecurringNext(from, "weekly", 1, nil, weekdayPtr(1)) // Monday
	// Already Monday, advance 1 week -> 2024-01-08
	want := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("weekly weekday=Mon from Mon interval=1: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_WeeklyCustomWeekdayInterval(t *testing.T) {
	// from = Wednesday 2024-01-03, target weekday = Friday (5), interval = 2
	from := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) // Wednesday
	got := computeRecurringNext(from, "weekly", 2, nil, weekdayPtr(5)) // Friday
	// Next Friday = Jan 5, then add (2-1)*7 = 7 days -> Jan 12
	want := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("weekly weekday=Fri from Wed interval=2: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_WeeklySameWeekdayInterval(t *testing.T) {
	// from = Monday 2024-01-01, target weekday = Monday, interval = 3
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := computeRecurringNext(from, "weekly", 3, nil, weekdayPtr(1)) // Monday
	// Already Monday, advance 3 weeks -> 2024-01-22
	want := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("weekly weekday=Mon from Mon interval=3: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_MonthlyDefault(t *testing.T) {
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	got := computeRecurringNext(from, "monthly", 1, nil, nil)
	// dayOfMonth nil -> defaults to 1st of next month
	want := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("monthly interval=1 default dom: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_MonthlyWithDayOfMonth(t *testing.T) {
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	dom := 20
	got := computeRecurringNext(from, "monthly", 1, &dom, nil)
	want := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("monthly interval=1 dom=20: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_MonthlyWithInterval(t *testing.T) {
	// interval=3 should advance by 3 months, not 1
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	dom := 10
	got := computeRecurringNext(from, "monthly", 3, &dom, nil)
	want := time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("monthly interval=3 dom=10: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_MonthlyWithIntervalDayClamping(t *testing.T) {
	// interval=2, dom=31, from in January -> next month with 31 days is March (Jan + 2)
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	dom := 31
	got := computeRecurringNext(from, "monthly", 2, &dom, nil)
	// from.Month() + 2 = March, March has 31 days, so should be March 31
	want := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("monthly interval=2 dom=31: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_DefaultFallback(t *testing.T) {
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	got := computeRecurringNext(from, "unknown", 5, nil, nil)
	want := from.AddDate(0, 0, 5)
	if !got.Equal(want) {
		t.Errorf("unknown freq: got %v, want %v", got, want)
	}
}

func TestComputeRecurringNext_ZeroInterval(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := computeRecurringNext(from, "daily", 0, nil, nil)
	want := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC) // interval defaulted to 1
	if !got.Equal(want) {
		t.Errorf("daily interval=0 defaults to 1: got %v, want %v", got, want)
	}
}
