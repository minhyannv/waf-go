package service

import (
	"encoding/json"
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
	// 从当前配置获取默认值
	systemConfig := &SystemConfig{
		WAF: WAFConfig{
			RateLimitWindow: s.config.WAF.RateLimitWindow,
			MaxRequests:     s.config.WAF.MaxRequests,
			EnableRateLimit: true,
			EnableBlacklist: true,
			EnableWhitelist: true,
		},
		Server: ServerConfig{
			Mode: s.config.Server.Mode,
		},
		Log: LogConfig{
			Level: s.config.Log.Level,
		},
	}

	// 尝试从Redis获取运行时配置
	configData, err := s.redis.Get(s.redis.Context(), "system:config").Result()
	if err == nil {
		var runtimeConfig SystemConfig
		if json.Unmarshal([]byte(configData), &runtimeConfig) == nil {
			// 合并运行时配置
			if runtimeConfig.WAF.RateLimitWindow > 0 {
				systemConfig.WAF.RateLimitWindow = runtimeConfig.WAF.RateLimitWindow
			}
			if runtimeConfig.WAF.MaxRequests > 0 {
				systemConfig.WAF.MaxRequests = runtimeConfig.WAF.MaxRequests
			}
			systemConfig.WAF.EnableRateLimit = runtimeConfig.WAF.EnableRateLimit
			systemConfig.WAF.EnableBlacklist = runtimeConfig.WAF.EnableBlacklist
			systemConfig.WAF.EnableWhitelist = runtimeConfig.WAF.EnableWhitelist

			if runtimeConfig.Server.Mode != "" {
				systemConfig.Server.Mode = runtimeConfig.Server.Mode
			}
			if runtimeConfig.Log.Level != "" {
				systemConfig.Log.Level = runtimeConfig.Log.Level
			}
		}
	}

	return systemConfig, nil
}

// UpdateSystemConfig 更新系统配置
func (s *ConfigService) UpdateSystemConfig(req *UpdateConfigRequest) (*SystemConfig, error) {
	// 获取当前配置
	currentConfig, err := s.GetSystemConfig()
	if err != nil {
		return nil, err
	}

	// 更新配置
	if req.WAF != nil {
		if req.WAF.RateLimitWindow > 0 {
			currentConfig.WAF.RateLimitWindow = req.WAF.RateLimitWindow
		}
		if req.WAF.MaxRequests > 0 {
			currentConfig.WAF.MaxRequests = req.WAF.MaxRequests
		}
		currentConfig.WAF.EnableRateLimit = req.WAF.EnableRateLimit
		currentConfig.WAF.EnableBlacklist = req.WAF.EnableBlacklist
		currentConfig.WAF.EnableWhitelist = req.WAF.EnableWhitelist
	}

	if req.Server != nil && req.Server.Mode != "" {
		currentConfig.Server.Mode = req.Server.Mode
	}

	if req.Log != nil && req.Log.Level != "" {
		currentConfig.Log.Level = req.Log.Level
	}

	// 保存到Redis
	configData, err := json.Marshal(currentConfig)
	if err != nil {
		return nil, err
	}

	err = s.redis.Set(s.redis.Context(), "system:config", configData, 0).Err()
	if err != nil {
		return nil, err
	}

	return currentConfig, nil
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
