package db

import (
	"fmt"
	"log"
	"waf-go/internal/config"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) *gorm.DB {
	// 确保配置已初始化
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	return db
}

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.Config) *redis.Client {
	// 确保配置已初始化
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return client
}
