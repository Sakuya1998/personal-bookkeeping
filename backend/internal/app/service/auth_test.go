package service

import (
	"context"
	"testing"
	"time"

	"personal-bookkeeping/internal/app/middleware"
	"personal-bookkeeping/internal/app/models"
	cch "personal-bookkeeping/internal/infra/cache"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ---------- generateToken ----------

func TestGenerateToken_ReturnsValidToken(t *testing.T) {
	s := &AuthService{
		Service:         &Service{},
		jwtSecret:       "test-secret",
		jwtExpireMinute: 60,
	}

	user := models.User{
		ID:       uuid.New(),
		Username: "testuser",
	}

	token, err := s.generateToken(user)
	if err != nil {
		t.Fatalf("generateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	// Verify the token can be parsed
	claims := &middleware.Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("token should be parseable: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("token should be valid")
	}
	if claims.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got %q", claims.Username)
	}
}

func TestGenerateToken_DifferentUsersGetDifferentTokens(t *testing.T) {
	s := &AuthService{
		Service:         &Service{},
		jwtSecret:       "test-secret",
		jwtExpireMinute: 60,
	}

	t1, _ := s.generateToken(models.User{ID: uuid.New(), Username: "user1"})
	t2, _ := s.generateToken(models.User{ID: uuid.New(), Username: "user2"})

	if t1 == t2 {
		t.Fatal("different users should get different tokens")
	}
}

func TestGenerateToken_WrongSecretFailsVerify(t *testing.T) {
	s := &AuthService{
		Service:         &Service{},
		jwtSecret:       "real-secret",
		jwtExpireMinute: 60,
	}

	token, _ := s.generateToken(models.User{ID: uuid.New(), Username: "test"})

	claims := &middleware.Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("wrong-secret"), nil
	})
	if err == nil {
		t.Fatal("should reject token signed with different secret")
	}
}

// ---------- BlacklistToken ----------

func TestBlacklistToken_WithCache(t *testing.T) {
	cache := newMockCache()
	s := &AuthService{
		Service: &Service{
			Cache: cache,
		},
	}

	jti := "test-jti"
	s.BlacklistToken(context.Background(), jti, time.Minute)

	// Verify the key was set using the actual key function
	key := cch.KeyTokenBlacklist(jti)
	val, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("expected cache entry for key %q, got error: %v", key, err)
	}
	if val != "1" {
		t.Fatalf(`expected "1", got %q`, val)
	}
}

func TestBlacklistToken_NilCache(t *testing.T) {
	s := &AuthService{
		Service: &Service{
			Cache: nil,
		},
	}

	// Should not panic
	s.BlacklistToken(context.Background(), "test-jti", time.Minute)
}

func TestBlacklistToken_UniqueKeys(t *testing.T) {
	cache := newMockCache()
	s := &AuthService{
		Service: &Service{
			Cache: cache,
		},
	}

	s.BlacklistToken(context.Background(), "jti-1", time.Minute)
	s.BlacklistToken(context.Background(), "jti-2", time.Minute)

	if _, err := cache.Get(context.Background(), cch.KeyTokenBlacklist("jti-1")); err != nil {
		t.Fatal("jti-1 should exist")
	}
	if _, err := cache.Get(context.Background(), cch.KeyTokenBlacklist("jti-2")); err != nil {
		t.Fatal("jti-2 should exist")
	}
}
