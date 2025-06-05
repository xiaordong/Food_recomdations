package dao

import (
	"Food_recommendation/Basic/model"
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

func CreateDishes(ctx context.Context, d model.Dishes) error {
	tx := DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Create(&d).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create dishes: %w", err)
	}
	if err := tx.Model(&model.Store{}).Where("id = ?", d.StoreID).
		Update("updated_at", time.Now()).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update store timestamp: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("transaction commit failed: %w", err)
	}
	return nil
}
func GetDishes(ctx context.Context, SID uint, MID uint) ([]model.Dishes, error) {
	var dishes []model.Dishes
	if SID == 0 {
		return dishes, errors.New("store ID is required")
	}
	query := DB.WithContext(ctx).
		Select("id", "store_id", "name", "price", "desc", "image_url", "available", "version"). // 指定查询字段
		Where("store_id = ?", SID)
	if MID == 0 {
		query = query.Where("available = true")
	}
	result := query.Find(&dishes)
	if result.Error != nil {
		return dishes, fmt.Errorf("database query failed: %w", result.Error)
	}
	return dishes, nil
}
func GetADishes(ctx context.Context, SID uint, DID uint) (model.Dishes, error) {
	var dish model.Dishes

	// 参数校验
	if SID == 0 {
		return dish, errors.New("store ID is required")
	}
	if DID == 0 {
		return dish, errors.New("dish ID is required")
	}

	// 构建查询（仅选择指定字段）
	result := DB.WithContext(ctx).
		Select("id", "store_id", "name", "price", "desc", "image_url", "available", "version"). // 指定数据库列名
		Where("id = ? AND store_id = ?", DID, SID).
		First(&dish)

	// 错误处理
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return dish, errors.New("dish not found")
		}
		return dish, fmt.Errorf("database query failed: %w", result.Error)
	}

	return dish, nil
}
func UpdateDishes(ctx context.Context, d model.Dishes) error {
	if d.ID == 0 {
		return errors.New("dish ID is required")
	}
	var originalDish model.Dishes
	if err := DB.WithContext(ctx).Where("id = ?", d.ID).First(&originalDish).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("dish not found")
		}
		return fmt.Errorf("query original dish failed: %w", err)
	}

	updates := map[string]interface{}{
		"name":      d.Name,
		"price":     d.Price,
		"desc":      d.Desc,
		"image_url": d.ImageURL,
		"available": d.Available,
		"version":   gorm.Expr("version + 1"), // 版本号自增
	}

	for key, value := range updates {
		switch v := value.(type) {
		case string:
			if v == "" {
				delete(updates, key) // 删除空字符串字段
			}
		case decimal.Decimal:
			if v.IsZero() {
				delete(updates, key) // 删除decimal零值字段
			}
		}
	}
	if len(updates) == 0 {
		return nil
	}
	result := DB.WithContext(ctx).Model(&model.Dishes{}).
		Where("id = ? AND version = ?", d.ID, originalDish.Version).
		Select( // 明确指定允许更新的字段（防止恶意字段注入）
			"name", "price", "desc", "image_url", "available", "version",
		).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("concurrent update conflict")
	}
	return nil
}
func DeleteDishes(ctx context.Context, SID uint, DID uint) error {
	if SID == 0 {
		return errors.New("store ID is required")
	}
	if DID == 0 {
		return errors.New("dish ID is required")
	}
	tx := DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction failed: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 检查菜品是否存在且属于指定店铺
	var dish model.Dishes
	if err := tx.Where("id = ? AND store_id = ?", DID, SID).First(&dish).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("dish not found or unauthorized")
		}
		return fmt.Errorf("query dish failed: %w", err)
	}
	result := tx.Unscoped().Delete(&dish)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("delete dish failed: %w", result.Error)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("commit transaction failed: %w", err)
	}
	return nil
}
