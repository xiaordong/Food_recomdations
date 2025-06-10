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

func LikeDish(ctx context.Context, userID, dishID uint, isLike bool) error {
	tx := DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			return
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
		}
	}()

	// 定义点赞记录
	like := model.Like{
		UserID: userID,
		DishID: dishID,
	}

	var err error
	if isLike {
		// 检查是否已存在点赞记录
		result := tx.Model(&model.Like{}).Where("user_id = ? AND dish_id = ?", userID, dishID).First(&like)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// 不存在记录，创建新点赞
				if err = tx.Create(&like).Error; err != nil {
					return fmt.Errorf("create like failed: %w", err)
				}
			} else {
				// 其他查询错误
				return fmt.Errorf("query like failed: %w", result.Error)
			}
		} else {
			// 记录已存在，不重复创建
			return nil
		}
	} else {
		// 取消点赞：删除已存在的点赞记录
		result := tx.Model(&model.Like{}).Where("user_id = ? AND dish_id = ?", userID, dishID).Delete(&like)
		if result.Error != nil {
			return fmt.Errorf("delete like failed: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			// 无记录可删除，视为成功
			return nil
		}
	}

	// 原子更新菜品点赞数，防止负数
	var delta int
	if isLike {
		delta = 1
	} else {
		delta = -1
	}

	// 检查菜品是否存在
	var dish model.Dishes
	if err := tx.First(&dish, dishID).Error; err != nil {
		return fmt.Errorf("dish not found: %w", err)
	}

	// 更新点赞数，确保不会小于0
	updateResult := tx.Model(&model.Dishes{}).
		Where("id = ? AND like_num >= ?", dishID, -delta).
		Update("like_num", gorm.Expr("like_num + ?", delta))

	if updateResult.Error != nil {
		return fmt.Errorf("update like num failed: %w", updateResult.Error)
	}
	if updateResult.RowsAffected == 0 {
		// 菜品不存在或点赞数会变为负数
		return fmt.Errorf("update like num failed: dishID=%d", dishID)
	}

	return nil
}

func RateDish(ctx context.Context, userID, dishID uint, score uint, commit string) error {
	tx := DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if err := tx.Commit().Error; err != nil {
			tx.Rollback()
		}
	}()

	// 1. 处理用户评分记录
	var rating model.Rating
	if err := tx.Where("user_id = ? AND dish_id = ?", userID, dishID).First(&rating).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("query rating failed: %w", err)
		}
		// 新增评分记录
		rating = model.Rating{UserID: userID, DishID: dishID, Num: score, Comment: commit}
		if err := tx.Create(&rating).Error; err != nil {
			return fmt.Errorf("create rating failed: %w", err)
		}
	} else {
		// 更新已有评分记录
		if err := tx.Model(&rating).Update("num", score).Error; err != nil {
			return fmt.Errorf("update rating failed: %w", err)
		}
	}

	// 2. 重新计算菜品评分（使用 SQL 聚合函数）
	var result struct {
		Sum uint `gorm:"column:rating_sum"`
		Num uint `gorm:"column:rating_num"`
	}
	if err := tx.Model(&model.Rating{}).
		Where("dish_id = ?", dishID).
		Select("SUM(num) AS rating_sum, COUNT(*) AS rating_num").
		Scan(&result).Error; err != nil {
		return fmt.Errorf("calculate rating failed: %w", err)
	}
	fmt.Printf("Calculated rating sum: %d, count: %d\n", result.Sum, result.Num)

	// 3. 更新菜品评分字段
	var avgRating float64
	if result.Num > 0 {
		avgRating = float64(result.Sum) / float64(result.Num)
	} else {
		avgRating = 0 // 默认评分
	}
	if err := tx.Model(&model.Dishes{}).
		Where("id = ?", dishID).
		Updates(map[string]interface{}{
			"rating_sum": result.Sum,
			"rating_num": result.Num,
			"avg_rating": avgRating,
		}).Error; err != nil {
		return fmt.Errorf("update dish rating failed: %w", err)
	}

	// 4. 更新店铺评分
	// 获取菜品所属店铺ID
	var dish model.Dishes
	if err := tx.First(&dish, dishID).Error; err != nil {
		return fmt.Errorf("dish not found: %w", err)
	}

	// 计算店铺的总评分和评论数
	var storeResult struct {
		TotalRatingSum uint `gorm:"column:total_rating_sum"`
		TotalRatingNum uint `gorm:"column:total_rating_num"`
	}
	if err := tx.Model(&model.Dishes{}).
		Where("store_id = ?", dish.StoreID).
		Select("SUM(rating_sum) AS total_rating_sum, SUM(rating_num) AS total_rating_num").
		Scan(&storeResult).Error; err != nil {
		return fmt.Errorf("calculate store rating failed: %w", err)
	}

	// 计算店铺平均评分
	var storeAvgRating float64
	if storeResult.TotalRatingNum > 0 {
		storeAvgRating = float64(storeResult.TotalRatingSum) / float64(storeResult.TotalRatingNum)
	} else {
		storeAvgRating = 0 // 没有评分时显示0分
	}

	// 更新店铺评分
	if err := tx.Model(&model.Store{}).
		Where("id = ?", dish.StoreID).
		Update("avg_rating", storeAvgRating).Error; err != nil {
		return fmt.Errorf("update store rating failed: %w", err)
	}

	return nil
}

func GetUserHistory(ctx context.Context, userID uint) ([]model.History, error) {
	var history []model.History
	result := DB.WithContext(ctx).Where("user_id = ?", userID).Find(&history)
	if result.Error != nil {
		return nil, result.Error
	}
	return history, nil
}

// GetAllDishes 获取所有菜品
func GetAllDishes(ctx context.Context) ([]model.Dishes, error) {
	var dishes []model.Dishes
	result := DB.WithContext(ctx).Find(&dishes)
	if result.Error != nil {
		return nil, result.Error
	}
	return dishes, nil
}
