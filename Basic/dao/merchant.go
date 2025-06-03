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
func CheckLogin(ctx context.Context, m model.Merchant) error {
	inputPass, _ := utils.Crypto(m.Password)
	if err := DB.WithContext(ctx).Where("merchant_name = ? AND password = ?", m.MerchantName, inputPass).First(&m).Error; err != nil {
		log.Println("Password is incorrect!")
		return errors.New("password is incorrect")
	}
	return nil
}
func GetProfile(ctx context.Context, merchantName string) (model.Merchant, error) {
	var merchant model.Merchant
	result := DB.WithContext(ctx).Where("merchant_name = ?", merchantName).First(&merchant)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return model.Merchant{}, fmt.Errorf("merchant not found")
		}
		return model.Merchant{}, fmt.Errorf("database error: %w", result.Error)
	}
	return merchant, nil
}
