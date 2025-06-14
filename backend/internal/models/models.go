package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID        uint      `json:"id" gorm:"primarykey"`                                   // 用户ID，主键
	Username  string    `json:"username" gorm:"uniqueIndex;not null;type:varchar(100)"` // 用户名，唯一
	Password  string    `json:"-" gorm:"not null"`                                      // 密码，加密存储，不返回给前端
	Email     string    `json:"email" gorm:"uniqueIndex;type:varchar(255)"`             // 邮箱地址，唯一
	Role      string    `json:"role" gorm:"not null;default:'viewer'"`                  // 用户角色：admin(超级管理员), tenant_admin(租户管理员), viewer(查看者)
	TenantID  uint      `json:"tenant_id"`                                              // 所属租户ID，0表示超级管理员
	Status    string    `json:"status" gorm:"not null;default:'active'"`                // 用户状态：active(激活), inactive(禁用)
	CreatedAt time.Time `json:"created_at"`                                             // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                                             // 更新时间
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`            // 关联的租户信息
}

// Tenant 租户表
type Tenant struct {
	ID        uint      `json:"id" gorm:"primarykey"`                        // 租户ID，主键
	Name      string    `json:"name" gorm:"not null"`                        // 租户名称
	Domain    string    `json:"domain" gorm:"uniqueIndex;type:varchar(255)"` // 租户域名，唯一标识
	Status    string    `json:"status" gorm:"not null;default:'active'"`     // 租户状态：active(激活), inactive(禁用)
	CreatedAt time.Time `json:"created_at"`                                  // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                                  // 更新时间
}

// Rule WAF规则表
type Rule struct {
	ID           uint      `json:"id" gorm:"primarykey"`                                                    // 规则ID，主键
	Name         string    `json:"name" gorm:"not null;type:varchar(255);uniqueIndex:idx_rule_name_tenant"` // 规则名称，在同一租户内唯一
	Description  string    `json:"description"`                                                             // 规则描述
	MatchType    string    `json:"match_type" gorm:"not null;type:varchar(50);index"`                       // 匹配类型：uri(URI路径), ip(IP地址), header(请求头), body(请求体), user_agent(用户代理)
	Pattern      string    `json:"pattern" gorm:"not null"`                                                 // 匹配模式，具体的匹配规则内容
	MatchMode    string    `json:"match_mode" gorm:"not null;type:varchar(50)"`                             // 匹配模式：exact(精确匹配), regex(正则匹配), contains(包含匹配)
	Action       string    `json:"action" gorm:"not null;type:varchar(50);index"`                           // 执行动作：block(阻断), allow(放行), log(仅记录)
	ResponseCode int       `json:"response_code" gorm:"default:403"`                                        // 阻断时返回的HTTP状态码，默认403
	ResponseMsg  string    `json:"response_msg"`                                                            // 阻断时返回的消息内容
	Priority     int       `json:"priority" gorm:"default:1;index"`                                         // 规则优先级，数字越大优先级越高
	Enabled      bool      `json:"enabled" gorm:"default:true;index"`                                       // 规则是否启用
	TenantID     uint      `json:"tenant_id" gorm:"uniqueIndex:idx_rule_name_tenant;index"`                 // 所属租户ID，0表示全局规则
	CreatedAt    time.Time `json:"created_at"`                                                              // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`                                                              // 更新时间
	Tenant       *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                             // 关联的租户信息
}

// Policy 策略表
type Policy struct {
	ID          uint      `json:"id" gorm:"primarykey"`                                                      // 策略ID，主键
	Name        string    `json:"name" gorm:"not null;type:varchar(255);uniqueIndex:idx_policy_name_tenant"` // 策略名称，在同一租户内唯一
	Description string    `json:"description"`                                                               // 策略描述
	Domain      string    `json:"domain" gorm:"type:varchar(255);index"`                                     // 应用域名，指定策略生效的域名
	RuleIDs     string    `json:"rule_ids" gorm:"type:text"`                                                 // 关联的规则ID列表，JSON数组格式存储
	Enabled     bool      `json:"enabled" gorm:"default:true;index"`                                         // 策略是否启用
	TenantID    uint      `json:"tenant_id" gorm:"uniqueIndex:idx_policy_name_tenant;index"`                 // 所属租户ID，0表示全局策略
	CreatedAt   time.Time `json:"created_at"`                                                                // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`                                                                // 更新时间
	Tenant      *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                               // 关联的租户信息
}

