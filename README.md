# LetsTalk 社交 App 后端服务

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
├── config.yaml                  # 配置文件
├── go.mod                      # Go模块依赖管理文件
├── go.sum                      # 依赖校验文件
├── internal/                    # 内部包（只能被本项目引用）
│   ├── config/                 # 配置模块（仅配置读取）
│   │   └── config.go           # 配置加载（使用Viper库读取YAML配置）
│   ├── handler/                # HTTP处理器层（类似MVC中的Controller）
│   │   ├── auth_handler.go     # 认证相关接口（登录、注册等）
│   │   ├── chat_handler.go     # 聊天消息接口（消息列表、好友管理等）
│   │   ├── interaction_handler.go  # 互动相关接口（通知列表、好友操作）
│   │   ├── moment_handler.go   # 动态相关接口（动态列表、发布）
│   │   ├── payment_handler.go  # 支付相关接口（钻石、会员）
│   │   ├── upload_handler.go   # 文件上传接口（头像、图片、音视频、文件）
│   │   ├── user_handler.go     # 用户相关接口（用户信息、关注列表等）
│   │   └── ws_handler.go       # WebSocket连接处理
│   ├── infra/                  # 基础设施层（数据库、Redis、迁移、重置）
│   │   ├── database.go         # 数据库连接（使用GORM连接PostgreSQL）
│   │   ├── migrate.go          # 数据库自动迁移（根据模型自动创建表）
│   │   ├── redis.go            # Redis连接（使用go-redis连接Redis）
│   │   └── reset.go            # 数据重置功能（删除数据库、日志、上传文件、Redis）
│   ├── middleware/             # 中间件（拦截器）
│   │   └── jwt.go              # JWT认证中间件（验证用户登录状态）
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
│   │   ├── helper_test.go    # 测试辅助工具
│   │   └── integration_test.go # 集成测试
│   └── websocket/            # WebSocket实时通信
│       ├── client.go         # WebSocket客户端连接管理
│       ├── hub.go            # WebSocket连接管理器（处理客户端注册/注销）
│       └── message.go        # WebSocket消息类型和数据结构定义
└── pkg/                      # 公共包（可被外部项目引用）
    └── response/             # 响应封装
        └── response.go       # 统一API响应格式
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
中间件处理 (middleware/jwt.go) - JWT认证（私有接口）
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
| bcrypt | 密码加密 | 安全的密码哈希算法 |

---

## 配置说明

编辑 `config.yaml`:

```yaml
# 系统配置
system:
  reset: 0                 # 重置标志(0=不重置, 1=重置所有数据)
  log_level: info          # 日志级别(debug/info/warn/error/fatal)
  sms_valid_minutes: 5     # 短信验证码有效期（分钟）

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
2. 如果配置 `reset=1`，执行数据重置
3. 连接数据库（PostgreSQL，不存在则自动创建）
4. 自动创建所有数据表（根据models.go中的定义）
5. 初始化Redis连接
6. 创建上传文件目录
7. 监听 8080 端口

---

## API接口文档

### 系统模块 `/api/v1/sys`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/sys/code-sms` | 获取手机验证码 | 否 |
| GET | `/api/v1/sys/code-alnum` | 获取图形验证码 | 否 |
| POST | `/api/v1/sys/register` | 用户注册 | 否 |
| POST | `/api/v1/sys/login-code` | 验证码登录 | 否 |
| POST | `/api/v1/sys/login-pwd` | 密码登录 | 否 |
| POST | `/api/v1/sys/logout` | 当前用户退出系统 | 是 |
| POST | `/api/v1/sys/reset-pwd` | 重置密码 | 否 |

