package dao

import (
	"Food_recommendation/Basic/model"
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strings"
	"time"
)

func CreateDishes(ctx context.Context, dish model.Dishes) error {
	tx := DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if err := tx.Commit().Error; err != nil {
			tx.Rollback()
		}
	}()

	// 保存原始标签并临时移除关联
	originalTags := dish.Tags
	dish.Tags = nil

	// 创建菜品
	if err := tx.Create(&dish).Error; err != nil {
		return fmt.Errorf("创建菜品失败: %w", err)
	}

	// 更新店铺时间戳
	if err := tx.Model(&model.Store{}).Where("id = ?", dish.StoreID).
		Update("updated_at", time.Now()).Error; err != nil {
		return fmt.Errorf("更新店铺时间戳失败: %w", err)
	}

	// 处理标签关联
	for _, tag := range originalTags {
		if strings.TrimSpace(tag.Name) == "" {
			continue
		}

		// 查询或创建标签
		var t model.Tag
		if err := tx.Where("name = ?", tag.Name).First(&t).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				t = model.Tag{Name: tag.Name}
				if err := tx.Create(&t).Error; err != nil {
					return fmt.Errorf("创建标签失败: %w", err)
				}
			} else {
				return fmt.Errorf("查询标签失败: %w", err)
			}
		}

		if err := tx.Exec("INSERT INTO dishes_tags (dishes_id, tag_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE dishes_id=dishes_id",
			dish.ID, t.ID).Error; err != nil {
			return fmt.Errorf("关联标签失败: %w", err)
		}
	}

	return nil
}

func GetDishes(ctx context.Context, SID uint, MID uint) ([]model.Dishes, error) {
	var dishes []model.Dishes
	if SID == 0 {
		return dishes, errors.New("store ID is required")
	}
	query := DB.WithContext(ctx).
		Select("id", "store_id", "name", "price", "desc", "image_url", "available"). // 指定查询字段
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

	// 预加载标签关联数据
	result := DB.WithContext(ctx).
		Select("dishes.id", "dishes.store_id", "dishes.name", "dishes.price", "dishes.desc", "dishes.image_url", "dishes.available").
		Preload("Tags", "name != ''"). // 预加载非空标签
		Where("dishes.id = ? AND dishes.store_id = ?", DID, SID).
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
