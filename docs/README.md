# TalkABC 后端服务文档

基于 Go + Gin + GORM + PostgreSQL + Redis 开发的社交应用后端 API 服务，支持 WebSocket 实时通信。

---

## 文档导航

根据您的角色选择对应的文档：

| 角色 | 文档 | 主要内容 |
|------|------|----------|
| 前端开发人员 | [面向前端开发人员.md](./面向前端开发人员.md) | API接口、WebSocket、响应格式、认证方式、资料收集流程 |
| 后端开发人员 | [面向后端开发人员.md](./面向后端开发人员.md) | 项目结构、架构说明、配置说明、日志模块、命令行参数、开发规范 |
| 后端测试人员 | [面向后端测试人员.md](./面向后端测试人员.md) | API接口、认证方式、测试类型与策略、核心测试场景、安全测试要点 |
| 后端运维人员 | [面向后端运维人员.md](./面向后端运维人员.md) | 配置说明、数据库配置、启动服务、构建脚本、数据重置、部署流程、监控告警 |
| 数据迁移人员 | [面向数据迁移人员.md](./面向数据迁移人员.md) | 数据库架构、数据迁移方案、迁移步骤、数据验证、回滚方案 |

---

## 快速开始

### 环境要求

| 工具 | 版本 | 说明 |
|------|------|------|
| Go | >= 1.21 | 编程语言 |
| PostgreSQL | >= 14 | 关系型数据库 |
| Redis | >= 6.0 | 缓存数据库 |

### 启动服务

```bash
# 进入项目目录
cd talkabc-backend

# 设置开发环境
export APP_ENV=dev

# 下载依赖
go mod tidy

# 启动服务
go run ./cmd/server
```

服务启动后：
- API地址：`http://localhost:8080/api/v1`
- Swagger文档：`http://localhost:8080/swagger/index.html`

### Apifox API管理

项目使用 **Apifox** 作为API管理平台，提供完整的API文档、测试和调试功能。

**Apifox配置文件：** `talkabc.apifox.json`

**导入方式：**
1. 打开Apifox客户端
2. 点击"导入/导出" → "导入数据" → "从文件导入"
3. 选择 `talkabc.apifox.json` 文件
4. 点击"确定"完成导入

**导入后配置：**
- 环境名称：`TalkABC 本地开发`
- 基础URL：`http://localhost:8080`
- 认证方式：Bearer Token
- 自动填充Token：开启

**Apifox特性：**
- 📋 **API文档**：完整的接口定义、参数说明、响应示例
- 🔄 **一键调试**：直接在Apifox中调用API，无需编写代码
- 📝 **Mock数据**：支持Mock响应，前端可独立开发
- 🧪 **测试用例**：可编写和运行自动化测试用例
- 📊 **接口监控**：支持接口可用性监控

**Apifox下载：** [https://www.apifox.com/download](https://www.apifox.com/download)

### 环境变量配置

项目支持通过环境变量覆盖配置文件中的设置，以下是常用环境变量：

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| APP_ENV | 运行环境（dev/test/prod） | dev |
| APP_CONFIG | 配置文件路径 | config/config.yaml |
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户名 | postgres |
| DB_PASSWORD | 数据库密码 | 空 |
| DB_NAME | 数据库名称 | talkabc |
| REDIS_HOST | Redis主机 | localhost |
| REDIS_PORT | Redis端口 | 6379 |
| REDIS_PASSWORD | Redis密码 | 空 |
| REDIS_DB | Redis数据库编号 | 0 |
| JWT_SECRET | JWT签名密钥 | 空 |

#### Linux/Mac 临时生效

```bash
export APP_ENV=dev
export DB_PASSWORD=your_password
export REDIS_PASSWORD=your_redis_password
export JWT_SECRET=your_jwt_secret_key
```

#### Linux 永久生效

```bash
# Bash
echo 'export APP_ENV=dev' >> ~/.bashrc
source ~/.bashrc

# Zsh
echo 'export APP_ENV=dev' >> ~/.zshrc
source ~/.zshrc
```

#### Mac 永久生效

```bash
# Bash
echo 'export APP_ENV=dev' >> ~/.bash_profile
source ~/.bash_profile

# Zsh
echo 'export APP_ENV=dev' >> ~/.zshrc
source ~/.zshrc
```

#### Windows 10/11 临时生效

**PowerShell：**

```powershell
$env:APP_ENV="dev"
$env:DB_PASSWORD="your_password"
```

**CMD：**

```cmd
set APP_ENV=dev
set DB_PASSWORD=your_password
```

#### Windows 10/11 永久生效

按 `Win + R` → 输入 `sysdm.cpl` → 高级 → 环境变量 → 新建系统/用户变量

---

## 项目结构

```
talkabc-backend/
├── cmd/server/                 # 程序入口
├── config/                    # 配置文件
├── internal/                  # 内部包
│   ├── handler/               # HTTP处理器层
│   ├── service/               # 业务逻辑层
│   ├── repository/            # 数据访问层
│   ├── model/                 # 数据模型层
│   ├── middleware/            # 中间件
│   ├── router/                # 路由配置
│   ├── infra/                 # 基础设施层
│   └── websocket/             # WebSocket通信
├── pkg/                       # 公共包
│   ├── logger/                # 日志模块
│   ├── response/              # 响应封装
│   ├── security/              # 安全检查
│   └── utils/                 # 工具函数
├── swagger/                   # Swagger文档
├── tests/                     # 测试代码
└── docs/                      # 项目文档
```

---

## 技术栈

| 技术 | 用途 |
|------|------|
| Gin | HTTP Web框架 |
| GORM | ORM数据库操作 |
| PostgreSQL | 关系型数据库 |
| Redis | 缓存/实时数据存储 |
| gorilla/websocket | WebSocket库 |
| JWT | 用户认证 |
| Viper | 配置管理 |
| zap + lumberjack | 结构化日志 |
| bcrypt | 密码加密 |
| snowflake | ID生成 |

---

## 核心特性

- **双令牌机制**：access_token + refresh_token，支持令牌刷新和轮转
- **企业级日志**：基于zap + lumberjack，支持结构化日志、文件切割、动态级别
- **多环境配置**：支持开发、测试、生产环境切换，环境变量引用
- **实时通信**：基于WebSocket的即时消息、在线状态、互动通知
- **内容安全**：昵称和签名的敏感词过滤、URL过滤、XSS过滤
- **资料收集**：用户注册后首次访问的资料收集流程

---

## API接口概览

### 认证模块 `/api/v1/auth`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login/code` | 验证码登录 |
| POST | `/api/v1/auth/login/password` | 密码登录 |
| POST | `/api/v1/auth/refresh-token` | 刷新访问令牌 |
| POST | `/api/v1/auth/logout` | 用户退出 |
| POST | `/api/v1/auth/change-phone` | 更换手机号 |

### 用户模块 `/api/v1/users`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/users` | 用户列表 |
| GET | `/api/v1/users/:uid` | 用户信息 |
| GET | `/api/v1/users/:uid/following` | 关注列表 |

### 个人资料模块 `/api/v1/profile`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/profile/status` | 检查资料收集状态 |
| POST | `/api/v1/profile/sign` | 设置个性签名 |
| POST | `/api/v1/profile/complete` | 完成资料收集 |

---

## 统一响应格式

所有API接口使用统一格式返回：

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

---

## 更多信息

详细信息请根据您的角色查看对应文档。