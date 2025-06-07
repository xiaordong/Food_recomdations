package dao

import (
	"Food_recommendation/Basic/model"
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
	// 连接池配置参数
	maxOpenConns    = 200              // 最大打开连接数（根据服务器配置调整）
	maxIdleConns    = 50               // 最大空闲连接数
	connMaxLifetime = 10 * time.Minute // 连接最大存活时间
)

func InitDB() *gorm.DB {
	dsn := "root:2132047479@tcp(127.0.0.1:3306)/food?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s"

	// 配置数据库连接池
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
		// 高并发场景关键配置
		DefaultStringSize:         256,  // 字符串字段默认长度
		DisableDatetimePrecision:  true, // 禁用 datetime 精度（提升性能）
		DontSupportRenameColumn:   true, // 禁止重命名列（避免潜在锁问题）
		DontSupportRenameIndex:    true, // 禁止重命名索引
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		panic(&InitError{Msg: "数据库连接失败", Err: err})
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(&InitError{Msg: "获取数据库连接池失败", Err: err})
	}

	// 配置连接池参数
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		panic(&InitError{Msg: "数据库连接健康检查失败", Err: err})
	}

	DB = db
	go runAutoMigrate()
	return DB
}

// 异步执行数据库迁移（不阻塞主流程）
func runAutoMigrate() {
	// 使用带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := DB.WithContext(ctx).AutoMigrate(
		&model.User{},
		&model.Search{},
		&model.Merchant{},
		&model.Store{},
		&model.Dishes{},
		&model.Tag{},
		&model.History{},
	)

	if err != nil {
		panic(&InitError{Msg: "数据库迁移失败", Err: err})
	}
}

type InitError struct {
	Msg string
	Err error
}

func (e *InitError) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}
