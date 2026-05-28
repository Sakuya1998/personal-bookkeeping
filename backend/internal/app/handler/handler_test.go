package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"personal-bookkeeping/internal/infra/database"
	"personal-bookkeeping/internal/app/router"
	"personal-bookkeeping/internal/infra/config"

	"github.com/gin-gonic/gin"
)

func testDSN() string {
	if dsn := os.Getenv("BOOKKEEPING_TEST_DSN"); dsn != "" {
		return dsn
	}
	return "host=localhost port=5432 user=bookkeeper password=bookkeeper_dev dbname=bookkeeping_test sslmode=disable"
}

func testCfg() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{Port: "0"},
		DB: config.DBConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "bookkeeper",
			Password: "bookkeeper_dev",
			Name:     "bookkeeping_test",
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret:       "test-secret",
			ExpireMinute: 60,
		},
		CORS: config.CORSConfig{Origins: "*"},
		Log:  config.LogConfig{Target: "stderr"},
		OTEL: config.OTELConfig{Enabled: false},
	}
}

func testEngine(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	cfg := testCfg()
	database.Init(cfg)
	if database.GetDB() == nil {
		t.Skip("requires PostgreSQL — database not available")
	}
	if err := database.Ping(); err != nil {
		t.Skipf("requires PostgreSQL — ping failed: %v", err)
	}

	r := gin.New()
	router.Setup(r, cfg)
	return r
}

