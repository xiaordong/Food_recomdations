package utils

import (
	"time"

	"gorm.io/gorm"
)

// 商家模型
type Admin struct {
	ID        uint      `gorm:"primary_key;auto_increment" json:"id"`
	AdminName string    `json:"adminName" gorm:"unique;type: varchar(64)"`
	Password  string    `json:"password" gorm:"type: varchar(64)"`
	Phone     string    `json:"phone" gorm:"type: varchar(64);default:null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Stores    []Store   `gorm:"foreignKey:MerchantID;references:AdminName" json:"stores,omitempty"`
}

// 店铺模型
type Store struct {
	ID          int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID  string     `gorm:"type:varchar(50);not null" json:"merchant_id"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Address     string     `gorm:"type:text" json:"address"`
	Description string     `gorm:"type:text" json:"description"`
	Active      bool       `gorm:"default:true" json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Dishes      []Dish     `gorm:"foreignKey:StoreID" json:"dishes,omitempty"`
	StoreTags   []StoreTag `gorm:"foreignKey:StoreID" json:"tags,omitempty"`
	AvgRating   float32    `gorm:"type:float;default:0" json:"avg_rating"`
}

// 标签模型
type Tag struct {
	ID   int32  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(50);not null;unique" json:"name"`
}

// 店铺标签关联模型
type StoreTag struct {
	ID      int32 `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID int32 `gorm:"not null" json:"store_id"`
	TagID   int32 `gorm:"not null" json:"tag_id"`
}

// 用户模型
type User struct {
	Username      string          `json:"username" gorm:"primary_key;type: varchar(64)"`
	Password      string          `json:"password" gorm:"type: varchar(64)"`
	Email         string          `json:"email" gorm:"type: varchar(64);default:null"`
	Phone         string          `json:"phone" gorm:"type: varchar(64);default:null"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	BrowseHistory []BrowseHistory `gorm:"foreignKey:Username" json:"browse_history,omitempty"`
	SearchHistory []SearchHistory `gorm:"foreignKey:Username" json:"search_history,omitempty"`
}

// 菜品分类模型
type Category struct {
	ID        int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID   int32     `gorm:"not null" json:"store_id"`
	Name      string    `gorm:"type:varchar(50);not null" json:"name"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 菜品模型
type Dish struct {
	ID          int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID     int32     `gorm:"not null" json:"store_id"`
	CategoryID  int32     `gorm:"not null" json:"category_id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Price       float32   `gorm:"type:float;not null" json:"price"`
	Description string    `gorm:"type:text" json:"description"`
	ImageURL    string    `gorm:"type:text" json:"image_url"`
	Available   bool      `gorm:"default:true" json:"available"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DishTags    []DishTag `gorm:"foreignKey:DishID" json:"tags,omitempty"`
	AvgRating   float32   `gorm:"type:float;default:0" json:"avg_rating"`
}

// 菜品标签关联模型
type DishTag struct {
	ID     int32 `gorm:"primaryKey;autoIncrement" json:"id"`
	DishID int32 `gorm:"not null" json:"dish_id"`
	TagID  int32 `gorm:"not null" json:"tag_id"`
}

// 用户浏览历史
type BrowseHistory struct {
	ID       int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string    `gorm:"not null" json:"username"`
	StoreID  int32     `gorm:"not null" json:"store_id"`
	DishID   int32     `gorm:"default:0" json:"dish_id"`
	BrowseAt time.Time `json:"browse_at"`
}

// 用户搜索历史
type SearchHistory struct {
	ID       int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string    `gorm:"not null" json:"username"`
	Keyword  string    `gorm:"type:text;not null" json:"keyword"`
	SearchAt time.Time `json:"search_at"`
}

// 初始化数据库表
func InitDB(db *gorm.DB) error {
	// 自动迁移模型
	err := db.AutoMigrate(
		&Admin{},
		&Store{},
		&Tag{},
		&StoreTag{},
		&User{},
		&Category{},
		&Dish{},
		&DishTag{},
		&BrowseHistory{},
		&SearchHistory{},
	)
	return err
}
