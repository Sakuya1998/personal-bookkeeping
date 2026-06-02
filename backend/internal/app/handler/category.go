package handler

import (
	"errors"
	"net/http"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
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
	Name     *string `json:"name" binding:"omitempty,max=50"`
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

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	roots, err := h.svc.ListCategories(lid, user.ID)
	if err != nil {
		InternalError(c, "failed to query categories")
		return
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

	var ledgerUUID *uuid.UUID
	if input.LedgerID != nil {
		parsed, err := uuid.Parse(*input.LedgerID)
		if err != nil {
			BadRequest(c, "invalid ledger_id format")
			return
		}
		ledgerUUID = &parsed
	}

	var parentUUID *uuid.UUID
	if input.ParentID != nil {
		parsed, err := uuid.Parse(*input.ParentID)
		if err != nil {
			BadRequest(c, "invalid parent_id format")
			return
		}
		parentUUID = &parsed
	}

	category, err := h.svc.CreateCategory(user.ID, input.Name, input.Type, input.Icon, input.Color, ledgerUUID, parentUUID)
	if err != nil {
		InternalError(c, "failed to create category")
		return
	}

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

	var input UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	cid, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, "invalid category id")
		return
	}

	var parentUUID *uuid.UUID
	if input.ParentID != nil {
		parsed, err := uuid.Parse(*input.ParentID)
		if err != nil {
			BadRequest(c, "invalid parent_id format")
			return
		}
		parentUUID = &parsed
	}

	category, err := h.svc.UpdateCategory(cid, user.ID, input.Name, input.Type, input.Icon, input.Color, parentUUID, input.IsActive)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			NotFound(c, "category not found")
			return
		}
		InternalError(c, "failed to update category")
		return
	}

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

	cid, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, "invalid category id")
		return
	}

	err = h.svc.DeleteCategory(cid, user.ID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			NotFound(c, "category not found")
			return
		}
		// Check for conflict errors (child categories or existing transactions)
		// The service layer returns ErrConflict wrapped with descriptive messages
		if errors.Is(err, service.ErrConflict) {
			Conflict(c, err.Error())
			return
		}
		InternalError(c, "failed to delete category")
		return
	}

	RespondJSON(c, http.StatusOK, nil)
}