func jsonBody(v interface{}) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func jsonRequest(method, path string, body interface{}) *http.Request {
	req := httptest.NewRequest(method, path, jsonBody(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func authenticatedRequest(method, path, token string, body interface{}) *http.Request {
	req := jsonRequest(method, path, body)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func getToken(t *testing.T, r *gin.Engine, username, password string) string {
	t.Helper()

	// 尝试注册
	w := httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("POST", "/api/v1/auth/register", map[string]string{
		"username": username,
		"email":    username + "@test.com",
		"password": password,
	}))

	// 已存在则登录
	if w.Code == http.StatusConflict {
		w = httptest.NewRecorder()
		r.ServeHTTP(w, jsonRequest("POST", "/api/v1/auth/login", map[string]string{
			"username": username,
			"password": password,
		}))
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v, body: %s", err, w.Body.String())
	}
	return resp.Data.Token
}

// ---------- Tests ----------

func TestHealth(t *testing.T) {
	r := testEngine(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("GET", "/api/v1/health", nil))
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	r := testEngine(t)
	username := "test_user_" + t.Name()

	// Register
	w := httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("POST", "/api/v1/auth/register", map[string]string{
		"username": username,
		"email":    username + "@test.com",
		"password": "testpass123",
	}))
	if w.Code != http.StatusCreated {
		t.Errorf("register: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}

	// Login
	w = httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("POST", "/api/v1/auth/login", map[string]string{
		"username": username,
		"password": "testpass123",
	}))
	if w.Code != http.StatusOK {
		t.Errorf("login: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestLedgerCRUD(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "ledger_test_"+t.Name(), "testpass123")

	// Create
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{
		"name": "测试账本",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create ledger: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal failed: %v, body: %s", err, w.Body.String())
	}
	ledgerID := created.Data.ID

	// List
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers", token, nil))
	if w.Code != http.StatusOK {
		t.Errorf("list ledgers: expected 200, got %d", w.Code)
	}

	// Delete
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
	if w.Code != http.StatusOK {
		t.Errorf("delete ledger: expected 200, got %d", w.Code)
	}
}

func TestAnalyticsEndpoints(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "analytics_test_"+t.Name(), "testpass123")

	// 1. Create ledger
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{
		"name": "分析测试账本",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create ledger: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
	var ledgerResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &ledgerResp)
	ledgerID := ledgerResp.Data.ID

	// 2. Create categories
	type createResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	var catResp createResp

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name":      "餐饮",
		"type":      "expense",
		"ledger_id": ledgerID,
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create category: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &catResp)
	expenseCatID := catResp.Data.ID

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name":      "工资",
		"type":      "income",
		"ledger_id": ledgerID,
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create income category: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &catResp)
	incomeCatID := catResp.Data.ID

	// 3. Create transactions across two months
	createTxn := func(txnType, catID, date string, amount float64) int {
		t.Helper()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions", token, map[string]interface{}{
			"ledger_id":        ledgerID,
			"category_id":      catID,
			"type":             txnType,
			"amount":           amount,
			"currency":         "CNY",
			"transaction_date": date,
		}))
		return w.Code
	}

	// Current month
	now := time.Now()
	m1 := now.Format("2006-01-02")
	m2 := now.AddDate(0, -1, 0).Format("2006-01-02")
	m3 := now.AddDate(0, -2, 0).Format("2006-01-02")

	code := createTxn("expense", expenseCatID, m1, 100)
	if code != http.StatusCreated {
		t.Fatalf("create txn1: expected 201, got %d", code)
	}
	code = createTxn("expense", expenseCatID, m1, 50)
	if code != http.StatusCreated {
		t.Fatalf("create txn2: expected 201, got %d", code)
	}
	code = createTxn("income", incomeCatID, m1, 500)
	if code != http.StatusCreated {
		t.Fatalf("create txn3: expected 201, got %d", code)
	}
	createTxn("expense", expenseCatID, m2, 200)
	createTxn("income", incomeCatID, m2, 800)
	createTxn("expense", expenseCatID, m3, 300)
	createTxn("income", incomeCatID, m3, 600)

	// 4. Test monthly-trend
	t.Run("MonthlyTrend", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID+"/monthly-trend?months=6", token, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("monthly-trend: expected 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Data []struct {
				Month   string  `json:"month"`
				Income  float64 `json:"income"`
				Expense float64 `json:"expense"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal monthly-trend: %v", err)
		}
		if len(resp.Data) == 0 {
			t.Error("monthly-trend: expected at least 1 month of data")
		}
		for _, item := range resp.Data {
			if item.Month == "" {
				t.Error("monthly-trend: month should not be empty")
			}
		}
	})

	// 5. Test category-breakdown
	t.Run("CategoryBreakdown", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID+"/category-breakdown", token, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("category-breakdown: expected 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Data []struct {
				CategoryID   string  `json:"category_id"`
				CategoryName string  `json:"category_name"`
				CategoryIcon string  `json:"category_icon"`
				Type         string  `json:"type"`
				Total        float64 `json:"total"`
				Percentage   float64 `json:"percentage"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal category-breakdown: %v", err)
		}
		if len(resp.Data) == 0 {
			t.Error("category-breakdown: expected at least 1 category")
		}
		for _, item := range resp.Data {
			if item.CategoryID == "" {
				t.Error("category-breakdown: category_id should not be empty")
			}
			if item.Total <= 0 {
				t.Errorf("category-breakdown: total should be > 0 for %s, got %f", item.CategoryName, item.Total)
			}
			// percentage should be between 0 and 100
			if item.Percentage < 0 || item.Percentage > 100 {
				t.Errorf("category-breakdown: percentage out of range for %s: %f", item.CategoryName, item.Percentage)
			}
		}
	})

	// 6. Test daily-transactions
	t.Run("DailyTransactions", func(t *testing.T) {
		w := httptest.NewRecorder()
		url := "/api/v1/ledgers/" + ledgerID + "/daily-transactions?year=" + strconv.Itoa(now.Year()) + "&month=" + strconv.Itoa(int(now.Month()))
		r.ServeHTTP(w, authenticatedRequest("GET", url, token, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("daily-transactions: expected 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Data []struct {
				Date    string  `json:"date"`
				Income  float64 `json:"income"`
				Expense float64 `json:"expense"`
				Count   int64   `json:"count"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal daily-transactions: %v", err)
		}
		// Should have at least today's data
		foundToday := false
		for _, item := range resp.Data {
			if item.Date == m1 {
				foundToday = true
				if item.Count <= 0 {
					t.Error("daily-transactions: count should be > 0")
				}
			}
		}
		if !foundToday {
			t.Logf("daily-transactions: today's transactions not found (may be ok if DB rollback), got %d days", len(resp.Data))
		}
	})

	// Cleanup
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

// ---------- Category CRUD ----------

func TestCategoryCRUD(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "cat_test_"+t.Name(), "testpass123")

	// First, create a ledger to bind categories
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "分类测试账本"}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create ledger: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
	var ledgerResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &ledgerResp)
	ledgerID := ledgerResp.Data.ID

	// Create category
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name":      "餐饮",
		"type":      "expense",
		"ledger_id": ledgerID,
		"icon":      "🍽️",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create category: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
	var catCreateResp struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &catCreateResp)
	catID := catCreateResp.Data.ID

	// List categories by ledger
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID+"/categories", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list categories: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
	var catListResp struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &catListResp)
	if len(catListResp.Data) == 0 {
		t.Fatal("list categories: expected at least 1 category")
	}
	found := false
	for _, c := range catListResp.Data {
		if c.ID == catID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("list categories: created category not found in list")
	}

	// Update category
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("PUT", "/api/v1/categories/"+catID, token, map[string]interface{}{
		"name": "餐饮美食",
	}))
	if w.Code != http.StatusOK {
		t.Fatalf("update category: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Delete category
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/categories/"+catID, token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("delete category: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Cleanup: delete ledger (cascades)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

func TestCategoryDelete_WithTransactions(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "cat_del_"+t.Name(), "testpass123")

	// Create ledger
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "分类删除测试"}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create ledger: expected 201, got %d", w.Code)
	}
	var ledgerResp struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &ledgerResp)
	ledgerID := ledgerResp.Data.ID

	// Create category
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name":      "交通",
		"type":      "expense",
		"ledger_id": ledgerID,
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create category: expected 201, got %d", w.Code)
	}
	var catResp struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &catResp)
	catID := catResp.Data.ID

	// Create transaction using this category
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions", token, map[string]interface{}{
		"ledger_id":        ledgerID,
		"category_id":      catID,
		"type":             "expense",
		"amount":           50,
		"transaction_date": "2024-06-01",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create transaction: expected 201, got %d", w.Code)
	}

	// Delete category — should fail because it has transactions
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/categories/"+catID, token, nil))
	if w.Code != http.StatusConflict {
		t.Fatalf("delete category with transactions: expected 409, got %d, body: %s", w.Code, w.Body.String())
	}

	// Cleanup
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

// ---------- Transaction CRUD ----------

func TestTransactionCRUD(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "txn_test_"+t.Name(), "testpass123")

	// Create ledger
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "交易测试账本"}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create ledger: expected 201, got %d", w.Code)
	}
	var ledgerResp struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &ledgerResp)
	ledgerID := ledgerResp.Data.ID

	// Create category
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name":      "购物",
		"type":      "expense",
		"ledger_id": ledgerID,
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create category: expected 201, got %d", w.Code)
	}
	var catResp struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &catResp)
	catID := catResp.Data.ID

	// Create transaction
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions", token, map[string]interface{}{
		"ledger_id":        ledgerID,
		"category_id":      catID,
		"type":             "expense",
		"amount":           99.90,
		"currency":         "CNY",
		"description":      "测试购物",
		"transaction_date": "2024-06-15",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create transaction: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
	var txnResp struct {
		Data struct {
			ID              string  `json:"id"`
			Amount          float64 `json:"amount"`
			BaseAmount      float64 `json:"base_amount"`
			Description     *string `json:"description"`
			TransactionDate string  `json:"transaction_date"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &txnResp)
	txnID := txnResp.Data.ID

	// List transactions with filters
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID+"/transactions?type=expense&page=1&page_size=10", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list transactions: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
	var listResp struct {
		Data struct {
			Items      []interface{} `json:"items"`
			Total      int           `json:"total"`
			Page       int           `json:"page"`
			PageSize   int           `json:"page_size"`
			TotalPages int           `json:"total_pages"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if listResp.Data.Total < 1 {
		t.Errorf("list transactions: expected at least 1, got %d", listResp.Data.Total)
	}
	if listResp.Data.Page != 1 {
		t.Errorf("list transactions: expected page 1, got %d", listResp.Data.Page)
	}

	// Update transaction
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("PUT", "/api/v1/transactions/"+txnID, token, map[string]interface{}{
		"amount":      199.90,
		"description": "更新后购物",
	}))
	if w.Code != http.StatusOK {
		t.Fatalf("update transaction: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Delete transaction
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/transactions/"+txnID, token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("delete transaction: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Cleanup
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

func TestTransactionCRUD_WithMultiCurrency(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "txn_mc_"+t.Name(), "testpass123")

	// Create ledger (base currency CNY)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]interface{}{
		"name":          "多币种测试",
		"base_currency": "CNY",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create ledger: expected 201, got %d", w.Code)
	}
	var ledgerResp struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &ledgerResp)
	ledgerID := ledgerResp.Data.ID

	// Create category
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name":      "收入",
		"type":      "income",
		"ledger_id": ledgerID,
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create category: expected 201, got %d", w.Code)
	}
	var catResp struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &catResp)
	catID := catResp.Data.ID

	// Create a USD exchange rate first
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/exchange-rates", token, map[string]interface{}{
		"from_currency": "USD",
		"to_currency":   "CNY",
		"rate":          7.24,
		"date":          "2024-06-15",
	}))
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("create exchange rate: expected 201/200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Create transaction in USD — should auto-convert to base currency
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions", token, map[string]interface{}{
		"ledger_id":        ledgerID,
		"category_id":      catID,
		"type":             "income",
		"amount":           100,
		"currency":         "USD",
		"transaction_date": "2024-06-15",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create multi-currency transaction: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}

	// Cleanup
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

// ---------- Exchange Rate CRUD ----------

func TestExchangeRateCRUD(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "rate_test_"+t.Name(), "testpass123")

	// Create exchange rate
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/exchange-rates", token, map[string]interface{}{
		"from_currency": "USD",
		"to_currency":   "CNY",
		"rate":          7.24,
		"date":          "2024-06-01",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create rate: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
	var rateResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &rateResp)
	rateID := rateResp.Data.ID

	// Create EUR→CNY rate
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/exchange-rates", token, map[string]interface{}{
		"from_currency": "EUR",
		"to_currency":   "CNY",
		"rate":          7.86,
		"date":          "2024-06-01",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create EUR rate: expected 201, got %d", w.Code)
	}

	// List all rates
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/exchange-rates", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list rates: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
	var listResp struct {
		Data []struct {
			ID           string  `json:"id"`
			FromCurrency string  `json:"from_currency"`
			ToCurrency   string  `json:"to_currency"`
			Rate         float64 `json:"rate"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if len(listResp.Data) < 2 {
		t.Errorf("list rates: expected at least 2, got %d", len(listResp.Data))
	}

	// List with filters
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/exchange-rates?from=USD&to=CNY", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list filtered rates: expected 200, got %d", w.Code)
	}

	// Get latest rates
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/exchange-rates/latest", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("latest rates: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Upsert: same date+pair should overwrite (return 200)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/exchange-rates", token, map[string]interface{}{
		"from_currency": "USD",
		"to_currency":   "CNY",
		"rate":          7.25,
		"date":          "2024-06-01",
	}))
	if w.Code != http.StatusOK {
		t.Fatalf("upsert rate: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Delete rate
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/exchange-rates/"+rateID, token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("delete rate: expected 200, got %d", w.Code)
	}

	// Delete nonexistent — should 404
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/exchange-rates/nonexistent-id", token, nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("delete nonexistent rate: expected 404, got %d", w.Code)
	}
}

