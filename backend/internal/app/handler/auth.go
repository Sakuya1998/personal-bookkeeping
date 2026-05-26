package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"personal-bookkeeping/internal/app/middleware"
	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"
	cch "personal-bookkeeping/internal/infra/cache"
	"personal-bookkeeping/internal/infra/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

type RegisterInput struct {
	Username string `json:"username" binding:"required,min=2,max=50" example:"alice"`
	Email    string `json:"email" binding:"required,email,max=100" example:"alice@example.com"`
	Password string `json:"password" binding:"required,min=6,max=100" example:"secret123"`
}

type LoginInput struct {
	Username string `json:"username" binding:"required" example:"alice"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

// Register  godoc
// @Summary      注册新用户
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body RegisterInput true "注册信息"
// @Success      201 {object} Response{data=authResponse}
// @Failure      400 {object} Response
// @Failure      409 {object} Response
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	var existing models.User
	if err := database.GetDB().Where("username = ? OR email = ?", input.Username, input.Email).First(&existing).Error; err == nil {
		Conflict(c, "username or email already exists")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		InternalError(c, "failed to hash password")
		return
	}

	user := models.User{
		ID:           uuid.New(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hash),
		IsActive:     true,
	}

	if err := database.GetDB().Create(&user).Error; err != nil {
		InternalError(c, "failed to create user")
		return
	}

	createDefaultLedgers(user)
	createDefaultCategories(user)

	token, err := h.generateToken(user)
	if err != nil {
		InternalError(c, "failed to generate token")
		return
	}

	RespondJSON(c, http.StatusCreated, gin.H{
		"token": token,
		"user":  toUserResponse(&user),
	})
}

// Login  godoc
// @Summary      用户登录
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body LoginInput true "登录信息"
// @Success      200 {object} Response{data=authResponse}
// @Failure      401 {object} Response
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	var user models.User
	if err := database.GetDB().Where("username = ?", input.Username).First(&user).Error; err != nil {
		Unauthorized(c, "invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		Unauthorized(c, "invalid username or password")
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		InternalError(c, "failed to generate token")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"token": token,
		"user":  toUserResponse(&user),
	})
}

// Me  godoc
// @Summary      当前用户信息
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} Response{data=userResponse}
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	RespondJSON(c, http.StatusOK, toUserResponse(user))
}

// Logout  godoc
// @Summary      登出（撤销当前 token）
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} Response
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	// Extract token jti from request
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		RespondJSON(c, http.StatusOK, gin.H{"message": "logged out"})
		return
	}

	claims := &middleware.Claims{}
	if _, _, err := (&jwt.Parser{}).ParseUnverified(parts[1], claims); err != nil || claims.ID == "" {
		// Can't extract jti — user still considered logged out
		RespondJSON(c, http.StatusOK, gin.H{"message": "logged out"})
		return
	}

	// Add token to blacklist (TTL = remaining token lifetime for safety)
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining <= 0 {
		remaining = time.Hour
	}
	if cache := database.GetCache(); cache != nil {
		_ = cache.Set(c.Request.Context(), cch.KeyTokenBlacklist(claims.ID), "1", remaining)
	}

	slog.Info("user logged out", "user_id", user.ID, "token_jti", claims.ID)
	RespondJSON(c, http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) generateToken(user models.User) (string, error) {
	claims := middleware.Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(h.cfg.JWT.ExpireMinute) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWT.Secret))
}

type authResponse struct {
	Token string       `json:"token" example:"eyJ..."`
	User  userResponse `json:"user"`
}

type userResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username  string `json:"username" example:"alice"`
	Email     string `json:"email" example:"alice@example.com"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

func toUserResponse(u *models.User) userResponse {
	return userResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Email:     u.Email,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

func createDefaultLedgers(user models.User) {
	defaultLedgers := []struct {
		name string
		desc string
		cur  string
	}{
		{"日常账本", "日常收支记录", "CNY"},
		{"投资账本", "投资理财记录", "USD"},
	}
	for _, l := range defaultLedgers {
		desc := l.desc
		database.GetDB().Create(&models.Ledger{
			ID:           uuid.New(),
			UserID:       user.ID,
			Name:         l.name,
			Description:  &desc,
			BaseCurrency: l.cur,
			SortOrder:    0,
		})
	}
}

func createDefaultCategories(user models.User) {
	defaultCategories := []struct {
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
	for _, cat := range defaultCategories {
		icon := cat.icon
		database.GetDB().Create(&models.Category{
			ID:       uuid.New(),
			UserID:   user.ID,
			Name:     cat.name,
			Type:     cat.typ,
			Icon:     &icon,
			IsActive: true,
		})
	}
}
