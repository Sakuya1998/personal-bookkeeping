package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"personal-bookkeeping/internal/app/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-jwt-secret-for-unit-tests"

// --- JWT Claims Unit Tests (no DB/cache needed) ---

func TestAuthClaims_GenerateAndParse(t *testing.T) {
	now := time.Now()
	userID := uuid.New().String()
	username := "testuser"

	claims := middleware.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	parsedClaims := &middleware.Claims{}
	parsedToken, err := jwt.ParseWithClaims(tokenStr, parsedClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse valid token: %v", err)
	}
	if !parsedToken.Valid {
		t.Fatal("expected token to be valid")
	}
	if parsedClaims.UserID != userID {
		t.Errorf("expected UserID %q, got %q", userID, parsedClaims.UserID)
	}
	if parsedClaims.Username != username {
		t.Errorf("expected Username %q, got %q", username, parsedClaims.Username)
	}
}

func TestAuthClaims_ExpiredToken(t *testing.T) {
	now := time.Now()
	claims := middleware.Claims{
		UserID:   uuid.New().String(),
		Username: "expireduser",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)), // 1 hour in the past
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	parsedClaims := &middleware.Claims{}
	_, err = jwt.ParseWithClaims(tokenStr, parsedClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestAuthClaims_TamperedToken(t *testing.T) {
	now := time.Now()
	claims := middleware.Claims{
		UserID:   uuid.New().String(),
		Username: "tampereduser",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Parse with a different secret (simulating tampered key)
	wrongSecret := "different-secret"
	parsedClaims := &middleware.Claims{}
	_, err = jwt.ParseWithClaims(tokenStr, parsedClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(wrongSecret), nil
	})
	if err == nil {
		t.Fatal("expected error for tampered token (wrong secret), got nil")
	}
}

func TestAuthClaims_NoneAlgorithm(t *testing.T) {
	// Security: token with "none" algorithm should be rejected
	claims := middleware.Claims{
		UserID:   uuid.New().String(),
		Username: "nonealgo",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	// Create unsigned token with "none" algorithm
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to create none-algo token: %v", err)
	}

	// Parse should fail when using HS256 secret
	parsedClaims := &middleware.Claims{}
	_, err = jwt.ParseWithClaims(tokenStr, parsedClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err == nil {
		t.Fatal("expected error for 'none' algorithm token, got nil — security risk!")
	}
}

func TestAuthClaims_EmptyClaims(t *testing.T) {
	// Edge case: token with empty claims should still be parseable but invalid for our use
	emptyClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, emptyClaims)
	tokenStr, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign empty claims: %v", err)
	}

	// Our middleware will parse it but UserID will be empty
	parsedClaims := &middleware.Claims{}
	parsedToken, err := jwt.ParseWithClaims(tokenStr, parsedClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("valid token with empty claims should parse: %v", err)
	}
	if !parsedToken.Valid {
		t.Fatal("expected token to be valid")
	}
	if parsedClaims.UserID != "" {
		t.Errorf("expected empty UserID, got %q", parsedClaims.UserID)
	}
}

// --- AuthRequired Handler Unit Tests (early-exit paths only, no DB/cache) ---

func setupTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", middleware.AuthRequired(testSecret), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestAuthRequired_MissingHeader(t *testing.T) {
	r := setupTestEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequired_InvalidFormat(t *testing.T) {
	r := setupTestEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz") // Not Bearer

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequired_EmptyToken(t *testing.T) {
	r := setupTestEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ") // Bearer with empty token

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
