package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// OCRResult 识别结果
type OCRResult struct {
	Text      string `json:"text"`
	Amount    float64 `json:"amount,omitempty"`
	Date      string  `json:"date,omitempty"`      // YYYY-MM-DD
	Merchant  string  `json:"merchant,omitempty"`
	RawText   string  `json:"raw_text"`
}

type ocrRegion struct {
	Text       string    `json:"text"`
	Confidence float64   `json:"confidence"`
	BBox       [][2]float64 `json:"bbox,omitempty"`
}

// ocrServiceResponse 对应 gunthercox/ocr-service 的响应格式
type ocrServiceResponse struct {
	Text    string      `json:"text"`
	Regions []ocrRegion `json:"regions"`
}

// RecognizeReceipt 识别小票图片内容
func RecognizeReceipt(endpoint string, file io.Reader, filename string) (*OCRResult, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}
	// Specify PaddleOCR engine
	writer.WriteField("engine", "paddleocr")
	writer.WriteField("lang", "en")
	writer.Close()

	resp, err := http.Post(endpoint+"/", writer.FormDataContentType(), &buf)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ocr service http %d: %s", resp.StatusCode, string(body))
	}

	var sr ocrServiceResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Build lines from regions
	lines := make([]string, 0, len(sr.Regions))
	for _, r := range sr.Regions {
		lines = append(lines, strings.TrimSpace(r.Text))
	}
	rawText := sr.Text
	if rawText == "" {
		rawText = strings.Join(lines, "\n")
	}

	result := &OCRResult{
		Text:    rawText,
		RawText: rawText,
	}

	// Parse amount from text
	if amt, ok := extractAmount(lines); ok {
		result.Amount = amt
	}

	// Parse date
	if d, ok := extractDate(lines); ok {
		result.Date = d
	}

	// Parse merchant name (first few lines, non-date non-amount)
	if m, ok := extractMerchant(lines); ok {
		result.Merchant = m
	}

	slog.Info("ocr completed",
		"merchant", result.Merchant,
		"amount", result.Amount,
		"date", result.Date,
		"lines", len(lines),
	)
	return result, nil
}

// amountRegex 匹配金额: ¥12.50, 12.50元, 12.50
var amountRegex = regexp.MustCompile(`(?:¥|￥|CNY|USD)?\s*(\d+\.\d{2})\s*(?:元)?`)

// extractAmount 从文本行中提取金额
func extractAmount(lines []string) (float64, bool) {
	var maxAmt float64
	found := false
	for _, line := range lines {
		matches := amountRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			amt, err := strconv.ParseFloat(matches[1], 64)
			if err == nil && amt > maxAmt {
				maxAmt = amt
				found = true
			}
		}
		// Also try simple float at end of line (total pattern)
		clean := strings.NewReplacer(" ", "", ",", "").Replace(line)
		if strings.Contains(strings.ToLower(clean), "total") || strings.Contains(clean, "合计") || strings.Contains(clean, "实付") {
			re := regexp.MustCompile(`(\d+\.\d{2})`)
			parts := re.FindStringSubmatch(clean)
			if len(parts) >= 2 {
				amt, err := strconv.ParseFloat(parts[1], 64)
				if err == nil && amt > maxAmt {
					maxAmt = amt
					found = true
				}
			}
		}
	}
	if found {
		return maxAmt, true
	}
	return 0, false
}

// dateRegex 匹配日期格式
var dateRegex = regexp.MustCompile(`(\d{4})[-/]?(\d{1,2})[-/]?(\d{1,2})`)

func extractDate(lines []string) (string, bool) {
	for _, line := range lines {
		matches := dateRegex.FindStringSubmatch(line)
		if len(matches) >= 4 {
			y, _ := strconv.Atoi(matches[1])
			m, _ := strconv.Atoi(matches[2])
			d, _ := strconv.Atoi(matches[3])
			if y >= 2000 && y <= 2100 && m >= 1 && m <= 12 && d >= 1 && d <= 31 {
				return fmt.Sprintf("%04d-%02d-%02d", y, m, d), true
			}
		}
	}
	// Fallback to today
	return time.Now().Format("2006-01-02"), false
}

// extractMerchant 提取商家名称（通常在前几行，不是日期和金额）
func extractMerchant(lines []string) (string, bool) {
	for _, line := range lines {
		clean := strings.TrimSpace(line)
		if clean == "" {
			continue
		}
		// Skip lines that look like dates, amounts, or common keywords
		if dateRegex.MatchString(clean) {
			continue
		}
		if amountRegex.MatchString(clean) {
			continue
		}
		skipWords := []string{"小票", "收据", "发票", "单号", "电话", "地址", "数量", "单价", "合计", "实付", "找零", "谢谢", "欢迎", "NO.", "Tel", "tel", "Addr"}
		shouldSkip := false
		for _, w := range skipWords {
			if strings.Contains(clean, w) {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}
		// First non-skipped line is likely the merchant name
		if len(clean) >= 2 && len(clean) <= 50 {
			return clean, true
		}
	}
	return "", false
}
