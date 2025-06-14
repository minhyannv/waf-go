package waf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"waf-go/internal/config"
	"waf-go/internal/logger"
	"waf-go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Engine struct {
	db     *gorm.DB
	redis  *redis.Client
	config *config.Config
	rules  map[uint]*models.Rule
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

// NewEngine 创建WAF引擎实例
func NewEngine(db *gorm.DB, redis *redis.Client, cfg *config.Config) *Engine {
	engine := &Engine{
		db:     db,
		redis:  redis,
		config: cfg,
		rules:  make(map[uint]*models.Rule),
	}

	// 加载规则
	engine.LoadRules()

	return engine
}

// LoadRules 加载所有启用的规则
func (e *Engine) LoadRules() error {
	var rules []models.Rule
	err := e.db.Where("enabled = ?", true).Find(&rules).Error
	if err != nil {
		return err
	}

	e.rules = make(map[uint]*models.Rule)
	for i := range rules {
		e.rules[rules[i].ID] = &rules[i]
	}

	logger.Info("规则加载完成", zap.Int("count", len(rules)))
	return nil
}

// ProcessRequest WAF中间件处理函数
func (e *Engine) ProcessRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqInfo := e.extractRequestInfo(c)

		// 检查白名单
		if e.checkWhitelist(reqInfo) {
			c.Next()
			return
		}

		// 检查黑名单
		if e.checkBlacklist(reqInfo) {
			e.blockRequest(c, reqInfo, "blacklist", "IP在黑名单中")
			return
		}

		// 检查速率限制
		if e.checkRateLimit(reqInfo) {
			e.blockRequest(c, reqInfo, "rate_limit", "请求过于频繁")
			return
		}

		// 匹配规则
		result := e.matchRules(reqInfo)
		if result.Matched {
			switch result.Action {
			case "block":
				e.logAttack(reqInfo, result)
				e.blockRequest(c, reqInfo, result.Rule.Name, result.ResponseMsg)
				return
			case "log":
				e.logAttack(reqInfo, result)
			}
		}

		c.Next()
	}
}

// extractRequestInfo 提取请求信息
func (e *Engine) extractRequestInfo(c *gin.Context) *RequestInfo {
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		headers[key] = strings.Join(values, ";")
	}

	body := ""
	if c.Request.ContentLength > 0 && c.Request.ContentLength < 1024*1024 { // 限制1MB
		if bodyBytes, err := c.GetRawData(); err == nil {
			body = string(bodyBytes)
		}
	}

	return &RequestInfo{
		ID:        generateRequestID(),
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Method:    c.Request.Method,
		URI:       c.Request.RequestURI,
		Headers:   headers,
		Body:      body,
		TenantID:  e.getTenantID(c.Request.Host),
	}
}

// checkWhitelist 检查白名单
func (e *Engine) checkWhitelist(req *RequestInfo) bool {
	var count int64
	e.db.Model(&models.WhiteList{}).Where(
		"enabled = ? AND tenant_id = ? AND ((type = ? AND value = ?) OR (type = ? AND value = ?) OR (type = ? AND value = ?))",
		true, req.TenantID, "ip", req.IP, "uri", req.URI, "user_agent", req.UserAgent,
	).Count(&count)

	return count > 0
}

// checkBlacklist 检查黑名单
func (e *Engine) checkBlacklist(req *RequestInfo) bool {
	var count int64
	e.db.Model(&models.BlackList{}).Where(
		"enabled = ? AND tenant_id = ? AND ((type = ? AND value = ?) OR (type = ? AND value = ?) OR (type = ? AND value = ?))",
		true, req.TenantID, "ip", req.IP, "uri", req.URI, "user_agent", req.UserAgent,
	).Count(&count)

	return count > 0
}

