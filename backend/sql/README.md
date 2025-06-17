# WAF系统数据库初始化脚本

## 概述

`complete_init_database.sql` 是WAF系统的完整数据库初始化脚本，包含了所有必要的表结构创建和测试数据插入。

## 功能特性

### 表结构
- **租户表** (tenants): 多租户支持
- **用户表** (users): 用户管理和认证
- **域名表** (domain): 简化的域名配置（HTTP/HTTPS + 证书）
- **策略表** (policies): WAF安全策略
- **规则表** (rules): 具体的安全规则
- **黑白名单表** (black_lists, white_lists): IP/URI/User-Agent过滤
- **关联表**: 域名-策略、策略-规则、域名-黑白名单关联
- **日志表** (attack_logs): 攻击日志记录
- **限流表** (rate_limits): 速率限制记录
- **Webhook表** (webhooks): 告警通知配置

### 测试数据
- **6个租户**: 包括全局、默认、电商、博客、API、测试租户
- **5个用户**: 不同角色的测试用户（密码均为 `admin123`）
- **9个域名**: 分布在不同租户的测试域名
- **10条规则**: 常见的WAF安全规则
- **6个策略**: 不同场景的安全策略组合
- **22条关联**: 策略-规则关联配置
- **20条攻击日志**: 模拟的攻击记录
- **黑白名单**: 测试用的IP/URI/User-Agent条目

## 使用方法

### 1. 直接执行脚本
```bash
# 在Docker环境中执行
docker exec -i waf-mysql mysql -u root -pwaf123456 waf < backend/sql/complete_init_database.sql

# 在本地MySQL中执行
mysql -u root -p waf < backend/sql/complete_init_database.sql
```

### 2. 验证初始化结果
脚本执行后会显示统计信息，包括：
- 各表的记录数量
- 租户和用户信息
- 域名配置信息

### 3. 登录测试
使用以下账户登录系统：
- 用户名: `admin`
- 密码: `admin123`
- 角色: 系统管理员

## 域名配置简化说明

相比之前的复杂配置，新的域名表只保留核心字段：
- `protocol`: HTTP或HTTPS协议选择
- `port`: 监听端口（HTTP默认80，HTTPS默认443）
- `ssl_certificate`: SSL证书内容（HTTPS时必需）
- `ssl_private_key`: SSL私钥内容（HTTPS时必需）
- `backend_url`: 后端服务地址
- `tenant_id`: 租户隔离
- `enabled`: 启用状态

## 注意事项

1. **字符集**: 所有表使用 `utf8mb4_unicode_ci` 字符集
2. **外键约束**: 启用了完整的外键约束，确保数据一致性
3. **索引优化**: 为常用查询字段添加了索引
4. **租户隔离**: 所有数据都按租户进行隔离
5. **密码安全**: 用户密码使用bcrypt加密存储

## 清理和重置

如需重新初始化，脚本会自动：
1. 删除所有现有表
2. 重新创建表结构
3. 插入测试数据
4. 显示初始化结果

## 版本信息

- 版本: 2.0 (简化版)
- 创建时间: 2025-06-15
- 适用于: WAF系统域名管理简化版本 