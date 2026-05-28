package service

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReportPeriod 报表周期
type ReportPeriod string

const (
	ReportMonthly  ReportPeriod = "monthly"
	ReportQuarterly ReportPeriod = "quarterly"
)

// ReportData 报表数据
type ReportData struct {
	LedgerName   string
	Period       string // e.g. "2026-05"
	PeriodLabel  string // e.g. "2026年5月"
	IncomeTotal  float64
	ExpenseTotal float64
	Balance      float64
	PrevIncome   float64 // previous period for comparison
	PrevExpense  float64
	Categories   []CategorySummary
	DailyAvg     float64
	DaysInPeriod int
}

// CategorySummary 分类汇总
type CategorySummary struct {
	Name       string
	Icon       string
	Type       string
	Total      float64
	Percentage float64
}

// BuildReportData 从 DB 查询并组装报表数据
func BuildReportData(db *gorm.DB, ledgerID, userID uuid.UUID, period ReportPeriod, periodStr string) (*ReportData, error) {
	// Parse period
	year, month, err := parsePeriod(periodStr)
	if err != nil {
		return nil, fmt.Errorf("invalid period %q: %w", periodStr, err)
	}

	var startDate, endDate string
	var daysInPeriod int

	switch period {
	case ReportMonthly:
		startDate = fmt.Sprintf("%04d-%02d-01", year, month)
		nextMonth := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
		endDate = nextMonth.Format("2006-01-02")
		daysInPeriod = nextMonth.AddDate(0, 0, -1).Day()
	case ReportQuarterly:
		quarterStart := time.Date(year, time.Month((month-1)/3*3+1), 1, 0, 0, 0, 0, time.UTC)
		startDate = quarterStart.Format("2006-01-02")
		quarterEnd := quarterStart.AddDate(0, 3, 0)
		endDate = quarterEnd.Format("2006-01-02")
		daysInPeriod = int(quarterEnd.Sub(quarterStart).Hours() / 24)
	}

	// Previous period for comparison
	var prevStart, prevEnd string
	switch period {
	case ReportMonthly:
		prevStart = time.Date(year, time.Month(month-1), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		prevEnd = startDate
	case ReportQuarterly:
		prevStart = time.Date(year, time.Month((month-1)/3*3+1-3), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		prevEnd = startDate
	}

	// Get ledger name (with ownership check)
	var ledgerName string
	result := db.Raw("SELECT name FROM ledgers WHERE id = ? AND user_id = ?", ledgerID, userID).Scan(&ledgerName)
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("ledger not found or access denied")
	}

	// This period totals
	var incomeTotal, expenseTotal float64
	db.Raw(`SELECT COALESCE(SUM(base_amount),0) FROM transactions
		WHERE ledger_id=? AND user_id=? AND type='income' AND transaction_date>=? AND transaction_date<?`,
		ledgerID, userID, startDate, endDate).Scan(&incomeTotal)
	db.Raw(`SELECT COALESCE(SUM(base_amount),0) FROM transactions
		WHERE ledger_id=? AND user_id=? AND type='expense' AND transaction_date>=? AND transaction_date<?`,
		ledgerID, userID, startDate, endDate).Scan(&expenseTotal)

	// Previous period totals for comparison
	var prevIncome, prevExpense float64
	db.Raw(`SELECT COALESCE(SUM(base_amount),0) FROM transactions
		WHERE ledger_id=? AND user_id=? AND type='income' AND transaction_date>=? AND transaction_date<?`,
		ledgerID, userID, prevStart, prevEnd).Scan(&prevIncome)
	db.Raw(`SELECT COALESCE(SUM(base_amount),0) FROM transactions
		WHERE ledger_id=? AND user_id=? AND type='expense' AND transaction_date>=? AND transaction_date<?`,
		ledgerID, userID, prevStart, prevEnd).Scan(&prevExpense)

	// Category breakdown
	type catRow struct {
		Name       string
		Icon       string
		Type       string
		Total      float64
	}
	var catRows []catRow
	db.Raw(`SELECT COALESCE(c.name,'') as name, COALESCE(c.icon,'') as icon, t.type,
		COALESCE(SUM(t.base_amount),0) as total
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.ledger_id=? AND t.user_id=? AND t.transaction_date>=? AND t.transaction_date<?
		GROUP BY t.category_id, c.name, c.icon, t.type
		ORDER BY total DESC`, ledgerID, userID, startDate, endDate).Scan(&catRows)

	var categories []CategorySummary
	for _, r := range catRows {
		pct := 0.0
		if r.Type == "expense" && expenseTotal > 0 {
			pct = r.Total / expenseTotal * 100
		} else if r.Type == "income" && incomeTotal > 0 {
			pct = r.Total / incomeTotal * 100
		}
		categories = append(categories, CategorySummary{
			Name:       r.Name,
			Icon:       r.Icon,
			Type:       r.Type,
			Total:      math.Round(r.Total*100) / 100,
			Percentage: math.Round(pct*10) / 10,
		})
	}

	// Period label
	var periodLabel string
	switch period {
	case ReportMonthly:
		periodLabel = fmt.Sprintf("%d年%d月", year, month)
	case ReportQuarterly:
		q := (month-1)/3 + 1
		periodLabel = fmt.Sprintf("%d年第%d季度", year, q)
	}

	dailyAvg := 0.0
	if daysInPeriod > 0 {
		dailyAvg = math.Round(expenseTotal/float64(daysInPeriod)*100) / 100
	}

	return &ReportData{
		LedgerName:   ledgerName,
		Period:       periodStr,
		PeriodLabel:  periodLabel,
		IncomeTotal:  math.Round(incomeTotal*100) / 100,
		ExpenseTotal: math.Round(expenseTotal*100) / 100,
		Balance:      math.Round((incomeTotal-expenseTotal)*100) / 100,
		PrevIncome:   math.Round(prevIncome*100) / 100,
		PrevExpense:  math.Round(prevExpense*100) / 100,
		Categories:   categories,
		DailyAvg:     dailyAvg,
		DaysInPeriod: daysInPeriod,
	}, nil
}

// GenerateReportPDF 生成 PDF 报表
func GenerateReportPDF(data *ReportData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()
	pageW, _ := pdf.GetPageSize()

	// Title
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetTextColor(22, 119, 255)
	pdf.CellFormat(pageW, 15, data.LedgerName, "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 14)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(pageW, 10, fmt.Sprintf("%s 财务报告", data.PeriodLabel), "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Summary section
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(pageW, 8, "收支概览", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Summary table
	colW := pageW / 4
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(240, 245, 255)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(colW, 7, "项目", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW, 7, "收入", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW, 7, "支出", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW, 7, "结余", "1", 1, "C", true, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(colW, 7, "本期", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colW, 7, fmt.Sprintf("¥%.2f", data.IncomeTotal), "1", 0, "C", false, 0, "")
	pdf.CellFormat(colW, 7, fmt.Sprintf("¥%.2f", data.ExpenseTotal), "1", 0, "C", false, 0, "")
	balanceColor := []int{82, 196, 26} // green
	if data.Balance < 0 {
		balanceColor = []int{255, 77, 79} // red
	}
	pdf.SetTextColor(balanceColor[0], balanceColor[1], balanceColor[2])
	pdf.CellFormat(colW, 7, fmt.Sprintf("¥%.2f", data.Balance), "1", 1, "C", false, 0, "")

	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(colW, 7, "上期", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colW, 7, fmt.Sprintf("¥%.2f", data.PrevIncome), "1", 0, "C", false, 0, "")
	pdf.CellFormat(colW, 7, fmt.Sprintf("¥%.2f", data.PrevExpense), "1", 1, "C", false, 0, "")
	pdf.Ln(2)

	// Change indicators
	incomeChange := calcChange(data.IncomeTotal, data.PrevIncome)
	expenseChange := calcChange(data.ExpenseTotal, data.PrevExpense)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(pageW, 6, fmt.Sprintf("收入环比: %s  |  支出环比: %s  |  日均支出: ¥%.2f",
		incomeChange, expenseChange, data.DailyAvg), "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Category breakdown
	if len(data.Categories) > 0 {
		pdf.SetFont("Helvetica", "B", 13)
		pdf.SetTextColor(50, 50, 50)
		pdf.CellFormat(pageW, 8, "分类明细", "", 1, "L", false, 0, "")
		pdf.Ln(2)

		catColW := pageW * 0.35
		typeColW := pageW * 0.15
		amtColW := pageW * 0.25
		pctColW := pageW * 0.25

		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(240, 245, 255)
		pdf.SetTextColor(50, 50, 50)
		pdf.CellFormat(catColW, 6, "分类", "1", 0, "C", true, 0, "")
		pdf.CellFormat(typeColW, 6, "类型", "1", 0, "C", true, 0, "")
		pdf.CellFormat(amtColW, 6, "金额", "1", 0, "C", true, 0, "")
		pdf.CellFormat(pctColW, 6, "占比", "1", 1, "C", true, 0, "")

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetFillColor(255, 255, 255)
		for _, c := range data.Categories {
			pdf.SetTextColor(50, 50, 50)
			icon := c.Icon
			if icon != "" {
				icon += " "
			}
			pdf.CellFormat(catColW, 6, truncate(icon+c.Name, 16), "1", 0, "L", false, 0, "")
			typeLabel := "支出"
			typeColor := []int{255, 77, 79}
			if c.Type == "income" {
				typeLabel = "收入"
				typeColor = []int{82, 196, 26}
			}
			pdf.SetTextColor(typeColor[0], typeColor[1], typeColor[2])
			pdf.CellFormat(typeColW, 6, typeLabel, "1", 0, "C", false, 0, "")
			pdf.SetTextColor(50, 50, 50)
			pdf.CellFormat(amtColW, 6, fmt.Sprintf("¥%.2f", c.Total), "1", 0, "R", false, 0, "")
			pdf.CellFormat(pctColW, 6, fmt.Sprintf("%.1f%%", c.Percentage), "1", 1, "C", false, 0, "")
		}
	}

	// Footer
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(180, 180, 180)
	pdf.CellFormat(pageW, 10, fmt.Sprintf("生成时间: %s | Personal Bookkeeping v3.0",
		time.Now().Format("2006-01-02 15:04")), "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return buf.Bytes(), nil
}

func parsePeriod(s string) (year, month int, err error) {
	if len(s) != 7 || s[4] != '-' {
		return 0, 0, fmt.Errorf("expected YYYY-MM")
	}
	_, err = fmt.Sscanf(s, "%d-%d", &year, &month)
	if err != nil || month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("invalid month")
	}
	return
}

func calcChange(current, previous float64) string {
	if previous == 0 {
		if current == 0 {
			return "持平"
		}
		return "+∞"
	}
	pct := (current - previous) / previous * 100
	sign := ""
	if pct > 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%.1f%%", sign, pct)
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// BuildReportData builds report data using the service's DB instance.
func (s *ReportService) BuildReportData(ledgerID, userID uuid.UUID, period ReportPeriod, periodStr string) (*ReportData, error) {
	return BuildReportData(s.DB, ledgerID, userID, period, periodStr)
}
