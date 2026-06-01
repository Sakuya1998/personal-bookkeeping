package middleware

import (
	"net/http"

	"personal-bookkeeping/internal/app/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LedgerAccess returns a middleware that checks if the authenticated user
// is a member of the ledger specified by the URL parameter "ledger_id".
// It injects "ledger_role" (string) and "ledger_member" (*models.LedgerMember) into the context.
func LedgerAccess(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ledgerIDStr := c.Param("ledger_id")
		if ledgerIDStr == "" {
			c.Next()
			return
		}

		ledgerID, err := uuid.Parse(ledgerIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid ledger_id"})
			return
		}

		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "user not authenticated"})
			return
		}

		var member models.LedgerMember
		if err := db.Where("ledger_id = ? AND user_id = ?", ledgerID, user.(*models.User).ID).
			First(&member).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "not a member of this ledger"})
			return
		}

		c.Set("ledger_role", member.Role)
		c.Set("ledger_member", &member)
		c.Next()
	}
}
