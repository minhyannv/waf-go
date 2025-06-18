-- WAF系统完整数据库初始化脚本
-- 包含表结构创建和测试数据插入
-- 版本: 2.0 (简化版)
-- 创建时间: 2025-06-15

-- 设置正确的字符集
SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci;
SET CHARACTER SET utf8mb4;

-- 使用数据库
USE waf;

-- 禁用外键检查和自动提交
SET FOREIGN_KEY_CHECKS = 0;
SET AUTOCOMMIT = 0;

-- 开始事务
START TRANSACTION;

-- =============================================================================
-- 清理现有数据和表结构
-- =============================================================================

-- 删除所有表（按依赖关系倒序）
DROP TABLE IF EXISTS `webhooks`;
DROP TABLE IF EXISTS `rate_limits`;
DROP TABLE IF EXISTS `attack_logs`;
DROP TABLE IF EXISTS `domain_white_lists`;
DROP TABLE IF EXISTS `domain_black_lists`;
DROP TABLE IF EXISTS `domain_policies`;
DROP TABLE IF EXISTS `policy_rules`;
DROP TABLE IF EXISTS `white_lists`;
DROP TABLE IF EXISTS `black_lists`;
DROP TABLE IF EXISTS `rules`;
DROP TABLE IF EXISTS `policies`;
DROP TABLE IF EXISTS `domains`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `tenants`;

-- =============================================================================
-- 创建表结构
-- =============================================================================

