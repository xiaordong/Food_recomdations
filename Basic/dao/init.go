package dao

import (
	"Food_recommendation/Basic/model"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	dsn := "root:2132047479@tcp(127.0.0.1:3306)/Food?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic("数据库连接失败,error:" + err.Error())
	}
	DB = db.Debug()
	AutoMigrate()
	return DB
}
func AutoMigrate() {
	err := DB.AutoMigrate(
		&model.User{},
		&model.Search{},
		&model.Merchant{},
		&model.Store{},
		&model.Dishes{},
		&model.Tag{},
		&model.History{},
	)
	if err != nil {
		panic("数据库迁移失败,error:" + err.Error())
	}
}
