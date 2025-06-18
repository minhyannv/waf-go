# WAF-Go

基于 Go 语言开发的 Web 应用防火墙（WAF）系统，提供域名管理、安全策略配置、攻击防护等功能。

## 功能特性

- **域名管理**：支持 HTTP/HTTPS 域名配置，SSL 证书管理
- **安全策略**：规则管理、策略配置、黑白名单
- **攻击防护**：SQL注入、XSS、恶意爬虫等攻击检测
- **实时监控**：攻击趋势分析、攻击日志记录
- **多租户**：支持多租户隔离管理

## 技术栈

- **后端**：Go + Gin + GORM + MySQL + Redis
- **前端**：Vue 3 + TypeScript + Element Plus

## 快速开始

### 环境要求
- Docker & Docker Compose
- Go 1.21+
- Node.js 18+

### 一键启动

```bash
# 克隆项目
git clone https://github.com/yourusername/waf-go.git
cd waf-go

# 启动所有服务
docker-compose up -d

# 访问系统
# 前端：http://localhost
# 后端：http://localhost:8080
```

### 开发环境

```bash
# 启动数据库
docker-compose up -d mysql redis

# 启动后端
cd backend
go run main.go

# 启动前端
cd frontend
npm install
npm run dev
```

## 默认配置

- **前端端口**：80
- **后端端口**：8080
- **MySQL端口**：3306
- **Redis端口**：6379
- **默认用户**：admin / admin123

## 项目结构

```
waf-go/
├── backend/          # 后端代码
├── frontend/         # 前端代码
├── docker-compose.yml
└── README.md
```

## 许可证

MIT License 