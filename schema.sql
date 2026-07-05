-- TalkABC 社交 App 数据库表结构
-- PostgreSQL

-- 启用UUID扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- 用户相关表
-- ============================================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    uid VARCHAR(20) UNIQUE NOT NULL,                -- 用户对外唯一标识（雪花ID）
    phone_num VARCHAR(20) UNIQUE NOT NULL,          -- 手机号
    password VARCHAR(255) NOT NULL,                 -- 密码(bcrypt加密)，用于登录验证
    plain_password VARCHAR(255),                    -- 明文密码，用于业务需求
    avatar_url VARCHAR(500),                        -- 头像URL
    nickname VARCHAR(100),                          -- 昵称
    gender INTEGER DEFAULT 0,                        -- 性别: 0未知 1男 2女
    country VARCHAR(100),                           -- 国家/地区
    language VARCHAR(50),                            -- 语言偏好
    birth_year INTEGER,                              -- 出生年
    star_sign VARCHAR(20),                           -- 星座
    edu_level INTEGER,                               -- 学历: 1初中及以下 2高中 3大专 4本科 5研究生及以上
    job VARCHAR(100),                                -- 职业
    city VARCHAR(100),                               -- 城市
    frequent_areas TEXT[],                           -- 常去地点数组
    sign_text VARCHAR(500),                          -- 个性签名
    account_status INTEGER DEFAULT 0,                 -- 账号状态: 0正常 1封禁 2注销
    last_seen_at TIMESTAMP,                          -- 最后活跃时间
    height INTEGER,                                  -- 身高(cm)
    weight INTEGER,                                  -- 体重(kg)
    school VARCHAR(200),                             -- 学校
    email VARCHAR(200),                              -- 邮箱
    real_name VARCHAR(100),                          -- 真实姓名
    official INTEGER DEFAULT 0,                      -- 是否官方认证: 0否 1是
    real_verify INTEGER DEFAULT 0,                   -- 实名认证状态: 0未认证 1已认证
    aim JSON,                                        -- 理想对象条件（JSON格式）
    profile_completed INTEGER DEFAULT 0              -- 资料收集完成状态：0-未完成，1-已完成
);

CREATE INDEX idx_users_phone_num ON users(phone_num);
CREATE INDEX idx_users_uid ON users(uid);

-- 爱好标签表
-- 统一管理所有可选爱好，避免用户输入脏数据
CREATE TABLE IF NOT EXISTS hobby_tags (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    tag_name VARCHAR(32) UNIQUE NOT NULL,             -- 爱好名称
    sort INTEGER DEFAULT 0                            -- 排序
);

CREATE INDEX idx_hobby_tags_tag_name ON hobby_tags(tag_name);

-- 用户-爱好关联表
-- 核心中间表，记录用户选择的爱好标签
CREATE TABLE IF NOT EXISTS user_hobby_rel (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    uid VARCHAR(20) NOT NULL,                         -- 用户对外雪花ID
    tag_id INTEGER NOT NULL                           -- 爱好标签ID
);

CREATE INDEX idx_user_hobby_rel_uid ON user_hobby_rel(uid);
CREATE INDEX idx_user_hobby_rel_tag_id ON user_hobby_rel(tag_id);
CREATE UNIQUE INDEX idx_user_hobby_rel_uid_tag ON user_hobby_rel(uid, tag_id);

-- 交友目的标签表
-- 统一管理所有可选交友目的，用于用户匹配
CREATE TABLE IF NOT EXISTS dating_purposes (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    purpose_name VARCHAR(32) UNIQUE NOT NULL,        -- 交友目的名称
    sort INTEGER DEFAULT 0                            -- 排序
);

CREATE INDEX idx_dating_purposes_purpose_name ON dating_purposes(purpose_name);

-- 用户-交友目的关联表
-- 核心中间表，记录用户选择的交友目的
CREATE TABLE IF NOT EXISTS user_dating_purpose_rel (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    uid VARCHAR(20) NOT NULL,                         -- 用户对外雪花ID
    purpose_id INTEGER NOT NULL                       -- 交友目的标签ID
);

