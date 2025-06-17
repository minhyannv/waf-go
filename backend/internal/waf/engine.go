package waf

import (
	"context"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"waf-go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// WAFEngine WAF引擎
type WAFEngine struct {
	db          *gorm.DB
	redisClient *redis.Client
	rules       []models.Rule
	blackList   []models.BlackList
	whiteList   []models.WhiteList
	lastUpdate  time.Time
}

type RequestInfo struct {
	ID        string
	IP        string
	UserAgent string
	Method    string
	URI       string
	Headers   map[string]string
	Body      string
	TenantID  uint
}

type MatchResult struct {
	Matched      bool
	Rule         *models.Rule
	MatchField   string
	MatchValue   string
	Action       string
	ResponseCode int
	ResponseMsg  string
}

// NewWAFEngine 创建新的WAF引擎
func NewWAFEngine(db *gorm.DB, redisClient *redis.Client) *WAFEngine {
	engine := &WAFEngine{
		db:          db,
		redisClient: redisClient,
		rules:       []models.Rule{},
		blackList:   []models.BlackList{},
		whiteList:   []models.WhiteList{},
		lastUpdate:  time.Time{},
	}

	// 初始化时加载所有规则
	engine.LoadRules()

	return engine
}

// LoadRules 加载所有启用的规则、黑名单和白名单
func (e *WAFEngine) LoadRules() {
	// 加载启用的规则
	if err := e.db.Where("enabled = ?", true).Find(&e.rules).Error; err != nil {
		log.Printf("Failed to load rules: %v", err)
		return
	}

	// 加载启用的黑名单
	if err := e.db.Where("enabled = ?", true).Find(&e.blackList).Error; err != nil {
		log.Printf("Failed to load blacklist: %v", err)
		return
	}

	// 加载启用的白名单
	if err := e.db.Where("enabled = ?", true).Find(&e.whiteList).Error; err != nil {
		log.Printf("Failed to load whitelist: %v", err)
		return
	}

	e.lastUpdate = time.Now()
	log.Printf("Loaded %d rules, %d blacklist items, %d whitelist items",
		len(e.rules), len(e.blackList), len(e.whiteList))
}

// GetDomainByHost 根据Host头获取域名配置
func (e *WAFEngine) GetDomainByHost(host string) (*models.Domain, error) {
	// 移除端口号
	hostWithoutPort := host
	if idx := strings.Index(host, ":"); idx != -1 {
		hostWithoutPort = host[:idx]
	}

	var domain models.Domain

	// 首先尝试精确匹配
	if err := e.db.Where("domain = ? AND enabled = ?", hostWithoutPort, true).First(&domain).Error; err == nil {
		return &domain, nil
	}

	// 如果没有精确匹配，尝试通配符匹配
	if err := e.db.Where("domain = ? AND enabled = ?", "*", true).First(&domain).Error; err == nil {
		return &domain, nil
	}

	return nil, fmt.Errorf("domain not found: %s", hostWithoutPort)
}

// GetDomainRules 获取域名对应的所有规则（通过域名->策略->规则的关联）
func (e *WAFEngine) GetDomainRules(domainID uint) ([]models.Rule, error) {
	var rules []models.Rule

	// 通过多对多关联获取域名对应的规则
	// 查询路径：domains -> domain_policies -> policies -> policy_rules -> rules
	err := e.db.Table("rules").
		Select("DISTINCT rules.*").
		Joins("JOIN policy_rules ON rules.id = policy_rules.rule_id").
		Joins("JOIN policies ON policy_rules.policy_id = policies.id").
		Joins("JOIN domain_policies ON policies.id = domain_policies.policy_id").
		Where("domain_policies.domain_id = ? AND domain_policies.enabled = ? AND policy_rules.enabled = ? AND policies.enabled = ? AND rules.enabled = ?",
			domainID, true, true, true, true).
		Order("policy_rules.priority DESC, rules.priority DESC").
		Find(&rules).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get domain rules: %v", err)
	}

	return rules, nil
}

