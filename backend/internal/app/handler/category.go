package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"
	cch "personal-bookkeeping/internal/infra/cache"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryHandler struct{}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

type CreateCategoryInput struct {
	LedgerID *string `json:"ledger_id"`
	Name     string  `json:"name" binding:"required,max=50" example:"餐饮"`
	Type     string  `json:"type" binding:"required,oneof=income expense" example:"expense"`
	Icon     *string `json:"icon" example:"🍽️"`
	Color    *string `json:"color"`
	ParentID *string `json:"parent_id"`
}

type UpdateCategoryInput struct {
	Name     *string `json:"name"`
	Type     *string `json:"type" binding:"omitempty,oneof=income expense"`
	Icon     *string `json:"icon"`
	Color    *string `json:"color"`
	ParentID *string `json:"parent_id"`
	IsActive *bool   `json:"is_active"`
}

// List  godoc
// @Summary      分类列表（树形）
// @Tags         categories
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Success      200 {object} Response
// @Router       /ledgers/{ledger_id}/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")

	// Try cache
	cchInstance := database.GetCache()
	if cchInstance != nil {
		key := cch.KeyCategoryList(user.ID.String())
		if cached, err := cchInstance.Get(c.Request.Context(), key); err == nil {
			var roots []models.Category
			if json.Unmarshal([]byte(cached), &roots) == nil {
				RespondJSON(c, http.StatusOK, roots)
				return
			}
		}
	}

	var categories []models.Category
	database.GetDB().Where("user_id = ? AND (ledger_id IS NULL OR ledger_id = ?)", user.ID, ledgerID).
		Order("sort_order asc, name asc").Find(&categories)

	categoryMap := make(map[uuid.UUID]*models.Category)
	for i := range categories {
		categories[i].Children = []models.Category{}
		categoryMap[categories[i].ID] = &categories[i]
	}
	var roots []models.Category
	for i := range categories {
		if categories[i].ParentID != nil {
			if parent, ok := categoryMap[*categories[i].ParentID]; ok {
				parent.Children = append(parent.Children, categories[i])
			}
		} else {
			roots = append(roots, categories[i])
		}
	}

	// Set cache
	if cchInstance != nil {
		key := cch.KeyCategoryList(user.ID.String())
		if data, err := json.Marshal(roots); err == nil {
			_ = cchInstance.Set(c.Request.Context(), key, string(data), 10*time.Minute)
		}
	}

	RespondJSON(c, http.StatusOK, roots)
}

// Create  godoc
// @Summary      创建分类
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body CreateCategoryInput true "分类信息"
// @Success      201 {object} Response
// @Router       /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	var input CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	category := models.Category{
		ID:       uuid.New(),
		UserID:   user.ID,
		Name:     input.Name,
		Type:     input.Type,
		Icon:     input.Icon,
		Color:    input.Color,
		IsActive: true,
	}

	if input.LedgerID != nil {
		if parsed, err := uuid.Parse(*input.LedgerID); err == nil {
			category.LedgerID = &parsed
		}
	}
	if input.ParentID != nil {
		if parsed, err := uuid.Parse(*input.ParentID); err == nil {
			category.ParentID = &parsed
		}
	}

	if err := database.GetDB().Create(&category).Error; err != nil {
		InternalError(c, "failed to create category")
		return
	}

	invalidateCategoryCache(user.ID)
	RespondJSON(c, http.StatusCreated, category)
}

// Update  godoc
// @Summary      更新分类
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path string               true "分类 ID"
// @Param        input body UpdateCategoryInput   true "更新内容"
// @Success      200 {object} Response
// @Router       /categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("id")

	var category models.Category
	if err := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "category not found")
			return
		}
		InternalError(c, "database error")
		return
	}

	var input UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Type != nil {
		updates["type"] = *input.Type
	}
	if input.Icon != nil {
		updates["icon"] = *input.Icon
	}
	if input.Color != nil {
		updates["color"] = *input.Color
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.ParentID != nil {
		if parsed, err := uuid.Parse(*input.ParentID); err == nil {
			updates["parent_id"] = parsed
		}
	}

	database.GetDB().Model(&category).Updates(updates)
	database.GetDB().First(&category, category.ID)

	invalidateCategoryCache(user.ID)

	RespondJSON(c, http.StatusOK, category)
}

// Delete  godoc
// @Summary      删除分类
// @Tags         categories
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "分类 ID"
// @Success      200 {object} Response
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("id")

	var count int64
	database.GetDB().Model(&models.Transaction{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		Conflict(c, "cannot delete category with existing transactions, deactivate it instead")
		return
	}

	result := database.GetDB().Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.Category{})
	if result.RowsAffected == 0 {
		NotFound(c, "category not found")
		return
	}

	invalidateCategoryCache(user.ID)
	RespondJSON(c, http.StatusOK, nil)
}

func invalidateCategoryCache(userID uuid.UUID) {
	c := database.GetCache()
	if c == nil {
		return
	}
	_ = c.Delete(context.Background(), cch.KeyCategoryList(userID.String()))
}
