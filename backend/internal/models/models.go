package models

import (
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// 核心实体表
// =============================================================================

// Tenant 租户表 - 多租户系统的核心
type Tenant struct {
	ID        uint      `json:"id" gorm:"primarykey;column:id"`                        // 租户ID，主键
	Name      string    `json:"name" gorm:"not null;type:varchar(100);column:name"`    // 租户名称
	Code      string    `json:"code" gorm:"uniqueIndex;type:varchar(50);column:code"`  // 租户代码，唯一标识
	Status    string    `json:"status" gorm:"not null;default:'active';column:status"` // 租户状态：active(激活), inactive(禁用)
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`                   // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`                   // 更新时间
}

// User 用户表 - 租户下的用户管理
type User struct {
	ID        uint      `json:"id" gorm:"primarykey;column:id"`                                         // 用户ID，主键
	Username  string    `json:"username" gorm:"uniqueIndex;not null;type:varchar(100);column:username"` // 用户名，全局唯一
	Password  string    `json:"-" gorm:"not null;column:password"`                                      // 密码，加密存储，不返回给前端
	Email     string    `json:"email" gorm:"uniqueIndex;type:varchar(255);column:email"`                // 邮箱地址，唯一
	Role      string    `json:"role" gorm:"not null;default:'viewer';column:role"`                      // 用户角色：admin(超级管理员), tenant_admin(租户管理员), viewer(查看者)
	TenantID  uint      `json:"tenant_id" gorm:"not null;index;column:tenant_id"`                       // 所属租户ID，0表示超级管理员
	Status    string    `json:"status" gorm:"not null;default:'active';column:status"`                  // 用户状态：active(激活), inactive(禁用)
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`                                    // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`                                    // 更新时间
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                            // 关联的租户信息
}

// Domain 域名配置表 - 简化版本
type Domain struct {
	ID             uint      `json:"id" gorm:"primarykey;column:id"`                                           // 域名配置ID，主键
	Domain         string    `json:"domain" gorm:"not null;uniqueIndex;type:varchar(255);column:domain"`       // 域名，全局唯一
	Protocol       string    `json:"protocol" gorm:"type:enum('http','https');default:'http';column:protocol"` // 协议：http 或 https
	Port           int       `json:"port" gorm:"default:80;column:port"`                                       // 监听端口
	SSLCertificate string    `json:"ssl_certificate" gorm:"type:text;column:ssl_certificate"`                  // SSL证书内容（PEM格式）
	SSLPrivateKey  string    `json:"ssl_private_key" gorm:"type:text;column:ssl_private_key"`                  // SSL私钥内容（PEM格式）
	BackendURL     string    `json:"backend_url" gorm:"not null;type:varchar(500);column:backend_url"`         // 后端服务地址
	TenantID       uint      `json:"tenant_id" gorm:"not null;index;column:tenant_id"`                         // 所属租户ID
	Enabled        bool      `json:"enabled" gorm:"default:true;index;column:enabled"`                         // 是否启用
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`                                      // 创建时间
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`                                      // 更新时间
	Tenant         *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                              // 关联的租户信息

	// 多对多关系 - 通过关联表连接
	Policies   []Policy    `json:"policies,omitempty" gorm:"many2many:domain_policies"`       // 域名关联的策略列表
	BlackLists []BlackList `json:"black_lists,omitempty" gorm:"many2many:domain_black_lists"` // 域名关联的黑名单列表
	WhiteLists []WhiteList `json:"white_lists,omitempty" gorm:"many2many:domain_white_lists"` // 域名关联的白名单列表
}

// TableName 指定Domain模型使用的表名
func (Domain) TableName() string {
	return "domains"
}

// Policy 策略表 - WAF安全策略定义
type Policy struct {
	ID          uint      `json:"id" gorm:"primarykey;column:id"`
	Name        string    `json:"name" gorm:"not null;type:varchar(255);uniqueIndex:idx_policy_name_tenant;column:name"`
	Description string    `json:"description" gorm:"column:description"`
	Enabled     bool      `json:"enabled" gorm:"default:true;index;column:enabled"`
	TenantID    uint      `json:"tenant_id" gorm:"uniqueIndex:idx_policy_name_tenant;index;column:tenant_id"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
	Tenant      *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Rules       []Rule    `json:"rules,omitempty" gorm:"many2many:policy_rules;"`
}

// Rule WAF规则表 - 具体的安全规则定义
type Rule struct {
	ID           uint      `json:"id" gorm:"primarykey;column:id"`                                                      // 规则ID，主键
	Name         string    `json:"name" gorm:"not null;type:varchar(255);uniqueIndex:idx_rule_name_tenant;column:name"` // 规则名称，在同一租户内唯一
	Description  string    `json:"description" gorm:"column:description"`                                               // 规则描述
	MatchType    string    `json:"match_type" gorm:"not null;type:varchar(50);index;column:match_type"`                 // 匹配类型：uri(URI路径), ip(IP地址), header(请求头), body(请求体), user_agent(用户代理)
	Pattern      string    `json:"pattern" gorm:"not null;column:pattern"`                                              // 匹配模式，具体的匹配规则内容
	MatchMode    string    `json:"match_mode" gorm:"not null;type:varchar(50);column:match_mode"`                       // 匹配模式：exact(精确匹配), regex(正则匹配), contains(包含匹配)
	Action       string    `json:"action" gorm:"not null;type:varchar(50);index;column:action"`                         // 执行动作：block(阻断), allow(放行), log(仅记录)
	ResponseCode int       `json:"response_code" gorm:"default:403;column:response_code"`                               // 阻断时返回的HTTP状态码，默认403
	ResponseMsg  string    `json:"response_msg" gorm:"column:response_msg"`                                             // 阻断时返回的消息内容
	Priority     int       `json:"priority" gorm:"default:1;index;column:priority"`                                     // 规则优先级，数字越大优先级越高
	Enabled      bool      `json:"enabled" gorm:"default:true;index;column:enabled"`                                    // 规则是否启用
	TenantID     uint      `json:"tenant_id" gorm:"uniqueIndex:idx_rule_name_tenant;index;column:tenant_id"`            // 所属租户ID，0表示全局规则
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`                                                 // 创建时间
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`                                                 // 更新时间
	Tenant       *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                                         // 关联的租户信息
}

// BlackList 黑名单表 - 黑名单条目定义
type BlackList struct {
	ID        uint      `json:"id" gorm:"primarykey;column:id"`                                                                         // 黑名单ID，主键
	Type      string    `json:"type" gorm:"not null;type:varchar(50);uniqueIndex:idx_blacklist_type_value_tenant;column:type"`          // 黑名单类型：ip(IP地址), uri(URI路径), user_agent(用户代理)
	Value     string    `json:"value" gorm:"not null;type:varchar(500);uniqueIndex:idx_blacklist_type_value_tenant;index;column:value"` // 黑名单值，具体的IP、URI或User-Agent
	Comment   string    `json:"comment" gorm:"column:comment"`                                                                          // 备注说明
	TenantID  uint      `json:"tenant_id" gorm:"uniqueIndex:idx_blacklist_type_value_tenant;index;column:tenant_id"`                    // 所属租户ID，0表示全局黑名单
	Enabled   bool      `json:"enabled" gorm:"default:true;index;column:enabled"`                                                       // 是否启用
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`                                                                    // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`                                                                    // 更新时间
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                                                            // 关联的租户信息
}

// WhiteList 白名单表 - 白名单条目定义
type WhiteList struct {
	ID        uint      `json:"id" gorm:"primarykey;column:id"`                                                                         // 白名单ID，主键
	Type      string    `json:"type" gorm:"not null;type:varchar(50);uniqueIndex:idx_whitelist_type_value_tenant;column:type"`          // 白名单类型：ip(IP地址), uri(URI路径), user_agent(用户代理)
	Value     string    `json:"value" gorm:"not null;type:varchar(500);uniqueIndex:idx_whitelist_type_value_tenant;index;column:value"` // 白名单值，具体的IP、URI或User-Agent
	Comment   string    `json:"comment" gorm:"column:comment"`                                                                          // 备注说明
	TenantID  uint      `json:"tenant_id" gorm:"uniqueIndex:idx_whitelist_type_value_tenant;index;column:tenant_id"`                    // 所属租户ID，0表示全局白名单
	Enabled   bool      `json:"enabled" gorm:"default:true;index;column:enabled"`                                                       // 是否启用
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`                                                                    // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`                                                                    // 更新时间
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                                                            // 关联的租户信息
}

// =============================================================================
// 多对多关联表
// =============================================================================

// DomainPolicy 域名策略关联表 - 域名(Domain) ↔ 策略(Policy): 多对多
type DomainPolicy struct {
	ID        uint      `json:"id" gorm:"primarykey;column:id"`                                           // 关联ID，主键
	DomainID  uint      `json:"domain_id" gorm:"not null;uniqueIndex:idx_domain_policy;column:domain_id"` // 域名ID
	PolicyID  uint      `json:"policy_id" gorm:"not null;uniqueIndex:idx_domain_policy;column:policy_id"` // 策略ID
	Priority  int       `json:"priority" gorm:"default:1;column:priority"`                                // 策略在该域名下的优先级，数字越大优先级越高
	Enabled   bool      `json:"enabled" gorm:"default:true;index;column:enabled"`                         // 是否启用此关联
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`                                      // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`                                      // 更新时间
	Domain    *Domain   `json:"domain,omitempty" gorm:"foreignKey:DomainID"`                              // 关联的域名配置
	Policy    *Policy   `json:"policy,omitempty" gorm:"foreignKey:PolicyID"`                              // 关联的策略
}

// PolicyRule 策略规则关联表 - 策略(Policy) ↔ 规则(Rule): 多对多
type PolicyRule struct {
	ID        uint      `json:"id" gorm:"primarykey;column:id"`
	PolicyID  uint      `json:"policy_id" gorm:"not null;uniqueIndex:idx_policy_rule;column:policy_id"`
	RuleID    uint      `json:"rule_id" gorm:"not null;uniqueIndex:idx_policy_rule;column:rule_id"`
	Priority  int       `json:"priority" gorm:"default:1;column:priority"`
	Enabled   bool      `json:"enabled" gorm:"default:true;index;column:enabled"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
	Policy    *Policy   `json:"policy,omitempty" gorm:"foreignKey:PolicyID"`
	Rule      *Rule     `json:"rule,omitempty" gorm:"foreignKey:RuleID"`
}

// DomainBlackList 域名黑名单关联表 - 域名(Domain) ↔ 黑名单(BlackList): 多对多
type DomainBlackList struct {
	ID          uint       `json:"id" gorm:"primarykey;column:id"`                                                      // 关联ID，主键
	DomainID    uint       `json:"domain_id" gorm:"not null;uniqueIndex:idx_domain_blacklist;column:domain_id"`         // 域名ID
	BlackListID uint       `json:"black_list_id" gorm:"not null;uniqueIndex:idx_domain_blacklist;column:black_list_id"` // 黑名单ID
	Priority    int        `json:"priority" gorm:"default:1;column:priority"`                                           // 黑名单在该域名下的优先级
	Enabled     bool       `json:"enabled" gorm:"default:true;index;column:enabled"`                                    // 是否启用此关联
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"`                                                 // 创建时间
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"`                                                 // 更新时间
	Domain      *Domain    `json:"domain,omitempty" gorm:"foreignKey:DomainID"`                                         // 关联的域名配置
	BlackList   *BlackList `json:"black_list,omitempty" gorm:"foreignKey:BlackListID"`                                  // 关联的黑名单
}

// DomainWhiteList 域名白名单关联表 - 域名(Domain) ↔ 白名单(WhiteList): 多对多
type DomainWhiteList struct {
	ID          uint       `json:"id" gorm:"primarykey;column:id"`                                                      // 关联ID，主键
	DomainID    uint       `json:"domain_id" gorm:"not null;uniqueIndex:idx_domain_whitelist;column:domain_id"`         // 域名ID
	WhiteListID uint       `json:"white_list_id" gorm:"not null;uniqueIndex:idx_domain_whitelist;column:white_list_id"` // 白名单ID
	Priority    int        `json:"priority" gorm:"default:1;column:priority"`                                           // 白名单在该域名下的优先级
	Enabled     bool       `json:"enabled" gorm:"default:true;index;column:enabled"`                                    // 是否启用此关联
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"`                                                 // 创建时间
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"`                                                 // 更新时间
	Domain      *Domain    `json:"domain,omitempty" gorm:"foreignKey:DomainID"`                                         // 关联的域名配置
	WhiteList   *WhiteList `json:"white_list,omitempty" gorm:"foreignKey:WhiteListID"`                                  // 关联的白名单
}

// =============================================================================
// 业务支撑表
// =============================================================================

// AttackLog 攻击日志表
type AttackLog struct {
	ID             uint      `json:"id" gorm:"primarykey;column:id"`                                                          // 攻击日志ID，主键
	RequestID      string    `json:"request_id" gorm:"type:varchar(255);index;column:request_id"`                             // 请求唯一标识符
	ClientIP       string    `json:"client_ip" gorm:"type:varchar(45);index:idx_attack_ip_time;index;column:client_ip"`       // 客户端IP地址
	UserAgent      string    `json:"user_agent" gorm:"column:user_agent"`                                                     // 用户代理字符串
	RequestMethod  string    `json:"request_method" gorm:"type:varchar(10);index;column:request_method"`                      // HTTP请求方法：GET, POST, PUT, DELETE等
	RequestURI     string    `json:"request_uri" gorm:"type:varchar(500);index:idx_attack_uri_time;index;column:request_uri"` // 请求URI路径
	RequestHeaders string    `json:"request_headers" gorm:"type:text;column:request_headers"`                                 // 请求头信息，JSON格式存储
	RequestBody    string    `json:"request_body" gorm:"type:text;column:request_body"`                                       // 请求体内容
	DomainID       uint      `json:"domain_id" gorm:"index;column:domain_id"`                                                 // 触发的域名ID
	Domain         string    `json:"domain" gorm:"type:varchar(255);index;column:domain"`                                     // 触发的域名
	RuleID         uint      `json:"rule_id" gorm:"index;column:rule_id"`                                                     // 触发的规则ID
	RuleName       string    `json:"rule_name" gorm:"type:varchar(255);index;column:rule_name"`                               // 触发的规则名称
	MatchField     string    `json:"match_field" gorm:"type:varchar(100);column:match_field"`                                 // 匹配的字段名称
	MatchValue     string    `json:"match_value" gorm:"type:text;column:match_value"`                                         // 匹配值
	Action         string    `json:"action" gorm:"type:varchar(50);column:action"`                                            // 执行动作
	ResponseCode   int       `json:"response_code" gorm:"column:response_code"`                                               // 响应状态码
	TenantID       uint      `json:"tenant_id" gorm:"index;column:tenant_id"`                                                 // 租户ID
	CreatedAt      time.Time `json:"created_at" gorm:"index;column:created_at"`                                               // 创建时间
}

// RateLimit 速率限制记录
type RateLimit struct {
	ID           uint      `json:"id" gorm:"primarykey;column:id"`                               // 限流记录ID，主键
	ClientIP     string    `json:"client_ip" gorm:"not null;type:varchar(45);column:client_ip"`  // 客户端IP地址
	RequestCount int       `json:"request_count" gorm:"not null;default:1;column:request_count"` // 请求次数
	WindowStart  time.Time `json:"window_start" gorm:"not null;column:window_start"`             // 时间窗口开始
	TenantID     uint      `json:"tenant_id" gorm:"not null;index;column:tenant_id"`             // 租户ID
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`                          // 创建时间
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`                          // 更新时间
	Tenant       *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                  // 关联的租户信息
}

// Webhook 告警配置
type Webhook struct {
	ID        uint      `json:"id" gorm:"primarykey;column:id"`                                                         // Webhook配置ID，主键
	Name      string    `json:"name" gorm:"not null;type:varchar(255);uniqueIndex:idx_webhook_name_tenant;column:name"` // Webhook名称，在同一租户内唯一
	URL       string    `json:"url" gorm:"not null;type:varchar(500);column:url"`                                       // Webhook回调URL地址
	Method    string    `json:"method" gorm:"not null;type:varchar(10);default:'POST';column:method"`                   // HTTP请求方法，默认POST
	Headers   string    `json:"headers" gorm:"type:text;column:headers"`                                                // 请求头信息，JSON格式存储
	Template  string    `json:"template" gorm:"type:text;column:template"`                                              // 消息模板，支持变量替换
	Events    string    `json:"events" gorm:"type:text;column:events"`                                                  // 触发事件类型列表，JSON数组格式存储
	Enabled   bool      `json:"enabled" gorm:"default:true;index;column:enabled"`                                       // 是否启用
	TenantID  uint      `json:"tenant_id" gorm:"uniqueIndex:idx_webhook_name_tenant;index;column:tenant_id"`            // 所属租户ID，0表示全局配置
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`                                                    // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`                                                    // 更新时间
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`                                            // 关联的租户信息
}

// 数据库迁移
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// 核心实体表
		&Tenant{},
		&User{},
		&Domain{},
		&Policy{},
		&Rule{},
		&BlackList{},
		&WhiteList{},
		// 多对多关联表
		&DomainPolicy{},
		&PolicyRule{},
		&DomainBlackList{},
		&DomainWhiteList{},
		// 业务支撑表
		&AttackLog{},
		&RateLimit{},
		&Webhook{},
	)
}
