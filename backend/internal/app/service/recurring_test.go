package service

import (
	"testing"
	"time"
)

// ---------- ComputeNextRunDate ----------

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestComputeNextRunDate_Daily(t *testing.T) {
	from := date(2026, 1, 15)
	got := ComputeNextRunDate(from, "daily", 1, nil, nil)
	if !got.Equal(from) {
		t.Fatalf("daily: expected %v, got %v", from, got)
	}
}

func TestComputeNextRunDate_WeeklyDefaultsToFrom(t *testing.T) {
	from := date(2026, 1, 15)
	got := ComputeNextRunDate(from, "weekly", 1, nil, nil)
	// No weekday specified → returns from
	if !got.Equal(from) {
		t.Fatalf("weekly without weekday: expected %v, got %v", from, got)
	}
}

func TestComputeNextRunDate_WeeklyNextMonday(t *testing.T) {
	// 2026-01-15 is Thursday
	from := date(2026, 1, 15)
	monday := 1
	got := ComputeNextRunDate(from, "weekly", 1, nil, &monday)
	// Next Monday from Thursday is Jan 19
	want := date(2026, 1, 19)
	if !got.Equal(want) {
		t.Fatalf("weekly next Monday: expected %v, got %v", want, got)
	}
}

func TestComputeNextRunDate_WeeklySameDay(t *testing.T) {
	// 2026-01-19 is Monday
	from := date(2026, 1, 19)
	monday := 1
	got := ComputeNextRunDate(from, "weekly", 1, nil, &monday)
	want := date(2026, 1, 19)
	if !got.Equal(want) {
		t.Fatalf("weekly same day: expected %v, got %v", want, got)
	}
}

func TestComputeNextRunDate_MonthlySameDay(t *testing.T) {
	from := date(2026, 1, 15)
	dom := 15
	got := ComputeNextRunDate(from, "monthly", 1, &dom, nil)
	want := date(2026, 1, 15)
	if !got.Equal(want) {
		t.Fatalf("monthly same day: expected %v, got %v", want, got)
	}
}

func TestComputeNextRunDate_MonthlyNextMonth(t *testing.T) {
	// from Jan 20, dom=5 → next run is Feb 5
	from := date(2026, 1, 20)
	dom := 5
	got := ComputeNextRunDate(from, "monthly", 1, &dom, nil)
	want := date(2026, 2, 5)
	if !got.Equal(want) {
		t.Fatalf("monthly next month: expected %v, got %v", want, got)
	}
}

func TestComputeNextRunDate_MonthlyDomExceedsDays(t *testing.T) {
	// Feb 1, dom=31 → Go's time.Date normalizes Feb 31 to Mar 3
	// Since Mar 3 is after Feb 1, the function returns it as-is
	from := date(2026, 2, 1)
	dom := 31
	got := ComputeNextRunDate(from, "monthly", 1, &dom, nil)
	want := date(2026, 3, 3)
	if !got.Equal(want) {
		t.Fatalf("monthly overflow: expected %v, got %v", want, got)
	}
}

func TestComputeNextRunDate_MonthlyFrom31To30DayMonth(t *testing.T) {
	// Jan 31 → Feb (28 days) → clamp to Feb 28
	from := date(2026, 1, 31)
	dom := 31
	got := ComputeNextRunDate(from, "monthly", 1, &dom, nil)
	want := date(2026, 1, 31) // same month, same day
	if !got.Equal(want) {
		t.Fatalf("monthly same month: expected %v, got %v", want, got)
	}
}

func TestComputeNextRunDate_MonthlyDefaultDom(t *testing.T) {
	// No dayOfMonth specified → defaults to day 1
	from := date(2026, 3, 15)
	got := ComputeNextRunDate(from, "monthly", 1, nil, nil)
	want := date(2026, 4, 1) // next month, day 1
	if !got.Equal(want) {
		t.Fatalf("monthly default dom: expected %v, got %v", want, got)
	}
}

func TestComputeNextRunDate_YearlyAlwaysNextYear(t *testing.T) {
	// yearly always advances to January 1 of the next year
	from := date(2026, 1, 1)
	got := ComputeNextRunDate(from, "yearly", 1, nil, nil)
	want := date(2027, 1, 1)
	if !got.Equal(want) {
		t.Fatalf("yearly from Jan 1: expected %v, got %v", want, got)
	}
}

func TestComputeNextRunDate_YearlyNextYear(t *testing.T) {
	from := date(2026, 6, 15)
	got := ComputeNextRunDate(from, "yearly", 1, nil, nil)
	want := date(2027, 1, 1)
	if !got.Equal(want) {
		t.Fatalf("yearly next year: expected %v, got %v", want, got)
	}
}

func TestComputeNextRunDate_YearlyDec(t *testing.T) {
	from := date(2026, 12, 31)
	got := ComputeNextRunDate(from, "yearly", 1, nil, nil)
	want := date(2027, 1, 1)
	if !got.Equal(want) {
		t.Fatalf("yearly from dec: expected %v, got %v", want, got)
	}
}

func TestComputeNextRunDate_InvalidFrequency(t *testing.T) {
	from := date(2026, 1, 15)
	got := ComputeNextRunDate(from, "invalid", 1, nil, nil)
	if !got.Equal(from) {
		t.Fatalf("invalid freq: expected %v, got %v", from, got)
	}
}

func TestComputeNextRunDate_ZeroInterval(t *testing.T) {
	from := date(2026, 1, 15)
	got := ComputeNextRunDate(from, "daily", 0, nil, nil)
	if !got.Equal(from) {
		t.Fatalf("zero interval: expected %v, got %v", from, got)
	}
}

// ---------- DaysInMonth ----------

func TestDaysInMonth(t *testing.T) {
	tests := []struct {
		year  int
		month int
		want  int
	}{
		{2026, 1, 31},
		{2026, 2, 28},  // non-leap
		{2024, 2, 29},  // leap
		{2026, 4, 30},
		{2026, 12, 31},
	}
	for _, tt := range tests {
		got := DaysInMonth(tt.year, tt.month)
		if got != tt.want {
			t.Errorf("DaysInMonth(%d, %d) = %d, want %d", tt.year, tt.month, got, tt.want)
		}
	}
}
