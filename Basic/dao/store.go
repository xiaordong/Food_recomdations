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
