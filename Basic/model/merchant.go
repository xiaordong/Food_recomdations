package model

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"log"
	"regexp"
	"strings"
	"time"
)

var node *snowflake.Node

func init() {
	var err error
	node, err = snowflake.NewNode(1) // nodeID 设为 1（需唯一）
	if err != nil {
		log.Fatalf("Failed to create snowflake node: %v", err)
	}
}

func GenID() snowflake.ID {
	return node.Generate()
}

type Merchant struct {
	ID           uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	MerchantName string `json:"MerchantName" gorm:"unique;not null;type:varchar(64);index"`
	Password     string `json:"Password" gorm:"not null;type:varchar(64)"`
	Phone        string `json:"Phone" gorm:"not null;type:varchar(20);uniqueIndex"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Version      uint    `gorm:"version;default:1" json:"version"`
	Stores       []Store `gorm:"foreignKey:MerchantID" json:"stores,omitempty"`
}

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

func (m *Merchant) BeforeCreate(tx *gorm.DB) error {
	m.ID = uint(GenID())
	if !phoneRegex.MatchString(m.Phone) {
		return errors.New("手机号格式不正确")
	}
	m.MerchantName = strings.TrimSpace(m.MerchantName)
	if m.MerchantName == "" {
		return errors.New("商户名不能为空")
	}
	if len(m.MerchantName) > 64 {
		return errors.New("商户名长度不能超过64个字符")
	}
	if len(m.Password) < 8 {
		return errors.New("密码长度必须至少8位")
	}
	var existingMerchant Merchant
	if err := tx.Where("phone = ?", m.Phone).First(&existingMerchant).Error; err == nil {
		return errors.New("该手机号已被注册")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("数据库查询错误: %w", err)
	}
	if err := tx.Where("merchant_name = ?", m.MerchantName).First(&existingMerchant).Error; err == nil {
		return errors.New("该商户名已被使用")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("数据库查询错误: %w", err)
	}
	return nil
}

type Store struct {
	ID          uint    `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	MerchantID  uint    `gorm:"not null;index" json:"merchantID"`
	Name        string  `gorm:"not null;type:varchar(32);index:,unique,where:merchant_id = merchant_id" json:"name"`
	Description string  `gorm:"type:varchar(255)" json:"description"`
	Active      bool    `json:"active" gorm:"default:true"`
	AvgRating   float64 `json:"avgRating" gorm:"default:5"`
	Address     string  `gorm:"type:varchar(64)" json:"address"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     uint     `gorm:"version;default:1" json:"version"`
	Merchant    Merchant `json:"-"`
	Dishes      []Dishes `gorm:"foreignKey:StoreID" json:"dishes,omitempty"`
}

func (s *Store) BeforeCreate(tx *gorm.DB) error {
	if s.Name == "" {
		return errors.New("门店名称不能为空")
	}
	if len(s.Name) > 32 {
		return errors.New("门店名称长度不能超过32个字符")
	}
	if len(s.Description) > 255 {
		return errors.New("门店描述长度不能超过255个字符")
	}
	var count int64
	if err := tx.Model(&Merchant{}).Where("id = ?", s.MerchantID).Count(&count).Error; err != nil {
		return fmt.Errorf("数据库查询错误: %w", err)
	}
	if count == 0 {
		return errors.New("关联的商户不存在")
	}
	return nil
}

type Dishes struct {
	ID        uint            `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	StoreID   uint            `gorm:"not null;index" json:"storeId"`
	Name      string          `gorm:"not null;type:varchar(32)" json:"name"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2)" json:"price"`
	Desc      string          `gorm:"type:varchar(255)" json:"desc"`
	ImageURL  string          `gorm:"type:varchar(255)" json:"imageUrl"`
	Available bool            `gorm:"default:true" json:"available"`
	AvgRating float64         `gorm:"default:5" json:"avgRating"`
	LikeNum   uint            `gorm:"default:0" json:"likeNum"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   uint  `gorm:"version;default:1" json:"version"`
	Store     Store `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Tags      []Tag `gorm:"many2many:dishes_tags;" json:"tags,omitempty"`
}

func (d *Dishes) BeforeCreate(tx *gorm.DB) error {
	if d.Name == "" {
		return errors.New("菜品名称不能为空")
	}
	if len(d.Name) > 32 {
		return errors.New("菜品名称长度不能超过32个字符")
	}
	if len(d.Desc) > 255 {
		return errors.New("菜品描述长度不能超过255个字符")
	}
	if d.Price.IsNegative() {
		return errors.New("菜品价格不能为负数")
	}
	var count int64
	if err := tx.Model(&Store{}).Where("id = ? AND active = true", d.StoreID).Count(&count).Error; err != nil {
		return fmt.Errorf("数据库查询错误: %w", err)
	}
	if count == 0 {
		return errors.New("关联的店铺不存在或未激活")
	}
	if err := tx.Model(&Dishes{}).Where("store_id = ? AND name = ?", d.StoreID, d.Name).Count(&count).Error; err != nil {
		return fmt.Errorf("数据库查询错误: %w", err)
	}
	if count > 0 {
		return errors.New("同一店铺下不能有同名菜品")
	}
	if d.AvgRating < 0 || d.AvgRating > 5 {
		return errors.New("评分必须在0到5之间")
	}
	if d.ImageURL != "" && !isValidURL(d.ImageURL) {
		return errors.New("图片URL格式无效")
	}
	return nil
}

func isValidURL(url string) bool {
	pattern := `^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([\/\w.-]*)*\/?$`
	return regexp.MustCompile(pattern).MatchString(url)
}

type Tag struct {
	ID     uint     `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Name   string   `gorm:"not null;type:varchar(12);uniqueIndex" json:"name"`
	Dishes []Dishes `gorm:"many2many:dishes_tags;" json:"-"`
}
