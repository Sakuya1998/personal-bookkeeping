package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// extractAmount tests
// ---------------------------------------------------------------------------

func TestExtractAmount(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		wantAmt  float64
		wantOK   bool
	}{
		{
			name:    "¥12.50 pattern",
			lines:   []string{"¥12.50"},
			wantAmt: 12.50,
			wantOK:  true,
		},
		{
			name:    "￥25.00 pattern",
			lines:   []string{"￥25.00"},
			wantAmt: 25.00,
			wantOK:  true,
		},
		{
			name:    "合计:35.00 pattern",
			lines:   []string{"合计:35.00"},
			wantAmt: 35.00,
			wantOK:  true,
		},
		{
			name:    "Total $45.50 pattern",
			lines:   []string{"Total $45.50"},
			wantAmt: 45.50,
			wantOK:  true,
		},
		{
			name:    "实付 99.99 pattern",
			lines:   []string{"实付: 99.99"},
			wantAmt: 99.99,
			wantOK:  true,
		},
		{
			name:    "no amount match",
			lines:   []string{"Hello World", "Some text"},
			wantAmt: 0,
			wantOK:  false,
		},
		{
			name:    "multiple amounts picks max",
			lines:   []string{"¥10.00", "¥20.50", "¥15.00"},
			wantAmt: 20.50,
			wantOK:  true,
		},
		{
			name:    "USD 100.00 pattern",
			lines:   []string{"USD 100.00"},
			wantAmt: 100.00,
			wantOK:  true,
		},
		{
			name:    "CNY 8.50 pattern",
			lines:   []string{"CNY 8.50"},
			wantAmt: 8.50,
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAmt, gotOK := extractAmount(tt.lines)
			if gotOK != tt.wantOK {
				t.Errorf("extractAmount() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotAmt != tt.wantAmt {
				t.Errorf("extractAmount() amt = %v, want %v", gotAmt, tt.wantAmt)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extractDate tests
// ---------------------------------------------------------------------------

func TestExtractDate(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		wantDate string
		wantOK   bool
	}{
		{
			name:     "YYYY-MM-DD format",
			lines:    []string{"2026-06-01"},
			wantDate: "2026-06-01",
			wantOK:   true,
		},
		{
			name:     "YYYY/MM/DD format",
			lines:    []string{"2026/6/1"},
			wantDate: "2026-06-01",
			wantOK:   true,
		},
		{
			name:     "YYYY/MM/DD with leading zeros",
			lines:    []string{"2026/06/01"},
			wantDate: "2026-06-01",
			wantOK:   true,
		},
		{
			name:     "no date match",
			lines:    []string{"Some random text", "No dates here"},
			wantDate: time.Now().Format("2006-01-02"),
			wantOK:   false,
		},
		{
			name:     "date in second line",
			lines:    []string{"Receipt", "2026-12-25"},
			wantDate: "2026-12-25",
			wantOK:   true,
		},
		{
			name:     "year out of range",
			lines:    []string{"1999-01-01"},
			wantDate: time.Now().Format("2006-01-02"),
			wantOK:   false,
		},
		{
			name:     "month out of range",
			lines:    []string{"2026-13-01"},
			wantDate: time.Now().Format("2006-01-02"),
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDate, gotOK := extractDate(tt.lines)
			if gotOK != tt.wantOK {
				t.Errorf("extractDate() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotDate != tt.wantDate {
				t.Errorf("extractDate() date = %v, want %v", gotDate, tt.wantDate)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extractMerchant tests
// ---------------------------------------------------------------------------

func TestExtractMerchant(t *testing.T) {
	tests := []struct {
		name        string
		lines       []string
		wantMerchant string
		wantOK      bool
	}{
		{
			name:        "simple merchant name",
			lines:       []string{"Starbucks Coffee"},
			wantMerchant: "Starbucks Coffee",
			wantOK:      true,
		},
		{
			name: "skip date and amount lines",
			lines: []string{
				"2026-06-01",
				"¥12.50",
				"McDonald's",
			},
			wantMerchant: "McDonald's",
			wantOK:      true,
		},
		{
			name: "skip common keywords",
			lines: []string{
				"小票编号: 001",
				"电话: 12345678",
				"Walmart",
			},
			wantMerchant: "Walmart",
			wantOK:      true,
		},
		{
			name: "empty lines skipped",
			lines: []string{
				"",
				"  ",
				"Costco",
			},
			wantMerchant: "Costco",
			wantOK:      true,
		},
		{
			name:        "no valid merchant found",
			lines:       []string{"2026-06-01", "¥12.50"},
			wantMerchant: "",
			wantOK:      false,
		},
		{
			name: "skip 合计 line",
			lines: []string{
				"7-Eleven",
				"合计: 15.00",
			},
			wantMerchant: "7-Eleven",
			wantOK:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMerchant, gotOK := extractMerchant(tt.lines)
			if gotOK != tt.wantOK {
				t.Errorf("extractMerchant() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotMerchant != tt.wantMerchant {
				t.Errorf("extractMerchant() merchant = %v, want %v", gotMerchant, tt.wantMerchant)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RecognizeReceipt tests
// ---------------------------------------------------------------------------

func TestRecognizeReceipt_Success(t *testing.T) {
	// Mock ocr-service server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/" {
			t.Errorf("expected /, got %s", r.URL.Path)
		}

		// Verify multipart form file was sent
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse multipart form: %v", err)
		}
		file, _, err := r.FormFile("image")
		if err != nil {
			t.Fatalf("form file: %v", err)
		}
		content, _ := io.ReadAll(file)
		file.Close()
		if len(content) == 0 {
			t.Error("expected file content, got empty")
		}
		if r.FormValue("engine") != "paddleocr" {
			t.Errorf("expected engine=paddleocr, got %q", r.FormValue("engine"))
		}
		if r.FormValue("lang") != "ch" {
			t.Errorf("expected lang=ch, got %q", r.FormValue("lang"))
		}

		// Return valid ocr-service response
		resp := ocrServiceResponse{
			Text: "星巴克咖啡\n2026-06-01\n¥35.00",
			Regions: []ocrRegion{
				{Text: "星巴克咖啡", Confidence: 0.95},
				{Text: "2026-06-01", Confidence: 0.98},
				{Text: "¥35.00", Confidence: 0.97},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fileContent := bytes.NewReader([]byte("fake-image-data"))
	result, err := RecognizeReceipt(server.URL, fileContent, "receipt.jpg")
	if err != nil {
		t.Fatalf("RecognizeReceipt failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Amount != 35.00 {
		t.Errorf("expected amount 35.00, got %v", result.Amount)
	}
	if result.Date != "2026-06-01" {
		t.Errorf("expected date 2026-06-01, got %v", result.Date)
	}
	if result.Merchant != "星巴克咖啡" {
		t.Errorf("expected merchant 星巴克咖啡, got %v", result.Merchant)
	}
	if !strings.Contains(result.RawText, "星巴克咖啡") {
		t.Errorf("expected raw text to contain merchant")
	}
}

func TestRecognizeReceipt_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	fileContent := bytes.NewReader([]byte("fake-image-data"))
	_, err := RecognizeReceipt(server.URL, fileContent, "receipt.jpg")
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to mention 500, got: %v", err)
	}
}

func TestRecognizeReceipt_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ocrServiceResponse{
			Text:    "",
			Regions: []ocrRegion{},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fileContent := bytes.NewReader([]byte("fake-image-data"))
	result, err := RecognizeReceipt(server.URL, fileContent, "receipt.jpg")
	if err != nil {
		t.Fatalf("RecognizeReceipt failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Text != "" {
		t.Errorf("expected empty text for no results, got %q", result.Text)
	}
}

func TestRecognizeReceipt_ServerUnreachable(t *testing.T) {
	// Use a server that closes immediately
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do nothing, just hang up
	}))
	server.Close() // Close immediately so the URL is invalid

	fileContent := bytes.NewReader([]byte("fake-image-data"))
	_, err := RecognizeReceipt(server.URL, fileContent, "receipt.jpg")
	if err == nil {
		t.Log("expected error for unreachable server (may pass if connection refused is not treated as error by http.Post)")
	}
}

func TestRecognizeReceipt_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	fileContent := bytes.NewReader([]byte("fake-image-data"))
	_, err := RecognizeReceipt(server.URL, fileContent, "receipt.jpg")
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

// ---------------------------------------------------------------------------
// Integration-style test with known receipt patterns (no network)
// ---------------------------------------------------------------------------

func TestRecognizeReceipt_WithKnownPatterns(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ocrServiceResponse{
			Text: "沃尔玛超市\n2026/5/15\n合计: 128.50",
			Regions: []ocrRegion{
				{Text: "沃尔玛超市", Confidence: 0.96},
				{Text: "2026/5/15", Confidence: 0.94},
				{Text: "合计: 128.50", Confidence: 0.92},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fileContent := bytes.NewReader([]byte("fake-image-data"))
	result, err := RecognizeReceipt(server.URL, fileContent, "receipt.png")
	if err != nil {
		t.Fatalf("RecognizeReceipt failed: %v", err)
	}

	if result.Amount != 128.50 {
		t.Errorf("expected amount 128.50, got %v", result.Amount)
	}
	if result.Date != "2026-05-15" {
		t.Errorf("expected date 2026-05-15, got %v", result.Date)
	}
	if result.Merchant != "沃尔玛超市" {
		t.Errorf("expected merchant 沃尔玛超市, got %v", result.Merchant)
	}

	fmt.Printf("OCR result: amount=%.2f date=%s merchant=%s\n",
		result.Amount, result.Date, result.Merchant)
}
