package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	WAF      WAFConfig      `mapstructure:"waf"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}

type WAFConfig struct {
	RateLimitWindow int `mapstructure:"rate_limit_window"`
	MaxRequests     int `mapstructure:"max_requests"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

var cfg *Config

func Init() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	// 绑定环境变量到配置项
	viper.BindEnv("database.dsn", "DATABASE_DSN")
	viper.BindEnv("redis.addr", "REDIS_ADDR")
	viper.BindEnv("jwt.secret", "JWT_SECRET")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("配置文件读取失败，使用默认配置: %v", err)
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatalf("配置解析失败: %v", err)
	}

	return cfg
}

func setDefaults() {
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.dsn", "root:password@tcp(localhost:3306)/waf?charset=utf8mb4&parseTime=True&loc=Local")
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.secret", "waf-secret-key")
	viper.SetDefault("jwt.expire", 3600)
	viper.SetDefault("waf.rate_limit_window", 60)
	viper.SetDefault("waf.max_requests", 100)
	viper.SetDefault("log.level", "info")
}

func Get() *Config {
	return cfg
}