// GetDomainBlackList 获取域名对应的黑名单（通过多对多关联）
func (e *WAFEngine) GetDomainBlackList(domainID uint) ([]models.BlackList, error) {
	var blackList []models.BlackList

	// 通过多对多关联获取域名对应的黑名单
	err := e.db.Table("black_lists").
		Select("DISTINCT black_lists.*").
		Joins("JOIN domain_black_lists ON black_lists.id = domain_black_lists.black_list_id").
		Where("domain_black_lists.domain_id = ? AND domain_black_lists.enabled = ? AND black_lists.enabled = ?",
			domainID, true, true).
		Order("domain_black_lists.priority DESC").
		Find(&blackList).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get domain blacklist: %v", err)
	}

	return blackList, nil
}

// GetDomainWhiteList 获取域名对应的白名单（通过多对多关联）
func (e *WAFEngine) GetDomainWhiteList(domainID uint) ([]models.WhiteList, error) {
	var whiteList []models.WhiteList

	// 通过多对多关联获取域名对应的白名单
	err := e.db.Table("white_lists").
		Select("DISTINCT white_lists.*").
		Joins("JOIN domain_white_lists ON white_lists.id = domain_white_lists.white_list_id").
		Where("domain_white_lists.domain_id = ? AND domain_white_lists.enabled = ? AND white_lists.enabled = ?",
			domainID, true, true).
		Order("domain_white_lists.priority DESC").
		Find(&whiteList).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get domain whitelist: %v", err)
	}

	return whiteList, nil
}

// CheckRequest 检查请求是否符合WAF规则
func (e *WAFEngine) CheckRequest(c *gin.Context) (*CheckResult, error) {
	// 获取域名配置
	host := c.Request.Host
	domain, err := e.GetDomainByHost(host)
	if err != nil {
		return &CheckResult{
			Action:     "allow",
			StatusCode: 200,
			Message:    fmt.Sprintf("Domain not configured: %s", host),
			Domain:     host,
		}, nil
	}

	result := &CheckResult{
		Action:     "allow",
		StatusCode: 200,
		Message:    "Request allowed",
		Domain:     domain.Domain,
		DomainID:   domain.ID,
		TenantID:   domain.TenantID,
	}

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	uri := c.Request.URL.Path

	// 1. 首先检查白名单（优先级最高）
	whiteList, err := e.GetDomainWhiteList(domain.ID)
	if err != nil {
		log.Printf("Failed to get domain whitelist: %v", err)
	} else {
		for _, item := range whiteList {
			if e.matchWhiteList(item, clientIP, uri, userAgent) {
				result.Action = "allow"
				result.Message = fmt.Sprintf("Whitelisted: %s", item.Comment)
				return result, nil
			}
		}
	}

	// 2. 检查黑名单
	blackList, err := e.GetDomainBlackList(domain.ID)
	if err != nil {
		log.Printf("Failed to get domain blacklist: %v", err)
	} else {
		for _, item := range blackList {
			if e.matchBlackList(item, clientIP, uri, userAgent) {
				result.Action = "block"
				result.StatusCode = 403
				result.Message = fmt.Sprintf("Blacklisted: %s", item.Comment)
				result.MatchedRule = &MatchedRule{
					ID:         0, // 黑名单没有规则ID
					Name:       fmt.Sprintf("黑名单-%s", item.Type),
					MatchField: item.Type,
					MatchValue: item.Value,
				}
				return result, nil
			}
		}
	}

	// 3. 检查速率限制
	if e.checkRateLimit(clientIP, domain.TenantID) {
		result.Action = "block"
		result.StatusCode = 429
		result.Message = "Rate limit exceeded"
		result.MatchedRule = &MatchedRule{
			ID:         0,
			Name:       "速率限制",
			MatchField: "rate_limit",
			MatchValue: clientIP,
		}
		return result, nil
	}

	// 4. 检查WAF规则
	rules, err := e.GetDomainRules(domain.ID)
	if err != nil {
		log.Printf("Failed to get domain rules: %v", err)
		// 如果获取规则失败，默认允许通过
		return result, nil
	}

	for _, rule := range rules {
		matched, matchValue := e.matchRule(rule, c)
		if matched {
			result.MatchedRule = &MatchedRule{
				ID:         rule.ID,
				Name:       rule.Name,
				MatchField: rule.MatchType,
				MatchValue: matchValue,
			}

			switch rule.Action {
			case "block":
				result.Action = "block"
				result.StatusCode = rule.ResponseCode
				if rule.ResponseMsg != "" {
					result.Message = rule.ResponseMsg
				} else {
					result.Message = fmt.Sprintf("Blocked by rule: %s", rule.Name)
				}
				return result, nil
			case "log":
				result.Action = "log"
				result.Message = fmt.Sprintf("Logged by rule: %s", rule.Name)
				// 继续检查其他规则
			case "allow":
				result.Action = "allow"
				result.Message = fmt.Sprintf("Allowed by rule: %s", rule.Name)
				return result, nil
			}
		}
	}

	return result, nil
}

