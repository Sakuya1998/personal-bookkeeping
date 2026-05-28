package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/infra/cache"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ListCategories 获取分类树形列表，带缓存
func (s *CategoryService) ListCategories(ledgerID, userID uuid.UUID) ([]models.Category, error) {
	// Try cache
	if s.Cache != nil {
		key := cache.KeyCategoryList(userID.String())
		if cached, err := s.Cache.Get(context.Background(), key); err == nil {
			var roots []models.Category
			if json.Unmarshal([]byte(cached), &roots) == nil {
				return roots, nil
			}
		}
	}

	var categories []models.Category
	if err := s.DB.Where("user_id = ? AND (ledger_id IS NULL OR ledger_id = ?)", userID, ledgerID).
		Order("sort_order asc, name asc").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}

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
	if s.Cache != nil {
		key := cache.KeyCategoryList(userID.String())
		if data, err := json.Marshal(roots); err == nil {
			_ = s.Cache.Set(context.Background(), key, string(data), 10*time.Minute)
		}
	}

	return roots, nil
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(userID uuid.UUID, name string, categoryType string, icon, color *string, ledgerID, parentID *uuid.UUID) (*models.Category, error) {
	category := models.Category{
		ID:       uuid.New(),
		UserID:   userID,
		Name:     name,
		Type:     categoryType,
		Icon:     icon,
		Color:    color,
		LedgerID: ledgerID,
		ParentID: parentID,
		IsActive: true,
	}

	if err := s.DB.Create(&category).Error; err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	s.invalidateCache(userID)
	return &category, nil
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(id, userID uuid.UUID, name, categoryType, icon, color *string, parentID *uuid.UUID, isActive *bool) (*models.Category, error) {
	var category models.Category
	if err := s.DB.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query category: %w", err)
	}

	updates := map[string]interface{}{}
	if name != nil {
		updates["name"] = *name
	}
	if categoryType != nil {
		updates["type"] = *categoryType
	}
	if icon != nil {
		updates["icon"] = *icon
	}
	if color != nil {
		updates["color"] = *color
	}
	if isActive != nil {
		updates["is_active"] = *isActive
	}
	if parentID != nil {
		updates["parent_id"] = *parentID
	}

	if err := s.DB.Model(&category).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	if err := s.DB.First(&category, category.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload category: %w", err)
	}

	s.invalidateCache(userID)
	return &category, nil
}

// DeleteCategory 删除分类，带子分类和交易验证
func (s *CategoryService) DeleteCategory(id, userID uuid.UUID) error {
	// Check child categories
	var childCount int64
	if err := s.DB.Model(&models.Category{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		return fmt.Errorf("failed to check child categories: %w", err)
	}
	if childCount > 0 {
		return fmt.Errorf("%w: cannot delete category with child categories, remove children first", ErrConflict)
	}

	// Check transaction references
	var txnCount int64
	if err := s.DB.Model(&models.Transaction{}).Where("category_id = ?", id).Count(&txnCount).Error; err != nil {
		return fmt.Errorf("failed to check transaction count: %w", err)
	}
	if txnCount > 0 {
		return fmt.Errorf("%w: cannot delete category with existing transactions, deactivate it instead", ErrConflict)
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Clean up recurring rules and budgets referencing this category
		if err := tx.Where("category_id = ?", id).Delete(&models.RecurringRule{}).Error; err != nil {
			return err
		}
		if err := tx.Where("category_id = ?", id).Delete(&models.Budget{}).Error; err != nil {
			return err
		}

		result := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Category{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.invalidateCache(userID)
	return nil
}

// invalidateCache 清除分类缓存
func (s *CategoryService) invalidateCache(userID uuid.UUID) {
	if s.Cache == nil {
		return
	}
	_ = s.Cache.Delete(context.Background(), cache.KeyCategoryList(userID.String()))
}
