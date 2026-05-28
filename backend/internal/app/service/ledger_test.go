package service

import (
	"testing"

	"personal-bookkeeping/internal/app/models"

	"github.com/google/uuid"
)

// ---------- splitTags ----------

func TestSplitTags_Empty(t *testing.T) {
	got := splitTags("")
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestSplitTags_Single(t *testing.T) {
	got := splitTags("hello")
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("expected ['hello'], got %v", got)
	}
}

func TestSplitTags_Multiple(t *testing.T) {
	got := splitTags("a,b,c")
	if len(got) != 3 {
		t.Fatalf("expected 3 parts, got %d: %v", len(got), got)
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("expected ['a','b','c'], got %v", got)
	}
}

func TestSplitTags_TrailingComma(t *testing.T) {
	got := splitTags("a,b,")
	if len(got) != 2 {
		t.Fatalf("expected 2 parts, got %d: %v", len(got), got)
	}
}

// ---------- trim ----------

func TestTrim(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"hello", "hello"},
		{"  hello", "hello"},
		{"hello\t", "hello"},
		{"\thello\t", "hello"},
		{"  hello world  ", "hello world"},
		{"   ", ""},
	}
	for _, tt := range tests {
		got := trim(tt.in)
		if got != tt.want {
			t.Errorf("trim(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// ---------- CSVRow ----------

func TestCSVRow(t *testing.T) {
	catID := uuid.New()
	txn := models.Transaction{
		ID:              uuid.New(),
		TransactionDate: "2026-06-01",
		Type:            "expense",
		Amount:          12.50,
		Currency:        "CNY",
		BaseAmount:      12.50,
		CategoryID:      catID,
	}

	row := CSVRow(txn)
	if row == "" {
		t.Fatal("CSVRow should not be empty")
	}
	if len(row) < 10 {
		t.Fatalf("CSVRow too short: %q", row)
	}
}

func TestCSVRow_WithDescription(t *testing.T) {
	desc := "coffee"
	txn := models.Transaction{
		ID:              uuid.New(),
		TransactionDate: "2026-06-01",
		Type:            "expense",
		Amount:          35.00,
		Currency:        "CNY",
		BaseAmount:      35.00,
		Description:     &desc,
		CategoryID:      uuid.New(),
	}

	row := CSVRow(txn)
	if row == "" {
		t.Fatal("CSVRow should not be empty")
	}
	// Description should be in the row
	if !contains(row, "coffee") {
		t.Fatalf("CSVRow should contain description 'coffee': %q", row)
	}
}

func TestCSVRow_NilDescription(t *testing.T) {
	txn := models.Transaction{
		ID:              uuid.New(),
		TransactionDate: "2026-06-01",
		Type:            "income",
		Amount:          10000.00,
		Currency:        "USD",
		BaseAmount:      72500.00,
		CategoryID:      uuid.New(),
	}

	row := CSVRow(txn)
	if row == "" {
		t.Fatal("CSVRow should not be empty")
	}
}

// ---------- CSVHeader ----------

func TestCSVHeader(t *testing.T) {
	headers := CSVHeader()
	if len(headers) != 8 {
		t.Fatalf("expected 8 headers, got %d: %v", len(headers), headers)
	}
	if headers[0] != "id" || headers[1] != "date" {
		t.Fatalf("expected first two headers 'id','date', got %v", headers[:2])
	}
}

// ---------- FormatAmount ----------

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, "0.00"},
		{12.5, "12.50"},
		{100, "100.00"},
		{0.01, "0.01"},
		{1234.56, "1234.56"},
		{-50.00, "-50.00"},
	}
	for _, tt := range tests {
		got := FormatAmount(tt.v)
		if got != tt.want {
			t.Errorf("FormatAmount(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

// ---------- stringsJoin ----------

func TestStringsJoin_Empty(t *testing.T) {
	got := stringsJoin(nil, ",")
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	got = stringsJoin([]string{}, ",")
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestStringsJoin_Single(t *testing.T) {
	got := stringsJoin([]string{"a"}, ",")
	if got != "a" {
		t.Fatalf("expected 'a', got %q", got)
	}
}

func TestStringsJoin_Multiple(t *testing.T) {
	got := stringsJoin([]string{"a", "b", "c"}, ",")
	if got != "a,b,c" {
		t.Fatalf("expected 'a,b,c', got %q", got)
	}
}

// Helper
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
