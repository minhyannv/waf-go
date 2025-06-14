-- 插入测试数据脚本（不指定ID，让MySQL自动分配）
-- 注意：按照外键依赖顺序插入数据

USE waf;

-- 禁用外键检查
SET FOREIGN_KEY_CHECKS = 0;

-- 清空所有表数据（按照外键依赖的逆序）
DELETE FROM webhooks;
DELETE FROM black_lists;
DELETE FROM white_lists;
DELETE FROM rate_limits;
DELETE FROM attack_logs;
DELETE FROM policies;
DELETE FROM rules;
DELETE FROM users;
DELETE FROM tenants;

-- 重置所有表的AUTO_INCREMENT
ALTER TABLE tenants AUTO_INCREMENT = 1;
ALTER TABLE users AUTO_INCREMENT = 1;
ALTER TABLE rules AUTO_INCREMENT = 1;
ALTER TABLE policies AUTO_INCREMENT = 1;
ALTER TABLE attack_logs AUTO_INCREMENT = 1;
ALTER TABLE rate_limits AUTO_INCREMENT = 1;
ALTER TABLE white_lists AUTO_INCREMENT = 1;
ALTER TABLE black_lists AUTO_INCREMENT = 1;
ALTER TABLE webhooks AUTO_INCREMENT = 1;

-- 重新启用外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- 1. 插入租户数据（让MySQL自动分配ID）
INSERT INTO tenants (name, domain, status, created_at, updated_at) VALUES
('全局租户', '', 'active', NOW(), NOW()),
('默认租户', 'default.com', 'active', NOW(), NOW()),
('电商平台', 'shop.example.com', 'active', NOW(), NOW()),
('博客网站', 'blog.example.com', 'active', NOW(), NOW()),
('API服务', 'api.example.com', 'active', NOW(), NOW()),
('测试环境', 'test.example.com', 'inactive', NOW(), NOW());

-- 2. 插入用户数据（依赖租户表）
-- 注意：所有用户的默认密码都是 admin123
INSERT INTO users (username, password, email, role, tenant_id, status, created_at, updated_at) VALUES
('admin', '$2a$10$n2yWetLC/dp9YQ2UxIyWG.jtLYrPybUi.QZLwuxD9Hp.MmulSj0BK', 'admin@example.com', 'admin', 2, 'active', NOW(), NOW()),
('tenant_admin_shop', '$2a$10$n2yWetLC/dp9YQ2UxIyWG.jtLYrPybUi.QZLwuxD9Hp.MmulSj0BK', 'admin@shop.example.com', 'tenant_admin', 3, 'active', NOW(), NOW()),
('tenant_admin_blog', '$2a$10$n2yWetLC/dp9YQ2UxIyWG.jtLYrPybUi.QZLwuxD9Hp.MmulSj0BK', 'admin@blog.example.com', 'tenant_admin', 4, 'active', NOW(), NOW()),
('viewer_api', '$2a$10$n2yWetLC/dp9YQ2UxIyWG.jtLYrPybUi.QZLwuxD9Hp.MmulSj0BK', 'viewer@api.example.com', 'viewer', 5, 'active', NOW(), NOW()),
('test_user', '$2a$10$n2yWetLC/dp9YQ2UxIyWG.jtLYrPybUi.QZLwuxD9Hp.MmulSj0BK', 'test@test.example.com', 'viewer', 6, 'inactive', NOW(), NOW());

-- 3. 插入WAF规则数据（依赖租户表）
INSERT INTO rules (name, description, match_type, pattern, match_mode, action, response_code, response_msg, priority, enabled, tenant_id, created_at, updated_at) VALUES
-- 全局规则（tenant_id = 1，全局租户）
('SQL注入防护', '检测常见的SQL注入攻击模式', 'body', '(union|select|insert|update|delete|drop|create|alter)\\s+', 'regex', 'block', 403, 'SQL注入攻击被阻止', 10, true, 1, NOW(), NOW()),
('XSS防护', '检测跨站脚本攻击', 'body', '<script[^>]*>.*?</script>', 'regex', 'block', 403, 'XSS攻击被阻止', 9, true, 1, NOW(), NOW()),
('管理员路径保护', '保护管理员访问路径', 'uri', '/admin', 'contains', 'block', 403, '禁止访问管理员页面', 8, true, 1, NOW(), NOW()),
('恶意爬虫拦截', '拦截恶意爬虫和扫描器', 'user_agent', '(sqlmap|nmap|nikto|dirb|gobuster)', 'regex', 'block', 403, '恶意爬虫被拦截', 7, true, 1, NOW(), NOW()),

