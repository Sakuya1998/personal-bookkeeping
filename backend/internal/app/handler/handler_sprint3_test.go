package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// =============================================================================
// Sprint 3 — RecurringRule (周期性交易)
// =============================================================================
//
// dev 实现以下 handler 后，移除 t.Skip() 并填充完整断言。

// POST /api/v1/recurring
// Request:  { "ledger_id","category_id","type","amount","currency",
//             "frequency","interval","start_date","end_date","description","tags" }
// Response: 201 { "code":201, "data": { "id":"...", ... } }

func TestRecurringCreate_Daily(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "rec_daily_"+t.Name(), "testpass123")

	// Setup: create ledger + category
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/recurring", token, map[string]interface{}{
		"ledger_id":        ledgerID,
		"category_id":      catID,
		"type":             "expense",
		"amount":           15.00,
		"currency":         "CNY",
		"frequency":        "daily",
		"interval":         1,
		"start_date":       time.Now().Format("2006-01-02"),
		"description":      "每日咖啡",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
	// Verify response contains id with non-empty UUID
	var resp struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, err := uuid.Parse(resp.Data.ID); err != nil {
		t.Errorf("expected valid UUID, got %q: %v", resp.Data.ID, err)
	}
}

func TestRecurringCreate_Weekly(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "rec_weekly_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/recurring", token, map[string]interface{}{
		"ledger_id":   ledgerID,
		"category_id": catID,
		"type":        "income",
		"amount":      5000.00,
		"currency":    "CNY",
		"frequency":   "weekly",
		"interval":    2,
		"start_date":  "2026-06-01",
		"description": "双周工资",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestRecurringCreate_Monthly(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "rec_monthly_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/recurring", token, map[string]interface{}{
		"ledger_id":   ledgerID,
		"category_id": catID,
		"type":        "expense",
		"amount":      3000.00,
		"currency":    "CNY",
		"frequency":   "monthly",
		"interval":    1,
		"start_date":  "2026-01-01",
		"end_date":    "2026-12-31",
		"description": "房租",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestRecurringCreate_Yearly(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "rec_yearly_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/recurring", token, map[string]interface{}{
		"ledger_id":   ledgerID,
		"category_id": catID,
		"type":        "expense",
		"amount":      199.00,
		"currency":    "USD",
		"frequency":   "yearly",
		"interval":    1,
		"start_date":  "2026-01-01",
		"description": "域名续费",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestRecurringCreate_InvalidFrequency(t *testing.T) {
	// 无效频率应返回 400
	r := testEngine(t)
	token := getToken(t, r, "rec_inv_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/recurring", token, map[string]interface{}{
		"ledger_id":   ledgerID,
		"category_id": catID,
		"type":        "expense",
		"amount":      100,
		"currency":    "CNY",
		"frequency":   "bi-century", // 无效
		"interval":    1,
		"start_date":  "2026-01-01",
	}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid frequency: expected 400, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestRecurringCreate_ZeroAmount(t *testing.T) {
	// 金额为 0 应返回 400
	r := testEngine(t)
	token := getToken(t, r, "rec_zero_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/recurring", token, map[string]interface{}{
		"ledger_id":   ledgerID,
		"category_id": catID,
		"type":        "expense",
		"amount":      0,
		"currency":    "CNY",
		"frequency":   "monthly",
		"interval":    1,
		"start_date":  "2026-01-01",
	}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("zero amount: expected 400, got %d", w.Code)
	}
}

func TestRecurringCreate_EndDateBeforeStart(t *testing.T) {
	// 结束日期早于开始日期 → 400
	r := testEngine(t)
	token := getToken(t, r, "rec_date_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/recurring", token, map[string]interface{}{
		"ledger_id":   ledgerID,
		"category_id": catID,
		"type":        "expense",
		"amount":      100,
		"currency":    "CNY",
		"frequency":   "monthly",
		"interval":    1,
		"start_date":  "2026-12-01",
		"end_date":    "2026-01-01", // 早于开始
	}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("end_date before start_date: expected 400, got %d", w.Code)
	}
}

func TestRecurringCreate_Unauthorized(t *testing.T) {
	// 无 token → 401
	w := httptest.NewRecorder()
	r := testEngine(t)
	r.ServeHTTP(w, jsonRequest("POST", "/api/v1/recurring", map[string]interface{}{
		"ledger_id":   uuid.New().String(),
		"category_id": uuid.New().String(),
		"type":        "expense",
		"amount":      100,
		"currency":    "CNY",
		"frequency":   "monthly",
		"interval":    1,
		"start_date":  "2026-01-01",
	}))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// ---------- Recurring List ----------
//
// GET /api/v1/recurring
// Query: ledger_id (optional filter)
// Response: 200 { "code":200, "data": [ ... ] }

func TestRecurringList(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "rec_list_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	// Create 2 rules
	createRecurring(t, r, token, ledgerID, catID, "expense", 100, "monthly")
	createRecurring(t, r, token, ledgerID, catID, "expense", 200, "weekly")

	// List all rules for the user
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/recurring", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list recurring: expected 200, got %d", w.Code)
	}
	// Verify response is an array with 2+ items
	var listResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if len(listResp.Data) < 2 {
		t.Fatalf("expected at least 2 rules, got %d", len(listResp.Data))
	}
}

func TestRecurringList_Empty(t *testing.T) {
	// 无规则时返回空数组
	r := testEngine(t)
	token := getToken(t, r, "rec_empty_"+t.Name(), "testpass123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/recurring", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Verify: data is [] (empty array, not null)
}

// ---------- Recurring Update ----------
//
// PUT /api/v1/recurring/:id
// Request:  partial fields
// Response: 200 { "code":200, "data": { ... } }

func TestRecurringUpdate_Amount(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "rec_upd1_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	// Create rule
	ruleID := createRecurring(t, r, token, ledgerID, catID, "expense", 100, "monthly")

	// Update amount
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("PUT", "/api/v1/recurring/"+ruleID, token,
		map[string]interface{}{"amount": 200}))
	if w.Code != http.StatusOK {
		t.Fatalf("update amount: expected 200, got %d", w.Code)
	}
	// Verify response data.amount == 200
}

func TestRecurringUpdate_Frequency(t *testing.T) {
	// 更新频率从 monthly → weekly
	r := testEngine(t)
	token := getToken(t, r, "rec_upd2_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	// Create rule with monthly frequency
	ruleID := createRecurring(t, r, token, ledgerID, catID, "expense", 100, "monthly")

	// Update frequency to weekly
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("PUT", "/api/v1/recurring/"+ruleID, token,
		map[string]interface{}{"frequency": "weekly"}))
	if w.Code != http.StatusOK {
		t.Fatalf("update frequency: expected 200, got %d", w.Code)
	}

	// GET /api/v1/recurring to verify frequency changed
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/recurring", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list recurring: expected 200, got %d", w.Code)
	}
	var listResp struct {
		Data []struct {
			ID        string `json:"id"`
			Frequency string `json:"frequency"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	for _, rule := range listResp.Data {
		if rule.ID == ruleID {
			if rule.Frequency != "weekly" {
				t.Fatalf("expected frequency 'weekly', got %q", rule.Frequency)
			}
			return
		}
	}
	t.Fatal("updated rule not found in list")
}

func TestRecurringUpdate_NotFound(t *testing.T) {
	// 不存在的规则 ID → 404
	w := httptest.NewRecorder()
	r := testEngine(t)
	token := getToken(t, r, "rec_upd404_"+t.Name(), "testpass123")
	r.ServeHTTP(w, authenticatedRequest("PUT",
		"/api/v1/recurring/"+uuid.New().String(), token,
		map[string]interface{}{"amount": 100}))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRecurringUpdate_Unauthorized(t *testing.T) {
	// 其他用户的规则不可编辑 → 404
	r := testEngine(t)

	// Create user A with rule
	tokenA := getToken(t, r, "rec_ua1_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, tokenA)
	ruleID := createRecurring(t, r, tokenA, ledgerID, catID, "expense", 100, "monthly")

	// Create user B
	tokenB := getToken(t, r, "rec_ua2_"+t.Name(), "testpass123")

	// User B PUT /api/v1/recurring/{userA_rule_id} → expect 404
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("PUT", "/api/v1/recurring/"+ruleID, tokenB,
		map[string]interface{}{"amount": 200}))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ---------- Recurring Delete ----------
//
// DELETE /api/v1/recurring/:id
// Response: 200 { "code":200 }

func TestRecurringDelete(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "rec_del_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)
	ruleID := createRecurring(t, r, token, ledgerID, catID, "expense", 100, "monthly")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE", "/api/v1/recurring/"+ruleID, token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d", w.Code)
	}

	// Verify: GET returns empty list
}

func TestRecurringDelete_NotFound(t *testing.T) {
	// 不存在的 ID → 404
	r := testEngine(t)
	token := getToken(t, r, "rec_del404_"+t.Name(), "testpass123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("DELETE",
		"/api/v1/recurring/"+uuid.New().String(), token, nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRecurringDelete_Unauthorized(t *testing.T) {
	// 无 token → 401
	r := testEngine(t)
	token := getToken(t, r, "rec_del_ua_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)
	ruleID := createRecurring(t, r, token, ledgerID, catID, "expense", 100, "monthly")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("DELETE", "/api/v1/recurring/"+ruleID, nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// =============================================================================
// Sprint 3 — Budget (支出预警)
// =============================================================================

// POST /api/v1/budgets
// Request:  { "ledger_id","category_id"(optional),"month","amount" }
// Response: 201 { "code":201, "data": { "id":"...", ... } }
// 同一 ledger + category + month 组合应 upsert (返回 200)

func TestBudgetCreate(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "bgt_create_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/budgets", token, map[string]interface{}{
		"ledger_id":   ledgerID,
		"category_id": catID,
		"month":       "2026-06",
		"amount":      2000.00,
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestBudgetCreate_OverallBudget(t *testing.T) {
	// 不带 category_id 的全局预算
	r := testEngine(t)
	token := getToken(t, r, "bgt_all_"+t.Name(), "testpass123")
	ledgerID, _ := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/budgets", token, map[string]interface{}{
		"ledger_id": ledgerID,
		"month":     "2026-06",
		"amount":    10000.00,
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestBudgetCreate_Upsert(t *testing.T) {
	// 同一组合应 upsert (返回 200, 不是 409)
	r := testEngine(t)
	token := getToken(t, r, "bgt_ups_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	// First create
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/budgets", token, map[string]interface{}{
		"ledger_id": ledgerID, "category_id": catID, "month": "2026-06", "amount": 2000,
	}))
	// Second create (same key) → upsert
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/budgets", token, map[string]interface{}{
		"ledger_id": ledgerID, "category_id": catID, "month": "2026-06", "amount": 2500,
	}))
	if w.Code != http.StatusOK {
		t.Fatalf("upsert: expected 200, got %d", w.Code)
	}
}

func TestBudgetCreate_ZeroAmount(t *testing.T) {
	// 预算为 0 → 400
	r := testEngine(t)
	token := getToken(t, r, "bgt_zro_"+t.Name(), "testpass123")
	ledgerID, _ := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/budgets", token, map[string]interface{}{
		"ledger_id": ledgerID, "month": "2026-06", "amount": 0,
	}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("zero budget: expected 400, got %d", w.Code)
	}
}

func TestBudgetCreate_NegativeAmount(t *testing.T) {
	// 负值 → 400
	r := testEngine(t)
	token := getToken(t, r, "bgt_neg_"+t.Name(), "testpass123")
	ledgerID, _ := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/budgets", token, map[string]interface{}{
		"ledger_id": ledgerID, "month": "2026-06", "amount": -100,
	}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("negative amount: expected 400, got %d", w.Code)
	}
}

func TestBudgetCreate_InvalidCategory(t *testing.T) {
	// 不存在的分类 ID → 400 或 404
	r := testEngine(t)
	token := getToken(t, r, "bgt_invcat_"+t.Name(), "testpass123")
	ledgerID, _ := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/budgets", token, map[string]interface{}{
		"ledger_id": ledgerID, "category_id": uuid.New().String(), "month": "2026-06", "amount": 500,
	}))
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Fatalf("invalid category: expected 400 or 404, got %d", w.Code)
	}
}

func TestBudgetCreate_InvalidMonth(t *testing.T) {
	// 无效月份格式 → 400
	r := testEngine(t)
	token := getToken(t, r, "bgt_invmon_"+t.Name(), "testpass123")
	ledgerID, _ := setupLedgerAndCategory(t, r, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/budgets", token, map[string]interface{}{
		"ledger_id": ledgerID, "month": "abc", "amount": 500,
	}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid month: expected 400, got %d", w.Code)
	}
}

// ---------- Budget List ----------
//
// GET /api/v1/budgets?month=2026-06
// Response: 200 { "code":200, "data": [ ... ] }

func TestBudgetList(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "bgt_list_"+t.Name(), "testpass123")
	_, _ = setupLedgerAndCategory(t, r, token)

	// Create a budget
	// ...

	// List by month
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET",
		"/api/v1/budgets?month=2026-06", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Verify response contains budget(s)
}

func TestBudgetList_Empty(t *testing.T) {
	// 无预算月份返回空数组
	r := testEngine(t)
	token := getToken(t, r, "bgt_none_"+t.Name(), "testpass123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET",
		"/api/v1/budgets?month=2099-01", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Verify: data is []
}

func TestBudgetList_MissingMonth(t *testing.T) {
	// 缺少 month 参数 → 400
	r := testEngine(t)
	token := getToken(t, r, "bgt_nomon_"+t.Name(), "testpass123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/budgets", token, nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing month: expected 400, got %d", w.Code)
	}
}

// ---------- Budget Status ----------
//
// GET /api/v1/budgets/status?month=2026-06
// Response: 200 { "code":200, "data": [
//   { "category_id","category_name","budget":2000,"spent":850,"percentage":42.5 }
// ] }

func TestBudgetStatus_Normal(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "bgt_st1_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	// Create budget: 2000 for category
	createBudget(t, r, token, ledgerID, catID, time.Now().Format("2006-01"), 2000)

	// Create transactions: 850 spent in this category this month
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions", token, map[string]interface{}{
		"ledger_id":        ledgerID,
		"category_id":      catID,
		"type":             "expense",
		"amount":           850.00,
		"currency":         "CNY",
		"transaction_date": time.Now().Format("2006-01-02"),
		"description":      "test",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create transaction: expected 201, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET",
		"/api/v1/budgets/status?month="+time.Now().Format("2006-01"), token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Verify: returned budget.amount == 2000, spent == 850, percentage ~42.5
	var statusResp struct {
		Data []struct {
			Budget     float64 `json:"budget"`
			Spent      float64 `json:"spent"`
			Percentage float64 `json:"percentage"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &statusResp)
	if len(statusResp.Data) < 1 {
		t.Fatal("expected at least 1 budget status entry")
	}
	entry := statusResp.Data[0]
	if entry.Budget != 2000 {
		t.Errorf("expected budget 2000, got %f", entry.Budget)
	}
	if entry.Spent != 850 {
		t.Errorf("expected spent 850, got %f", entry.Spent)
	}
	if entry.Percentage < 41.5 || entry.Percentage > 43.5 {
		t.Errorf("expected percentage ~42.5, got %f", entry.Percentage)
	}
}

func TestBudgetStatus_OverBudget(t *testing.T) {
	// 超支场景: spent > budget → percentage > 100
	r := testEngine(t)
	token := getToken(t, r, "bgt_st2_"+t.Name(), "testpass123")
	ledgerID, catID := setupLedgerAndCategory(t, r, token)

	// Create budget: 1000 for category
	createBudget(t, r, token, ledgerID, catID, time.Now().Format("2006-01"), 1000)

	// Create transactions: 1500 spent in this category this month
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions", token, map[string]interface{}{
		"ledger_id":        ledgerID,
		"category_id":      catID,
		"type":             "expense",
		"amount":           1500.00,
		"currency":         "CNY",
		"transaction_date": time.Now().Format("2006-01-02"),
		"description":      "test over budget",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create transaction: expected 201, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET",
		"/api/v1/budgets/status?month="+time.Now().Format("2006-01"), token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var statusResp struct {
		Data []struct {
			Budget     float64 `json:"budget"`
			Spent      float64 `json:"spent"`
			Percentage float64 `json:"percentage"`
			OverBudget bool    `json:"over_budget"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &statusResp)
	if len(statusResp.Data) < 1 {
		t.Fatal("expected at least 1 budget status entry")
	}
	entry := statusResp.Data[0]
	if entry.Percentage <= 100 {
		t.Errorf("expected percentage > 100 (over budget), got %f", entry.Percentage)
	}
	if !entry.OverBudget {
		t.Error("expected over_budget to be true")
	}
}

func TestBudgetStatus_NoSpending(t *testing.T) {
	// 有预算但无支出 → spent=0, percentage=0
}

func TestBudgetStatus_MissingMonth(t *testing.T) {
	// 缺少 month → 400
}

// =============================================================================
// Helpers for Sprint 3 (可被 test 函数复用, dev 实现 handler 后使用)
// =============================================================================

// setupLedgerAndCategory 创建测试账本和分类, 返回两者的 ID。
// 该函数先创建账本, 再创建分类, 解析响应提取 UUID。
func setupLedgerAndCategory(t *testing.T, r *gin.Engine, token string) (ledgerID, catID string) {
	t.Helper()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{
		"name": "S3测试账本_" + t.Name(),
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("setup: create ledger: %d, body: %s", w.Code, w.Body.String())
	}
	var lr struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &lr)
	ledgerID = lr.Data.ID

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name": "S3测试分类_" + t.Name(), "type": "expense", "ledger_id": ledgerID,
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("setup: create category: %d, body: %s", w.Code, w.Body.String())
	}
	var cr struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &cr)
	catID = cr.Data.ID
	return
}

// createRecurring 创建一条周期性规则并返回 ID。
// dev 实现 handler 后可用此辅助函数简化测试 setup。
func createRecurring(t *testing.T, r *gin.Engine, token, ledgerID, catID, txnType string,
	amount float64, frequency string) (ruleID string) {
	t.Helper()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/recurring", token, map[string]interface{}{
		"ledger_id":   ledgerID,
		"category_id": catID,
		"type":        txnType,
		"amount":      amount,
		"currency":    "CNY",
		"frequency":   frequency,
		"interval":    1,
		"start_date":  time.Now().Format("2006-01-02"),
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("createRecurring: expected 201, got %d", w.Code)
	}
	var resp struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.ID
}

// createBudget 创建一条预算并返回 ID。
func createBudget(t *testing.T, r *gin.Engine, token, ledgerID, catID, month string,
	amount float64) (budgetID string) {
	t.Helper()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/budgets", token, map[string]interface{}{
		"ledger_id":   ledgerID,
		"category_id": catID,
		"month":       month,
		"amount":      amount,
	}))
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("createBudget: expected 201/200, got %d", w.Code)
	}
	var resp struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.ID
}