// AttackLog 攻击日志表
type AttackLog struct {
	ID             uint      `json:"id" gorm:"primarykey"`                                                                                    // 攻击日志ID，主键
	RequestID      string    `json:"request_id" gorm:"type:varchar(255);index"`                                                               // 请求唯一标识符
	ClientIP       string    `json:"client_ip" gorm:"type:varchar(45);index:idx_attack_ip_time;index"`                                        // 客户端IP地址
	UserAgent      string    `json:"user_agent"`                                                                                              // 用户代理字符串
	RequestMethod  string    `json:"request_method" gorm:"type:varchar(10);index"`                                                            // HTTP请求方法：GET, POST, PUT, DELETE等
	RequestURI     string    `json:"request_uri" gorm:"type:varchar(500);index:idx_attack_uri_time;index"`                                    // 请求URI路径
	RequestHeaders string    `json:"request_headers" gorm:"type:text"`                                                                        // 请求头信息，JSON格式存储
	RequestBody    string    `json:"request_body" gorm:"type:text"`                                                                           // 请求体内容
	RuleID         uint      `json:"rule_id" gorm:"index"`                                                                                    // 触发的规则ID
	RuleName       string    `json:"rule_name" gorm:"type:varchar(255);index"`                                                                // 触发的规则名称
	MatchField     string    `json:"match_field" gorm:"type:varchar(100)"`                                                                    // 匹配的字段名称
	MatchValue     string    `json:"match_value"`                                                                                             // 匹配的具体值
	Action         string    `json:"action" gorm:"type:varchar(50);index"`                                                                    // 执行的动作：block, allow, log
	ResponseCode   int       `json:"response_code" gorm:"index"`                                                                              // 响应状态码
	TenantID       uint      `json:"tenant_id" gorm:"index:idx_attack_tenant_time;index"`                                                     // 所属租户ID
	CreatedAt      time.Time `json:"created_at" gorm:"index:idx_attack_ip_time;index:idx_attack_uri_time;index:idx_attack_tenant_time;index"` // 攻击发生时间
	Rule           *Rule     `json:"rule,omitempty" gorm:"foreignKey:RuleID"`                                                                 // 关联的规则信息
	Tenant         *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                                                             // 关联的租户信息
}

// RateLimit 速率限制记录
type RateLimit struct {
	ID        uint      `json:"id" gorm:"primarykey"`                              // 限流记录ID，主键
	Key       string    `json:"key" gorm:"uniqueIndex;not null;type:varchar(255)"` // 限流标识符，通常是IP地址或其他唯一标识
	Count     int       `json:"count" gorm:"default:0"`                            // 当前时间窗口内的请求计数
	Window    time.Time `json:"window" gorm:"index"`                               // 时间窗口开始时间
	CreatedAt time.Time `json:"created_at"`                                        // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                                        // 更新时间
}

// WhiteList 白名单
type WhiteList struct {
	ID        uint      `json:"id" gorm:"primarykey"`                                                                      // 白名单ID，主键
	Type      string    `json:"type" gorm:"not null;type:varchar(50);uniqueIndex:idx_whitelist_type_value_tenant"`         // 白名单类型：ip(IP地址), uri(URI路径), user_agent(用户代理)
	Value     string    `json:"value" gorm:"not null;type:varchar(500);uniqueIndex:idx_whitelist_type_value_tenant;index"` // 白名单值，具体的IP、URI或User-Agent
	Comment   string    `json:"comment"`                                                                                   // 备注说明
	TenantID  uint      `json:"tenant_id" gorm:"uniqueIndex:idx_whitelist_type_value_tenant;index"`                        // 所属租户ID，0表示全局白名单
	Enabled   bool      `json:"enabled" gorm:"default:true;index"`                                                         // 是否启用
	CreatedAt time.Time `json:"created_at"`                                                                                // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                                                                                // 更新时间
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                                               // 关联的租户信息
}

// BlackList 黑名单
type BlackList struct {
	ID        uint      `json:"id" gorm:"primarykey"`                                                                      // 黑名单ID，主键
	Type      string    `json:"type" gorm:"not null;type:varchar(50);uniqueIndex:idx_blacklist_type_value_tenant"`         // 黑名单类型：ip(IP地址), uri(URI路径), user_agent(用户代理)
	Value     string    `json:"value" gorm:"not null;type:varchar(500);uniqueIndex:idx_blacklist_type_value_tenant;index"` // 黑名单值，具体的IP、URI或User-Agent
	Comment   string    `json:"comment"`                                                                                   // 备注说明
	TenantID  uint      `json:"tenant_id" gorm:"uniqueIndex:idx_blacklist_type_value_tenant;index"`                        // 所属租户ID，0表示全局黑名单
	Enabled   bool      `json:"enabled" gorm:"default:true;index"`                                                         // 是否启用
	CreatedAt time.Time `json:"created_at"`                                                                                // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                                                                                // 更新时间
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                                               // 关联的租户信息
}

// Webhook 告警配置
type Webhook struct {
	ID        uint      `json:"id" gorm:"primarykey"`                                                       // Webhook配置ID，主键
	Name      string    `json:"name" gorm:"not null;type:varchar(255);uniqueIndex:idx_webhook_name_tenant"` // Webhook名称，在同一租户内唯一
	URL       string    `json:"url" gorm:"not null;type:varchar(500)"`                                      // Webhook回调URL地址
	Method    string    `json:"method" gorm:"not null;type:varchar(10);default:'POST'"`                     // HTTP请求方法，默认POST
	Headers   string    `json:"headers" gorm:"type:text"`                                                   // 请求头信息，JSON格式存储
	Template  string    `json:"template" gorm:"type:text"`                                                  // 消息模板，支持变量替换
	Events    string    `json:"events" gorm:"type:text"`                                                    // 触发事件类型列表，JSON数组格式存储
	Enabled   bool      `json:"enabled" gorm:"default:true;index"`                                          // 是否启用
	TenantID  uint      `json:"tenant_id" gorm:"uniqueIndex:idx_webhook_name_tenant;index"`                 // 所属租户ID，0表示全局配置
	CreatedAt time.Time `json:"created_at"`                                                                 // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                                                                 // 更新时间
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                                // 关联的租户信息
}

// 数据库迁移
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Tenant{},
		&Rule{},
		&Policy{},
		&AttackLog{},
		&RateLimit{},
		&WhiteList{},
		&BlackList{},
		&Webhook{},
	)
}
