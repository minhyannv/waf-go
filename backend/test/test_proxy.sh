#!/bin/bash

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# 测试配置
WAF_HOST="localhost:8081"
BACKEND_HOST="localhost:3001"  # 修改后端端口

echo -e "${GREEN}开始WAF代理测试${NC}"

# 清理可能存在的进程
echo "清理可能存在的进程..."
lsof -ti:3001 | xargs kill -9 2>/dev/null || true

# 获取认证令牌
echo "获取认证令牌..."
TOKEN=$(curl -s -X POST "http://$WAF_HOST/api/v1/auth/login" \
-H "Content-Type: application/json" \
-d '{
    "username": "admin",
    "password": "admin123"
}' | jq -r '.data.token')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}获取认证令牌失败${NC}"
    exit 1
fi

echo "认证令牌: $TOKEN"

# 1. 启动后端服务
echo "启动测试后端服务..."
PORT=3001 go run test_backend.go &
BACKEND_PID=$!
sleep 2

# 等待后端服务启动
echo "等待后端服务就绪..."
for i in {1..5}; do
    if curl -s "http://$BACKEND_HOST/health" > /dev/null; then
        break
    fi
    sleep 1
done

# 2. 添加测试域名
echo "添加测试域名配置..."
DOMAIN_ID=$(curl -s -X POST "http://$WAF_HOST/api/v1/domains" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d '{
    "domain": "test.example.com",
    "protocol": "http",
    "port": 80,
    "backend_url": "http://localhost:3001",
    "enabled": true,
    "tenant_id": 1
}' | jq -r '.data.id')

echo "创建的域名ID: $DOMAIN_ID"

echo -e "\n测试HTTP代理..."

# 3. 测试正常请求
echo "测试正常请求..."
curl -H "Host: test.example.com" "http://$WAF_HOST/test" -v

# 4. 测试SQL注入攻击
echo -e "\n测试SQL注入攻击..."
curl -H "Host: test.example.com" "http://$WAF_HOST/test?id=1%27%20OR%20%271%27=%271" -v

# 5. 测试XSS攻击
echo -e "\n测试XSS攻击..."
curl -H "Host: test.example.com" "http://$WAF_HOST/test?input=<script>alert(1)</script>" -v

echo -e "\n\n测试域名管理..."

# 6. 获取域名配置
echo "获取域名配置..."
curl -s -H "Authorization: Bearer $TOKEN" "http://$WAF_HOST/api/v1/domains/$DOMAIN_ID"

# 7. 更新域名配置
echo "更新域名配置..."
curl -s -X PUT "http://$WAF_HOST/api/v1/domains/$DOMAIN_ID" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d '{
    "backend_url": "http://localhost:3001",
    "enabled": true
}'

# 8. 禁用域名
echo "禁用域名..."
curl -s -X POST "http://$WAF_HOST/api/v1/domains/$DOMAIN_ID/toggle" \
-H "Authorization: Bearer $TOKEN"

# 9. 测试已禁用域名
echo "测试已禁用域名..."
curl -H "Host: test.example.com" "http://$WAF_HOST/test" -v

# 10. 启用域名
echo "启用域名..."
curl -s -X POST "http://$WAF_HOST/api/v1/domains/$DOMAIN_ID/toggle" \
-H "Authorization: Bearer $TOKEN"

# 11. 清理测试环境
echo "清理测试环境..."
curl -s -X DELETE "http://$WAF_HOST/api/v1/domains/$DOMAIN_ID" \
-H "Authorization: Bearer $TOKEN"

# 12. 停止后端服务
kill $BACKEND_PID

echo -e "${GREEN}HTTP测试完成${NC}" 