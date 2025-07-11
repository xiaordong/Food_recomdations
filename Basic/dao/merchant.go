package dao

import (
	"Food_recommendation/Basic/model"
	"Food_recommendation/utils"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
)

func MerchantCreate(ctx context.Context, m model.Merchant) error {
	var count int64
	if err := DB.WithContext(ctx).Model(&model.Merchant{}).
		Where("merchant_name = ?", m.MerchantName).Count(&count).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if count > 0 {
		return errors.New("merchant_name already exists")
	}
	password, _ := utils.Crypto(m.Password)
	m.Password = password
	if err := DB.Create(&m).Error; err != nil {
		log.Println("Create Merchant Error!")
		return err
	}
	log.Println("Create Merchant Success!")
	return nil
}
func CheckLogin(ctx context.Context, m model.Merchant) (model.Merchant, error) {
	inputPass, _ := utils.Crypto(m.Password)
	if err := DB.WithContext(ctx).Where("merchant_name = ? AND password = ?", m.MerchantName, inputPass).First(&m).Error; err != nil {
		log.Println("Password is incorrect!")
		return m, errors.New("password is incorrect")
	}
	fmt.Println(m)
	return m, nil
}
func GetProfile(ctx context.Context, id string) (model.Merchant, error) {
	// 查询商家信息并预加载关联的店铺
	var merchant model.Merchant
	result := DB.WithContext(ctx).
		Preload("Stores"). // 预加载商家关联的所有店铺
		Where("id = ?", id).
		First(&merchant)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return model.Merchant{}, fmt.Errorf("merchant not found")
		}
		return model.Merchant{}, fmt.Errorf("database error: %w", result.Error)
	}

	return merchant, nil
}
func UpdateProfile(ctx context.Context, m model.Merchant) error {
	var existingMerchant model.Merchant
	if err := DB.WithContext(ctx).First(&existingMerchant, m.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("merchant not found")
		}
		return fmt.Errorf("database error: %w", err)
	}
	m.Version = existingMerchant.Version
	if m.MerchantName != existingMerchant.MerchantName {
		var count int64
		if err := DB.WithContext(ctx).Model(&model.Merchant{}).
			Where("merchant_name = ? AND id != ?", m.MerchantName, m.ID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("database error: %w", err)
		}
		if count > 0 {
			return errors.New("merchant name already exists")
		}
	}
	if m.Phone != existingMerchant.Phone {
		var count int64
		if err := DB.WithContext(ctx).Model(&model.Merchant{}).Where("phone = ? AND id != ?", m.Phone, m.ID).Count(&count).Error; err != nil {
			return fmt.Errorf("database error: %w", err)
		}
		if count > 0 {
			return errors.New("phone number already exists")
		}
	}
	updateFields := map[string]interface{}{}
	if m.MerchantName != "" {
		updateFields["merchant_name"] = m.MerchantName
	}
	if m.Phone != "" {
		updateFields["phone"] = m.Phone
	}

	updateFields["version"] = gorm.Expr("version + 1")
	result := DB.WithContext(ctx).Model(&model.Merchant{}).
		Where("id = ? AND version = ?", m.ID, m.Version).
		Updates(updateFields)

	if result.Error != nil {
		return fmt.Errorf("update failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("merchant information has been modified, please refresh and try again")
	}
	return nil
}
