# TalkABC 社交 App 后端服务

基于 Go + Gin + GORM + PostgreSQL + Redis 开发的社交应用后端 API 服务，支持 WebSocket 实时通信。

---

## 目录

- [项目结构](#项目结构)
- [架构说明](#架构说明)
- [配置说明](#配置说明)
- [数据库配置](#数据库配置)
- [启动服务](#启动服务)
- [API接口文档](#api接口文档)
- [WebSocket实时通信](#websocket实时通信)
- [统一响应格式](#统一响应格式)
- [认证方式](#认证方式)
- [日志模块](#日志模块)
- [数据重置](#数据重置)
- [命令行参数](#命令行参数)

---

## 项目结构

```
backend/
├── cmd/                          # 程序入口目录
│   └── server/
│       └── main.go              # 程序入口文件
├── config/                       # 配置文件目录
│   └── logger.yaml              # 日志独立配置文件
├── config.yaml                  # 主配置文件
├── go.mod                      # Go模块依赖管理文件
├── go.sum                      # 依赖校验文件
├── internal/                    # 内部包（只能被本项目引用）
│   ├── config/                 # 配置模块（仅配置读取）
│   │   └── config.go           # 配置加载（使用Viper库读取YAML配置）
│   ├── handler/                # HTTP处理器层（类似MVC中的Controller）
│   │   ├── auth_login_handler.go    # 登录相关接口
│   │   ├── auth_register_handler.go # 注册/密码重置接口
│   │   ├── chat_handler.go          # 聊天消息接口（消息列表、好友管理等）
│   │   ├── interaction_handler.go   # 互动相关接口（通知列表、好友操作）
│   │   ├── moment_handler.go        # 动态相关接口（动态列表、发布）
│   │   ├── payment_handler.go       # 支付相关接口（钻石、会员）
│   │   ├── profile_handler.go       # 个人资料接口（完善资料、偏好设置）
│   │   ├── sms_handler.go           # 短信验证码接口
│   │   ├── sys_handler.go           # 系统接口（日志级别管理）
│   │   ├── upload_handler.go        # 文件上传接口（头像、图片、音视频、文件）
│   │   ├── user_handler.go          # 用户相关接口（用户信息、关注列表等）
│   │   └── ws_handler.go            # WebSocket连接处理
│   ├── infra/                  # 基础设施层（数据库、Redis、迁移、重置）
│   │   ├── database.go         # 数据库连接（使用GORM连接PostgreSQL）
│   │   ├── migrate.go          # 数据库自动迁移（根据模型自动创建表）
│   │   ├── redis.go            # Redis连接（使用go-redis连接Redis）
│   │   └── reset.go            # 数据重置功能（删除数据库、日志、上传文件、Redis）
│   ├── middleware/             # 中间件（拦截器）
│   │   ├── jwt.go              # JWT认证中间件（验证用户登录状态）
│   │   └── logger.go           # 请求日志中间件（统一记录请求信息）
│   ├── model/                 # 数据模型层（对应数据库表结构）
│   │   └── models.go          # 所有数据模型定义
│   ├── repository/            # 数据访问层（直接操作数据库）
│   │   ├── auth_repository.go
│   │   ├── chat_repository.go
│   │   ├── interaction_repository.go
│   │   ├── moment_repository.go
│   │   ├── payment_repository.go
│   │   └── user_repository.go
│   ├── router/               # 路由配置
│   │   └── router.go         # Gin路由注册（使用路由分组统一管理中间件）
│   ├── service/              # 业务逻辑层（核心业务处理）
│   │   ├── auth_service.go
│   │   ├── chat_service.go
│   │   ├── interaction_service.go
│   │   ├── moment_service.go
│   │   ├── payment_service.go
│   │   └── user_service.go
│   ├── test/                 # 测试代码
│   │   ├── helper_test.go    # 测试辅助工具（统一初始化日志、Redis）
│   │   ├── integration_test.go # 集成测试
│   │   ├── auth_login_test.go  # 登录模块测试
│   │   ├── auth_register_test.go # 注册模块测试
│   │   ├── sms_test.go         # 短信验证码测试
│   │   ├── user_test.go        # 用户模块测试
│   │   └── moment_test.go      # 动态模块测试
│   └── websocket/            # WebSocket实时通信
│       ├── client.go         # WebSocket客户端连接管理
│       ├── hub.go            # WebSocket连接管理器（处理客户端注册/注销）
│       └── message.go        # WebSocket消息类型和数据结构定义
└── pkg/                      # 公共包（可被外部项目引用）
    ├── logger/               # 企业级日志模块（zap + lumberjack）
    │   └── logger.go         # 日志核心实现
    ├── response/             # 响应封装
    │   └── response.go       # 统一API响应格式
    └── utils/                # 工具函数
        └── snowflake.go      # 雪花ID生成器
```

---

## 架构说明

### 分层架构

项目采用经典的三层架构设计，每层有明确的职责：

```
┌─────────────────────────────────────────────────────────┐
│                    Handler 层                           │
│   (接收请求、参数验证、调用Service、返回响应)              │
├─────────────────────────────────────────────────────────┤
│                    Service 层                           │
│   (处理业务逻辑、数据转换、调用Repository)                │
├─────────────────────────────────────────────────────────┤
│                   Repository 层                         │
│   (直接与数据库交互、CRUD操作)                           │
├─────────────────────────────────────────────────────────┤
│                      数据库                              │
│               PostgreSQL + Redis                        │
└─────────────────────────────────────────────────────────┘
```

### 请求处理流程

```
客户端请求
    ↓
路由匹配 (router.go)
    ↓
请求日志中间件 (middleware/logger.go) - 生成request_id，记录请求开始
    ↓
CORS中间件 (cors) - 跨域处理
    ↓
JWT认证中间件 (middleware/jwt.go) - 私有接口验证
    ↓
Handler处理 (handler/*.go) - 参数验证、调用Service
    ↓
Service处理 (service/*.go) - 业务逻辑
    ↓
Repository处理 (repository/*.go) - 数据库操作
    ↓
数据库响应
    ↓
逐层返回
    ↓
请求日志中间件 - 记录请求结束（状态码、耗时）
    ↓
客户端收到JSON响应
```

### WebSocket处理流程

```
客户端WebSocket连接请求
    ↓
WS Handler验证JWT token
    ↓
创建Client并注册到Hub
    ↓
启动ReadPump和WritePump协程
    ↓
接收客户端消息 → 解析 → 分发到对应处理函数
    ↓
处理业务逻辑 → 通过Hub推送给目标客户端
    ↓
客户端收到实时消息
```

### 技术栈

| 技术 | 用途 | 说明 |
|------|------|------|
| Gin | HTTP Web框架 | 高性能路由和中间件，支持路由分组 |
| GORM | ORM数据库操作 | 简化数据库操作，自动迁移 |
| PostgreSQL | 关系型数据库 | 存储用户、消息、动态等数据 |
| Redis | 缓存/实时数据存储 | 存储验证码、在线状态、会话管理 |
| go-redis/v8 | Redis客户端 | Go语言Redis驱动 |
| gorilla/websocket | WebSocket库 | 实现实时双向通信 |
| JWT | 用户认证 | 无状态令牌认证 |
| Viper | 配置管理 | 读取YAML配置文件和命令行参数 |
| bcrypt | 密码加密 | 安全的密码哈希算法（cost=10） |
| zap | 结构化日志 | 高性能、结构化日志库 |
| lumberjack | 日志文件切割 | 自动按大小/时间分割日志文件 |
| snowflake | ID生成 | 分布式唯一ID生成器 |

---

## 配置说明

### 主配置文件 (config.yaml)

编辑 `config.yaml`:

```yaml
# 系统配置
system:
  reset: 0                 # 重置标志(0=不重置, 1=重置所有数据)

# 日志配置（已独立到 config/logger.yaml，此处保留为兼容）
logger:
  level: debug          # 日志级别: debug, info, warn, error, fatal
  format: console       # 日志格式: console, json
  output: both          # 输出方式: console, file, both
  file_path: ./logs/app.log
  max_size: 100         # 单个日志文件最大大小(MB)
  max_backups: 30       # 保留的最大日志文件数
  max_age: 7            # 日志文件保留天数
  compress: true        # 是否压缩归档日志

# 安全配置
security:
  sms_valid_minutes: 5              # 短信验证码有效期（分钟）
  sms_cooldown_seconds: 60          # 短信发送冷却时间（秒）
  sms_hourly_limit: 10              # 每小时发送次数限制
  ip_register_hourly_limit: 10      # 注册时每小时每个IP发送次数限制
  ip_login_minute_limit: 10         # 登录时每分钟每个IP登录次数限制
  login_failure_lock_minutes: 5     # 登录失败锁定时间（分钟）

# 短信服务商配置
sms_provider:
  default: "aliyun"              # 默认短信服务商：aliyun, huawei, tencent
  aliyun:
    access_key_id: ""            # 阿里云AccessKey ID
    access_key_secret: ""        # 阿里云AccessKey Secret
    region_id: "cn-hangzhou"     # 阿里云区域ID
    sign_name: ""                # 短信签名
    template_code: ""            # 短信模板Code
  huawei:
    app_key: ""                  # 华为云AppKey
    app_secret: ""               # 华为云AppSecret
    sign_name: ""                # 短信签名
    template_id: ""              # 短信模板ID
  tencent:
    secret_id: ""                # 腾讯云SecretID
    secret_key: ""               # 腾讯云SecretKey
    region_id: "ap-guangzhou"    # 腾讯云区域ID
    sign_name: ""                # 短信签名
    template_id: ""              # 短信模板ID

# 服务器配置
server:
  port: 8080              # 服务监听端口

# 数据库配置
database:
  host: localhost         # 数据库服务器地址
  port: 5432             # PostgreSQL默认端口
  user: postgres          # 数据库用户名
  password: admin         # 数据库密码
  dbname: talkabc        # 数据库名称
  sslmode: disable        # 禁用SSL（开发环境）

# JWT认证配置
jwt:
  secret: your_secret_key # JWT签名密钥（生产环境应使用复杂密钥）
  expires_hour: 24        # Token过期时间（小时）

# 文件上传配置
upload:
  avatar_path: ./uploads/avatars    # 头像上传目录
  moment_path: ./uploads/moments     # 动态上传目录
  message_path: ./uploads/messages   # 消息文件上传目录

# Redis配置
redis:
  host: localhost                  # Redis服务器地址
  port: 6379                      # Redis默认端口（版本5）
  password: ""                    # Redis密码（无密码时为空）
  db: 0                           # Redis数据库编号

# CORS配置
cors:
  origins:
    - "*"                         # 允许所有来源（生产环境应改为具体域名）
  methods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  headers:
    - Origin
    - Content-Type
    - Authorization
    - Accept
    - X-Requested-With
  credentials: true               # 是否允许携带凭证
```

### 日志配置文件 (config/logger.yaml)

日志配置已独立为单独文件，优先级高于 config.yaml 中的 logger 配置：

```yaml
level: debug          # 日志级别: debug, info, warn, error, fatal
format: console       # 日志格式: console（人类可读）, json（机器可读）
output: both          # 输出方式: console, file, both
file_path: ./logs/app.log  # 日志文件路径
max_size: 100         # 单个日志文件最大大小(MB)
max_backups: 30       # 保留的最大日志文件数
max_age: 7            # 日志文件保留天数
compress: true        # 是否压缩归档日志
```

---

## 数据库配置

### PostgreSQL

确保 PostgreSQL 已安装并运行：

```sql
-- 创建数据库（程序会自动检测并创建，无需手动操作）
CREATE DATABASE talkabc;
```

### Redis

确保 Redis 5.x 已安装并运行在本机 6379 端口：

```bash
# Windows 启动Redis
redis-server.exe

# 验证Redis连接
redis-cli ping
# 返回 PONG 表示连接成功
```

---

## 启动服务

```bash
# 进入项目目录
cd backend

# 下载依赖（根据go.mod安装所有依赖包）
go mod tidy

# 启动服务（会自动创建数据表）
go run cmd/server/main.go

# 或编译后运行
go build -o talkabc.exe ./cmd/server
./talkabc.exe
```

服务启动后会：
1. 加载配置文件（config.yaml）
2. 加载日志配置文件（config/logger.yaml）
3. 如果配置 `reset=1`，执行数据重置
4. 连接数据库（PostgreSQL，不存在则自动创建）
5. 自动创建所有数据表（根据models.go中的定义）
6. 初始化Redis连接
7. 初始化短信网关
8. 创建上传文件目录
9. 监听 8080 端口

---

## API接口文档

### 认证模块 `/api/v1/auth`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/auth/code-sms` | 获取手机验证码 | 否 |
| POST | `/api/v1/auth/code-sms/verify` | 验证手机验证码 | 否 |
| GET | `/api/v1/auth/code-alnum` | 获取图形验证码 | 否 |
| POST | `/api/v1/auth/code-alnum/verify` | 验证图形验证码 | 否 |
| POST | `/api/v1/auth/register` | 用户注册 | 否 |
| POST | `/api/v1/auth/login/code` | 验证码登录 | 否 |
| POST | `/api/v1/auth/login/password` | 密码登录 | 否 |
| POST | `/api/v1/auth/refresh-token` | 刷新访问令牌 | 否 |
| POST | `/api/v1/auth/logout` | 用户退出 | 是 |
| POST | `/api/v1/auth/change-phone` | 更换手机号 | 是 |
| POST | `/api/v1/auth/reset-password/initiate` | 发起密码重置 | 否 |
| GET | `/api/v1/auth/reset-password/validate` | 验证重置Token | 否 |
| POST | `/api/v1/auth/reset-password/complete` | 完成密码重置 | 否 |

### 用户模块 `/api/v1/users`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/users` | 用户列表（发现页） | 是 |
| GET | `/api/v1/users/:uid` | 用户信息 | 是 |
| GET | `/api/v1/users/:uid/following` | 关注列表 | 是 |
| GET | `/api/v1/users/:uid/fans` | 粉丝列表 | 是 |
| POST | `/api/v1/users/:uid/greet` | 打招呼 | 是 |
| POST | `/api/v1/users/:uid/notification/:flag` | 设置通知 | 是 |

### 个人资料模块 `/api/v1/profile`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/profile/me` | 完善个人信息 | 是 |
| POST | `/api/v1/profile/preferences` | 设置理想对象条件 | 是 |

### 聊天消息模块 `/api/v1/messages`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/messages/system` | 系统消息列表 | 是 |
| GET | `/api/v1/messages/latest` | 最新消息列表 | 是 |
| GET | `/api/v1/messages/:uid` | 与特定用户的消息列表 | 是 |
| POST | `/api/v1/messages/top/:uid/:flag` | 置顶/取消置顶消息 | 是 |
| DELETE | `/api/v1/messages/:uid` | 清空聊天记录 | 是 |

### 动态模块 `/api/v1/moments`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/moments/latest` | 最新动态列表 | 是 |
| GET | `/api/v1/users/:uid/moments` | 特定用户的动态列表 | 是 |
| GET | `/api/v1/moments/:mid/comments` | 动态评论列表 | 是 |
| POST | `/api/v1/moments` | 发布动态 | 是 |

### 文件上传模块 `/api/v1/uploads`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/users/avatar` | 上传头像 | 是 |
| POST | `/api/v1/uploads/image` | 上传图片文件 | 是 |
| POST | `/api/v1/uploads/audio` | 上传音频文件 | 是 |
| POST | `/api/v1/uploads/video` | 上传视频文件 | 是 |
| POST | `/api/v1/uploads/file` | 上传文件 | 是 |

### 互动通知模块 `/api/v1/notifications`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/notifications/praise` | 赞我的列表 | 是 |
| GET | `/api/v1/notifications/comment` | 评论我的列表 | 是 |
| GET | `/api/v1/notifications/friend` | 加我的列表 | 是 |
| GET | `/api/v1/notifications/visit` | 访问我的列表 | 是 |
| GET | `/api/v1/notifications/like` | 喜欢我的列表 | 是 |

### 好友模块 `/api/v1/friendships`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/friendships/:uid/:flag` | 添加/取消好友 | 是 |
| POST | `/api/v1/friendships/agree/:uid/:flag` | 同意/拒绝好友申请 | 是 |

### 礼物模块 `/api/v1/gifts`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/gifts/send/:uid/:giftid` | 赠送礼物 | 是 |

### 广告模块 `/api/v1/ads`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/ads/latest` | 最新广告位 | 是 |

### 钻石模块 `/api/v1/diamonds`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/diamonds/buy/:did` | 购买钻石 | 是 |
| GET | `/api/v1/diamonds/stock` | 钻石余额 | 是 |
| GET | `/api/v1/diamonds/history` | 购买历史 | 是 |

### 会员模块 `/api/v1/memberships`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/memberships/buy/:vid` | 购买会员 | 是 |
| GET | `/api/v1/memberships/history` | 购买历史 | 是 |

### 系统模块 `/api/v1/system`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/system/log-level` | 获取当前日志级别 | 是 |
| POST | `/api/v1/system/log-level` | 动态修改日志级别 | 是 |

### WebSocket模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/ws?token=xxx&deviceId=xxx` | WebSocket连接 | 是（token参数） |
| GET | `/api/v1/onlinestatus` | 获取在线状态 | 是 |

### WebSocket消息类型（客户端发送）

以下操作通过 WebSocket 发送，不再提供 HTTP 接口：

| 消息类型 | 说明 | 参数 |
|----------|------|------|
| `send_text` | 发送文本消息 | `to_uid`, `text` |
| `send_image` | 发送图片消息 | `to_uid`, `file_url` |
| `send_video` | 发送视频消息 | `to_uid`, `file_url` |
| `send_voice` | 发送语音消息 | `to_uid`, `file_url` |
| `send_file` | 发送文件消息 | `to_uid`, `file_url` |
| `send_withdraw` | 撤回消息 | `to_uid`, `msg_id` |
| `focus_user` | 关注/取消关注用户 | `to_uid`, `flag` |
| `block_user` | 拉黑/取消拉黑用户 | `to_uid`, `flag` |
| `like_user` | 喜欢/取消喜欢用户 | `to_uid`, `flag` |
| `praise_moment` | 点赞/取消点赞动态 | `moment_id` |
| `comment_moment` | 评论动态 | `moment_id`, `text` |
| `report_moment` | 举报动态 | `moment_id` |

---

## WebSocket实时通信

### 连接方式

客户端通过 WebSocket 协议连接服务器，需在 URL 参数中携带 JWT token：

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=your_jwt_token&deviceId=your_device_id');
```

### 消息协议

服务端推送的消息采用 JSON 格式，包含以下字段：

```json
{
  "type": "chat",      // 消息类型
  "from_uid": "1",     // 发送者用户ID（字符串类型）
  "to_uid": "2",       // 接收者用户ID（字符串类型）
  "data": {}           // 消息数据
}
```

### 消息类型

| 类型 | 说明 | 数据结构 |
|------|------|----------|
| `chat` | 聊天消息 | `{"id": 1, "text": "消息内容", "file_url": "图片/音视频URL", "msg_type": 1, "send_time": 1678901234}` |
| `withdraw` | 撤回消息 | `{"msg_id": 123}` |
| `friend_request` | 好友申请 | `{"from_uid": "1", "from_name": "用户名", "from_avatar": "头像URL"}` |
| `comment` | 评论通知 | `{"moment_id": 1, "from_uid": "2", "from_name": "用户名", "text": "评论内容"}` |
| `praise` | 点赞通知 | `{"moment_id": 1, "from_uid": "2", "from_name": "用户名"}` |
| `online_status` | 在线状态变更 | `{"uid": "1", "online": 1}` |
| `offline_status` | 离线状态变更 | `{"uid": "1"}` |
| `system` | 系统消息 | `{"text": "系统消息内容"}` |

### 消息类型说明

**msg_type（聊天消息类型）：**
- 1 - 文本消息
- 2 - 图片消息
- 3 - 语音消息
- 4 - 视频消息
- 5 - 文件消息

**online（在线状态）：**
- 0 - 离线
- 1 - 在线

### 在线状态管理

系统通过 Redis 存储和管理用户在线状态：

**Redis Key设计：**
```
online:user:{uid}  # Set结构，存储用户所有在线设备ID
```

**心跳机制：**
- 客户端每 30s 发送 ping 消息
- 服务端收到 ping 后更新 Redis Key 过期时间为 90s
- 超过 90s 未收到 ping，Key 自动过期，判定用户离线

**多端登录：**
- 支持手机、平板、网页同时在线
- 每个设备有独立的 deviceId
- 任意设备发送 ping 会续期整组在线状态

---

## 统一响应格式

所有API接口都使用以下统一格式返回数据：

```json
{
  "code": 0,           // 状态码：0表示成功，其他值表示失败
  "msg": "success",   // 消息：成功时为"success"，失败时为错误信息
  "data": {}          // 数据：成功时返回数据，失败时为null
}
```

### 响应示例

**成功响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "uid": "12345678901234567890",
    "nickname": "Tom"
  }
}
```

**失败响应：**
```json
{
  "code": 1001,
  "msg": "用户不存在",
  "data": null
}
```

---

## 认证方式

除公开接口外，其他接口需要在请求头中携带JWT令牌：

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### JWT认证流程

```
1. 用户登录（POST /api/v1/auth/login/code 或 /api/v1/auth/login/password）
2. 服务器验证后返回 access_token 和 refresh_token
3. 客户端在后续请求的Header中携带 access_token
4. 服务器通过JWT中间件验证令牌
5. 验证通过后，将用户Uid存入上下文，继续处理请求
6. access_token过期后，使用refresh_token获取新令牌（POST /api/v1/auth/refresh-token）
```

### 双令牌机制

系统采用双令牌机制提升安全性：

| 令牌类型 | 有效期 | 用途 | 存储位置 |
|----------|--------|------|----------|
| access_token | 2小时 | 日常API访问认证 | 请求头Authorization |
| refresh_token | 7天 | 获取新的access_token | 客户端本地存储 |

**刷新令牌流程：**
```
1. 客户端检测access_token过期（401错误）
2. 使用refresh_token调用POST /api/v1/auth/refresh-token
3. 服务器验证refresh_token有效性（格式、签名、Redis存储校验）
4. 生成新的access_token和refresh_token（令牌轮转）
5. 旧令牌自动失效
6. 返回新令牌给客户端
```

**安全规则：**
- refresh_token存储在Redis中，退出登录时立即失效
- 每次刷新生成新的refresh_token（令牌轮转）
- JWT中间件验证请求中的token与Redis中存储的token一致，防止旧token滥用
- 操作日志记录所有刷新令牌操作，不可删除

### 用户Uid说明

- 用户对外唯一标识使用雪花ID（字符串类型，varchar(20)）
- 数据库自增ID仅内部使用，不对外暴露
- 防止爬虫遍历用户数据

---

## 日志模块

### 核心设计

项目采用企业级日志方案，基于 **zap + lumberjack** 实现：

| 特性 | 说明 |
|------|------|
| 全局单例 | 整个项目只初始化一次 logger，禁止到处 new 日志对象 |
| 独立配置 | 通过 `config/logger.yaml` 统一控制日志行为 |
| 分层输出 | 同时输出到控制台和文件 |
| 自动切割 | 按大小/时间分割日志，自动清理过期日志 |
| 链路追踪 | 支持从 context 中提取 request_id/trace_id |
| 动态级别 | 运行时可通过 API 动态修改日志级别 |

### 日志级别

支持以下日志级别（从低到高）：

| 级别 | 说明 | 适用场景 |
|------|------|----------|
| `debug` | 调试信息 | 开发环境，详细的调试日志 |
| `info` | 一般信息 | 默认级别，记录关键操作和状态 |
| `warn` | 警告信息 | 可能的问题，但不影响正常运行 |
| `error` | 错误信息 | 操作失败或异常情况 |
| `fatal` | 致命错误 | 严重错误，程序将退出 |

### 日志配置

通过独立配置文件 `config/logger.yaml` 设置：

```yaml
level: debug          # 日志级别
format: console       # console/json
output: both          # console/file/both
file_path: ./logs/app.log
max_size: 100         # MB
max_backups: 30       # 保留备份数
max_age: 7            # 保留天数
compress: true        # 是否压缩
```

### 请求日志中间件

系统自动记录每个HTTP请求的完整信息：

```
INFO  Request start - method: POST, path: /api/v1/auth/login/code, client_ip: 192.168.1.100
INFO  Request end - method: POST, path: /api/v1/auth/login/code, status: 200, latency: 15.3ms
```

### 动态日志级别

支持通过 API 实时修改日志级别：

```bash
# 获取当前日志级别
curl -H "Authorization: Bearer xxx" http://localhost:8080/api/v1/system/log-level

# 修改日志级别为 debug
curl -H "Authorization: Bearer xxx" -X POST -d '{"level":"debug"}' http://localhost:8080/api/v1/system/log-level
```

### 日志输出

日志同时输出到：
1. **控制台** - 实时查看，人类可读格式
2. **文件** - `./logs/app.log`，自动按大小切割

### 日志格式

**Console格式（开发环境）：**
```
[2026-07-03 10:30:45.123] [INFO] [service:talkabc] [env:development] [request_id:abc123] Request processed
```

**JSON格式（生产环境）：**
```json
{
  "time": "2026-07-03T10:30:45.123Z",
  "level": "info",
  "service": "talkabc",
  "env": "development",
  "request_id": "abc123",
  "message": "Request processed"
}
```

### 敏感信息脱敏

系统提供工具函数对敏感信息进行脱敏处理：

| 函数 | 说明 | 示例 |
|------|------|------|
| `logger.MaskToken` | Token脱敏 | `abc123def456` → `abc1***456` |
| `logger.MaskSensitive` | 通用敏感信息脱敏 | 根据长度动态隐藏中间部分 |

---

## 数据重置

### 重置功能说明

当 `system.reset=1` 时，启动程序会自动执行以下重置操作：

| 操作 | 说明 |
|------|------|
| 删除数据库 | 删除并重建 PostgreSQL 数据库 |
| 删除日志 | 删除 `./logs` 目录及所有日志文件 |
| 删除上传文件 | 删除 `./uploads` 目录及所有上传文件 |
| 清空Redis | 清空 Redis 当前数据库中的所有数据 |

### 使用方式

**配置文件方式：**
```yaml
system:
  reset: 1
```

**命令行方式：**
```bash
./talkabc.exe --system-reset=1
```

### 注意事项

1. **谨慎使用**：重置操作会删除所有数据，无法恢复
2. **开发环境**：建议仅在开发环境使用此功能
3. **重置后**：数据库会自动重建并执行迁移，无需手动操作
4. **自动重置**：重置完成后程序会继续正常启动

---

## 命令行参数

### 参数说明

程序支持通过命令行参数覆盖配置文件中的设置，参数优先级：**命令行参数 > 配置文件 > 默认值**。

| 参数 | 类型 | 默认值 | 说明 | 配置文件对应 |
|------|------|--------|------|-------------|
| `--system-reset` | int | 0 | 系统重置标志(0=不重置, 1=重置) | `system.reset` |
| `--server-port` | int | 8080 | 服务器监听端口 | `server.port` |
| `--database-host` | string | localhost | 数据库主机地址 | `database.host` |
| `--database-port` | int | 5432 | 数据库端口 | `database.port` |
| `--database-user` | string | postgres | 数据库用户名 | `database.user` |
| `--database-password` | string | "" | 数据库密码 | `database.password` |
| `--database-name` | string | talkabc | 数据库名称 | `database.dbname` |
| `--database-sslmode` | string | disable | 数据库SSL模式 | `database.sslmode` |
| `--jwt-secret` | string | talkabc_secret_key | JWT签名密钥 | `jwt.secret` |
| `--jwt-expireshour` | int | 24 | JWT过期时间（小时） | `jwt.expires_hour` |
| `--redis-host` | string | localhost | Redis主机地址 | `redis.host` |
| `--redis-port` | int | 6379 | Redis端口 | `redis.port` |
| `--redis-password` | string | "" | Redis密码 | `redis.password` |
| `--redis-db` | int | 0 | Redis数据库编号 | `redis.db` |
| `--upload-avatarpath` | string | ./uploads/avatars | 头像上传路径 | `upload.avatar_path` |
| `--upload-momentpath` | string | ./uploads/moments | 动态上传路径 | `upload.moment_path` |
| `--upload-messagepath` | string | ./uploads/messages | 消息文件上传路径 | `upload.message_path` |

### 使用示例

**使用配置文件启动：**
```bash
./talkabc.exe
```

**覆盖单个配置：**
```bash
./talkabc.exe --server-port=8081
```

**覆盖多个配置：**
```bash
./talkabc.exe --server-port=8081 --database-host=192.168.1.100 --database-password=123456
```

**完全通过命令行配置（无需配置文件）：**
```bash
./talkabc.exe --database-host=localhost --database-user=postgres --database-password=admin --database-name=talkabc
```

**重置数据后启动：**
```bash
./talkabc.exe --system-reset=1
```

**查看帮助信息：**
```bash
./talkabc.exe --help
```

### 优先级说明

配置的加载优先级从高到低：

1. **命令行参数**（最高优先级）
2. **配置文件**（`config.yaml`）
3. **默认值**（最低优先级）

例如：如果配置文件中 `server.port` 设为 8080，但命令行传入 `--server-port=8081`，则实际使用 8081。

---

## 数据库表结构

详见 [schema.sql](./schema.sql) 文件。

---

## 项目变更记录

- 1. 创建 `infra` 包，将数据库、Redis、迁移、重置功能从 `config` 包分离
- 2. 用户模型 `User` 的在线状态由 Redis 维护，移除数据库字段
- 3. 用户模型 `User` 新增字符串类型 `Uid`（雪花ID），防止爬虫遍历
- 4. 验证码 `VerificationCode` 模型从 PostgreSQL 移到 Redis，有效期由 `system.sms_valid_minutes` 控制
- 5. 收发消息、撤回消息、关注、拉黑、点赞、评论、举报等操作改为 WebSocket 实现
- 6. 文件上传接口统一到 `upload_handler.go`
- 7. 使用路由分组统一管理 JWT 中间件
- 8. 新增爱好字典表 `hobby_tags` 和用户-爱好关联表 `user_hobby_rel`
- 9. 会员表 `vip` 改名为 `member`
- 10. 密码加密方式从自定义盐+SHA256改为 bcrypt
- 11. 用户最后活跃时间 `last_seen_at` 在用户每次通过WebSocket发送消息时更新
- 12. 所有handler、service、repository方法增加详细注释说明业务流程
- 13. 所有测试代码增加注释说明测试内容
- 14. 使用路由分组统一添加JWT中间件，避免每个路由单独指定
- 15. 新增交友目的标签表 `dating_purposes` 和用户-交友目的关联表 `user_dating_purpose_rel`
- 16. 更新礼物初始数据，参考抖音礼物品类和价格体系（22种礼物）
- 17. 修复config.go中viper默认值键名与结构体字段不匹配的问题
- 18. 修复路由重复前缀问题（/api/v1/api/v1），添加fallback路由处理
- 19. 修复短信验证码每小时发送限制逻辑，使用SETNX初始化Redis计数器
- 20. 实现企业级日志系统（zap + lumberjack），支持结构化日志、文件切割、日志分级
- 21. 日志配置独立为 `config/logger.yaml` 文件，便于维护
- 22. 新增请求日志中间件，自动记录每个HTTP请求的方法、路径、状态码、耗时
- 23. 支持运行时动态修改日志级别（通过API接口）
- 24. 实现上下文链路追踪，自动从context提取request_id/trace_id
- 25. 新增敏感信息脱敏工具函数（MaskToken、MaskSensitive）
- 26. 分离单元测试和集成测试，集成测试移到 `internal/test` 目录
- 27. 所有Handler方法添加请求参数日志记录
- 28. Handler层拆分为独立文件（auth_login、auth_register、profile、sys）
- 29. 新增系统API接口（获取/修改日志级别）
- 30. 实现双令牌机制（access_token + refresh_token），支持令牌刷新和轮转
- 31. 新增刷新令牌接口 POST /api/v1/auth/refresh-token
- 32. 新增更换手机号接口 POST /api/v1/auth/change-phone（需JWT认证）
- 33. JWT中间件增强：验证Redis中存储的token与请求token一致，防止旧token滥用
- 34. 退出登录时清除access_token和refresh_token，确保完全退出
- 35. 操作日志增加refresh_token、change_phone操作类型记录