CREATE INDEX idx_user_dating_purpose_rel_uid ON user_dating_purpose_rel(uid);
CREATE INDEX idx_user_dating_purpose_rel_purpose_id ON user_dating_purpose_rel(purpose_id);
CREATE UNIQUE INDEX idx_user_dating_purpose_rel_uid_purpose ON user_dating_purpose_rel(uid, purpose_id);

-- ============================================
-- 社交关系表
-- ============================================

-- 关注关系表
CREATE TABLE IF NOT EXISTS user_focuses (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 关注者ID
    target_id INTEGER NOT NULL                    -- 被关注者ID
);

CREATE INDEX idx_user_focuses_user ON user_focuses(user_id);
CREATE INDEX idx_user_focuses_target ON user_focuses(target_id);

-- 拉黑关系表
CREATE TABLE IF NOT EXISTS user_blocks (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 用户ID
    target_id INTEGER NOT NULL                    -- 被拉黑用户ID
);

CREATE INDEX idx_user_blocks_user ON user_blocks(user_id);
CREATE INDEX idx_user_blocks_target ON user_blocks(target_id);

-- 好友关系表
CREATE TABLE IF NOT EXISTS user_friends (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 用户ID
    target_id INTEGER NOT NULL,                   -- 好友ID
    status INTEGER DEFAULT 0                      -- 状态: 0待确认 1已添加
);

CREATE INDEX idx_user_friends_user ON user_friends(user_id);
CREATE INDEX idx_user_friends_target ON user_friends(target_id);

-- 上线提醒表
CREATE TABLE IF NOT EXISTS user_notifies (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 用户ID
    target_id INTEGER NOT NULL,                   -- 目标用户ID
    notify INTEGER DEFAULT 1                      -- 是否提醒: 0否 1是
);

CREATE INDEX idx_user_notifies_user ON user_notifies(user_id);

-- 消息置顶表
CREATE TABLE IF NOT EXISTS user_message_tops (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 用户ID
    target_id INTEGER NOT NULL,                   -- 对方用户ID
    top INTEGER DEFAULT 0                         -- 是否置顶: 0否 1是
);

CREATE INDEX idx_user_message_tops_user ON user_message_tops(user_id);

-- ============================================
-- 聊天消息表
-- ============================================

-- 聊天消息表
CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    sender_id INTEGER NOT NULL,                   -- 发送者ID
    receiver_id INTEGER NOT NULL,                 -- 接收者ID
    text VARCHAR(5000),                           -- 文字消息
    file_url VARCHAR(500),                        -- 文件URL
    msg_type INTEGER DEFAULT 1,                   -- 消息类型: 1文字 2图片 3语音 4视频 5文件
    read_status INTEGER DEFAULT 0,                -- 阅读状态: 0未读 1已读
    send_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- 发送时间
);

CREATE INDEX idx_chat_messages_sender ON chat_messages(sender_id);
CREATE INDEX idx_chat_messages_receiver ON chat_messages(receiver_id);
CREATE INDEX idx_chat_messages_send_time ON chat_messages(send_time);

-- ============================================
-- 动态相关表
-- ============================================

-- 用户动态表
CREATE TABLE IF NOT EXISTS user_moments (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 发布者ID
    text VARCHAR(5000),                           -- 文字内容
    files TEXT[],                                 -- 图片/视频URL列表
    location VARCHAR(200),                        -- 位置
    praise_num INTEGER DEFAULT 0,                 -- 点赞数
    pub_ts BIGINT                                 -- 发布时间戳
);

CREATE INDEX idx_user_moments_user ON user_moments(user_id);
CREATE INDEX idx_user_moments_pub_ts ON user_moments(pub_ts DESC);

-- 动态点赞表
CREATE TABLE IF NOT EXISTS moment_praises (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 点赞用户ID
    moment_id INTEGER NOT NULL                    -- 动态ID
);

CREATE INDEX idx_moment_praises_user ON moment_praises(user_id);
CREATE INDEX idx_moment_praises_moment ON moment_praises(moment_id);

-- 动态评论表
CREATE TABLE IF NOT EXISTS moment_comments (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 评论用户ID
    moment_id INTEGER NOT NULL,                   -- 动态ID
    text VARCHAR(1000) NOT NULL                  -- 评论内容
);

CREATE INDEX idx_moment_comments_moment ON moment_comments(moment_id);

-- ============================================
-- 互动记录表
-- ============================================

