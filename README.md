# WAF 管理系统

一个基于 Go + Vue 3 的 Web 应用防火墙（WAF）系统，提供实时威胁检测、规则管理、攻击日志记录等功能。

## 🚀 功能特性

### 核心 WAF 功能
- ✅ 多种匹配模式：URI、IP、请求头、请求体、User-Agent
- ✅ 灵活匹配算法：精确匹配、正则表达式、包含匹配
- ✅ 多种响应动作：拦截、记录、允许
- ✅ 速率限制：基于滑动窗口的请求频率控制
- ✅ 黑白名单：IP、URI、User-Agent 级别的访问控制
- ✅ 热更新：规则修改后无需重启服务

### 管理功能
- ✅ 用户认证：JWT 令牌认证
- ✅ 角色权限：管理员、租户管理员、查看者
- ✅ 多租户支持：不同租户隔离管理
- ✅ 实时仪表盘：攻击统计、趋势图表、实时监控
- ✅ 日志查询：详细的攻击日志记录和查询
- ✅ 规则管理：可视化规则创建和管理
- ✅ 黑名单管理：支持 IP、URI、User-Agent 黑名单
- ✅ 攻击统计：Top 攻击 IP、URI、User-Agent 统计
- ✅ 规则分布：攻击规则触发统计和可视化

### 仪表盘功能
- ✅ 攻击趋势分析：24小时攻击趋势图表
- ✅ 攻击规则分布：规则触发次数统计
- ✅ Top 攻击统计：IP、URI、User-Agent 排行
- ✅ 实时监控：最近1小时攻击实时数据
- ✅ 响应式布局：适配不同屏幕尺寸

## 🏗 技术架构

### 后端技术栈
- **Go 1.21+** - 编程语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **MySQL 8.0** - 数据存储
- **Redis 7** - 缓存和速率限制
- **JWT** - 身份认证
- **Zap** - 日志框架

### 前端技术栈
- **Vue 3** - 前端框架
- **TypeScript** - 类型检查
- **Element Plus** - UI 组件库
- **Vue Router** - 路由管理
- **Pinia** - 状态管理
- **Axios** - HTTP 客户端
- **ECharts** - 图表可视化

### 部署技术栈
- **Docker** - 容器化部署
- **Docker Compose** - 多容器编排
- **Nginx** - 前端静态文件服务

## 📦 项目结构

```
waf-go/
├── backend/                    # 后端代码
│   ├── main.go                # 主入口文件
│   ├── go.mod                 # Go 模块定义
│   ├── Dockerfile             # 后端 Docker 镜像
│   ├── config/                # 配置文件
│   │   └── config.yaml        # 默认配置
│   ├── internal/              # 内部包
│   │   ├── config/            # 配置管理
│   │   ├── database/          # 数据库连接
│   │   ├── models/            # 数据模型
│   │   ├── service/           # 业务逻辑
│   │   ├── handler/           # HTTP 处理器
│   │   ├── middleware/        # 中间件
│   │   ├── router/            # 路由配置
│   │   ├── logger/            # 日志管理
│   │   ├── utils/             # 工具函数
│   │   └── waf/               # WAF 核心引擎
│   ├── sql/                   # SQL 脚本
│   │   ├── final_test_data.sql # 完整测试数据
│   │   └── init_database.sql   # 基础初始化数据
│   └── test_api.sh            # API 测试脚本
├── frontend/                  # 前端代码
│   ├── Dockerfile             # 前端 Docker 镜像
│   ├── src/
│   │   ├── views/             # 页面组件
│   │   ├── layouts/           # 布局组件
│   │   ├── api/               # API 接口
│   │   ├── utils/             # 工具函数
│   │   └── router/            # 路由配置
│   ├── package.json
│   └── vite.config.ts
└── docker-compose.yml         # Docker Compose 配置
```

## 🚀 快速开始

### 方式一：Docker Compose 部署（推荐）

1. **克隆项目**
```bash
git clone <repository-url>
cd waf-go
```

2. **启动所有服务**
```bash
docker-compose up -d
```

3. **访问系统**
- 前端界面：http://localhost
- 后端API：http://localhost:8081

4. **默认账户**
- 用户名：`admin`
- 密码：`admin123`

### 方式二：本地开发部署

#### 环境要求
- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- Redis 6.0+

#### 后端启动

1. **安装依赖**
```bash
cd backend
go mod tidy
```

2. **配置数据库**
   - 创建 MySQL 数据库：`waf`
   - 修改 `config/config.yaml` 中的数据库连接信息
   - 启动 Redis 服务

3. **初始化数据库**
```bash
# 执行基础初始化
mysql -u root -p waf < sql/init_database.sql
```

4. **运行服务**
```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动

#### 前端启动

1. **安装依赖**
```bash
cd frontend
npm install
```

2. **启动开发服务器**
```bash
npm run dev
```

前端将在 `http://localhost:5173` 启动

## 🔧 配置说明

### Docker Compose 配置

系统使用 Docker Compose 进行容器化部署，包含以下服务：

