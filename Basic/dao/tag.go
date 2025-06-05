package dao

import (
	"Food_recommendation/Basic/model"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

func AddTags(ctx context.Context, tags []string) error {
	if len(tags) == 0 {
		return errors.New("tags list is empty")
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
	tagModels := make([]model.Tag, 0, len(tags))
	for _, name := range tags {
		name = strings.TrimSpace(name)
		if name == "" || len(name) > 12 {
			continue
		}
		tagModels = append(tagModels, model.Tag{
			Name: name,
		})
	}
	result := tx.Model(&model.Tag{}).CreateInBatches(tagModels, 100)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("insert tags failed: %w", result.Error)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("commit transaction failed: %w", err)
	}
	return nil
}
func GetTags(ctx context.Context) ([]string, error) {
	var tags []model.Tag
	result := DB.WithContext(ctx).
		Model(&model.Tag{}).
		Limit(100).      // 限制最多返回100条记录
		Order("id ASC"). // 按ID升序排列（可选）
		Find(&tags)

	if result.Error != nil {
		return nil, fmt.Errorf("query tags failed: %w", result.Error)
	}
	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	return tagNames, nil
}
func ChooseTag(ctx context.Context, dishID uint, tags []string) error {
	if dishID == 0 {
		return errors.New("dish ID is required")
	}
	if len(tags) > 3 {
		return errors.New("maximum 3 tags allowed")
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

	var tagIDs []uint
	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" || len(tagName) > 12 {
			continue // 跳过空标签或超长标签
		}

		var tag model.Tag
		if err := tx.Where("name = ?", tagName).First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newTag := model.Tag{
					Name: tagName,
				}
				if err := tx.Create(&newTag).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("create tag failed: %w", err)
				}
				tagIDs = append(tagIDs, newTag.ID)
			} else {
				tx.Rollback()
				return fmt.Errorf("query tag failed: %w", err)
			}
		} else {
			// 标签已存在，使用现有ID
			tagIDs = append(tagIDs, tag.ID)
		}
	}

	// 清除菜品原有的所有标签关系
	if err := tx.Model(&model.Dishes{}).Association("Tags").Clear(); err != nil {
		tx.Rollback()
		return fmt.Errorf("clear dish tags failed: %w", err)
	}

	// 为菜品设置新的标签关系
	if len(tagIDs) > 0 {
		var dish model.Dishes
		if err := tx.First(&dish, dishID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("find dish failed: %w", err)
		}

		var tags []model.Tag
		if err := tx.Find(&tags, tagIDs).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("find tags failed: %w", err)
		}

		if err := tx.Model(&dish).Association("Tags").Append(tags); err != nil {
			tx.Rollback()
			return fmt.Errorf("append tags to dish failed: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}