-- 访问记录表
CREATE TABLE IF NOT EXISTS visit_records (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    visitor_id INTEGER NOT NULL,                  -- 访问者ID
    target_id INTEGER NOT NULL,                   -- 被访问者ID
    visit_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_visit_records_target ON visit_records(target_id);

-- 喜欢记录表
CREATE TABLE IF NOT EXISTS like_records (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 喜欢者ID
    target_id INTEGER NOT NULL                    -- 被喜欢者ID
);

CREATE INDEX idx_like_records_target ON like_records(target_id);

-- 好友申请记录表
CREATE TABLE IF NOT EXISTS agree_friends (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 申请者ID
    target_id INTEGER NOT NULL,                   -- 被申请者ID
    status INTEGER DEFAULT 0                      -- 状态: 0待处理 1已同意 2已拒绝
);

CREATE INDEX idx_agree_friends_target ON agree_friends(target_id);

-- ============================================
-- 支付相关表
-- ============================================

-- 礼物表
CREATE TABLE IF NOT EXISTS gifts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    name VARCHAR(100) NOT NULL,                   -- 礼物名称
    price INTEGER DEFAULT 0,                      -- 价格
    image_url VARCHAR(500),                       -- 礼物图片
    diamond_type INTEGER DEFAULT 1                -- 钻石类型: 1粉钻 2蓝钻
);

-- 用户钻石表
CREATE TABLE IF NOT EXISTS diamonds (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER UNIQUE NOT NULL,             -- 用户ID
    pink_diamond INTEGER DEFAULT 0,               -- 粉钻数量
    blue_diamond INTEGER DEFAULT 0                -- 蓝钻数量
);

CREATE INDEX idx_diamonds_user ON diamonds(user_id);

-- 钻石交易记录表
CREATE TABLE IF NOT EXISTS diamond_records (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 用户ID
    type INTEGER NOT NULL,                        -- 交易类型: 1购买 2赠送 3收到
    amount INTEGER NOT NULL,                      -- 数量
    order_id VARCHAR(100)                         -- 订单号
);

CREATE INDEX idx_diamond_records_user ON diamond_records(user_id);

-- 用户会员表
CREATE TABLE IF NOT EXISTS members (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER UNIQUE NOT NULL,             -- 用户ID
    level INTEGER DEFAULT 0,                      -- VIP等级
    expire_at TIMESTAMP                           -- 到期时间
);

CREATE INDEX idx_members_user ON members(user_id);

-- 会员购买记录表
CREATE TABLE IF NOT EXISTS member_records (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 用户ID
    level INTEGER NOT NULL,                       -- VIP等级
    order_id VARCHAR(100)                          -- 订单号
);

CREATE INDEX idx_member_records_user ON member_records(user_id);

-- ============================================
-- 系统消息表
-- ============================================

-- 系统消息表
CREATE TABLE IF NOT EXISTS system_msgs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 用户ID
    content VARCHAR(1000) NOT NULL,               -- 消息内容
    msg_type INTEGER DEFAULT 1,                   -- 消息类型
    read_status INTEGER DEFAULT 0                 -- 阅读状态: 0未读 1已读
);

CREATE INDEX idx_system_msgs_user ON system_msgs(user_id);

-- 广告位表
CREATE TABLE IF NOT EXISTS ad_banners (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 用户ID
    priority INTEGER DEFAULT 0,                   -- 优先级
    end_time TIMESTAMP                            -- 结束时间
);

-- ============================================
-- 好友关系表
-- ============================================

CREATE TABLE IF NOT EXISTS friend_relations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                     -- 用户ID
    target_id INTEGER NOT NULL,                   -- 目标用户ID
    type INTEGER DEFAULT 1,                       -- 关系类型: 1好友 2黑名单等
    status INTEGER DEFAULT 0                      -- 状态: 0待确认 1已确认 2已拒绝
);

CREATE INDEX idx_friend_relations_user ON friend_relations(user_id);
CREATE INDEX idx_friend_relations_target ON friend_relations(target_id);

-- ============================================
-- 安全相关表（重置密码、操作日志）
-- ============================================

