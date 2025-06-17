package service

import (
	"waf-go/internal/proxy"
	"waf-go/internal/waf"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// Services 服务集合
type Services struct {
	authService           *AuthService
	userService           *UserService
	tenantService         *TenantService
	domainService         *DomainService
	policyService         *PolicyService
	ruleService           *RuleService
	logService            *LogService
	dashboardService      *DashboardService
	whiteListService      *WhiteListService
	blackListService      *BlackListService
	configService         *ConfigService
	tenantSecurityService *TenantSecurityService
	wafEngine             *waf.WAFEngine
}

// NewServices 创建服务集合
func NewServices(db *gorm.DB, rdb *redis.Client) *Services {
	proxyManager := proxy.NewProxyManager()
	wafEngine := waf.NewWAFEngine(db, rdb)
	return &Services{
		authService:           NewAuthService(db),
		userService:           NewUserService(db),
		tenantService:         NewTenantService(db),
		domainService:         NewDomainService(db, proxyManager),
		policyService:         NewPolicyService(db),
		ruleService:           NewRuleService(db, wafEngine),
		logService:            NewLogService(db),
		dashboardService:      NewDashboardService(db),
		whiteListService:      NewWhiteListService(db),
		blackListService:      NewBlackListService(db),
		configService:         NewConfigService(db, rdb, nil),
		tenantSecurityService: NewTenantSecurityService(db),
		wafEngine:             wafEngine,
	}
}

// GetUserService 获取用户服务
func (s *Services) GetUserService() *UserService {
	return s.userService
}

// GetTenantService 获取租户服务
func (s *Services) GetTenantService() *TenantService {
	return s.tenantService
}

// GetDomainService 获取域名服务
func (s *Services) GetDomainService() *DomainService {
	return s.domainService
}

// GetPolicyService 获取策略服务
func (s *Services) GetPolicyService() *PolicyService {
	return s.policyService
}

// GetRuleService 获取规则服务
func (s *Services) GetRuleService() *RuleService {
	return s.ruleService
}

// GetLogService 获取日志服务
func (s *Services) GetLogService() *LogService {
	return s.logService
}

// GetDashboardService 获取仪表盘服务
func (s *Services) GetDashboardService() *DashboardService {
	return s.dashboardService
}

func (s *Services) GetWhiteListService() *WhiteListService {
	return s.whiteListService
}

func (s *Services) GetBlackListService() *BlackListService {
	return s.blackListService
}

func (s *Services) GetConfigService() *ConfigService {
	return s.configService
}

func (s *Services) GetTenantSecurityService() *TenantSecurityService {
	return s.tenantSecurityService
}

func (s *Services) GetWAFEngine() *waf.WAFEngine {
	return s.wafEngine
}

func (s *Services) GetAuthService() *AuthService {
	return s.authService
}
