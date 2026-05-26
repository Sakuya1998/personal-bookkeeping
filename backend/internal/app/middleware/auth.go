package middleware

import (
	"net/http"
	"strings"

	"personal-bookkeeping/internal/app/repository"
	models "personal-bookkeeping/internal/app/model"
	cch "personal-bookkeeping/internal/infra/cache"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "invalid authorization format"})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "invalid or expired token"})
			return
		}

		// Check token blacklist (logout revocation)
		if cache := database.GetCache(); cache != nil {
			exists, _ := cache.Exists(c.Request.Context(), cch.KeyTokenBlacklist(claims.ID))
			if exists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "token revoked"})
				return
			}
		}

		var user models.User
		if err := database.GetDB().First(&user, "id = ?", claims.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "user not found"})
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}