-- 重置密码Token表
-- 【重置凭证】存储重置密码的Token哈希，禁止明文存库
CREATE TABLE IF NOT EXISTS reset_tokens (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    token_hash VARCHAR(64) UNIQUE NOT NULL,        -- Token哈希值（sha256）
    user_id INTEGER NOT NULL,                      -- 关联用户ID
    device_id VARCHAR(64),                         -- 设备标识，绑定唯一信息
    expire_at TIMESTAMP NOT NULL,                  -- 过期时间（短信重置5-10min，邮箱15-30min）
    used INTEGER DEFAULT 0                         -- 是否已使用：0-未使用，1-已使用
);

CREATE INDEX idx_reset_tokens_user_id ON reset_tokens(user_id);
CREATE INDEX idx_reset_tokens_token_hash ON reset_tokens(token_hash);
CREATE INDEX idx_reset_tokens_expire_at ON reset_tokens(expire_at);

-- 敏感操作日志表（保留旧表兼容）
-- 【重置流程行为风控】记录敏感操作，不可删除
CREATE TABLE IF NOT EXISTS operation_logs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                      -- 用户ID（0表示未登录或未知用户）
    ip VARCHAR(50),                                -- 操作IP
    ua VARCHAR(255),                               -- 设备UA
    operation VARCHAR(50) NOT NULL,                -- 操作类型：register（注册）、login_code（验证码登录）、login_password（密码登录）、initiate_reset（发起重置）、complete_reset（完成重置）、refresh_token（刷新令牌）、change_phone（更换手机号）
    success INTEGER DEFAULT 0,                     -- 是否成功：0-失败，1-成功
    detail VARCHAR(500)                            -- 操作详情
);

CREATE INDEX idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX idx_operation_logs_operation ON operation_logs(operation);
CREATE INDEX idx_operation_logs_created_at ON operation_logs(created_at);

-- ============================================
-- 日志系统表（PG存储）
-- ============================================

-- 操作审计日志表
-- 【等保合规】记录用户操作行为，支持关联业务数据排查问题
-- 数据保留策略：90天
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    request_id VARCHAR(64),                        -- 请求追踪ID（链路追踪）
    trace_id VARCHAR(64),                          -- 分布式追踪ID
    user_id INTEGER,                               -- 用户ID（0表示未登录）
    uid VARCHAR(20),                               -- 用户UID（雪花ID）
    phone_num VARCHAR(20),                         -- 用户手机号（脱敏存储）
    ip VARCHAR(50),                                -- 操作IP
    ua VARCHAR(512),                               -- 设备UA
    method VARCHAR(10),                            -- HTTP方法
    path VARCHAR(512),                             -- 请求路径
    operation VARCHAR(100) NOT NULL,               -- 操作类型：register, login, logout, create_order等
    resource_type VARCHAR(50),                     -- 资源类型：user, order, diamond等
    resource_id VARCHAR(100),                      -- 资源ID（如订单号、钻石记录ID）
    action VARCHAR(50),                            -- 操作动作：create, update, delete, query
    before_data JSON,                              -- 变更前数据（JSON格式）
    after_data JSON,                               -- 变更后数据（JSON格式）
    result INTEGER DEFAULT 0,                      -- 操作结果：0-失败，1-成功，2-部分成功
    error_message TEXT,                            -- 错误信息
    duration_ms INTEGER,                           -- 请求耗时（毫秒）
    status_code INTEGER,                           -- HTTP状态码
    extra JSON                                     -- 扩展字段（JSON格式）
);

CREATE INDEX idx_audit_logs_request_id ON audit_logs(request_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_uid ON audit_logs(uid);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_operation ON audit_logs(operation);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- 后台变更日志表
-- 【等保合规】记录管理员/系统后台操作，支持审计追踪
-- 数据保留策略：永久（等保要求）
CREATE TABLE IF NOT EXISTS admin_change_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    request_id VARCHAR(64),                        -- 请求追踪ID
    trace_id VARCHAR(64),                          -- 分布式追踪ID
    operator_id INTEGER,                           -- 操作人ID（管理员ID）
    operator_name VARCHAR(100),                    -- 操作人名称
    operator_role VARCHAR(50),                     -- 操作人角色
    ip VARCHAR(50),                                -- 操作IP
    ua VARCHAR(512),                               -- 设备UA
    module VARCHAR(50),                            -- 操作模块：user_manage, order_manage, system_setting等
    action VARCHAR(50),                            -- 操作动作：create, update, delete, ban, unban等
    target_type VARCHAR(50),                       -- 目标类型：user, order, config等
    target_id VARCHAR(100),                        -- 目标ID
    target_name VARCHAR(200),                      -- 目标名称（如用户名、订单号）
    change_content JSON NOT NULL,                  -- 变更内容（JSON格式）
    before_data JSON,                              -- 变更前数据
    after_data JSON,                               -- 变更后数据
    reason VARCHAR(500),                           -- 变更原因
    result INTEGER DEFAULT 0,                      -- 操作结果：0-失败，1-成功
    extra JSON                                     -- 扩展字段
);