### 用户模块 `/api/v1/user`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/user/users` | 用户列表（发现页） | 是 |
| GET | `/api/v1/user/info/:uid` | 特定用户的信息 | 是 |
| GET | `/api/v1/user/focuslist/:uid` | 特定用户的关注列表 | 是 |
| GET | `/api/v1/user/fanslist/:uid` | 特定用户的粉丝列表 | 是 |
| POST | `/api/v1/user/notify/:uid/:flag` | 取消/上线提醒当前用户 | 是 |
| POST | `/api/v1/user/greet/:uid` | 打招呼对特定用户 | 是 |
| POST | `/api/v1/user/upload-avatar` | 上传头像 | 是 |
| POST | `/api/v1/user/collect-myinfo` | 完善个人信息 | 是 |
| POST | `/api/v1/user/collect-aiminfo` | 设置理想对象条件 | 是 |
| GET | `/api/v1/user/adbanner` | 最新广告位用户列表 | 是 |
| POST | `/api/v1/user/gift/:uid/:giftid` | 向特定用户赠送特定礼物 | 是 |
| GET | `/api/v1/user/praise-me` | 赞我的用户列表 | 是 |
| GET | `/api/v1/user/comment-me` | 评论我动态的用户列表 | 是 |
| GET | `/api/v1/user/add-me` | 加我好友的用户列表 | 是 |
| GET | `/api/v1/user/visit-me` | 访问我的用户列表 | 是 |
| GET | `/api/v1/user/like-me` | 喜欢我的用户列表 | 是 |
| POST | `/api/v1/user/agree-friend/:uid/:flag` | 取消/同意好友申请 | 是 |
| POST | `/api/v1/user/add/:uid/:flag` | 取消/添加好友 | 是 |

### 聊天消息模块 `/api/v1/msg`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/msg/sysmsgs` | 当前用户的系统消息列表 | 是 |
| GET | `/api/v1/msg/latest` | 当前用户的最新消息列表 | 是 |
| GET | `/api/v1/msg/:uid` | 与特定用户的消息列表 | 是 |
| POST | `/api/v1/msg/pintop/:uid/:flag` | 取消/置顶用户消息 | 是 |
| POST | `/api/v1/msg/clear/:uid` | 清空聊天记录 | 是 |

### 动态模块 `/api/v1/moment`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/moment/latest` | 所有用户的最新动态列表 | 是 |
| GET | `/api/v1/moment/:uid/latest` | 特定用户的动态列表 | 是 |
| GET | `/api/v1/moment/:mid/comments` | 某个动态的所有评论列表 | 是 |

### 文件上传模块 `/api/v1/upload`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/upload/image` | 上传图片文件 | 是 |
| POST | `/api/v1/upload/audio` | 上传音频文件 | 是 |
| POST | `/api/v1/upload/video` | 上传视频文件 | 是 |
| POST | `/api/v1/upload/file` | 上传文件 | 是 |

### 钻石模块 `/api/v1/diamond`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/diamond/buy/:did` | 购买钻石 | 是 |
| GET | `/api/v1/diamond/stock` | 我的钻石余额 | 是 |
| GET | `/api/v1/diamond/history` | 我的钻石购买历史 | 是 |

### 会员模块 `/api/v1/member`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/member/buy/:vid` | 升级会员 | 是 |
| GET | `/api/v1/member/history` | 我的会员购买历史 | 是 |

### WebSocket模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/ws?token=xxx` | WebSocket连接 | 是（token参数） |

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
const ws = new WebSocket('ws://localhost:8080/ws?token=your_jwt_token');
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
1. 用户登录（POST /sys/login-code 或 /sys/login-pwd）
2. 服务器验证后返回JWT令牌
3. 客户端在后续请求的Header中携带令牌
4. 服务器通过JWT中间件验证令牌
5. 验证通过后，将用户Uid存入上下文，继续处理请求
```

### 用户Uid说明

- 用户对外唯一标识使用雪花ID（字符串类型，varchar(20)）
- 数据库自增ID仅内部使用，不对外暴露
- 防止爬虫遍历用户数据

---

## 日志模块

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

通过配置文件或命令行参数设置日志级别：

**配置文件方式：**
```yaml
system:
  log_level: info
```

**命令行方式：**
```bash
./talkabc.exe --system-log-level=debug
```

### 日志输出

日志同时输出到：
1. **控制台** - 实时查看
2. **文件** - `./logs/app_YYYY-MM-DD.log`

### 日志格式

```
[2026-06-30 10:30:45.123] [INFO] [config/database.go:78] Database connected successfully
```

格式说明：
- `[时间戳]` - 精确到毫秒
- `[日志级别]` - DEBUG/INFO/WARN/ERROR/FATAL
- `[文件名:行号]` - 调用位置
- `消息内容` - 日志正文

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
| `--system-log-level` | string | info | 日志级别 | `system.log_level` |
| `--system-sms-valid-minutes` | int | 5 | 短信验证码有效期（分钟） | `system.sms_valid_minutes` |
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
