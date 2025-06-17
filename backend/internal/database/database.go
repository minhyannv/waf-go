package database

import (
	"context"
	"log"

	"waf-go/internal/config"
	"waf-go/internal/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
)

// Init 初始化数据库连接
func Init(cfg config.DatabaseConfig) *gorm.DB {
	var err error

	DB, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 强制设置连接字符集为UTF8MB4
	DB.Exec("SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci")
	DB.Exec("SET CHARACTER SET utf8mb4")
	DB.Exec("SET character_set_connection=utf8mb4")
	DB.Exec("SET character_set_client=utf8mb4")
	DB.Exec("SET character_set_results=utf8mb4")

	// 自动迁移数据库表
	if err := models.AutoMigrate(DB); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建默认管理员用户
	createDefaultAdmin()

	return DB
}

// InitRedis 初始化Redis连接
func InitRedis(cfg config.RedisConfig) *redis.Client {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	_, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("连接Redis失败: %v", err)
	}

	return RDB
}

// createDefaultAdmin 创建默认管理员用户
func createDefaultAdmin() {
	var count int64
	DB.Model(&models.User{}).Count(&count)

	if count == 0 {
		// 创建默认租户
		tenant := &models.Tenant{
			Name:   "Default",
			Code:   "default",
			Status: "active",
		}
		DB.Create(tenant)

		// 创建默认管理员
		admin := &models.User{
			Username: "admin",
			Password: "$2a$10$n2yWetLC/dp9YQ2UxIyWG.jtLYrPybUi.QZLwuxD9Hp.MmulSj0BK", // password: admin123
			Email:    "admin@example.com",
			Role:     "admin",
			TenantID: tenant.ID,
			Status:   "active",
		}
		DB.Create(admin)

		log.Println("默认管理员账户已创建: admin / admin123")
	}
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// GetRedis 获取Redis实例
func GetRedis() *redis.Client {
	return RDB
}