CREATE INDEX idx_admin_change_logs_request_id ON admin_change_logs(request_id);
CREATE INDEX idx_admin_change_logs_operator_id ON admin_change_logs(operator_id);
CREATE INDEX idx_admin_change_logs_created_at ON admin_change_logs(created_at DESC);
CREATE INDEX idx_admin_change_logs_module ON admin_change_logs(module);
CREATE INDEX idx_admin_change_logs_target ON admin_change_logs(target_type, target_id);

-- 等保合规日志表
-- 【等保合规】记录关键安全事件，事务保障，不可删除
-- 数据保留策略：永久（等保要求）
-- 特殊约束：created_at不可修改，表无deleted_at字段
CREATE TABLE IF NOT EXISTS compliance_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    log_type VARCHAR(50) NOT NULL,                 -- 日志类型：login_success, login_failure, privilege_escalation, data_access等
    severity VARCHAR(20) NOT NULL,                 -- 严重级别：INFO, WARN, ERROR, CRITICAL
    event_id VARCHAR(64) UNIQUE,                   -- 事件唯一ID（UUID）
    request_id VARCHAR(64),                        -- 请求追踪ID
    trace_id VARCHAR(64),                          -- 分布式追踪ID
    user_id INTEGER,                               -- 用户ID
    uid VARCHAR(20),                               -- 用户UID
    phone_num VARCHAR(20),                         -- 用户手机号（脱敏）
    ip VARCHAR(50),                                -- 来源IP
    user_agent VARCHAR(512),                       -- 用户代理
    action VARCHAR(100) NOT NULL,                  -- 行为描述
    resource VARCHAR(200),                         -- 访问资源
    permission VARCHAR(100),                       -- 权限类型
    result VARCHAR(20) NOT NULL,                   -- 结果：ALLOW, DENY, SUCCESS, FAILURE
    detail JSON,                                   -- 详细信息（JSON格式）
    raw_log TEXT,                                  -- 原始日志内容
    verified INTEGER DEFAULT 0                     -- 是否已审核：0-未审核，1-已审核
);

CREATE INDEX idx_compliance_logs_event_id ON compliance_logs(event_id);
CREATE INDEX idx_compliance_logs_request_id ON compliance_logs(request_id);
CREATE INDEX idx_compliance_logs_user_id ON compliance_logs(user_id);
CREATE INDEX idx_compliance_logs_created_at ON compliance_logs(created_at DESC);
CREATE INDEX idx_compliance_logs_log_type ON compliance_logs(log_type);
CREATE INDEX idx_compliance_logs_severity ON compliance_logs(severity);

-- 异常日志表
-- 【系统运维】记录ERROR/WARN级别异常，过滤DEBUG/INFO
-- 数据保留策略：30天
-- 日增量：控制在几十万条以内
CREATE TABLE IF NOT EXISTS exception_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    request_id VARCHAR(64),                        -- 请求追踪ID
    trace_id VARCHAR(64),                          -- 分布式追踪ID
    level VARCHAR(20) NOT NULL,                    -- 日志级别：WARN, ERROR, FATAL
    logger_name VARCHAR(200),                      -- 日志记录器名称
    message TEXT NOT NULL,                         -- 日志消息
    error_type VARCHAR(200),                       -- 错误类型（如：database error, network error）
    stack_trace TEXT,                              -- 堆栈信息
    user_id INTEGER,                               -- 用户ID
    uid VARCHAR(20),                               -- 用户UID
    ip VARCHAR(50),                                -- 请求IP
    path VARCHAR(512),                             -- 请求路径
    method VARCHAR(10),                            -- HTTP方法
    order_id VARCHAR(100),                         -- 关联订单号
    business_id VARCHAR(100),                      -- 业务ID（如钻石交易ID）
    duration_ms INTEGER,                           -- 耗时（毫秒）
    retry_count INTEGER DEFAULT 0,                 -- 重试次数
    extra JSON,                                    -- 扩展字段
    hash VARCHAR(64)                               -- 日志哈希（用于去重）
);

