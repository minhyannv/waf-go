# WAF 测试文档

## 概述

本文档介绍如何测试 WAF 系统的功能，包括本地代理真实域名（如百度）并模拟各种攻击流量。

## 目录

- [攻击模拟脚本](#攻击模拟脚本)
- [本地代理百度测试](#本地代理百度测试)
- [常见攻击类型](#常见攻击类型)
- [测试步骤详解](#测试步骤详解)

---

## 攻击模拟脚本

### 使用方法

```bash
cd backend/examples
go run attack_simulator.go http://目标地址[:端口]
```

### 示例

```bash
# 测试本地WAF
go run attack_simulator.go http://localhost:8081

# 测试代理的百度
go run attack_simulator.go http://www.baidu.com:8081
```

---

## 本地代理百度测试

### 测试目标

- 让WAF作为反向代理，代理百度（www.baidu.com）
- 本地hosts将www.baidu.com指向WAF
- WAF后端配置为百度真实IP
- 用Go脚本/浏览器/curl访问www.baidu.com，模拟正常和攻击流量，观察WAF是否拦截

### 步骤详解

#### 步骤一：本地hosts配置

编辑 `/etc/hosts` 文件，添加如下内容（需要管理员权限）：

```bash
# macOS/Linux
sudo vim /etc/hosts

# Windows
# 编辑 C:\Windows\System32\drivers\etc\hosts
```

添加以下内容：
```
127.0.0.1 www.baidu.com
```

这样所有对 www.baidu.com 的请求都会被转发到本机（即WAF）。

#### 步骤二：WAF后台添加域名和后端

1. **登录WAF管理后台**
   - 访问 http://localhost:5174/
   - 使用管理员账户登录

2. **添加域名**
   - 进入"域名管理"页面
   - 添加域名：`www.baidu.com`
   - 配置后端源站为百度真实IP（可选用如下之一）：
     ```
     220.181.38.150
     220.181.38.149
     ```
   - 端口为80（http），协议为http

3. **配置策略和规则**
   - 进入"策略管理"页面
   - 创建或编辑策略，关联域名 `www.baidu.com`
   - 添加SQL注入、XSS等检测规则

#### 步骤三：启动WAF服务

确保WAF监听本地8081端口：

```bash
cd backend
go run main.go
```

看到以下输出表示启动成功：
```
HTTP服务器启动在端口 8081
```

#### 步骤四：发起正常和攻击请求

##### 1. 正常请求

```bash
curl -H "Host: www.baidu.com" http://127.0.0.1:8081/
```

或直接在浏览器访问 http://www.baidu.com:8081/

##### 2. SQL注入攻击请求

```bash
curl -H "Host: www.baidu.com" "http://127.0.0.1:8081/?wd=1' OR '1'='1"
```

##### 3. XSS攻击请求

```bash
curl -H "Host: www.baidu.com" "http://127.0.0.1:8081/?wd=<script>alert('xss')</script>"
```

##### 4. 使用Go脚本批量模拟

```bash
cd backend/examples
go run attack_simulator.go http://www.baidu.com:8081
```

#### 步骤五：查看WAF拦截效果

1. **登录WAF前端管理后台**
   - 访问 http://localhost:5174/

2. **查看攻击日志**
   - 打开"攻击日志"页面
   - 检查是否有刚才的SQL注入、XSS等攻击记录
   - 检查正常请求是否未被拦截

3. **查看仪表板统计**
   - 打开"仪表板"页面
   - 查看攻击趋势、Top规则、Top IP等统计信息

---

## 常见攻击类型

### SQL注入攻击

```bash
# 基础SQL注入
curl "http://目标地址/?id=1' OR '1'='1"

# 联合查询注入
curl "http://目标地址/?id=1' UNION SELECT 1,2,3--"

# 布尔盲注
curl "http://目标地址/?id=1' AND 1=1--"
```

### XSS攻击

```bash
# 反射型XSS
curl "http://目标地址/?q=<script>alert('xss')</script>"

# 存储型XSS
curl -X POST "http://目标地址/comment" \
  -d "content=<script>alert('stored xss')</script>"

# DOM型XSS
curl "http://目标地址/?hash=javascript:alert('dom xss')"
```

### 命令注入

```bash
# 基础命令注入
curl "http://目标地址/?cmd=;ls -la"

# 管道命令注入
curl "http://目标地址/?cmd=|cat /etc/passwd"
```

### 路径遍历

```bash
# 基础路径遍历
curl "http://目标地址/?file=../../../etc/passwd"

# 编码绕过
curl "http://目标地址/?file=%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd"
```

---

## 测试步骤详解

### 1. 环境准备

```bash
# 1. 启动MySQL和Redis
docker-compose up -d

# 2. 启动后端服务
cd backend
go run main.go

# 3. 启动前端服务
cd frontend
npm run dev
```

### 2. 配置测试环境

```bash
# 1. 修改hosts文件
sudo vim /etc/hosts
# 添加: 127.0.0.1 www.baidu.com

# 2. 清理DNS缓存
# macOS
sudo dscacheutil -flushcache
# Linux
sudo systemctl restart systemd-resolved
```

### 3. 执行测试

```bash
# 1. 正常流量测试
curl -H "Host: www.baidu.com" http://127.0.0.1:8081/

# 2. 攻击流量测试
cd backend/examples
go run attack_simulator.go http://www.baidu.com:8081

# 3. 手动测试特定攻击
curl -H "Host: www.baidu.com" "http://127.0.0.1:8081/?wd=1' OR '1'='1"
```

### 4. 验证结果

1. **检查攻击日志**
   - 登录WAF管理后台
   - 查看"攻击日志"页面
   - 确认攻击请求被正确记录

2. **检查仪表板**
   - 查看"仪表板"页面
   - 确认攻击统计正确更新

3. **检查规则命中**
   - 查看"规则管理"页面
   - 确认相关规则被正确触发

---

## 注意事项

### 1. 环境配置

- hosts文件修改后，建议清理浏览器DNS缓存或重启浏览器
- 确保WAF服务正常运行在8081端口
- 确保MySQL和Redis服务正常运行

### 2. 测试限制

- 百度有防爬虫机制，部分请求可能被百度自身拦截，与WAF无关
- 某些攻击可能被目标网站自身的安全机制拦截
- 测试时注意不要对生产环境发起真实攻击

### 3. 性能考虑

- 大量并发攻击测试可能影响WAF性能
- 建议在测试环境进行，避免影响生产服务
- 监控WAF服务器的CPU和内存使用情况

### 4. 恢复环境

测试结束后，记得将 `/etc/hosts` 中的 `127.0.0.1 www.baidu.com` 删除，恢复正常访问：

```bash
sudo vim /etc/hosts
# 删除或注释掉: 127.0.0.1 www.baidu.com
```

---

## 故障排除

### 1. WAF服务启动失败

```bash
# 检查端口占用
lsof -i:8081

# 杀死占用进程
kill -9 <PID>

# 重新启动WAF
cd backend
go run main.go
```

### 2. 攻击日志未记录

- 检查WAF规则是否启用
- 检查域名配置是否正确
- 检查后端源站是否可达
- 查看WAF服务日志

### 3. 请求被拒绝

- 检查黑白名单配置
- 检查IP限制规则
- 检查请求频率限制
- 查看WAF拦截日志

---

## 扩展测试

### 1. 性能测试

```bash
# 使用ab进行压力测试
ab -n 1000 -c 10 http://www.baidu.com:8081/

# 使用wrk进行并发测试
wrk -t12 -c400 -d30s http://www.baidu.com:8081/
```

### 2. 自定义攻击脚本

可以修改 `attack_simulator.go` 文件，添加更多攻击类型：

```go
// 添加新的攻击类型
{
    Name: "命令注入",
    Path: "/?cmd=;ls -la",
    Params: map[string]string{},
},
{
    Name: "路径遍历",
    Path: "/?file=../../../etc/passwd",
    Params: map[string]string{},
},
```

### 3. 自动化测试

可以编写自动化测试脚本，批量执行各种攻击并验证结果。

---

## 总结

通过以上测试，可以验证WAF系统的以下功能：

1. **代理功能**：正确转发请求到后端服务器
2. **检测功能**：识别并拦截各种攻击
3. **日志功能**：记录攻击详情和统计信息
4. **管理功能**：通过Web界面管理规则和配置

建议定期进行此类测试，确保WAF系统正常运行和防护效果。 