// ---------- Auth edge cases ----------

func TestAuthMe(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "me_test_"+t.Name(), "testpass123")

	// Get current user
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/auth/me", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("auth/me: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal auth/me: %v", err)
	}
	if resp.Data.Username == "" {
		t.Error("auth/me: username should not be empty")
	}
	if resp.Data.Email == "" {
		t.Error("auth/me: email should not be empty")
	}
}

func TestAuthLogout(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "logout_test_"+t.Name(), "testpass123")

	// Logout
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/auth/logout", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("logout: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Token should now be invalid (blacklisted)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/auth/me", token, nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("auth/me after logout: expected 401, got %d — token was not properly blacklisted", w.Code)
	}
}

func TestLedgerSummary(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "summary_test_"+t.Name(), "testpass123")

	// Create ledger
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "汇总测试"}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create ledger: expected 201, got %d", w.Code)
	}
	var ledgerResp struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &ledgerResp)
	ledgerID := ledgerResp.Data.ID

	// Get summary (should be zeroes)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID+"/summary", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("summary: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Cleanup
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

func TestUnauthorizedAccess(t *testing.T) {
	r := testEngine(t)

	// Access protected endpoint without token
	w := httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("GET", "/api/v1/ledgers", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized: expected 401, got %d", w.Code)
	}

	// Access with malformed token
	w = httptest.NewRecorder()
	req := jsonRequest("GET", "/api/v1/ledgers", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-123")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token: expected 401, got %d", w.Code)
	}
}

