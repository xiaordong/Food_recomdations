package model

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

type User struct {
	ID        uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Username  string `gorm:"unique;not null;type:varchar(64);index" json:"username"`
	Password  string `json:"Password" gorm:"not null;type:varchar(64)"`
	Phone     string `json:"Phone" gorm:"not null;type:varchar(20);uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Searches  []Search `gorm:"foreignKey:UserID" json:"searches,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if !phoneRegex.MatchString(u.Phone) {
		return errors.New("手机号格式不正确")
	}
	u.Username = strings.TrimSpace(u.Username)
	if u.Username == "" {
		return errors.New("用户名不能为空")
	}
	if len(u.Username) > 64 {
		return errors.New("用户名长度不能超过64个字符")
	}
	if len(u.Password) < 8 {
		return errors.New("密码长度必须至少8位")
	}
	var existingUser User
	if err := tx.Where("phone = ?", u.Phone).First(&existingUser).Error; err == nil {
		return errors.New("该手机号已被注册")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("数据库查询错误: %w", err)
	}
	if err := tx.Where("username = ?", u.Username).First(&existingUser).Error; err == nil {
		return errors.New("该用户名已被使用")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("数据库查询错误: %w", err)
	}
	return nil
}

type Search struct {
	ID        uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	UserID    uint   `gorm:"not null;index" json:"userId"`
	User      User   `gorm:"foreignKey:UserID" json:"-"`
	Key       string `gorm:"unique;not null;type:varchar(64);index" json:"key"`
	CreatedAt time.Time
}

func (s *Search) BeforeCreate(tx *gorm.DB) error {
	s.Key = strings.TrimSpace(s.Key)
	if s.Key == "" {
		return errors.New("搜索关键词不能为空")
	}
	if len(s.Key) > 64 {
		return errors.New("搜索关键词长度不能超过64个字符")
	}
	var count int64
	if err := tx.Model(&User{}).Where("id = ?", s.UserID).Count(&count).Error; err != nil {
		return fmt.Errorf("数据库查询错误: %w", err)
	}
	if count == 0 {
		return errors.New("关联的用户不存在")
	}
	if s.ID == 0 {
		var existingSearch Search
		if err := tx.Where("key = ?", s.Key).First(&existingSearch).Error; err == nil {
			return errors.New("该搜索关键词已存在")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("数据库查询错误: %w", err)
		}
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	return nil
}

type History struct {
	ID        uint  `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	UserID    uint  `gorm:"not null;index" json:"userId"`
	User      User  `gorm:"foreignKey:UserID" json:"-"`
	StoreID   uint  `gorm:"not null;index" json:"storeId"`
	Store     Store `gorm:"foreignKey:StoreID" json:"-"`
	CreatedAt time.Time
}

func (h *History) BeforeCreate(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&User{}).Where("id = ?", h.UserID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("关联的用户不存在")
	}
	if err := tx.Model(&Store{}).Where("id = ? AND active = true", h.StoreID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("关联的店铺不存在或未激活")
	}
	return nil
}

type ShowMerchant struct {
	Img        string `json:"img"`        // 菜品图片URL
	DishesName string `json:"dishesName"` // 菜品名称
	StoreName  string `json:"storeName"`  // 店铺名称
	Rating     string `json:"rating"`     // 评分（注意：实际是float64类型，JSON中转为string）
	Link       string `json:"link"`       // 链接（实际是店铺ID）
}
type Statue struct {
	DishesID uint `gorm:"not null;index" json:"dishesId"`
}
