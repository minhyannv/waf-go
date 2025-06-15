# WAF-Go

WAF-Go 是一个基于 Go 语言开发的 Web 应用防火墙（WAF）系统，提供了简单易用的 Web 界面，支持域名管理、安全策略配置、攻击防护等功能。

## 功能特性

### 域名管理
- 支持 HTTP/HTTPS 域名配置
- SSL 证书和私钥管理
- 后端服务器配置
- 域名状态管理（启用/禁用）

### 安全策略
- 规则管理
- 策略配置
- 黑白名单
- 攻击日志

### 系统管理
- 多租户支持
- 用户管理
- 权限控制
- 系统配置

## 技术栈

### 后端
- Go 1.21
- Gin Web Framework
- GORM
- MySQL
- Redis

### 前端
- Vue 3
- TypeScript
- Element Plus
- Vite
- Axios

## 快速开始

### 环境要求
- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- Redis 7.0+

### 开发环境搭建

1. 克隆项目
```bash
git clone https://github.com/yourusername/waf-go.git
cd waf-go
```

2. 启动开发环境
```bash
# 启动数据库和缓存服务
docker-compose up -d mysql redis

# 启动后端服务
cd backend
go mod download
go run main.go

# 启动前端服务
cd frontend
npm install
npm run dev
```

### 生产环境部署

使用 Docker Compose 一键部署：

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f
```

默认端口：
- 前端：80
- 后端：8080
- MySQL：3306
- Redis：6379

## 配置说明

### 后端配置

配置文件位置：`backend/config/config.yaml`

```yaml
server:
  port: ":8080"
  mode: "debug"  # debug 或 release

database:
  dsn: "user:password@tcp(mysql:3306)/waf?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  addr: "redis:6379"
  password: ""
  db: 0

jwt:
  secret: "your-jwt-secret"
  expire: 3600

waf:
  rate_limit_window: 60
  max_requests: 100

log:
  level: "debug"
```

### 前端配置

环境变量配置文件：`frontend/.env`

```env
VITE_API_BASE_URL=http://localhost:8080
```

## 开发指南

### 目录结构

```
.
├── backend                 # 后端代码
│   ├── cmd                # 命令行工具
│   ├── config             # 配置文件
│   ├── internal           # 内部包
│   │   ├── handler       # HTTP 处理器
│   │   ├── middleware    # 中间件
│   │   ├── models        # 数据模型
│   │   ├── service       # 业务逻辑
│   │   └── utils         # 工具函数
│   └── sql               # SQL 文件
├── frontend               # 前端代码
│   ├── src
│   │   ├── api          # API 请求
│   │   ├── components   # 组件
│   │   ├── router       # 路由配置
│   │   ├── store        # 状态管理
│   │   ├── utils        # 工具函数
│   │   └── views        # 页面
│   └── public            # 静态资源
└── docker-compose.yml    # Docker 编排配置
```

### API 文档

后端 API 文档使用 Swagger 生成，访问地址：`http://localhost:8080/swagger/index.html`

## 贡献指南

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的改动 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开一个 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解更多细节。 