// matchWhiteList 检查是否匹配白名单
func (e *WAFEngine) matchWhiteList(item models.WhiteList, clientIP, uri, userAgent string) bool {
	return e.matchListItem(item.Type, item.Value, clientIP, uri, userAgent)
}

// matchBlackList 检查是否匹配黑名单
func (e *WAFEngine) matchBlackList(item models.BlackList, clientIP, uri, userAgent string) bool {
	return e.matchListItem(item.Type, item.Value, clientIP, uri, userAgent)
}

// matchListItem 通用的列表项匹配函数
func (e *WAFEngine) matchListItem(itemType, value, clientIP, uri, userAgent string) bool {
	switch itemType {
	case "ip":
		return e.matchIP(value, clientIP)
	case "uri":
		return strings.Contains(uri, value)
	case "user_agent":
		return strings.Contains(strings.ToLower(userAgent), strings.ToLower(value))
	}
	return false
}

// matchIP 检查IP是否匹配（支持CIDR）
func (e *WAFEngine) matchIP(pattern, clientIP string) bool {
	// 如果包含/，按CIDR处理
	if strings.Contains(pattern, "/") {
		_, cidr, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		ip := net.ParseIP(clientIP)
		return ip != nil && cidr.Contains(ip)
	}

	// 否则精确匹配
	return pattern == clientIP
}

// checkRateLimit 检查速率限制
func (e *WAFEngine) checkRateLimit(clientIP string, tenantID uint) bool {
	// 使用Redis实现滑动窗口速率限制
	key := fmt.Sprintf("rate_limit:%d:%s", tenantID, clientIP)
	window := 60       // 60秒窗口
	maxRequests := 100 // 最大请求数

	ctx := context.Background()

	// 获取当前计数
	count, err := e.redisClient.Get(ctx, key).Int()
	if err != nil && err.Error() != "redis: nil" {
		log.Printf("Redis error: %v", err)
		return false
	}

	if count >= maxRequests {
		return true
	}

	// 增加计数
	pipe := e.redisClient.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(window)*time.Second)
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("Redis pipeline error: %v", err)
	}

	return false
}

// matchRule 检查请求是否匹配规则
func (e *WAFEngine) matchRule(rule models.Rule, c *gin.Context) (bool, string) {
	var value string

	switch rule.MatchType {
	case "uri":
		value = c.Request.URL.Path
	case "ip":
		value = c.ClientIP()
	case "header":
		// pattern格式: "header_name" 或 "header_name:header_value"
		if strings.Contains(rule.Pattern, ":") {
			parts := strings.SplitN(rule.Pattern, ":", 2)
			headerName, expectedValue := parts[0], parts[1]
			actualValue := c.GetHeader(headerName)
			return e.performMatch(rule.MatchMode, expectedValue, actualValue), actualValue
		} else {
			value = c.GetHeader(rule.Pattern)
		}
	case "body":
		// 读取请求体
		if c.Request.Body != nil {
			bodyBytes, err := c.GetRawData()
			if err == nil {
				value = string(bodyBytes)
			}
		}
	case "user_agent":
		value = c.GetHeader("User-Agent")
	default:
		return false, ""
	}

	matched := e.performMatch(rule.MatchMode, rule.Pattern, value)
	return matched, value
}

