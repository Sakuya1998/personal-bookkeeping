package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"personal-bookkeeping/internal/app/middleware"
	"personal-bookkeeping/internal/app/models"
	cch "personal-bookkeeping/internal/infra/cache"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ---------- sentinel errors ----------

var (
	ErrUserExists     = errors.New("username or email already exists")
	ErrInvalidCreds   = errors.New("invalid username or password")
	ErrEmailInUse     = errors.New("email already in use by another user")
	ErrTokenGenFailed = errors.New("failed to generate token")
)

// ---------- AuthService ----------

type AuthService struct {
	*Service
	jwtSecret       string
	jwtExpireMinute int
}

func NewAuthService(s *Service, jwtSecret string, jwtExpireMinute int) *AuthService {
	return &AuthService{
		Service:         s,
		jwtSecret:       jwtSecret,
		jwtExpireMinute: jwtExpireMinute,
	}
}

// Register creates a new user with default ledgers/categories and returns a JWT token.
func (s *AuthService) Register(username, email, password string) (*models.User, string, error) {
	// Check existing
	var existing models.User
	if err := s.DB.Where("username = ? OR email = ?", username, email).First(&existing).Error; err == nil {
		return nil, "", ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	user := models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		IsActive:     true,
	}

	if err := s.DB.Create(&user).Error; err != nil {
		return nil, "", fmt.Errorf("create user: %w", err)
	}

	s.createDefaultLedgers(user.ID)
	s.createDefaultCategories(user.ID)

	token, err := s.generateToken(user)
	if err != nil {
		return &user, "", ErrTokenGenFailed
	}

	return &user, token, nil
}

// Login verifies credentials and returns a JWT token.
func (s *AuthService) Login(username, password string) (*models.User, string, error) {
	var user models.User
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, "", ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCreds
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", ErrTokenGenFailed
	}

	return &user, token, nil
}

// ChangePassword validates old password and updates to new password.
func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	var user models.User
	if err := s.DB.First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCreds
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.DB.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", string(hash)).Error; err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

// ChangeEmail updates user email after checking uniqueness.
func (s *AuthService) ChangeEmail(userID uuid.UUID, newEmail string) error {
	var existing models.User
	if err := s.DB.Where("email = ? AND id <> ?", newEmail, userID).First(&existing).Error; err == nil {
		return ErrEmailInUse
	}

	if err := s.DB.Model(&models.User{}).Where("id = ?", userID).Update("email", newEmail).Error; err != nil {
		return fmt.Errorf("update email: %w", err)
	}

	return nil
}

// ---------- internal helpers ----------

func (s *AuthService) generateToken(user models.User) (string, error) {
	claims := middleware.Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpireMinute) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) createDefaultLedgers(userID uuid.UUID) {
	defaults := []struct {
		name string
		desc string
		cur  string
	}{
		{"日常账本", "日常收支记录", "CNY"},
		{"投资账本", "投资理财记录", "USD"},
	}
	for _, l := range defaults {
		desc := l.desc
		s.DB.Create(&models.Ledger{
			ID:           uuid.New(),
			UserID:       userID,
			Name:         l.name,
			Description:  &desc,
			BaseCurrency: l.cur,
			SortOrder:    0,
		})
	}
}

func (s *AuthService) createDefaultCategories(userID uuid.UUID) {
	defaults := []struct {
		name string
		typ  string
		icon string
	}{
		{"餐饮", "expense", "🍽️"}, {"交通", "expense", "🚗"},
		{"购物", "expense", "🛒"}, {"居住", "expense", "🏠"},
		{"娱乐", "expense", "🎮"}, {"通讯", "expense", "📱"},
		{"医疗", "expense", "💊"}, {"教育", "expense", "📚"},
		{"工资", "income", "💰"}, {"奖金", "income", "🎁"},
		{"投资", "income", "📈"}, {"其他", "income", "📋"},
		{"其他", "expense", "📋"},
	}
	for _, cat := range defaults {
		icon := cat.icon
		s.DB.Create(&models.Category{
			ID:       uuid.New(),
			UserID:   userID,
			Name:     cat.name,
			Type:     cat.typ,
			Icon:     &icon,
			IsActive: true,
		})
	}
}

// BlacklistToken stores a JWT jti in the cache with the given TTL for logout revocation.
func (s *AuthService) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) {
	if s.Cache == nil {
		return
	}
	_ = s.Cache.Set(ctx, cch.KeyTokenBlacklist(jti), "1", ttl)
}
