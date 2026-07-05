package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), m)
	}
	return json.Unmarshal(bytes, m)
}

// User 用户模型
// 对应数据库中的 users 表，存储用户基本信息
type User struct {
	gorm.Model
	Uid           string         `gorm:"type:varchar(20);unique;not null;index"` // 用户对外唯一标识（雪花ID），防止爬虫遍历
	PhoneNum      string         `gorm:"unique;not null"`                        // 手机号，用于登录和验证
	Password      string         `gorm:"not null"`                               // 密码（bcrypt加密）
	AvatarURL     string                                                         // 头像URL
	Nickname      string                                                         // 昵称
	Gender        int                                                            // 性别：0-未知，1-男，2-女
	Country       string                                                         // 国家/地区
	Language      string                                                         // 语言偏好
	BirthYear     int                                                            // 出生年份
	StarSign      string                                                         // 星座
	EduLevel      int                                                            // 教育程度：1-初中及以下，2-高中，3-大专，4-本科，5-研究生及以上
	Job           string                                                         // 职业
	City          string                                                         // 城市
	FrequentAreas pq.StringArray `gorm:"type:text[]"`                           // 常去地点数组
	SignText      string                                                         // 个性签名
	AccountStatus int                                                            // 账号状态：0-正常，1-封禁，2-注销
	LastSeenAt    time.Time                                                      // 最后活跃时间
	Height        int                                                            // 身高（cm）
	Weight        int                                                            // 体重（kg）
	School        string                                                         // 学校
	Email         string                                                         // 邮箱
	RealName      string                                                         // 真实姓名
	Official      int                 // 是否官方认证：0-否，1-是
	RealVerify       int                 // 实名认证状态：0-未认证，1-已认证
	Aim              JSONMap             `gorm:"type:json"`                         // 理想对象条件（JSON格式）
	ProfileCompleted int                 `gorm:"default:0"`                        // 资料收集完成状态：0-未完成，1-已完成
}

// FriendRelation 好友关系模型
// 对应数据库中的 friend_relations 表
type FriendRelation struct {
	gorm.Model              // 内置模型字段
	UserID    uint          // 用户ID
	TargetID  uint          // 目标用户ID
	Type      int           // 关系类型：1-好友，2-黑名单等
	Status    int           // 状态：0-待确认，1-已确认，2-已拒绝
}

// ChatMessage 聊天消息模型
// 对应数据库中的 chat_messages 表
type ChatMessage struct {
	gorm.Model                // 内置模型字段
	SenderID   uint          // 发送者用户ID
	ReceiverID  uint          // 接收者用户ID
	Text        string        // 消息文本内容
	FileURL     string        // 文件URL（图片、语音等）
	MsgType     int           // 消息类型：1-文本，2-图片，3-语音，4-视频，5-文件
	ReadStatus  int           // 已读状态：0-未读，1-已读
	SendTime    time.Time     // 发送时间
}

// UserMoment 用户动态/朋友圈模型
// 对应数据库中的 user_moments 表
type UserMoment struct {
	gorm.Model                 // 内置模型字段
	UserID    uint             // 发布动态的用户ID
	Text      string           // 动态文字内容
	Files     pq.StringArray `gorm:"type:text[]"` // 动态图片/视频URL数组
	Location  string           // 动态发布的地理位置
	PraiseNum int              // 点赞数量
	PubTS     int64            // 发布时间戳
}

// MomentPraise 动态点赞模型
// 对应数据库中的 moment_praises 表
type MomentPraise struct {
	gorm.Model     // 内置模型字段
	UserID   uint   // 点赞的用户ID
	MomentID uint   // 被点赞的动态ID
}

// MomentComment 动态评论模型
// 对应数据库中的 moment_comments 表
type MomentComment struct {
	gorm.Model    // 内置模型字段
	UserID   uint  // 评论的用户ID
	MomentID uint  // 被评论的动态ID
	Text     string // 评论内容
}

// Gift 礼物模型
// 对应数据库中的 gifts 表
type Gift struct {
	gorm.Model           // 内置模型字段
	Name        string   // 礼物名称
	Price       int      // 礼物价格（钻石数）
	ImageURL    string   // 礼物图片URL
	DiamondType int      // 钻石类型：1-粉钻，2-蓝钻
}

// Diamond 用户钻石账户模型
// 对应数据库中的 diamonds 表
type Diamond struct {
	gorm.Model           // 内置模型字段
	UserID      uint     // 用户ID
	PinkDiamond int      // 粉钻数量
	BlueDiamond int      // 蓝钻数量
}

// DiamondRecord 钻石交易记录模型
// 对应数据库中的 diamond_records 表
type DiamondRecord struct {
	gorm.Model           // 内置模型字段
	UserID      uint     // 用户ID
	Type        int      // 交易类型：1-购买，2-赠送，3-收到
	Amount      int      // 交易数量
	OrderID     string   // 订单号
}

// Member 用户会员模型
// 对应数据库中的 members 表
type Member struct {
	gorm.Model           // 内置模型字段
	UserID    uint       // 用户ID
	Level     int        // 会员等级：1-月度，2-季度，3-年度，99-永久
	ExpireAt  time.Time  // 到期时间
}

// MemberRecord 会员购买记录模型
// 对应数据库中的 member_records 表
type MemberRecord struct {
	gorm.Model           // 内置模型字段
	UserID    uint       // 用户ID
	Level     int        // 会员等级
	OrderID   string     // 订单号
}

