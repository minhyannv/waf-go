# WAF-Go

基于 Go 语言开发的 Web 应用防火墙（WAF）系统，提供域名管理、安全策略配置、攻击防护等功能。

## 功能特性

### 🔒 安全防护
- **攻击检测**：SQL注入、XSS、恶意爬虫、恶意IP等攻击检测
- **规则引擎**：支持URI、IP、请求头、请求体、User-Agent等多种匹配类型
- **策略管理**：灵活的规则组合和策略配置
- **黑白名单**：IP黑白名单管理，快速封禁/放行

### 🌐 域名管理
- **多域名支持**：支持HTTP/HTTPS域名配置
- **SSL证书**：SSL证书管理和自动续期
- **端口配置**：支持自定义端口配置
- **协议支持**：HTTP/HTTPS协议切换

### 📊 监控分析
- **实时监控**：攻击趋势分析和实时告警
- **日志记录**：详细的攻击日志和访问记录
- **数据统计**：攻击类型、来源IP、时间分布等统计
- **可视化**：直观的图表展示和数据分析

### 👥 多租户
- **租户隔离**：多租户数据隔离管理
- **权限控制**：基于角色的权限管理
- **资源分配**：租户级别的资源配置

## 技术栈

### 后端
- **语言**：Go 1.21+
- **框架**：Gin (Web框架)
- **ORM**：GORM (数据库操作)
- **数据库**：MySQL 8.0+
- **缓存**：Redis 7.0+
- **认证**：JWT (身份认证)

### 前端
- **框架**：Vue 3 + TypeScript
- **UI库**：Element Plus
- **构建工具**：Vite
- **路由**：Vue Router
- **状态管理**：Pinia

## 快速开始

### 环境要求
- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- Redis 7.0+

### 一键启动

```bash
# 克隆项目
git clone https://github.com/minhyannv/waf-go.git
cd waf-go

# 启动所有服务
docker-compose up -d

# 访问系统
# 前端：http://localhost
# 后端API：http://localhost:8080
```

### 开发环境

```bash
# 启动数据库服务
docker-compose up -d mysql redis

# 启动后端服务
cd backend
go mod tidy
go run main.go

# 启动前端服务
cd frontend
npm install
npm run dev
```

## 配置说明

### 默认端口
- **前端**：80
- **后端API**：8080
- **MySQL**：3306
- **Redis**：6379

### 默认账户
- **管理员**：admin / admin123
- **数据库**：root / waf123456

### 环境变量
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=waf123456
DB_NAME=waf_go

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=your-jwt-secret
JWT_EXPIRE=24h
```

## 项目结构

```
waf-go/
├── backend/                 # 后端代码
│   ├── cmd/                # 命令行入口
│   ├── config/             # 配置文件
│   ├── internal/           # 内部包
│   │   ├── config/         # 配置管理
│   │   ├── database/       # 数据库连接
│   │   ├── handler/        # HTTP处理器
│   │   ├── middleware/     # 中间件
│   │   ├── models/         # 数据模型
│   │   ├── router/         # 路由配置
│   │   ├── service/        # 业务逻辑
│   │   ├── utils/          # 工具函数
│   │   └── waf/            # WAF引擎
│   ├── sql/                # SQL脚本
│   └── main.go             # 主程序入口
├── frontend/               # 前端代码
│   ├── src/
│   │   ├── api/            # API接口
│   │   ├── components/     # 组件
│   │   ├── layouts/        # 布局组件
│   │   ├── router/         # 路由配置
│   │   ├── stores/         # 状态管理
│   │   ├── utils/          # 工具函数
│   │   └── views/          # 页面组件
│   ├── public/             # 静态资源
│   └── package.json        # 依赖配置
├── docker-compose.yml      # Docker编排文件
└── README.md              # 项目说明
```

## API 文档

### 认证相关
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/logout` - 用户登出
- `GET /api/v1/auth/profile` - 获取用户信息

### 域名管理
- `GET /api/v1/domains` - 获取域名列表
- `POST /api/v1/domains` - 创建域名
- `PUT /api/v1/domains/:id` - 更新域名
- `DELETE /api/v1/domains/:id` - 删除域名

### 策略管理
- `GET /api/v1/policies` - 获取策略列表
- `POST /api/v1/policies` - 创建策略
- `PUT /api/v1/policies/:id` - 更新策略
- `DELETE /api/v1/policies/:id` - 删除策略

### 规则管理
- `GET /api/v1/rules` - 获取规则列表
- `POST /api/v1/rules` - 创建规则
- `PUT /api/v1/rules/:id` - 更新规则
- `DELETE /api/v1/rules/:id` - 删除规则

### 日志管理
- `GET /api/v1/logs` - 获取攻击日志
- `GET /api/v1/logs/:id` - 获取日志详情

## 部署说明

### Docker 部署
```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 生产环境
1. 修改 `docker-compose.yml` 中的环境变量
2. 配置 SSL 证书
3. 设置防火墙规则
4. 配置数据库备份
5. 设置监控告警

## 开发指南

### 代码规范
- 使用 Go 官方代码规范
- 遵循 Vue 3 组合式 API 最佳实践
- 使用 TypeScript 进行类型检查
- 添加必要的注释和文档

### 测试
```bash
# 后端测试
cd backend
go test ./...

# 前端测试
cd frontend
npm run test
```

### 构建
```bash
# 后端构建
cd backend
go build -o waf-server main.go

# 前端构建
cd frontend
npm run build
```

## 常见问题

### Q: 如何修改默认端口？
A: 修改 `docker-compose.yml` 文件中的端口映射配置。

### Q: 如何添加新的规则类型？
A: 在 `backend/internal/models/models.go` 中添加新的匹配类型，并在前端添加相应的UI组件。

### Q: 如何配置SSL证书？
A: 将证书文件放在 `frontend/public/certs/` 目录下，并在域名配置中指定证书路径。

### Q: 如何备份数据？
A: 使用 `mysqldump` 命令备份数据库，定期备份 `backend/sql/` 目录下的数据。

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目地址：https://github.com/minhyannv/waf-go
- 问题反馈：https://github.com/minhyannv/waf-go/issues
- 邮箱：your-email@example.com

## 更新日志

### v1.0.0 (2024-06-18)
- 🎉 初始版本发布
- ✨ 支持域名管理、策略配置、攻击防护
- ✨ 实现多租户隔离管理
- ✨ 提供实时监控和日志分析
- 🐛 修复策略编辑功能相关问题
- 🐛 解决Element Plus组件验证错误 