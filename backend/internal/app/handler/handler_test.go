package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"personal-bookkeeping/internal/app/repository"
	routes "personal-bookkeeping/internal/app/router"
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
	routes.Setup(r, cfg)
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