- **MySQL 8.0**：数据存储，端口 3306
- **Redis 7**：缓存服务，端口 6379  
- **WAF Backend**：后端服务，端口 8081
- **WAF Frontend**：前端服务，端口 80

### 后端配置文件 (config/config.yaml)

```yaml
server:
  port: ":8080"
  mode: "debug"

database:
  dsn: "root:password@tcp(localhost:3306)/waf?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

jwt:
  secret: "waf-secret-key-change-in-production"
  expire: 3600

waf:
  rate_limit_window: 60    # 速率限制时间窗口（秒）
  max_requests: 100        # 最大请求数

log:
  level: "debug"
```

## 📝 API 接口

### 认证接口
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/auth/userinfo` - 获取用户信息

### 仪表盘接口
- `GET /api/v1/dashboard/stats` - 获取统计数据
- `GET /api/v1/dashboard/attack-trend` - 获取攻击趋势
- `GET /api/v1/dashboard/rule-distribution` - 获取规则分布
- `GET /api/v1/dashboard/top-attack-ips` - 获取Top攻击IP
- `GET /api/v1/dashboard/top-attack-uris` - 获取Top攻击URI
- `GET /api/v1/dashboard/top-attack-user-agents` - 获取Top攻击User-Agent
- `GET /api/v1/dashboard/realtime-attacks` - 获取实时攻击数据

### 规则管理
- `GET /api/v1/rules` - 获取规则列表
- `POST /api/v1/rules` - 创建规则
- `PUT /api/v1/rules/:id` - 更新规则
- `DELETE /api/v1/rules/:id` - 删除规则
- `PATCH /api/v1/rules/:id/toggle` - 切换规则状态

### 日志查询
- `GET /api/v1/logs/attacks` - 获取攻击日志
- `GET /api/v1/logs/attacks/:id` - 获取日志详情

### 黑名单管理
- `GET /api/v1/blacklists` - 获取黑名单列表
- `POST /api/v1/blacklists` - 添加黑名单
- `PUT /api/v1/blacklists/:id` - 更新黑名单
- `DELETE /api/v1/blacklists/:id` - 删除黑名单
- `PATCH /api/v1/blacklists/:id/toggle` - 切换黑名单状态

## 🧪 测试

### WAF 功能测试
1. 启动后端服务
2. 访问受保护的路径：
   - 正常访问：`curl http://localhost:8081/app/test`
   - 触发拦截：`curl http://localhost:8081/app/admin`

### Docker 环境测试
```bash
# 测试前端访问
curl http://localhost

# 测试后端API
curl http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

## 🚀 部署

### 生产环境部署

1. **修改配置**
```bash
# 修改 docker-compose.yml 中的密码和密钥
# 修改 JWT_SECRET 为生产环境密钥
```

2. **启动服务**
```bash
docker-compose up -d
```

3. **查看服务状态**
```bash
docker-compose ps
docker-compose logs -f
```

4. **停止服务**
```bash
docker-compose down
```

### 数据持久化

Docker Compose 配置包含数据卷持久化：
- MySQL 数据：`mysql_data` 卷
- Redis 数据：`redis_data` 卷

## 📊 数据库结构

主要数据表：
- `users` - 用户表
- `tenants` - 租户表  
- `rules` - WAF 规则表
- `policies` - 策略表
- `attack_logs` - 攻击日志表
- `white_lists` - 白名单表
- `black_lists` - 黑名单表
- `rate_limits` - 速率限制表
- `webhooks` - Webhook 配置表

## 👥 默认测试账户

系统预置了多个测试账户，所有账户的默认密码都是 `admin123`：

- **admin** - 超级管理员
- **tenant_admin_shop** - 电商平台租户管理员
- **tenant_admin_blog** - 博客网站租户管理员  
- **viewer_api** - API服务查看者
- **test_user** - 测试用户（已禁用）

## 🛡️ 安全建议

### 生产环境安全配置

1. **修改默认密码**
```sql
-- 更新admin用户密码
UPDATE users SET password = '$2a$10$your_new_password_hash' WHERE username = 'admin';
```

2. **修改JWT密钥**
```yaml
jwt:
  secret: "your-production-secret-key-at-least-32-characters"
```

3. **数据库安全**
```yaml
# 修改数据库密码
MYSQL_ROOT_PASSWORD: your_secure_password
MYSQL_PASSWORD: your_secure_password
```

4. **网络安全**
- 使用HTTPS
- 配置防火墙规则
- 限制数据库访问IP

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证，详情请参阅 [LICENSE](LICENSE) 文件。

## 🎯 后续规划

- [ ] 规则优先级和冲突处理
- [ ] 地理位置封禁
- [ ] 机器学习威胁检测
- [ ] Webhook 告警集成
- [ ] 性能监控和分析
- [ ] 集群部署支持
- [ ] 更多图表和可视化
- [ ] API 限流和熔断
- [ ] 日志导出和备份
- [ ] 移动端适配

## 📞 支持

如果您在使用过程中遇到问题，请：

1. 查看 [Issues](../../issues) 中是否有相似问题
2. 创建新的 Issue 描述问题
3. 提供详细的错误日志和环境信息

## 🌟 致谢

感谢所有为这个项目做出贡献的开发者！ 