-- 租户表
CREATE TABLE `tenants` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '租户名称',
  `code` varchar(50) NOT NULL COMMENT '租户代码',
  `description` text COMMENT '租户描述',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '租户状态：active(激活), inactive(禁用)',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `code` (`code`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- 用户表
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(100) NOT NULL COMMENT '用户名',
  `password` varchar(255) NOT NULL COMMENT '密码',
  `email` varchar(255) DEFAULT NULL COMMENT '邮箱',
  `role` varchar(50) NOT NULL DEFAULT 'viewer' COMMENT '角色：admin, tenant_admin, viewer',
  `tenant_id` bigint unsigned NOT NULL COMMENT '所属租户ID',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态：active, inactive',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_status` (`status`),
  CONSTRAINT `fk_users_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 域名配置表（简化版）
CREATE TABLE `domains` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `domain` varchar(255) NOT NULL COMMENT '域名',
  `protocol` enum('http','https') NOT NULL DEFAULT 'http' COMMENT '协议类型',
  `port` int NOT NULL DEFAULT '80' COMMENT '端口号',
  `ssl_certificate` text COMMENT 'SSL证书内容',
  `ssl_private_key` text COMMENT 'SSL私钥内容',
  `backend_url` varchar(500) NOT NULL COMMENT '后端服务地址',
  `tenant_id` bigint unsigned NOT NULL COMMENT '租户ID',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `domain` (`domain`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_enabled` (`enabled`),
  CONSTRAINT `fk_domains_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='域名配置表';

-- 策略表
CREATE TABLE `policies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '策略名称',
  `description` text COMMENT '策略描述',
  `tenant_id` bigint unsigned NOT NULL COMMENT '所属租户ID',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_enabled` (`enabled`),
  CONSTRAINT `fk_policies_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='策略表';

-- 规则表
CREATE TABLE `rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '规则名称',
  `description` text COMMENT '规则描述',
  `match_type` varchar(50) NOT NULL COMMENT '匹配类型：uri, ip, header, body, user_agent',
  `pattern` text NOT NULL COMMENT '匹配模式，具体的匹配规则内容',
  `match_mode` varchar(50) NOT NULL COMMENT '匹配模式：exact, regex, contains',
  `action` varchar(50) NOT NULL COMMENT '动作：block, allow, log',
  `response_code` int NOT NULL DEFAULT '403' COMMENT '阻断时返回的HTTP状态码',
  `response_msg` text COMMENT '阻断时返回的消息内容',
  `priority` int NOT NULL DEFAULT '1' COMMENT '规则优先级',
  `tenant_id` bigint unsigned NOT NULL COMMENT '所属租户ID',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_rule_name_tenant` (`name`,`tenant_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_enabled` (`enabled`),
  KEY `idx_match_type` (`match_type`),
  KEY `idx_action` (`action`),
  KEY `idx_priority` (`priority`),
  CONSTRAINT `fk_rules_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则表';

-- 黑名单表
CREATE TABLE `black_lists` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(50) NOT NULL COMMENT '类型：ip, uri, user_agent',
  `value` varchar(500) NOT NULL COMMENT '值',
  `comment` text COMMENT '备注',
  `tenant_id` bigint unsigned NOT NULL COMMENT '所属租户ID',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_blacklist_type_value_tenant` (`type`,`value`,`tenant_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_enabled` (`enabled`),
  CONSTRAINT `fk_black_lists_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='黑名单表';

-- 白名单表
CREATE TABLE `white_lists` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(50) NOT NULL COMMENT '类型：ip, uri, user_agent',
  `value` varchar(500) NOT NULL COMMENT '值',
  `comment` text COMMENT '备注',
  `tenant_id` bigint unsigned NOT NULL COMMENT '所属租户ID',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_whitelist_type_value_tenant` (`type`,`value`,`tenant_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_enabled` (`enabled`),
  CONSTRAINT `fk_white_lists_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='白名单表';

-- 域名策略关联表 - 域名(Domain) ↔ 策略(Policy): 多对多
CREATE TABLE IF NOT EXISTS `domain_policies` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID，主键',
    `domain_id` bigint unsigned NOT NULL COMMENT '域名ID',
    `policy_id` bigint unsigned NOT NULL COMMENT '策略ID',
    `priority` int NOT NULL DEFAULT '1' COMMENT '策略在该域名下的优先级，数字越大优先级越高',
    `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用此关联',
    `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
    `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_domain_policy` (`domain_id`,`policy_id`),
    KEY `idx_domain_policies_enabled` (`enabled`),
    KEY `fk_domain_policies_domain_id` (`domain_id`),
    KEY `fk_domain_policies_policy_id` (`policy_id`),
    CONSTRAINT `fk_domain_policies_domain_id` FOREIGN KEY (`domain_id`) REFERENCES `domains` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_domain_policies_policy_id` FOREIGN KEY (`policy_id`) REFERENCES `policies` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='域名策略关联表';

-- 策略规则关联表 - 策略(Policy) ↔ 规则(Rule): 多对多
CREATE TABLE IF NOT EXISTS `policy_rules` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID，主键',
    `policy_id` bigint unsigned NOT NULL COMMENT '策略ID',
    `rule_id` bigint unsigned NOT NULL COMMENT '规则ID',
    `priority` int NOT NULL DEFAULT '1' COMMENT '规则在该策略下的优先级，数字越大优先级越高',
    `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用此关联',
    `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
    `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_policy_rule` (`policy_id`,`rule_id`),
    KEY `idx_policy_rules_enabled` (`enabled`),
    KEY `fk_policy_rules_policy_id` (`policy_id`),
    KEY `fk_policy_rules_rule_id` (`rule_id`),
    CONSTRAINT `fk_policy_rules_policy_id` FOREIGN KEY (`policy_id`) REFERENCES `policies` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_policy_rules_rule_id` FOREIGN KEY (`rule_id`) REFERENCES `rules` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='策略规则关联表';

-- 攻击日志表
CREATE TABLE `attack_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `domain_id` bigint unsigned DEFAULT NULL COMMENT '域名ID',
  `domain` varchar(255) DEFAULT NULL COMMENT '攻击时的域名',
  `rule_id` bigint unsigned DEFAULT NULL COMMENT '规则ID',
  `rule_name` varchar(255) DEFAULT NULL COMMENT '触发的规则名称',
  `client_ip` varchar(45) NOT NULL COMMENT '客户端IP',
  `user_agent` text COMMENT 'User-Agent',
  `request_method` varchar(10) NOT NULL COMMENT '请求方法',
  `request_uri` varchar(1000) NOT NULL COMMENT '请求URI',
  `request_headers` text COMMENT '请求头',
  `request_body` text COMMENT '请求体',
  `match_type` varchar(50) NOT NULL COMMENT '匹配类型',
  `match_field` varchar(100) DEFAULT NULL COMMENT '匹配的字段名称',
  `match_value` text NOT NULL COMMENT '匹配值',
  `action` varchar(50) NOT NULL COMMENT '执行动作',
  `response_code` int NOT NULL COMMENT '响应状态码',
  `tenant_id` bigint unsigned NOT NULL COMMENT '租户ID',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_client_ip` (`client_ip`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_domain_id` (`domain_id`),
  KEY `idx_rule_id` (`rule_id`),
  KEY `idx_rule_name` (`rule_name`),
  KEY `idx_domain` (`domain`),
  CONSTRAINT `fk_attack_logs_domain` FOREIGN KEY (`domain_id`) REFERENCES `domains` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_attack_logs_rule` FOREIGN KEY (`rule_id`) REFERENCES `rules` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_attack_logs_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='攻击日志表';

-- 速率限制表
CREATE TABLE `rate_limits` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `client_ip` varchar(45) NOT NULL COMMENT '客户端IP',
  `request_count` int NOT NULL DEFAULT '1' COMMENT '请求次数',
  `window_start` datetime(3) NOT NULL COMMENT '时间窗口开始',
  `tenant_id` bigint unsigned NOT NULL COMMENT '租户ID',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ip_window_tenant` (`client_ip`,`window_start`,`tenant_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  CONSTRAINT `fk_rate_limits_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='速率限制表';

-- Webhook配置表
CREATE TABLE `webhooks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT 'Webhook名称',
  `url` varchar(500) NOT NULL COMMENT 'Webhook URL',
  `method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT 'HTTP方法',
  `headers` text COMMENT '请求头（JSON格式）',
  `events` text NOT NULL COMMENT '触发事件（JSON数组）',
  `tenant_id` bigint unsigned NOT NULL COMMENT '租户ID',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_enabled` (`enabled`),
  CONSTRAINT `fk_webhooks_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Webhook配置表';

-- =============================================================================
-- 插入测试数据
-- =============================================================================

-- 插入租户数据
INSERT INTO tenants (id, name, code, description, status, created_at, updated_at) VALUES
(1, '全局租户', 'global', '系统全局租户，用于全局配置', 'active', NOW(), NOW()),
(2, '默认租户', 'default', '默认租户，用于一般用户', 'active', NOW(), NOW()),
(3, '电商平台', 'ecommerce', '电商平台租户', 'active', NOW(), NOW()),
(4, '博客网站', 'blog', '博客网站租户', 'active', NOW(), NOW()),
(5, 'API服务', 'api', 'API服务租户', 'active', NOW(), NOW()),
(6, '测试环境', 'test', '测试环境租户', 'inactive', NOW(), NOW());

-- 插入用户数据（密码都是admin123的bcrypt哈希）
INSERT INTO users (id, username, password, email, role, tenant_id, status, created_at, updated_at) VALUES
(1, 'admin', '$2a$10$Osar2M.FuYjIa6bt2Zh1SOUIgPshIsacwGZxUTZfNB5QFZiIbI6XS', 'admin@waf.com', 'admin', 2, 'active', NOW(), NOW()),
(2, 'tenant_admin_shop', '$2a$10$Osar2M.FuYjIa6bt2Zh1SOUIgPshIsacwGZxUTZfNB5QFZiIbI6XS', 'shop@waf.com', 'tenant_admin', 3, 'active', NOW(), NOW()),
(3, 'tenant_admin_blog', '$2a$10$Osar2M.FuYjIa6bt2Zh1SOUIgPshIsacwGZxUTZfNB5QFZiIbI6XS', 'blog@waf.com', 'tenant_admin', 4, 'active', NOW(), NOW()),
(4, 'viewer_api', '$2a$10$Osar2M.FuYjIa6bt2Zh1SOUIgPshIsacwGZxUTZfNB5QFZiIbI6XS', 'api@waf.com', 'viewer', 5, 'active', NOW(), NOW()),
(5, 'test_user', '$2a$10$Osar2M.FuYjIa6bt2Zh1SOUIgPshIsacwGZxUTZfNB5QFZiIbI6XS', 'test@waf.com', 'viewer', 6, 'inactive', NOW(), NOW());

-- 插入域名配置数据
INSERT INTO domains (id, domain, protocol, port, backend_url, tenant_id, enabled, created_at, updated_at) VALUES
-- 默认租户域名
(1, 'www.example.com', 'http', 80, 'http://localhost:3000', 2, 1, NOW(), NOW()),
(2, 'secure.example.com', 'https', 443, 'http://localhost:3000', 2, 1, NOW(), NOW()),

-- 电商平台域名
(3, 'shop.example.com', 'https', 443, 'http://localhost:4000', 3, 1, NOW(), NOW()),
(4, 'shop-api.example.com', 'https', 443, 'http://localhost:4001', 3, 1, NOW(), NOW()),

-- 博客网站域名
(5, 'blog.example.com', 'http', 80, 'http://localhost:5000', 4, 1, NOW(), NOW()),
(6, 'blog-secure.example.com', 'https', 443, 'http://localhost:5001', 4, 1, NOW(), NOW()),

-- API服务域名
(7, 'api.example.com', 'https', 443, 'http://localhost:6000', 5, 1, NOW(), NOW()),
(8, 'api-v2.example.com', 'https', 443, 'http://localhost:6001', 5, 1, NOW(), NOW()),

-- 测试环境域名
(9, 'test.example.com', 'http', 80, 'http://localhost:7000', 6, 0, NOW(), NOW());

-- 插入WAF规则数据
INSERT INTO rules (id, name, description, match_type, pattern, match_mode, action, response_code, response_msg, priority, tenant_id, enabled, created_at, updated_at) VALUES
(1, 'SQL注入防护', '防止SQL注入攻击', 'body', '(?i)(union|select|insert|update|delete|drop|create|alter|exec|script)', 'regex', 'block', 403, '检测到SQL注入攻击', 10, 1, 1, NOW(), NOW()),
(2, 'XSS防护', '防止跨站脚本攻击', 'body', '(?i)(<script|javascript:|on\\w+\\s*=)', 'regex', 'block', 403, '检测到XSS攻击', 9, 1, 1, NOW(), NOW()),
(3, '管理员路径保护', '保护管理员访问路径', 'uri', '/admin', 'contains', 'block', 403, '禁止访问管理员路径', 8, 2, 1, NOW(), NOW()),
(4, '敏感文件保护', '保护敏感文件访问', 'uri', '\\.(env|config|ini|log)$', 'regex', 'block', 403, '禁止访问敏感文件', 7, 2, 1, NOW(), NOW()),
(5, '恶意爬虫拦截', '拦截恶意爬虫', 'user_agent', 'bot', 'contains', 'block', 403, '检测到恶意爬虫', 6, 3, 1, NOW(), NOW()),
(6, 'API频率限制', 'API访问频率限制', 'uri', '/api/', 'contains', 'log', 200, '', 5, 3, 1, NOW(), NOW()),
(7, '博客垃圾评论防护', '防止垃圾评论', 'body', 'spam', 'contains', 'block', 403, '检测到垃圾内容', 4, 4, 1, NOW(), NOW()),
(8, '文件上传保护', '限制文件上传类型', 'uri', '/upload', 'contains', 'log', 200, '', 3, 4, 1, NOW(), NOW()),
(9, 'API密钥验证', 'API密钥验证规则', 'header', 'X-API-Key', 'contains', 'allow', 200, '', 2, 5, 1, NOW(), NOW()),
(10, '测试规则', '测试环境规则', 'uri', '/test', 'contains', 'log', 200, '', 1, 6, 0, NOW(), NOW());

-- 插入安全策略数据
INSERT INTO policies (id, name, description, tenant_id, enabled, created_at, updated_at) VALUES
(1, '全局基础安全策略', '包含基础的SQL注入和XSS防护', 1, 1, NOW(), NOW()),
(2, '电商平台安全策略', '电商平台专用安全策略', 3, 1, NOW(), NOW()),
(3, '博客网站安全策略', '博客网站专用安全策略', 4, 1, NOW(), NOW()),
(4, 'API服务安全策略', 'API服务专用安全策略', 5, 1, NOW(), NOW()),
(5, '严格安全策略', '高安全级别策略', 2, 1, NOW(), NOW()),
(6, '测试环境策略', '测试环境专用策略', 6, 0, NOW(), NOW());

-- 插入策略规则关联数据
INSERT INTO policy_rules (policy_id, rule_id, priority, enabled, created_at, updated_at) VALUES
-- 全局基础安全策略
(1, 1, 10, 1, NOW(), NOW()), -- SQL注入防护
(1, 2, 9, 1, NOW(), NOW()),  -- XSS防护

-- 电商平台安全策略
(2, 1, 10, 1, NOW(), NOW()), -- SQL注入防护
(2, 2, 9, 1, NOW(), NOW()),  -- XSS防护
(2, 3, 8, 1, NOW(), NOW()),  -- 管理员路径保护
(2, 4, 7, 1, NOW(), NOW()),  -- 敏感文件保护
(2, 5, 6, 1, NOW(), NOW()),  -- 恶意爬虫拦截
(2, 6, 5, 1, NOW(), NOW()),  -- API频率限制

-- 博客网站安全策略
(3, 1, 10, 1, NOW(), NOW()), -- SQL注入防护
(3, 2, 9, 1, NOW(), NOW()),  -- XSS防护
(3, 7, 8, 1, NOW(), NOW()),  -- 博客垃圾评论防护
(3, 8, 7, 1, NOW(), NOW()),  -- 文件上传保护

-- API服务安全策略
(4, 1, 10, 1, NOW(), NOW()), -- SQL注入防护
(4, 2, 9, 1, NOW(), NOW()),  -- XSS防护
(4, 6, 8, 1, NOW(), NOW()),  -- API频率限制
(4, 9, 7, 1, NOW(), NOW()),  -- API密钥验证

-- 严格安全策略
(5, 1, 10, 1, NOW(), NOW()), -- SQL注入防护
(5, 2, 9, 1, NOW(), NOW()),  -- XSS防护
(5, 3, 8, 1, NOW(), NOW()),  -- 管理员路径保护
(5, 4, 7, 1, NOW(), NOW()),  -- 敏感文件保护
(5, 5, 6, 1, NOW(), NOW()),  -- 恶意爬虫拦截

-- 测试环境策略
(6, 10, 5, 0, NOW(), NOW()); -- 测试规则

-- 插入域名策略关联数据
INSERT INTO domain_policies (domain_id, policy_id, priority, enabled, created_at, updated_at) VALUES
-- www.example.com 应用全局基础安全策略
(1, 1, 1, 1, NOW(), NOW()),

-- secure.example.com 应用全局基础安全策略和严格安全策略
(2, 1, 1, 1, NOW(), NOW()),
(2, 5, 1, 1, NOW(), NOW()),

-- shop.example.com 应用电商平台安全策略
(3, 2, 1, 1, NOW(), NOW()),

-- shop-api.example.com 应用电商平台安全策略
(4, 2, 1, 1, NOW(), NOW()),

-- blog.example.com 应用博客网站安全策略
(5, 3, 1, 1, NOW(), NOW()),

-- blog-secure.example.com 应用博客网站安全策略
(6, 3, 1, 1, NOW(), NOW()),

-- api.example.com 应用API服务安全策略
(7, 4, 1, 1, NOW(), NOW()),

-- api-v2.example.com 应用API服务安全策略
(8, 4, 1, 1, NOW(), NOW()),

-- test.example.com 应用测试环境策略
(9, 6, 1, 1, NOW(), NOW());

-- 插入攻击日志数据
INSERT INTO attack_logs (id, domain_id, domain, rule_id, rule_name, client_ip, user_agent, request_method, request_uri, request_headers, request_body, match_type, match_field, match_value, action, response_code, tenant_id, created_at) VALUES
(1, 1, 'www.example.com', 1, 'SQL注入防护', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', 'POST', '/login', '{"Content-Type": "application/json"}', '{"username": "admin", "password": "1\' OR \'1\'=\'1"}', 'body', 'request_body', 'union|select', 'block', 403, 2, DATE_SUB(NOW(), INTERVAL 1 HOUR)),
(2, 1, 'www.example.com', 2, 'XSS防护', '192.168.1.101', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', 'GET', '/search?q=<script>alert(1)</script>', '{"User-Agent": "Mozilla/5.0"}', '', 'uri', 'request_uri', '<script', 'block', 403, 2, DATE_SUB(NOW(), INTERVAL 2 HOUR)),
(3, 3, 'shop.example.com', 3, '管理员路径保护', '10.0.0.50', 'curl/7.68.0', 'GET', '/admin/users', '{"Authorization": "Bearer token"}', '', 'uri', 'request_uri', '/admin', 'block', 403, 3, DATE_SUB(NOW(), INTERVAL 3 HOUR)),
(4, 3, 'shop.example.com', 5, '恶意爬虫拦截', '203.0.113.10', 'Googlebot/2.1', 'GET', '/products', '{"User-Agent": "Googlebot/2.1"}', '', 'user_agent', 'user_agent', 'bot', 'block', 403, 3, DATE_SUB(NOW(), INTERVAL 4 HOUR)),
(5, 5, 'blog.example.com', 7, '博客垃圾评论防护', '198.51.100.20', 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0)', 'POST', '/comments', '{"Content-Type": "application/json"}', '{"content": "This is spam content"}', 'body', 'request_body', 'spam', 'block', 403, 4, DATE_SUB(NOW(), INTERVAL 5 HOUR)),
(6, 7, 'api.example.com', 6, 'API频率限制', '172.16.0.100', 'PostmanRuntime/7.28.0', 'GET', '/api/v1/data', '{"X-API-Key": "invalid"}', '', 'uri', 'request_uri', '/api/', 'log', 200, 5, DATE_SUB(NOW(), INTERVAL 6 HOUR)),
(7, 1, 'www.example.com', 1, 'SQL注入防护', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', 'POST', '/register', '{"Content-Type": "application/json"}', '{"email": "test@test.com", "sql": "DROP TABLE users"}', 'body', 'request_body', 'drop', 'block', 403, 2, DATE_SUB(NOW(), INTERVAL 7 HOUR)),
(8, 2, 'secure.example.com', 2, 'XSS防护', '192.168.1.102', 'Mozilla/5.0 (Linux; Android 10)', 'GET', '/profile?name=<img src=x onerror=alert(1)>', '{}', '', 'uri', 'request_uri', 'onerror', 'block', 403, 2, DATE_SUB(NOW(), INTERVAL 8 HOUR)),
(9, 4, 'shop-api.example.com', 6, 'API频率限制', '10.0.0.51', 'axios/0.21.1', 'POST', '/api/orders', '{"Content-Type": "application/json"}', '{"product_id": 123}', 'uri', 'request_uri', '/api/', 'log', 201, 3, DATE_SUB(NOW(), INTERVAL 9 HOUR)),
(10, 6, 'test.example.com', 8, '文件上传保护', '198.51.100.21', 'Mozilla/5.0 (Windows NT 10.0)', 'POST', '/upload', '{"Content-Type": "multipart/form-data"}', 'file upload data', 'uri', 'request_uri', '/upload', 'log', 200, 4, DATE_SUB(NOW(), INTERVAL 10 HOUR)),
(11, 1, 'www.example.com', 1, 'SQL注入防护', '203.0.113.11', 'sqlmap/1.4.7', 'GET', '/news?id=1 UNION SELECT password FROM users', '{}', '', 'uri', 'request_uri', 'union', 'block', 403, 2, DATE_SUB(NOW(), INTERVAL 11 HOUR)),
(12, 3, 'shop.example.com', 4, '敏感文件保护', '172.16.0.101', 'wget/1.20.3', 'GET', '/.env', '{}', '', 'uri', 'request_uri', '.env', 'block', 403, 3, DATE_SUB(NOW(), INTERVAL 12 HOUR)),
(13, 5, 'blog.example.com', 2, 'XSS防护', '192.168.1.103', 'Mozilla/5.0 (compatible; MSIE 9.0)', 'POST', '/contact', '{"Content-Type": "application/x-www-form-urlencoded"}', 'message=<script>document.cookie</script>', 'body', 'request_body', '<script', 'block', 403, 4, DATE_SUB(NOW(), INTERVAL 13 HOUR)),
(14, 7, 'api.example.com', 9, 'API密钥验证', '10.0.0.52', 'Python-requests/2.25.1', 'GET', '/api/v2/users', '{"X-API-Key": "valid-key-123"}', '', 'header', 'request_headers', 'X-API-Key', 'allow', 200, 5, DATE_SUB(NOW(), INTERVAL 14 HOUR)),
(15, 1, 'www.example.com', 2, 'XSS防护', '198.51.100.22', 'Mozilla/5.0 (X11; Linux x86_64)', 'GET', '/search?q=javascript:alert(document.domain)', '{}', '', 'uri', 'request_uri', 'javascript:', 'block', 403, 2, DATE_SUB(NOW(), INTERVAL 15 HOUR)),
(16, 3, 'shop.example.com', 5, '恶意爬虫拦截', '203.0.113.12', 'Baiduspider/2.0', 'GET', '/sitemap.xml', '{}', '', 'user_agent', 'user_agent', 'spider', 'block', 403, 3, DATE_SUB(NOW(), INTERVAL 16 HOUR)),
(17, 4, 'shop-api.example.com', 6, 'API频率限制', '172.16.0.102', 'okhttp/4.9.0', 'GET', '/api/products?limit=100', '{}', '', 'uri', 'request_uri', '/api/', 'log', 200, 3, DATE_SUB(NOW(), INTERVAL 17 HOUR)),
(18, 6, 'test.example.com', 7, '博客垃圾评论防护', '192.168.1.104', 'Mozilla/5.0 (iPad; CPU OS 14_0)', 'POST', '/comments', '{"Content-Type": "application/json"}', '{"text": "Buy cheap products now! spam spam spam"}', 'body', 'request_body', 'spam', 'block', 403, 4, DATE_SUB(NOW(), INTERVAL 18 HOUR)),
(19, 8, 'api-v2.example.com', 9, 'API密钥验证', '10.0.0.53', 'Go-http-client/1.1', 'POST', '/api/v2/auth', '{"X-API-Key": "test-key-456"}', '{"username": "api_user"}', 'header', 'request_headers', 'X-API-Key', 'allow', 200, 5, DATE_SUB(NOW(), INTERVAL 19 HOUR)),
(20, 2, 'secure.example.com', 1, 'SQL注入防护', '198.51.100.23', 'Mozilla/5.0 (compatible; Bingbot/2.0)', 'GET', '/admin?cmd=cat /etc/passwd', '{}', '', 'uri', 'request_uri', 'cat', 'block', 403, 2, DATE_SUB(NOW(), INTERVAL 20 HOUR));

-- 插入速率限制记录
INSERT INTO rate_limits (id, client_ip, request_count, window_start, tenant_id, created_at, updated_at) VALUES
(1, '192.168.1.100', 45, DATE_SUB(NOW(), INTERVAL 1 MINUTE), 2, NOW(), NOW()),
(2, '10.0.0.50', 120, DATE_SUB(NOW(), INTERVAL 1 MINUTE), 3, NOW(), NOW()),
(3, '198.51.100.20', 30, DATE_SUB(NOW(), INTERVAL 1 MINUTE), 4, NOW(), NOW()),
(4, '172.16.0.100', 200, DATE_SUB(NOW(), INTERVAL 1 MINUTE), 5, NOW(), NOW()),
(5, '203.0.113.10', 15, DATE_SUB(NOW(), INTERVAL 1 MINUTE), 3, NOW(), NOW());

-- 插入黑名单数据
INSERT INTO black_lists (id, type, value, comment, tenant_id, enabled, created_at, updated_at) VALUES
(1, 'ip', '192.168.1.100', '恶意IP地址', 1, 1, NOW(), NOW()),
(2, 'ip', '10.0.0.50', '可疑IP地址', 2, 1, NOW(), NOW()),
(3, 'uri', '/admin/delete', '危险管理操作', 2, 1, NOW(), NOW()),
(4, 'uri', '/system/config', '系统配置路径', 3, 1, NOW(), NOW()),
(5, 'user_agent', 'sqlmap', 'SQL注入工具', 1, 1, NOW(), NOW()),
(6, 'user_agent', 'nikto', '漏洞扫描工具', 1, 1, NOW(), NOW()),
(7, 'ip', '203.0.113.0/24', '恶意IP段', 3, 1, NOW(), NOW()),
(8, 'uri', '/.git/', 'Git仓库路径', 4, 1, NOW(), NOW()),
(9, 'user_agent', 'masscan', '端口扫描工具', 4, 1, NOW(), NOW()),
(10, 'ip', '198.51.100.0/24', '测试IP段', 5, 1, NOW(), NOW()),
(11, 'uri', '/debug', '调试路径', 6, 0, NOW(), NOW());

-- 插入白名单数据
INSERT INTO white_lists (id, type, value, comment, tenant_id, enabled, created_at, updated_at) VALUES
(1, 'ip', '127.0.0.1', '本地回环地址', 1, 1, NOW(), NOW()),
(2, 'ip', '172.16.0.0/16', '内网IP段', 1, 1, NOW(), NOW()),
(3, 'uri', '/api/health', '健康检查接口', 2, 1, NOW(), NOW()),
(4, 'uri', '/public/', '公共资源路径', 2, 1, NOW(), NOW()),
(5, 'user_agent', 'Googlebot', '谷歌爬虫', 3, 1, NOW(), NOW()),
(6, 'user_agent', 'Bingbot', '必应爬虫', 3, 1, NOW(), NOW()),
(7, 'ip', '10.0.0.0/8', '私有网络', 4, 1, NOW(), NOW()),
(8, 'uri', '/robots.txt', '爬虫协议文件', 4, 1, NOW(), NOW()),
(9, 'user_agent', 'curl', 'curl工具', 5, 1, NOW(), NOW()),
(10, 'ip', '192.168.0.0/16', '局域网段', 5, 1, NOW(), NOW()),
(11, 'uri', '/test/', '测试路径', 6, 0, NOW(), NOW());

-- 插入Webhook配置数据
INSERT INTO webhooks (id, name, url, method, headers, events, tenant_id, enabled, created_at, updated_at) VALUES
(1, 'Slack告警', 'https://hooks.slack.com/services/xxx/yyy/zzz', 'POST', '{"Content-Type": "application/json"}', '["attack_blocked", "rate_limit_exceeded"]', 2, 1, NOW(), NOW()),
(2, '邮件通知', 'https://api.mailgun.com/v3/domain/messages', 'POST', '{"Authorization": "Basic xxx"}', '["attack_blocked"]', 3, 1, NOW(), NOW()),
(3, '业务系统集成', 'https://api.business.com/waf/alerts', 'POST', '{"X-API-Key": "business-key"}', '["attack_blocked", "policy_triggered"]', 4, 1, NOW(), NOW()),
(4, '监控系统', 'https://monitoring.example.com/webhooks/waf', 'POST', '{"Authorization": "Bearer monitor-token"}', '["attack_blocked", "rate_limit_exceeded", "rule_triggered"]', 5, 1, NOW(), NOW()),
(5, '测试Webhook', 'https://httpbin.org/post', 'POST', '{"Content-Type": "application/json"}', '["test_event"]', 6, 0, NOW(), NOW());

-- 重新启用外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- 提交事务
COMMIT;

-- 重新启用自动提交
SET AUTOCOMMIT = 1;

-- =============================================================================
-- 数据统计和验证
-- =============================================================================

-- 显示初始化结果
SELECT '=== WAF系统数据库初始化完成 ===' as status;
SELECT '租户数量' as item, COUNT(*) as count FROM tenants
UNION ALL
SELECT '用户数量' as item, COUNT(*) as count FROM users
UNION ALL
SELECT '域名数量' as item, COUNT(*) as count FROM domains
UNION ALL
SELECT '策略数量' as item, COUNT(*) as count FROM policies
UNION ALL
SELECT '规则数量' as item, COUNT(*) as count FROM rules
UNION ALL
SELECT '黑名单数量' as item, COUNT(*) as count FROM black_lists
UNION ALL
SELECT '白名单数量' as item, COUNT(*) as count FROM white_lists
UNION ALL
SELECT '攻击日志数量' as item, COUNT(*) as count FROM attack_logs
UNION ALL
SELECT '速率限制记录' as item, COUNT(*) as count FROM rate_limits
UNION ALL
SELECT 'Webhook配置' as item, COUNT(*) as count FROM webhooks
UNION ALL
SELECT '策略规则关联' as item, COUNT(*) as count FROM policy_rules
UNION ALL
SELECT '域名策略关联' as item, COUNT(*) as count FROM domain_policies;

-- 显示租户和用户信息
SELECT '=== 租户和用户信息 ===' as info;
SELECT t.name as tenant_name, u.username, u.role, u.status 
FROM tenants t 
LEFT JOIN users u ON t.id = u.tenant_id 
ORDER BY t.id, u.id;

-- 显示域名配置信息
SELECT '=== 域名配置信息 ===' as info;
SELECT d.domain, d.protocol, d.port, d.backend_url, t.name as tenant_name, d.enabled
FROM domains d
JOIN tenants t ON d.tenant_id = t.id
ORDER BY d.id;

SELECT '=== 初始化完成，系统可以正常使用 ===' as final_status; 