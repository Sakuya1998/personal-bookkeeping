package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ---------- Batch Delete ----------

func TestTransactionBatchDelete_Success(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "batchdel_"+t.Name(), "testpass123")

	// Setup: create ledger -> category -> 3 transactions
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "批删测试"}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create ledger: expected 201, got %d", w.Code)
	}
	var lr struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &lr)
	ledgerID := lr.Data.ID

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name": "餐饮", "type": "expense", "ledger_id": ledgerID,
	}))
	var cr struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &cr)
	catID := cr.Data.ID

	today := time.Now().Format("2006-01-02")
	var ids []string
	for i := 0; i < 3; i++ {
		w = httptest.NewRecorder()
		r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions", token, map[string]interface{}{
			"ledger_id": ledgerID, "category_id": catID,
			"type": "expense", "amount": 10.0, "transaction_date": today,
		}))
		var tr struct{ Data struct{ ID string } `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &tr)
		ids = append(ids, tr.Data.ID)
	}

	// Delete first 2
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions/batch-delete", token, map[string]interface{}{
		"ids": ids[:2],
	}))
	if w.Code != http.StatusOK {
		t.Fatalf("batch-delete: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Deleted int64 `json:"deleted"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data.Deleted != 2 {
		t.Errorf("expected deleted=2, got %d", resp.Data.Deleted)
	}

	// Cleanup
	r.ServeHTTP(httptest.NewRecorder(), authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

func TestTransactionBatchDelete_EmptyIDs(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "batchdel_empty_"+t.Name(), "testpass123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions/batch-delete", token, map[string]interface{}{
		"ids": []string{},
	}))
	// empty IDs should fail validation
	if w.Code != http.StatusBadRequest {
		t.Errorf("empty ids: expected 400, got %d", w.Code)
	}
}

func TestTransactionBatchDelete_Unauthorized(t *testing.T) {
	r := testEngine(t)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("POST", "/api/v1/transactions/batch-delete", map[string]interface{}{
		"ids": []string{uuid.New().String()},
	}))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("unauthorized: expected 401, got %d", w.Code)
	}
}

// ---------- Batch Update ----------

func TestTransactionBatchUpdate_Success(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "batchupd_"+t.Name(), "testpass123")

	// Setup
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "批更测试"}))
	var lr struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &lr)
	ledgerID := lr.Data.ID

	// Create two categories
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name": "餐饮", "type": "expense", "ledger_id": ledgerID,
	}))
	var cr1 struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &cr1)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name": "交通", "type": "expense", "ledger_id": ledgerID,
	}))
	var cr2 struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &cr2)

	today := time.Now().Format("2006-01-02")
	var ids []string
	for i := 0; i < 2; i++ {
		w = httptest.NewRecorder()
		r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions", token, map[string]interface{}{
			"ledger_id": ledgerID, "category_id": cr1.Data.ID,
			"type": "expense", "amount": 20.0, "transaction_date": today,
		}))
		var tr struct{ Data struct{ ID string } `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &tr)
		ids = append(ids, tr.Data.ID)
	}

	// Batch update to new category
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("PUT", "/api/v1/transactions/batch-update", token, map[string]interface{}{
		"ids":         ids,
		"category_id": cr2.Data.ID,
	}))
	if w.Code != http.StatusOK {
		t.Fatalf("batch-update: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Updated int64 `json:"updated"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data.Updated != 2 {
		t.Errorf("expected updated=2, got %d", resp.Data.Updated)
	}

	// Cleanup
	r.ServeHTTP(httptest.NewRecorder(), authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

func TestTransactionBatchUpdate_InvalidCategory(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "batchupd_inv_"+t.Name(), "testpass123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("PUT", "/api/v1/transactions/batch-update", token, map[string]interface{}{
		"ids":         []string{uuid.New().String()},
		"category_id": uuid.New().String(),
	}))
	// should fail because either transaction doesn't exist or category doesn't exist
	if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Errorf("invalid category: expected 400 or 404, got %d", w.Code)
	}
}

func TestTransactionBatchUpdate_Unauthorized(t *testing.T) {
	r := testEngine(t)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, jsonRequest("PUT", "/api/v1/transactions/batch-update", map[string]interface{}{
		"ids":         []string{uuid.New().String()},
		"category_id": uuid.New().String(),
	}))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("unauthorized: expected 401, got %d", w.Code)
	}
}

// ---------- Export ----------

func TestExportCSV(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "export_csv_"+t.Name(), "testpass123")

	// Setup
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "导出CSV测试"}))
	var lr struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &lr)
	ledgerID := lr.Data.ID

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID+"/export?format=csv", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("export csv: expected 200, got %d", w.Code)
	}
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/csv; charset=utf-8" {
		t.Errorf("expected text/csv, got %s", contentType)
	}
	if len(w.Body.Bytes()) == 0 {
		t.Error("expected non-empty body")
	}

	r.ServeHTTP(httptest.NewRecorder(), authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

func TestExportJSON(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "export_json_"+t.Name(), "testpass123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "导出JSON测试"}))
	var lr struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &lr)
	ledgerID := lr.Data.ID

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID+"/export?format=json", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("export json: expected 200, got %d", w.Code)
	}

	r.ServeHTTP(httptest.NewRecorder(), authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

func TestExport_UnsupportedFormat(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "export_bad_"+t.Name(), "testpass123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "导出格式测试"}))
	var lr struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &lr)
	ledgerID := lr.Data.ID

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID+"/export?format=xml", token, nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("unsupported format: expected 400, got %d", w.Code)
	}

	r.ServeHTTP(httptest.NewRecorder(), authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

// ---------- Tags ----------

func TestTagsList(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "tags_list_"+t.Name(), "testpass123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "标签测试"}))
	var lr struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &lr)
	ledgerID := lr.Data.ID

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/categories", token, map[string]interface{}{
		"name": "餐饮", "type": "expense", "ledger_id": ledgerID,
	}))
	var cr struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &cr)
	catID := cr.Data.ID

	today := time.Now().Format("2006-01-02")
	// Create transactions with tags
	for i, tag := range []string{"午餐,外卖", "午餐", "通勤,地铁"} {
		_ = i
		w = httptest.NewRecorder()
		r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/transactions", token, map[string]interface{}{
			"ledger_id": ledgerID, "category_id": catID,
			"type": "expense", "amount": 10.0, "transaction_date": today,
			"tags": []string{tag},
		}))
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID+"/tags", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("tags list: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data []string `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data) == 0 {
		t.Error("expected at least 1 tag")
	}

	r.ServeHTTP(httptest.NewRecorder(), authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

func TestTagsList_Empty(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "tags_empty_"+t.Name(), "testpass123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("POST", "/api/v1/ledgers", token, map[string]string{"name": "空标签测试"}))
	var lr struct{ Data struct{ ID string } `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &lr)
	ledgerID := lr.Data.ID

	w = httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+ledgerID+"/tags", token, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("tags empty: expected 200, got %d", w.Code)
	}

	var resp struct {
		Data []string `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data == nil {
		t.Error("expected data to be array (not nil)")
	}

	r.ServeHTTP(httptest.NewRecorder(), authenticatedRequest("DELETE", "/api/v1/ledgers/"+ledgerID, token, nil))
}

func TestTagsList_UnauthorizedLedger(t *testing.T) {
	r := testEngine(t)
	token := getToken(t, r, "tags_unauth_"+t.Name(), "testpass123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, authenticatedRequest("GET", "/api/v1/ledgers/"+uuid.New().String()+"/tags", token, nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("unauthorized ledger: expected 404, got %d", w.Code)
	}
}
