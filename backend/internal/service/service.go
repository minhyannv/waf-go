package service

import (
	"waf-go/internal/config"
	"waf-go/internal/waf"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Services struct {
	User      *UserService
	Rule      *RuleService
	Policy    *PolicyService
	Log       *LogService
	Dashboard *DashboardService
	Auth      *AuthService
	Webhook   *WebhookService
	WhiteList *WhiteListService
	BlackList *BlackListService
	Config    *ConfigService
	WAFEngine *waf.Engine
}

func NewServices(db *gorm.DB, redis *redis.Client, cfg *config.Config) *Services {
	wafEngine := waf.NewEngine(db, redis, cfg)

	return &Services{
		User:      NewUserService(db),
		Rule:      NewRuleService(db, wafEngine),
		Policy:    NewPolicyService(db),
		Log:       NewLogService(db),
		Dashboard: NewDashboardService(db),
		Auth:      NewAuthService(db),
		Webhook:   NewWebhookService(db),
		WhiteList: NewWhiteListService(db),
		BlackList: NewBlackListService(db),
		Config:    NewConfigService(db, redis, cfg),
		WAFEngine: wafEngine,
	}
}
