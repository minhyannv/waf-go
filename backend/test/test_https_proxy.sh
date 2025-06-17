#!/bin/bash

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# 测试配置
WAF_HOST="localhost:8443"
BACKEND_HOST="localhost:3002"  # 修改后端端口

# 清理可能存在的进程
echo "清理可能存在的进程..."
lsof -ti:3002 | xargs kill -9 2>/dev/null || true

# 获取认证令牌
echo "获取认证令牌..."
TOKEN=$(curl -s -X POST "http://localhost:8081/api/v1/auth/login" \
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

# 生成测试证书
echo "生成测试SSL证书..."
DOMAIN="secure-$(date +%s).example.com"
openssl req -x509 -newkey rsa:2048 -keyout test.key -out test.crt -days 365 -nodes \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=Test/CN=$DOMAIN"

# 读取证书内容
CERT=$(cat test.crt | sed 's/$/\\n/' | tr -d '\n')
KEY=$(cat test.key | sed 's/$/\\n/' | tr -d '\n')

echo -e "${GREEN}开始HTTPS代理测试${NC}"

# 1. 启动后端服务
echo "启动测试后端服务..."
PORT=3002 go run test_backend.go &
BACKEND_PID=$!
sleep 2

# 2. 添加HTTPS测试域名
echo "添加HTTPS测试域名配置..."
DOMAIN_ID=$(curl -s -X POST "http://localhost:8081/api/v1/domains" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d "{
    \"domain\": \"$DOMAIN\",
    \"protocol\": \"https\",
    \"port\": 443,
    \"backend_url\": \"http://$BACKEND_HOST\",
    \"ssl_certificate\": \"$CERT\",
    \"ssl_private_key\": \"$KEY\",
    \"enabled\": true,
    \"tenant_id\": 2
}" | jq -r '.data.id')

if [ -z "$DOMAIN_ID" ] || [ "$DOMAIN_ID" = "null" ]; then
    echo -e "${RED}创建域名失败${NC}"
    kill $BACKEND_PID 2>/dev/null || true
    rm test.crt test.key
    exit 1
fi

echo "创建的域名ID: $DOMAIN_ID, 域名: $DOMAIN"

echo -e "\n\n测试HTTPS代理..."

# 3. 测试HTTPS代理
echo "测试正常HTTPS请求..."
curl -k -H "Host: $DOMAIN" "https://$WAF_HOST/test" -v

echo -e "\n测试HTTPS SQL注入攻击..."
curl -k -H "Host: $DOMAIN" "https://$WAF_HOST/login?username=admin'%20OR%20'1'='1" -v

echo -e "\n测试HTTPS XSS攻击..."
curl -k -H "Host: $DOMAIN" "https://$WAF_HOST/search?q=<script>alert(1)</script>" -v

# 4. 测试HTTPS域名管理
echo -e "\n\n测试HTTPS域名管理..."

echo "获取域名配置..."
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8081/api/v1/domains/$DOMAIN_ID"

echo -e "\n更新HTTPS域名配置..."
curl -X PUT "http://localhost:8081/api/v1/domains/$DOMAIN_ID" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d "{
    \"backend_url\": \"http://$BACKEND_HOST\",
    \"enabled\": true
}"

echo -e "\n禁用HTTPS域名..."
curl -X POST "http://localhost:8081/api/v1/domains/$DOMAIN_ID/toggle" \
-H "Authorization: Bearer $TOKEN"

echo -e "\n测试已禁用HTTPS域名..."
curl -k -H "Host: $DOMAIN" "https://$WAF_HOST/test" -v

echo -e "\n启用HTTPS域名..."
curl -X POST "http://localhost:8081/api/v1/domains/$DOMAIN_ID/toggle" \
-H "Authorization: Bearer $TOKEN"

# 5. 清理
echo -e "\n\n清理测试环境..."
curl -X DELETE "http://localhost:8081/api/v1/domains/$DOMAIN_ID" \
-H "Authorization: Bearer $TOKEN"
rm test.crt test.key

# 停止后端服务
kill $BACKEND_PID 2>/dev/null || true

echo -e "${GREEN}HTTPS测试完成${NC}" 