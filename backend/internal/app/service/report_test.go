package service

import (
	"testing"
)

// ---------------------------------------------------------------------------
// parsePeriod tests
// ---------------------------------------------------------------------------

func TestParsePeriod_Valid(t *testing.T) {
	year, month, err := parsePeriod("2026-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if year != 2026 {
		t.Errorf("expected year 2026, got %d", year)
	}
	if month != 6 {
		t.Errorf("expected month 6, got %d", month)
	}
}

func TestParsePeriod_FirstMonth(t *testing.T) {
	year, month, err := parsePeriod("2025-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if year != 2025 {
		t.Errorf("expected year 2025, got %d", year)
	}
	if month != 1 {
		t.Errorf("expected month 1, got %d", month)
	}
}

func TestParsePeriod_LastMonth(t *testing.T) {
	year, month, err := parsePeriod("2026-12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if year != 2026 {
		t.Errorf("expected year 2026, got %d", year)
	}
	if month != 12 {
		t.Errorf("expected month 12, got %d", month)
	}
}

func TestParsePeriod_InvalidMonth(t *testing.T) {
	_, _, err := parsePeriod("2026-13")
	if err == nil {
		t.Fatal("expected error for month 13")
	}
}

func TestParsePeriod_WrongFormat(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{"empty string", ""},
		{"too short", "2026-6"},
		{"too long", "2026-06-01"},
		{"no dash", "202606"},
		{"letters", "abc-def"},
		{"reversed", "06-2026"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parsePeriod(tt.s)
			if err == nil {
				t.Errorf("expected error for input %q", tt.s)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// calcChange tests
// ---------------------------------------------------------------------------

func TestCalcChange(t *testing.T) {
	tests := []struct {
		name     string
		current  float64
		previous float64
		want     string
	}{
		{
			name:     "increase 100%",
			current:  200,
			previous: 100,
			want:     "+100.0%",
		},
		{
			name:     "decrease 50%",
			current:  50,
			previous: 100,
			want:     "-50.0%",
		},
		{
			name:     "both zero",
			current:  0,
			previous: 0,
			want:     "持平",
		},
		{
			name:     "from zero to positive",
			current:  100,
			previous: 0,
			want:     "+∞",
		},
		{
			name:     "no change",
			current:  100,
			previous: 100,
			want:     "0.0%",
		},
		{
			name:     "decrease 100%",
			current:  0,
			previous: 100,
			want:     "-100.0%",
		},
		{
			name:     "small increase",
			current:  101.5,
			previous: 100,
			want:     "+1.5%",
		},
		{
			name:     "from zero to negative (edge case)",
			current:  -50,
			previous: 0,
			want:     "+∞",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcChange(tt.current, tt.previous)
			if got != tt.want {
				t.Errorf("calcChange(%v, %v) = %q, want %q", tt.current, tt.previous, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// truncate tests
// ---------------------------------------------------------------------------

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string unchanged",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "12345",
			maxLen: 5,
			want:   "12345",
		},
		{
			name:   "one char over",
			input:  "123456",
			maxLen: 5,
			want:   "1234…",
		},
		{
			name:   "long string truncated",
			input:  "this is a very long string",
			maxLen: 10,
			want:   "this is a…",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "unicode characters",
			input:  "星巴克咖啡店",
			maxLen: 4,
			want:   "星巴克…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BuildReportData (skipped - needs DB)
// ---------------------------------------------------------------------------

func TestBuildReportData(t *testing.T) {
	t.Skip("skipping BuildReportData test — requires real DB connection")
}

// ---------------------------------------------------------------------------
// GenerateReportPDF (skipped - needs fpdf)
// ---------------------------------------------------------------------------

func TestGenerateReportPDF(t *testing.T) {
	t.Skip("skipping GenerateReportPDF test — requires fpdf library")
}
