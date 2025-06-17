package service

import (
	"waf-go/internal/config"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ConfigService struct {
	db     *gorm.DB
	redis  *redis.Client
	config *config.Config
}

type SystemConfig struct {
	WAF    WAFConfig    `json:"waf"`
	Server ServerConfig `json:"server"`
	Log    LogConfig    `json:"log"`
}

type WAFConfig struct {
	RateLimitWindow int  `json:"rate_limit_window"`
	MaxRequests     int  `json:"max_requests"`
	EnableRateLimit bool `json:"enable_rate_limit"`
	EnableBlacklist bool `json:"enable_blacklist"`
	EnableWhitelist bool `json:"enable_whitelist"`
}

type ServerConfig struct {
	Mode string `json:"mode"`
}

type LogConfig struct {
	Level string `json:"level"`
}

type UpdateConfigRequest struct {
	WAF    *WAFConfig    `json:"waf,omitempty"`
	Server *ServerConfig `json:"server,omitempty"`
	Log    *LogConfig    `json:"log,omitempty"`
}

func NewConfigService(db *gorm.DB, redis *redis.Client, cfg *config.Config) *ConfigService {
	return &ConfigService{
		db:     db,
		redis:  redis,
		config: cfg,
	}
}

// GetSystemConfig 获取系统配置
func (s *ConfigService) GetSystemConfig() (*SystemConfig, error) {
	return &SystemConfig{
		WAF: WAFConfig{
			RateLimitWindow: s.config.WAF.RateLimitWindow,
			MaxRequests:     s.config.WAF.MaxRequests,
			EnableRateLimit: s.config.WAF.EnableRateLimit,
			EnableBlacklist: s.config.WAF.EnableBlacklist,
			EnableWhitelist: s.config.WAF.EnableWhitelist,
		},
		Server: ServerConfig{
			Mode: s.config.Server.Mode,
		},
		Log: LogConfig{
			Level: "info",
		},
	}, nil
}

// UpdateSystemConfig 更新系统配置
func (s *ConfigService) UpdateSystemConfig(req *UpdateConfigRequest) (*SystemConfig, error) {
	if req.WAF != nil {
		s.config.WAF.RateLimitWindow = req.WAF.RateLimitWindow
		s.config.WAF.MaxRequests = req.WAF.MaxRequests
		s.config.WAF.EnableRateLimit = req.WAF.EnableRateLimit
		s.config.WAF.EnableBlacklist = req.WAF.EnableBlacklist
		s.config.WAF.EnableWhitelist = req.WAF.EnableWhitelist
	}

	if req.Server != nil {
		s.config.Server.Mode = req.Server.Mode
	}

	return s.GetSystemConfig()
}

// ResetSystemConfig 重置系统配置为默认值
func (s *ConfigService) ResetSystemConfig() (*SystemConfig, error) {
	// 删除Redis中的运行时配置
	s.redis.Del(s.redis.Context(), "system:config")

	// 返回默认配置
	return s.GetSystemConfig()
}

// GetConfigStats 获取配置统计信息
func (s *ConfigService) GetConfigStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取当前配置
	config, err := s.GetSystemConfig()
	if err != nil {
		return nil, err
	}

	stats["current_config"] = config
	stats["config_source"] = "runtime" // 或 "default"

	// 检查Redis连接状态
	_, err = s.redis.Ping(s.redis.Context()).Result()
	stats["redis_connected"] = err == nil

	return stats, nil
}