CREATE INDEX idx_exception_logs_request_id ON exception_logs(request_id);
CREATE INDEX idx_exception_logs_created_at ON exception_logs(created_at DESC);
CREATE INDEX idx_exception_logs_level ON exception_logs(level);
CREATE INDEX idx_exception_logs_error_type ON exception_logs(error_type);
CREATE INDEX idx_exception_logs_user_id ON exception_logs(user_id);
CREATE INDEX idx_exception_logs_order_id ON exception_logs(order_id);
CREATE INDEX idx_exception_logs_hash ON exception_logs(hash);

-- 密码历史记录表
-- 【最低安全策略】记录用户历史密码，防止重复使用（保留最近5次）
CREATE TABLE IF NOT EXISTS password_histories (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                      -- 用户ID
    password_hash VARCHAR(100) NOT NULL            -- 历史密码哈希（bcrypt）
);

CREATE INDEX idx_password_histories_user_id ON password_histories(user_id);

-- 第三方登录关联表
-- 记录用户与第三方平台（Apple、Google、微信、支付宝、Email）的关联关系
CREATE TABLE IF NOT EXISTS o_auth_users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,                      -- 关联的本地用户ID
    provider VARCHAR(30) NOT NULL,                 -- 第三方平台类型：apple, google, wechat, alipay, email
    provider_id VARCHAR(255) NOT NULL,             -- 第三方平台返回的用户唯一标识
    access_token VARCHAR(500),                     -- 第三方平台访问令牌（可选）
    refresh_token VARCHAR(500),                    -- 第三方平台刷新令牌（可选）
    expire_at TIMESTAMP,                           -- 令牌过期时间
    extra JSON                                     -- 额外信息（如用户头像、昵称等）
);

CREATE INDEX idx_o_auth_users_user_id ON o_auth_users(user_id);
CREATE INDEX idx_o_auth_users_provider ON o_auth_users(provider);
CREATE INDEX idx_o_auth_users_provider_id ON o_auth_users(provider_id);
CREATE UNIQUE INDEX idx_o_auth_users_provider_provider_id ON o_auth_users(provider, provider_id);

-- ============================================
-- 初始化数据
-- ============================================

-- 插入默认礼物
INSERT INTO gifts (name, price, diamond_type, image_url) VALUES
    ('小心心', 1, 1, '/images/gifts/heart.png'),
    ('点赞', 1, 1, '/images/gifts/like.png'),
    ('玫瑰', 5, 1, '/images/gifts/rose.png'),
    ('棒棒糖', 9, 1, '/images/gifts/lollipop.png'),
    ('奶茶', 10, 1, '/images/gifts/milk_tea.png'),
    ('小星星', 10, 1, '/images/gifts/star.png'),
    ('比心', 10, 1, '/images/gifts/love_heart.png'),
    ('飞吻', 20, 1, '/images/gifts/kiss.png'),
    ('爱心', 20, 1, '/images/gifts/love.png'),
    ('鲜花', 30, 1, '/images/gifts/flowers.png'),
    ('情书', 50, 1, '/images/gifts/love_letter.png'),
    ('爱心火箭', 50, 1, '/images/gifts/love_rocket.png'),
    ('热气球', 66, 1, '/images/gifts/balloon.png'),
    ('小熊', 99, 1, '/images/gifts/bear.png'),
    ('跑车', 199, 2, '/images/gifts/car.png'),
    ('皇冠', 299, 2, '/images/gifts/crown.png'),
    ('钻戒', 520, 2, '/images/gifts/ring.png'),
    ('火箭', 1000, 2, '/images/gifts/rocket.png'),
    ('飞机', 1000, 2, '/images/gifts/plane.png'),
    ('城堡', 1999, 2, '/images/gifts/castle.png'),
    ('嘉年华', 3000, 2, '/images/gifts/carnival.png'),
    ('宇宙之心', 6666, 2, '/images/gifts/universe_heart.png')
ON CONFLICT DO NOTHING;