// SystemMsg 系统消息模型
// 对应数据库中的 system_msgs 表
type SystemMsg struct {
	gorm.Model          // 内置模型字段
	UserID      uint     // 接收消息的用户ID
	Content     string   // 消息内容
	MsgType     int      // 消息类型：1-系统通知，2-活动通知等
	ReadStatus  int      // 已读状态：0-未读，1-已读
}

// AdBanner 广告横幅模型
// 对应数据库中的 ad_banners 表
type AdBanner struct {
	gorm.Model          // 内置模型字段
	UserID   uint        // 广告对应的用户（推广者）
	Priority int         // 优先级，数字越大越靠前
	EndTime  time.Time  // 下线时间
}

// VisitRecord 访问记录模型
// 记录用户访问其他用户主页的轨迹
type VisitRecord struct {
	gorm.Model           // 内置模型字段
	VisitorID uint        // 访问者ID
	TargetID  uint        // 被访问的用户ID
	VisitTime time.Time   // 访问时间
}

// LikeRecord 喜欢记录模型
// 记录用户对其他用户表示喜欢的数据
type LikeRecord struct {
	gorm.Model     // 内置模型字段
	UserID   uint   // 喜欢方用户ID
	TargetID uint   // 被喜欢方用户ID
}

// AgreeFriend 同意好友请求模型
// 记录用户处理好友请求的结果
type AgreeFriend struct {
	gorm.Model     // 内置模型字段
	UserID   uint   // 用户ID
	TargetID uint   // 目标用户ID
	Status   int    // 处理状态：0-待处理，1-已同意，2-已拒绝
}

// UserNotify 用户通知设置模型
// 记录用户对特定联系人的通知设置
type UserNotify struct {
	gorm.Model     // 内置模型字段
	UserID   uint   // 用户ID
	TargetID uint   // 目标用户ID
	Notify   int    // 是否接收通知：0-关闭，1-开启
}

// UserBlock 用户拉黑记录模型
// 记录用户拉黑的其他用户
type UserBlock struct {
	gorm.Model     // 内置模型字段
	UserID   uint   // 用户ID
	TargetID uint   // 被拉黑的用户ID
}

// UserFocus 用户关注记录模型
// 记录用户关注的其他用户
type UserFocus struct {
	gorm.Model     // 内置模型字段
	UserID   uint   // 用户ID
	TargetID uint   // 被关注的用户ID
}

// UserFriend 用户好友关系模型
// 记录用户之间的好友关系
type UserFriend struct {
	gorm.Model     // 内置模型字段
	UserID   uint   // 用户ID
	TargetID uint   // 好友用户ID
	Status   int    // 好友状态：0-待确认，1-已添加
}

// UserMessageTop 用户消息置顶记录模型
// 记录用户置顶的聊天会话
type UserMessageTop struct {
	gorm.Model     // 内置模型字段
	UserID   uint   // 用户ID
	TargetID uint   // 置顶会话对方的用户ID
	Top      int    // 置顶状态：0-取消置顶，1-置顶
}

// HobbyTag 爱好标签模型
// 统一管理所有可选爱好，避免用户输入脏数据
type HobbyTag struct {
	gorm.Model
	TagName string `gorm:"unique;size:32"` // 爱好名称
	Sort    int    `gorm:"default:0"`      // 排序
}

// UserHobbyRel 用户-爱好关联模型
// 记录用户选择的爱好标签
type UserHobbyRel struct {
	gorm.Model
	Uid   string `gorm:"size:20;index"` // 用户对外雪花ID
	TagID uint   `gorm:"index"`         // 爱好标签ID
}

// DatingPurpose 交友目的标签模型
// 统一管理所有可选交友目的，用于用户匹配
type DatingPurpose struct {
	gorm.Model
	PurposeName string `gorm:"unique;size:32"` // 交友目的名称
	Sort        int    `gorm:"default:0"`      // 排序
}

// UserDatingPurposeRel 用户-交友目的关联模型
// 记录用户选择的交友目的标签
type UserDatingPurposeRel struct {
	gorm.Model
	Uid       string `gorm:"size:20;index"` // 用户对外雪花ID
	PurposeID uint   `gorm:"index"`         // 交友目的标签ID
}

// ResetToken 重置密码Token模型
// 【重置凭证】存储重置密码的Token哈希，禁止明文存库
type ResetToken struct {
	gorm.Model
	TokenHash  string    `gorm:"unique;size:64;not null"` // Token哈希值（sha256）
	UserID     uint      `gorm:"not null;index"`          // 关联用户ID
	DeviceID   string    `gorm:"size:64"`                 // 设备标识，绑定唯一信息
	ExpireAt   time.Time `gorm:"not null"`                // 过期时间（短信重置5-10min，邮箱15-30min）
	Used       int       `gorm:"default:0"`               // 是否已使用：0-未使用，1-已使用
}

// OperationLog 敏感操作日志模型
// 【重置流程行为风控】记录敏感操作，不可删除
type OperationLog struct {
	gorm.Model
	UserID    uint   `gorm:"not null;index"` // 用户ID
	IP        string `gorm:"size:50"`        // 操作IP
	UA        string `gorm:"size:255"`       // 设备UA
	Operation string `gorm:"size:50"`        // 操作类型：initiate_reset（发起重置）、complete_reset（完成重置）
	Success   int    `gorm:"default:0"`      // 是否成功：0-失败，1-成功
	Detail    string `gorm:"size:500"`       // 操作详情
}

// PasswordHistory 密码历史记录模型
// 【最低安全策略】记录用户历史密码，防止重复使用
type PasswordHistory struct {
	gorm.Model
	UserID       uint   `gorm:"not null;index"` // 用户ID
	PasswordHash string `gorm:"size:100;not null"` // 历史密码哈希（bcrypt）
}