// checkRateLimit 检查速率限制
func (e *Engine) checkRateLimit(req *RequestInfo) bool {
	key := fmt.Sprintf("rate_limit:%s", req.IP)
	window := time.Now().Unix() / int64(e.config.WAF.RateLimitWindow)

	rateLimitKey := fmt.Sprintf("%s:%d", key, window)

	count, err := e.redis.Get(context.Background(), rateLimitKey).Int()
	if err != nil && err != redis.Nil {
		logger.Error("Redis查询失败", zap.Error(err))
		return false
	}

	if count >= e.config.WAF.MaxRequests {
		return true
	}

	// 增加计数
	pipe := e.redis.Pipeline()
	pipe.Incr(context.Background(), rateLimitKey)
	pipe.Expire(context.Background(), rateLimitKey, time.Duration(e.config.WAF.RateLimitWindow)*time.Second)
	_, err = pipe.Exec(context.Background())
	if err != nil {
		logger.Error("Redis操作失败", zap.Error(err))
	}

	return false
}

// matchRules 匹配规则
func (e *Engine) matchRules(req *RequestInfo) *MatchResult {
	for _, rule := range e.rules {
		if rule.TenantID != 0 && rule.TenantID != req.TenantID {
			continue
		}

		matched, field, value := e.matchRule(rule, req)
		if matched {
			return &MatchResult{
				Matched:      true,
				Rule:         rule,
				MatchField:   field,
				MatchValue:   value,
				Action:       rule.Action,
				ResponseCode: rule.ResponseCode,
				ResponseMsg:  rule.ResponseMsg,
			}
		}
	}

	return &MatchResult{Matched: false}
}

// matchRule 匹配单个规则
func (e *Engine) matchRule(rule *models.Rule, req *RequestInfo) (bool, string, string) {
	var targetValue string
	var fieldName string

	switch rule.MatchType {
	case "uri":
		targetValue = req.URI
		fieldName = "uri"
	case "ip":
		targetValue = req.IP
		fieldName = "ip"
	case "user_agent":
		targetValue = req.UserAgent
		fieldName = "user_agent"
	case "header":
		// 匹配所有头部
		for key, value := range req.Headers {
			if e.matchPattern(rule, key+":"+value) {
				return true, "header", key + ":" + value
			}
		}
		return false, "", ""
	case "body":
		targetValue = req.Body
		fieldName = "body"
	default:
		return false, "", ""
	}

	return e.matchPattern(rule, targetValue), fieldName, targetValue
}

// matchPattern 匹配模式
func (e *Engine) matchPattern(rule *models.Rule, value string) bool {
	switch rule.MatchMode {
	case "exact":
		return value == rule.Pattern
	case "contains":
		return strings.Contains(value, rule.Pattern)
	case "regex":
		if regex, err := regexp.Compile(rule.Pattern); err == nil {
			return regex.MatchString(value)
		}
		return false
	default:
		return false
	}
}

// logAttack 记录攻击日志
func (e *Engine) logAttack(req *RequestInfo, result *MatchResult) {
	headersJSON, _ := json.Marshal(req.Headers)

	log := &models.AttackLog{
		RequestID:      req.ID,
		ClientIP:       req.IP,
		UserAgent:      req.UserAgent,
		RequestMethod:  req.Method,
		RequestURI:     req.URI,
		RequestHeaders: string(headersJSON),
		RequestBody:    req.Body,
		RuleID:         result.Rule.ID,
		RuleName:       result.Rule.Name,
		MatchField:     result.MatchField,
		MatchValue:     result.MatchValue,
		Action:         result.Action,
		ResponseCode:   result.ResponseCode,
		TenantID:       req.TenantID,
	}

	if err := e.db.Create(log).Error; err != nil {
		logger.Error("保存攻击日志失败", zap.Error(err))
	}
}

// blockRequest 阻止请求
func (e *Engine) blockRequest(c *gin.Context, req *RequestInfo, ruleName, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"code":       403,
		"message":    message,
		"rule":       ruleName,
		"request_id": req.ID,
	})
	c.Abort()
}

// getTenantID 根据域名获取租户ID
func (e *Engine) getTenantID(host string) uint {
	var tenant models.Tenant
	err := e.db.Where("domain = ? OR domain = ?", host, "*").First(&tenant).Error
	if err != nil {
		return 1 // 默认租户
	}
	return tenant.ID
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