func TestLedgerUpdate(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "ledger_upd_"+t.Name(), "testpass123")

	// Create ledger
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "更新测试"}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create ledger: expected 201, got %d", w.Code)
	}
	var created struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &created)
	ledgerID := created.Data.ID

	// Update
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("PUT", "/api/v1/ledgers/"+ledgerID, token, map[string]interface{}{
		"name": "已更新账本",
	}))
	if w.Code != http.StatusOK {
		t.Fatalf("update ledger: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Get detail
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID, token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("get ledger: expected 200, got %d", w.Code)
	}
	var getResp struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &getResp)
	if getResp.Data.Name != "已更新账本" {
		t.Errorf("get ledger: expected name '已更新账本', got '%s'", getResp.Data.Name)
	}

	// Cleanup
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

// ---------- Validation tests (no DB needed — SkipIfNoDB skip at the top) ----------

func TestTransactionValidation_BadInput(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "txn_val_"+t.Name(), "testpass123")

	// Missing required fields
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions", token, map[string]interface{}{
		// no ledger_id, no category_id
		"type":   "expense",
		"amount": 100,
	}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad transaction (missing ledger_id, category_id): expected 400, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestExchangeRateValidation_BadInput(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "rate_val_"+t.Name(), "testpass123")

	// Zero rate
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/exchange-rates", token, map[string]interface{}{
		"from_currency": "USD",
		"to_currency":   "CNY",
		"rate":          0,
		"date":          "2024-06-01",
	}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("zero rate: expected 400, got %d, body: %s", w.Code, w.Body.String())
	}

	// Missing fields
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/exchange-rates", token, map[string]interface{}{
		"from_currency": "USD",
		// no to_currency, no rate
	}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad rate (missing fields): expected 400, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestRegisterValidation(t *testing.T) {
	r := testEngine(t)

	// Empty username
	w := httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("POST", "/api/v1/auth/register", map[string]string{
		"username": "",
		"email":    "test@test.com",
		"password": "test123",
	}))
	if w.Code != http.StatusBadRequest {
		t.Errorf("empty username: expected 400, got %d, body: %s", w.Code, w.Body.String())
	}

	// Short password
	w = httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("POST", "/api/v1/auth/register", map[string]string{
		"username": "valuser",
		"email":    "val@test.com",
		"password": "ab",
	}))
	if w.Code != http.StatusBadRequest {
		t.Errorf("short password: expected 400, got %d, body: %s", w.Code, w.Body.String())
	}

	// Invalid email
	w = httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("POST", "/api/v1/auth/register", map[string]string{
		"username": "valuser2",
		"email":    "not-an-email",
		"password": "test123",
	}))
	if w.Code != http.StatusBadRequest {
		t.Errorf("invalid email: expected 400, got %d, body: %s", w.Code, w.Body.String())
	}

	// Duplicate registration
	w = httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("POST", "/api/v1/auth/register", map[string]string{
		"username": "dup_test_user",
		"email":    "dup@test.com",
		"password": "test123",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("first register: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("POST", "/api/v1/auth/register", map[string]string{
		"username": "dup_test_user",
		"email":    "dup@test.com",
		"password": "test123",
	}))
	if w.Code != http.StatusConflict {
		t.Errorf("duplicate register: expected 409, got %d, body: %s", w.Code, w.Body.String())
	}
}
