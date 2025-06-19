#!/bin/bash

# WAF百度代理测试脚本
# 使用方法: ./test_baidu_proxy.sh

echo "=== WAF百度代理测试脚本 ==="
echo ""

# 检查hosts配置
echo "1. 检查hosts配置..."
if grep -q "127.0.0.1 www.baidu.com" /etc/hosts; then
    echo "✓ hosts配置正确"
else
    echo "✗ 请先配置hosts文件:"
    echo "  sudo vim /etc/hosts"
    echo "  添加: 127.0.0.1 www.baidu.com"
    exit 1
fi

# 检查WAF服务
echo ""
echo "2. 检查WAF服务..."
if curl -s http://localhost:8081/app/test > /dev/null; then
    echo "✓ WAF服务运行正常"
else
    echo "✗ WAF服务未运行，请先启动:"
    echo "  cd backend && go run main.go"
    exit 1
fi

# 测试正常请求
echo ""
echo "3. 测试正常请求..."
echo "请求: http://www.baidu.com:8081/"
response=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: www.baidu.com" http://127.0.0.1:8081/)
echo "响应状态码: $response"

# 测试SQL注入攻击
echo ""
echo "4. 测试SQL注入攻击..."
echo "请求: http://www.baidu.com:8081/?wd=1' OR '1'='1"
response=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: www.baidu.com" "http://127.0.0.1:8081/?wd=1' OR '1'='1")
echo "响应状态码: $response"

# 测试XSS攻击
echo ""
echo "5. 测试XSS攻击..."
echo "请求: http://www.baidu.com:8081/?wd=<script>alert('xss')</script>"
response=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: www.baidu.com" "http://127.0.0.1:8081/?wd=<script>alert('xss')</script>")
echo "响应状态码: $response"

# 测试命令注入
echo ""
echo "6. 测试命令注入..."
echo "请求: http://www.baidu.com:8081/?cmd=;ls -la"
response=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: www.baidu.com" "http://127.0.0.1:8081/?cmd=;ls -la")
echo "响应状态码: $response"

echo ""
echo "=== 测试完成 ==="
echo ""
echo "请登录WAF管理后台查看攻击日志:"
echo "  http://localhost:5174/"
echo ""
echo "查看攻击日志页面，确认是否有上述攻击的记录。" 