package config

import (
	"log"
	"os"
	"strconv"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	config *Config
	once   sync.Once
)

// Config 系统配置
type Config struct {
	Server   ServerConfig   `yaml:"server" json:"server"`
	Database DatabaseConfig `yaml:"mysql" json:"database"`
	Redis    RedisConfig    `yaml:"redis" json:"redis"`
	WAF      WAFConfig      `yaml:"waf" json:"waf"`
	JWT      JWTConfig      `yaml:"jwt" json:"jwt"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Mode      string `yaml:"mode" json:"mode"`
	HTTPPort  int    `yaml:"http_port" json:"http_port"`
	HTTPSPort int    `yaml:"https_port" json:"https_port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	DBName   string `yaml:"database" json:"dbname"`
	DSN      string `yaml:"dsn" json:"dsn"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
	Addr     string `yaml:"addr" json:"addr"`
}

// WAFConfig WAF配置
type WAFConfig struct {
	RateLimitWindow int  `yaml:"rate_limit_window" json:"rate_limit_window"`
	MaxRequests     int  `yaml:"max_requests" json:"max_requests"`
	EnableRateLimit bool `yaml:"enable_rate_limit" json:"enable_rate_limit"`
	EnableBlacklist bool `yaml:"enable_blacklist" json:"enable_blacklist"`
	EnableWhitelist bool `yaml:"enable_whitelist" json:"enable_whitelist"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `yaml:"secret" json:"secret"`
	Expire int    `yaml:"expire" json:"expire"`
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	once.Do(func() {
		// 默认配置
		config = &Config{
			Server: ServerConfig{
				Mode:      "debug",
				HTTPPort:  8081,
				HTTPSPort: 8443,
			},
			Database: DatabaseConfig{
				Host:     "mysql",
				Port:     3306,
				User:     "waf",
				Password: "waf123456",
				DBName:   "waf",
				DSN:      "",
			},
			Redis: RedisConfig{
				Host:     "redis",
				Port:     6379,
				Password: "",
				DB:       0,
				Addr:     "",
			},
			WAF: WAFConfig{
				RateLimitWindow: 60,
				MaxRequests:     100,
				EnableRateLimit: true,
				EnableBlacklist: true,
				EnableWhitelist: true,
			},
			JWT: JWTConfig{
				Secret: "waf-secret-key-change-in-production",
				Expire: 86400, // 24小时
			},
		}

		// 从配置文件加载
		if err := loadFromFile(); err != nil {
			log.Printf("从配置文件加载失败: %v", err)
		}

		// 从环境变量加载
		loadFromEnv()
	})

	return config
}

// loadFromFile 从配置文件加载配置
func loadFromFile() error {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv() {
	// MySQL配置
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := getEnvInt("MYSQL_PORT", 0); port != 0 {
		config.Database.Port = port
	}
	if user := os.Getenv("MYSQL_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("MYSQL_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if dbname := os.Getenv("MYSQL_DATABASE"); dbname != "" {
		config.Database.DBName = dbname
	}

	// Redis配置
	if host := os.Getenv("REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}
	if port := getEnvInt("REDIS_PORT", 0); port != 0 {
		config.Redis.Port = port
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}
	if db := getEnvInt("REDIS_DB", -1); db != -1 {
		config.Redis.DB = db
	}

	// JWT配置
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.JWT.Secret = secret
	}

	// 服务器配置
	if port := getEnvInt("SERVER_PORT", 0); port != 0 {
		config.Server.HTTPPort = port
	}
}

// Get 获取配置实例
func Get() *Config {
	if config == nil {
		LoadConfig()
	}
	return config
}

// Init 初始化配置
func Init() error {
	LoadConfig()
	return nil
}

// getEnvInt 获取整数类型的环境变量，如果不存在或解析失败则返回默认值
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
