package dao

import (
	"Food_recommendation/Basic/model"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

func CreateStore(ctx context.Context, store model.Store) error {
	// 开启事务
	tx := DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	//检查映射的商家是否存在
	var merchant model.Merchant
	if err := tx.Model(&model.Merchant{}).Where("id = ?", store.MerchantID).First(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return fmt.Errorf("merchant not found: %w", err)
		}
		tx.Rollback()
		return fmt.Errorf("database error: %w", err)
	}
	//店铺名是否存在
	var count int64
	if err := tx.Model(&model.Store{}).
		Where("merchant_id = ? AND name = ?", store.MerchantID, store.Name).
		Count(&count).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("check store name failed: %w", err)
	}
	if count > 0 {
		tx.Rollback()
		return errors.New("store name already exists under this merchant")
	}

	// 3. 创建店铺记录
	fmt.Println(store)
	if err := tx.Create(&store).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create store: %w", err)
	}

	//提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("transaction commit failed: %w", err)
	}
	return nil
}
func MyStore(ctx context.Context, merchantID uint) ([]model.Store, error) {
	var stores []model.Store
	result := DB.WithContext(ctx).Where(&model.Store{MerchantID: merchantID}).Find(&stores)

	if result.Error != nil {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return stores, nil
}
func SearchStore(ctx context.Context, sid uint, mid uint) (model.Store, error) {
	var store model.Store
	result := DB.WithContext(ctx).Where(&model.Store{ID: sid, MerchantID: mid}).First(&store)
	if result.Error != nil {
		return model.Store{}, fmt.Errorf("database error: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.Store{}, gorm.ErrRecordNotFound
	}
	return store, nil
}
func UpdateStore(ctx context.Context, store model.Store) error {
	if store.ID == 0 || store.MerchantID == 0 {
		return errors.New("store ID and merchant ID are required")
	}

	var originalStore model.Store
	result := DB.WithContext(ctx).Where("id = ? AND merchant_id = ?", store.ID, store.MerchantID).First(&originalStore)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("store not found or unauthorized")
		}
		return fmt.Errorf("database query failed: %w", result.Error)
	}
	updateFields := make(map[string]interface{})
	if store.Description != originalStore.Description && store.Description != "" {
		updateFields["description"] = store.Description
	}
	if store.Active != originalStore.Active {
		updateFields["active"] = store.Active
	}
	if len(updateFields) == 0 {
		return nil
	}

	updateFields["version"] = gorm.Expr("version + 1")
	result = DB.WithContext(ctx).Model(&model.Store{}).Where("id = ? AND version = ?", store.ID, originalStore.Version).Updates(updateFields)
	if result.Error != nil {
		return fmt.Errorf("update failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("store information has been modified, please refresh and try again") // 版本冲突
	}
	return nil
}
func DeleteStore(ctx context.Context, sid uint, mid uint) error {
	tx := DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction failed: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 1. 校验
	var store model.Store
	result := tx.Where("id = ? AND merchant_id = ?", sid, mid).First(&store)
	if result.Error != nil {
		tx.Rollback()
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("store not found or unauthorized")
		}
		return fmt.Errorf("query store failed: %w", result.Error)
	}
	// 2. 检查店铺是否有关联数据
	var dishCount int64
	tx.Model(&model.Dishes{}).Where("store_id = ?", sid).Count(&dishCount)
	if dishCount > 0 {
		tx.Rollback()
		return errors.New("cannot delete store with associated dishes")
	}
	// 3. 执行删除操作（物理删除，因为前提是店铺无菜品，软删除意义不大反而浪费）
	result = tx.Unscoped().Delete(&store)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("delete store failed: %w", result.Error)
	}
	// 4. 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}
	return nil
}
func Check(ctx context.Context, SID uint, MID uint) bool {
	if SID == 0 || MID == 0 {
		return false // 非法参数直接返回false
	}
	var count int64
	err := DB.WithContext(ctx).Model(&model.Store{}).Where("id = ? AND merchant_id = ?", SID, MID).Count(&count).Error
	return err == nil && count > 0 // 无错误且存在记录时返回true
}
