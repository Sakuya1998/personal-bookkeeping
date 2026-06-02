package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/infra/cache"
	"personal-bookkeeping/internal/infra/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
		if cacheInst := cache.GetDefault(); cacheInst != nil {
			exists, _ := cacheInst.Exists(c.Request.Context(), cache.KeyTokenBlacklist(claims.ID))
			if exists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "token revoked"})
				return
			}
		}

		// Get user — try cache first, fall back to DB
		user, err := getUserWithCache(c.Request.Context(), claims.UserID, claims.Username)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "user not found"})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// userCacheTTL is how long a user lookup stays cached (reduces N+1 DB queries).
const userCacheTTL = 5 * time.Minute

// getUserWithCache returns the user model, caching the result for userCacheTTL.
// On cache miss, queries DB and populates cache.
// Non-existent users are cached with a short TTL (cache.NullSentinel) to prevent
// cache penetration.
func getUserWithCache(ctx context.Context, userID, username string) (*models.User, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// Check cache
	if cacheInst := cache.GetDefault(); cacheInst != nil {
		if data, err := cacheInst.Get(ctx, "user:"+userID); err == nil {
			if data == cache.NullSentinel {
				// Cached miss — DB already confirmed this user doesn't exist
				return nil, gorm.ErrRecordNotFound
			}
			if data != "" {
				// Fast path: cache hit — build user from cached fields
				return &models.User{
					ID:       uid,
					Username: username,
					IsActive: true,
				}, nil
			}
		}
	}

	// Slow path: query DB
	var user models.User
	if err := database.GetDB().First(&user, "id = ?", uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Cache the miss to absorb repeated requests for non-existent users
			if cacheInst := cache.GetDefault(); cacheInst != nil {
				_ = cacheInst.Set(ctx, "user:"+userID, cache.NullSentinel, cache.NullCacheTTL)
			}
			return nil, err
		}
		return nil, err
	}

	// Populate cache
	if cacheInst := cache.GetDefault(); cacheInst != nil {
		_ = cacheInst.Set(ctx, "user:"+userID, "1", userCacheTTL)
	}

	return &user, nil
}
