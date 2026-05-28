package handler

import (
	"net/http"
	"strings"
	"time"

	"personal-bookkeeping/internal/app/middleware"
	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type RegisterInput struct {
	Username string `json:"username" binding:"required,min=2,max=50" example:"alice"`
	Email    string `json:"email" binding:"required,email,max=100" example:"alice@example.com"`
	Password string `json:"password" binding:"required,min=8,max=100" example:"secret123"`
}

type LoginInput struct {
	Username string `json:"username" binding:"required" example:"alice"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required,min=8,max=100"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=100"`
}

type ChangeEmailInput struct {
	Email string `json:"email" binding:"required,email,max=100"`
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

	user, token, err := h.svc.Register(input.Username, input.Email, input.Password)
	if err != nil {
		if err == service.ErrUserExists {
			Conflict(c, "username or email already exists")
			return
		}
		InternalError(c, "failed to register")
		return
	}

	RespondJSON(c, http.StatusCreated, gin.H{
		"token": token,
		"user":  toUserResponse(user),
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

	user, token, err := h.svc.Login(input.Username, input.Password)
	if err != nil {
		if err == service.ErrInvalidCreds {
			Unauthorized(c, "invalid username or password")
			return
		}
		InternalError(c, "failed to login")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"token": token,
		"user":  toUserResponse(user),
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

	// Add token to blacklist (TTL = remaining token lifetime)
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining <= 0 {
		remaining = time.Hour
	}
	h.svc.BlacklistToken(c.Request.Context(), claims.ID, remaining)

	RespondJSON(c, http.StatusOK, gin.H{"message": "logged out"})
}

// ChangePassword  godoc
// @Summary      修改密码
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body ChangePasswordInput true "新旧密码"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      401 {object} Response
// @Router       /auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if err := h.svc.ChangePassword(user.ID, input.OldPassword, input.NewPassword); err != nil {
		if err == service.ErrInvalidCreds {
			Unauthorized(c, "invalid old password")
			return
		}
		InternalError(c, "failed to change password")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{"message": "password updated"})
}

// ChangeEmail  godoc
// @Summary      修改邮箱
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body ChangeEmailInput true "新邮箱"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      409 {object} Response
// @Router       /auth/email [put]
func (h *AuthHandler) ChangeEmail(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var input ChangeEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if err := h.svc.ChangeEmail(user.ID, input.Email); err != nil {
		if err == service.ErrEmailInUse {
			Conflict(c, "email already in use by another user")
			return
		}
		InternalError(c, "failed to change email")
		return
	}

	RespondJSON(c, http.StatusOK, toUserResponse(&models.User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     input.Email,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}))
}

// ---------- response types ----------

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