// performMatch 执行匹配操作
func (e *WAFEngine) performMatch(matchMode, pattern, value string) bool {
	switch matchMode {
	case "exact":
		return pattern == value
	case "contains":
		return strings.Contains(strings.ToLower(value), strings.ToLower(pattern))
	case "regex":
		matched, err := regexp.MatchString(pattern, value)
		if err != nil {
			log.Printf("Regex error: %v", err)
			return false
		}
		return matched
	default:
		return false
	}
}

// LogAttack 记录攻击日志
func (e *WAFEngine) LogAttack(c *gin.Context, result *CheckResult) {
	if result.MatchedRule == nil {
		return
	}

	// 读取请求体
	var requestBody string
	if c.Request.Body != nil {
		if bodyBytes, err := c.GetRawData(); err == nil {
			requestBody = string(bodyBytes)
		}
	}

	// 获取请求头
	headersMap := make(map[string]string)
	for name, values := range c.Request.Header {
		if len(values) > 0 {
			headersMap[name] = values[0]
		}
	}

	// 简单的JSON序列化
	requestHeaders := "{"
	for k, v := range headersMap {
		requestHeaders += fmt.Sprintf(`"%s":"%s",`, k, v)
	}
	if len(requestHeaders) > 1 {
		requestHeaders = requestHeaders[:len(requestHeaders)-1] // 移除最后的逗号
	}
	requestHeaders += "}"

	// 获取域名
	var domain string
	if result.DomainID > 0 {
		var d models.Domain
		if err := e.db.First(&d, result.DomainID).Error; err == nil {
			domain = d.Domain
		}
	}

	attackLog := models.AttackLog{
		RequestID:      generateRequestID(),
		ClientIP:       c.ClientIP(),
		UserAgent:      c.GetHeader("User-Agent"),
		RequestMethod:  c.Request.Method,
		RequestURI:     c.Request.URL.Path,
		RequestHeaders: requestHeaders,
		RequestBody:    requestBody,
		DomainID:       result.DomainID,
		Domain:         domain,
		RuleID:         result.MatchedRule.ID,
		RuleName:       result.MatchedRule.Name,
		MatchField:     result.MatchedRule.MatchField,
		MatchValue:     result.MatchedRule.MatchValue,
		Action:         result.Action,
		ResponseCode:   result.StatusCode,
		TenantID:       result.TenantID,
		CreatedAt:      time.Now(),
	}

	if err := e.db.Create(&attackLog).Error; err != nil {
		log.Printf("Failed to log attack: %v", err)
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// CheckResult WAF检查结果
type CheckResult struct {
	Action      string       `json:"action"`       // allow, block, log
	StatusCode  int          `json:"status_code"`  // HTTP状态码
	Message     string       `json:"message"`      // 响应消息
	Domain      string       `json:"domain"`       // 匹配的域名
	DomainID    uint         `json:"domain_id"`    // 域名ID
	TenantID    uint         `json:"tenant_id"`    // 租户ID
	MatchedRule *MatchedRule `json:"matched_rule"` // 匹配的规则
}

// MatchedRule 匹配的规则信息
type MatchedRule struct {
	ID         uint   `json:"id"`          // 规则ID
	Name       string `json:"name"`        // 规则名称
	MatchField string `json:"match_field"` // 匹配字段
	MatchValue string `json:"match_value"` // 匹配值
}