-- 电商平台专用规则
('购物车保护', '保护购物车相关接口', 'uri', '/cart', 'contains', 'log', 200, '', 6, true, 3, NOW(), NOW()),
('支付接口保护', '保护支付相关接口', 'uri', '/payment', 'contains', 'block', 403, '支付接口访问受限', 9, true, 3, NOW(), NOW()),

-- 博客网站专用规则
('评论垃圾信息过滤', '过滤评论中的垃圾信息', 'body', '(viagra|casino|poker|lottery)', 'regex', 'block', 403, '垃圾评论被拦截', 5, true, 4, NOW(), NOW()),
('文章编辑保护', '保护文章编辑功能', 'uri', '/edit', 'contains', 'log', 200, '', 4, true, 4, NOW(), NOW()),

-- API服务专用规则
('API频率限制', '限制API访问频率', 'uri', '/api/v1', 'contains', 'log', 200, '', 3, true, 5, NOW(), NOW()),
('API密钥验证', '验证API密钥格式', 'header', 'X-API-Key', 'contains', 'log', 200, '', 2, true, 5, NOW(), NOW());

-- 4. 插入策略数据（依赖租户表和规则表）
INSERT INTO policies (name, description, domain, rule_ids, enabled, tenant_id, created_at, updated_at) VALUES
('全局安全策略', '适用于所有域名的基础安全策略', '', '[1,2,3,4]', true, 1, NOW(), NOW()),
('电商平台安全策略', '电商网站专用安全策略', 'shop.example.com', '[1,2,3,4,5,6]', true, 3, NOW(), NOW()),
('博客网站安全策略', '博客网站专用安全策略', 'blog.example.com', '[1,2,3,4,7,8]', true, 4, NOW(), NOW()),
('API服务安全策略', 'API服务专用安全策略', 'api.example.com', '[1,2,9,10]', true, 5, NOW(), NOW()),
('测试环境策略', '测试环境使用的策略', 'test.example.com', '[1,2]', false, 6, NOW(), NOW());

-- 5. 插入攻击日志数据（依赖租户表和规则表）
INSERT INTO attack_logs (request_id, client_ip, user_agent, request_method, request_uri, request_headers, request_body, rule_id, rule_name, match_field, match_value, action, response_code, tenant_id, created_at) VALUES
-- 最近1小时的攻击日志（用于实时监控）
('req_001', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'POST', '/login', '{"Content-Type":"application/json"}', '{"username":"admin","password":"admin\' OR 1=1--"}', 1, 'SQL注入防护', 'body', 'OR 1=1--', 'block', 403, 3, DATE_SUB(NOW(), INTERVAL 15 MINUTE)),
('req_002', '10.0.0.50', 'sqlmap/1.6.12', 'GET', '/admin/users', '{"User-Agent":"sqlmap/1.6.12"}', '', 4, '恶意爬虫拦截', 'user_agent', 'sqlmap', 'block', 403, 3, DATE_SUB(NOW(), INTERVAL 25 MINUTE)),
('req_003', '172.16.0.80', 'Mozilla/5.0 (compatible; Baiduspider/2.0)', 'GET', '/admin/config', '{"User-Agent":"Mozilla/5.0 (compatible; Baiduspider/2.0)"}', '', 3, '管理员路径保护', 'uri', '/admin', 'block', 403, 4, DATE_SUB(NOW(), INTERVAL 35 MINUTE)),
('req_004', '203.0.113.45', 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36', 'POST', '/comment', '{"Content-Type":"application/json"}', '{"content":"<script>alert(\"XSS\")</script>"}', 2, 'XSS防护', 'body', '<script>', 'block', 403, 4, DATE_SUB(NOW(), INTERVAL 45 MINUTE)),
('req_005', '198.51.100.20', 'nikto/2.1.6', 'GET', '/admin/login', '{"User-Agent":"nikto/2.1.6"}', '', 4, '恶意爬虫拦截', 'user_agent', 'nikto', 'block', 403, 5, DATE_SUB(NOW(), INTERVAL 55 MINUTE)),

-- 最近几小时的攻击日志
('req_006', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', 'POST', '/payment/process', '{"Content-Type":"application/json"}', '{"amount":"100.00"}', 6, '支付接口保护', 'uri', '/payment', 'block', 403, 3, DATE_SUB(NOW(), INTERVAL 2 HOUR)),
('req_007', '10.0.0.50', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', 'POST', '/cart/add', '{"Content-Type":"application/json"}', '{"product_id":"123"}', 5, '购物车保护', 'uri', '/cart', 'log', 200, 3, DATE_SUB(NOW(), INTERVAL 3 HOUR)),
('req_008', '172.16.0.80', 'Mozilla/5.0 (compatible; Googlebot/2.1)', 'POST', '/blog/comment', '{"Content-Type":"application/json"}', '{"content":"Buy cheap viagra online!"}', 7, '评论垃圾信息过滤', 'body', 'viagra', 'block', 403, 4, DATE_SUB(NOW(), INTERVAL 4 HOUR)),
('req_009', '203.0.113.45', 'PostmanRuntime/7.29.2', 'GET', '/api/v1/users', '{"X-API-Key":"invalid_key"}', '', 10, 'API密钥验证', 'header', 'X-API-Key', 'log', 200, 5, DATE_SUB(NOW(), INTERVAL 5 HOUR)),
('req_010', '198.51.100.20', 'curl/7.68.0', 'GET', '/api/v1/data', '{"User-Agent":"curl/7.68.0"}', '', 9, 'API频率限制', 'uri', '/api/v1', 'log', 200, 5, DATE_SUB(NOW(), INTERVAL 6 HOUR)),

-- 最近几天的攻击日志
('req_011', '192.168.1.101', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', 'POST', '/login', '{"Content-Type":"application/json"}', '{"username":"admin","password":"admin\' UNION SELECT * FROM users--"}', 1, 'SQL注入防护', 'body', 'UNION SELECT', 'block', 403, 3, DATE_SUB(NOW(), INTERVAL 1 DAY)),
('req_012', '10.0.0.51', 'dirb/2.22', 'GET', '/admin/backup', '{"User-Agent":"dirb/2.22"}', '', 4, '恶意爬虫拦截', 'user_agent', 'dirb', 'block', 403, 4, DATE_SUB(NOW(), INTERVAL 1 DAY)),
('req_013', '172.16.0.81', 'Mozilla/5.0 (X11; Ubuntu; Linux x86_64)', 'GET', '/admin/settings', '{"User-Agent":"Mozilla/5.0 (X11; Ubuntu; Linux x86_64)"}', '', 3, '管理员路径保护', 'uri', '/admin', 'block', 403, 5, DATE_SUB(NOW(), INTERVAL 2 DAY)),
('req_014', '203.0.113.46', 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1)', 'POST', '/search', '{"Content-Type":"application/json"}', '{"query":"<img src=x onerror=alert(1)>"}', 2, 'XSS防护', 'body', '<img src=x onerror=', 'block', 403, 4, DATE_SUB(NOW(), INTERVAL 2 DAY)),
('req_015', '198.51.100.21', 'gobuster/3.1.0', 'GET', '/admin/users', '{"User-Agent":"gobuster/3.1.0"}', '', 4, '恶意爬虫拦截', 'user_agent', 'gobuster', 'block', 403, 3, DATE_SUB(NOW(), INTERVAL 3 DAY)),

-- 更多攻击日志数据，分布在不同时间段
('req_016', '192.168.1.102', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', 'POST', '/api/login', '{"Content-Type":"application/json"}', '{"user":"admin","pass":"admin\' OR \'1\'=\'1"}', 1, 'SQL注入防护', 'body', 'OR \'1\'=\'1', 'block', 403, 5, DATE_SUB(NOW(), INTERVAL 4 DAY)),
('req_017', '10.0.0.52', 'nmap/7.80', 'GET', '/admin/dashboard', '{"User-Agent":"nmap/7.80"}', '', 4, '恶意爬虫拦截', 'user_agent', 'nmap', 'block', 403, 3, DATE_SUB(NOW(), INTERVAL 5 DAY)),
('req_018', '172.16.0.82', 'Mozilla/5.0 (compatible; Yahoo! Slurp)', 'GET', '/admin/reports', '{"User-Agent":"Mozilla/5.0 (compatible; Yahoo! Slurp)"}', '', 3, '管理员路径保护', 'uri', '/admin', 'block', 403, 4, DATE_SUB(NOW(), INTERVAL 6 DAY)),
('req_019', '203.0.113.47', 'Mozilla/5.0 (Android 10; Mobile)', 'POST', '/feedback', '{"Content-Type":"application/json"}', '{"message":"<iframe src=javascript:alert(1)></iframe>"}', 2, 'XSS防护', 'body', '<iframe src=javascript:', 'block', 403, 4, DATE_SUB(NOW(), INTERVAL 7 DAY)),
('req_020', '198.51.100.22', 'curl/7.68.0', 'GET', '/api/v1/admin', '{"User-Agent":"curl/7.68.0"}', '', 9, 'API频率限制', 'uri', '/api/v1', 'log', 200, 5, DATE_SUB(NOW(), INTERVAL 8 DAY));

-- 6. 插入速率限制记录
INSERT INTO rate_limits (`key`, count, `window`, created_at, updated_at) VALUES
('192.168.1.100', 45, DATE_SUB(NOW(), INTERVAL 1 MINUTE), NOW(), NOW()),
('10.0.0.50', 120, DATE_SUB(NOW(), INTERVAL 1 MINUTE), NOW(), NOW()),
('172.16.0.80', 30, DATE_SUB(NOW(), INTERVAL 1 MINUTE), NOW(), NOW()),
('203.0.113.45', 80, DATE_SUB(NOW(), INTERVAL 1 MINUTE), NOW(), NOW()),
('198.51.100.20', 200, DATE_SUB(NOW(), INTERVAL 1 MINUTE), NOW(), NOW());

-- 7. 插入白名单数据（依赖租户表）
INSERT INTO white_lists (type, value, comment, tenant_id, enabled, created_at, updated_at) VALUES
-- 全局白名单
('ip', '127.0.0.1', '本地回环地址', 1, true, NOW(), NOW()),
('ip', '192.168.1.0/24', '内网IP段', 1, true, NOW(), NOW()),
('user_agent', 'Googlebot', '谷歌搜索引擎爬虫', 1, true, NOW(), NOW()),
('user_agent', 'Baiduspider', '百度搜索引擎爬虫', 1, true, NOW(), NOW()),
('uri', '/health', '健康检查接口', 1, true, NOW(), NOW()),

-- 电商平台白名单
('ip', '203.0.113.100', '支付网关IP', 3, true, NOW(), NOW()),
('uri', '/api/webhook/payment', '支付回调接口', 3, true, NOW(), NOW()),

-- 博客网站白名单
('ip', '198.51.100.100', '内容分发网络IP', 4, true, NOW(), NOW()),
('uri', '/rss', 'RSS订阅接口', 4, true, NOW(), NOW()),

-- API服务白名单
('ip', '172.16.0.100', '合作伙伴API调用IP', 5, true, NOW(), NOW()),
('user_agent', 'PartnerApp/1.0', '合作伙伴应用', 5, true, NOW(), NOW());

-- 8. 插入黑名单数据（依赖租户表）
INSERT INTO black_lists (type, value, comment, tenant_id, enabled, created_at, updated_at) VALUES
-- 全局黑名单
('ip', '192.0.2.100', '已知恶意IP', 1, true, NOW(), NOW()),
('ip', '198.51.100.50', '扫描器IP', 1, true, NOW(), NOW()),
('user_agent', 'sqlmap', 'SQL注入工具', 1, true, NOW(), NOW()),
('user_agent', 'nikto', '漏洞扫描工具', 1, true, NOW(), NOW()),
('user_agent', 'nmap', '端口扫描工具', 1, true, NOW(), NOW()),

-- 电商平台黑名单
('ip', '203.0.113.200', '恶意刷单IP', 3, true, NOW(), NOW()),
('user_agent', 'BadBot/1.0', '恶意爬虫', 3, true, NOW(), NOW()),

-- 博客网站黑名单
('ip', '198.51.100.200', '垃圾评论IP', 4, true, NOW(), NOW()),
('user_agent', 'SpamBot', '垃圾信息机器人', 4, true, NOW(), NOW()),

-- API服务黑名单
('ip', '172.16.0.200', '恶意API调用IP', 5, true, NOW(), NOW()),
('user_agent', 'AttackBot/2.0', '攻击机器人', 5, true, NOW(), NOW());

-- 9. 插入Webhook配置数据（依赖租户表）
INSERT INTO webhooks (name, url, method, headers, template, events, enabled, tenant_id, created_at, updated_at) VALUES
-- 全局Webhook
('安全告警通知', 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX', 'POST', '{"Content-Type":"application/json"}', '{"text":"WAF告警: {{.message}}","username":"WAF-Bot"}', '["attack","threshold"]', true, 1, NOW(), NOW()),
('邮件告警服务', 'https://api.sendgrid.com/v3/mail/send', 'POST', '{"Authorization":"Bearer SG.xxx","Content-Type":"application/json"}', '{"personalizations":[{"to":[{"email":"admin@example.com"}]}],"from":{"email":"waf@example.com"},"subject":"WAF安全告警","content":[{"type":"text/plain","value":"{{.message}}"}]}', '["attack"]', true, 1, NOW(), NOW()),

-- 电商平台Webhook
('电商平台告警', 'https://shop.example.com/api/webhook/security', 'POST', '{"Content-Type":"application/json","X-API-Key":"shop_webhook_key"}', '{"event":"security_alert","data":{"message":"{{.message}}","ip":"{{.client_ip}}","time":"{{.timestamp}}"}}', '["attack","threshold"]', true, 3, NOW(), NOW()),

-- 博客网站Webhook
('博客安全通知', 'https://blog.example.com/webhook/security', 'POST', '{"Content-Type":"application/json"}', '{"type":"security","message":"{{.message}}","details":{"ip":"{{.client_ip}}","uri":"{{.request_uri}}"}}', '["attack"]', true, 4, NOW(), NOW()),

-- API服务Webhook
('API监控告警', 'https://api.example.com/internal/webhook/monitor', 'POST', '{"Content-Type":"application/json","Authorization":"Bearer api_monitor_token"}', '{"alert_type":"waf","severity":"{{.severity}}","message":"{{.message}}","metadata":{"client_ip":"{{.client_ip}}","user_agent":"{{.user_agent}}"}}', '["attack","threshold"]', true, 5, NOW(), NOW());

-- 显示插入结果统计
SELECT 'tenants' as table_name, COUNT(*) as record_count FROM tenants
UNION ALL
SELECT 'users', COUNT(*) FROM users
UNION ALL
SELECT 'rules', COUNT(*) FROM rules
UNION ALL
SELECT 'policies', COUNT(*) FROM policies
UNION ALL
SELECT 'attack_logs', COUNT(*) FROM attack_logs
UNION ALL
SELECT 'rate_limits', COUNT(*) FROM rate_limits
UNION ALL
SELECT 'white_lists', COUNT(*) FROM white_lists
UNION ALL
SELECT 'black_lists', COUNT(*) FROM black_lists
UNION ALL
SELECT 'webhooks', COUNT(*) FROM webhooks; 