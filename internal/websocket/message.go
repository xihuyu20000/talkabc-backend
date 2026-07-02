package websocket

// MessageType WebSocket消息类型定义
type MessageType string

const (
	// ========== 服务端推送消息类型 ==========
	MessageTypeChat          MessageType = "chat"          // 聊天消息（文本/图片/语音/视频/文件）
	MessageTypeWithdraw      MessageType = "withdraw"      // 消息撤回通知
	MessageTypeFriendReq     MessageType = "friend_request" // 好友请求通知
	MessageTypeComment       MessageType = "comment"       // 评论通知
	MessageTypePraise        MessageType = "praise"        // 点赞通知
	MessageTypeOnline        MessageType = "online_status" // 上线状态通知
	MessageTypeOffline       MessageType = "offline_status" // 下线状态通知
	MessageTypeSystem        MessageType = "system"        // 系统消息

	// ========== 客户端发送消息类型 ==========
	MessageTypeSendText      MessageType = "send_text"      // 发送文本消息
	MessageTypeSendImage     MessageType = "send_image"     // 发送图片消息
	MessageTypeSendVideo     MessageType = "send_video"     // 发送视频消息
	MessageTypeSendVoice     MessageType = "send_voice"     // 发送语音消息
	MessageTypeSendFile      MessageType = "send_file"      // 发送文件消息
	MessageTypeSendWithdraw  MessageType = "send_withdraw"  // 撤回消息

	// ========== 客户端操作消息类型 ==========
	MessageTypeFocusUser     MessageType = "focus_user"     // 关注/取消关注用户
	MessageTypeBlockUser     MessageType = "block_user"     // 拉黑/取消拉黑用户
	MessageTypeLikeUser      MessageType = "like_user"      // 喜欢/取消喜欢用户
	MessageTypePraiseMoment  MessageType = "praise_moment"  // 点赞/取消点赞动态
	MessageTypeCommentMoment MessageType = "comment_moment" // 评论动态
	MessageTypeReportMoment  MessageType = "report_moment"  // 举报动态
)

// WSMessage WebSocket消息结构体
// 所有WebSocket消息的统一格式
type WSMessage struct {
	Type    MessageType `json:"type"`     // 消息类型
	FromUID string      `json:"from_uid"` // 发送者UID
	ToUID   string      `json:"to_uid"`   // 接收者UID
	Data    interface{} `json:"data"`     // 消息数据（根据消息类型不同而变化）
}

// ChatMessageData 聊天消息数据结构
// 用于文本、图片、语音、视频、文件等聊天消息
type ChatMessageData struct {
	ID       uint   `json:"id"`        // 消息ID（服务端返回时填充）
	Text     string `json:"text"`      // 文本内容（仅文本消息使用）
	FileURL  string `json:"file_url"`  // 文件URL（图片/语音/视频/文件消息使用）
	MsgType  int    `json:"msg_type"`  // 消息类型：1-文本，2-图片，3-语音，4-视频，5-文件
	SendTime int64  `json:"send_time"` // 发送时间戳（秒）
}

// WithdrawMessageData 消息撤回数据结构
type WithdrawMessageData struct {
	MsgID uint `json:"msg_id"` // 被撤回的消息ID
}

// FriendRequestData 好友请求数据结构
type FriendRequestData struct {
	FromUID    string `json:"from_uid"`    // 请求发送者UID
	FromName   string `json:"from_name"`   // 请求发送者昵称
	FromAvatar string `json:"from_avatar"` // 请求发送者头像URL
}

// CommentData 评论数据结构
type CommentData struct {
	MomentID uint   `json:"moment_id"` // 被评论的动态ID
	FromUID  string `json:"from_uid"`  // 评论者UID
	FromName string `json:"from_name"` // 评论者昵称
	Text     string `json:"text"`      // 评论内容
}

// PraiseData 点赞数据结构
type PraiseData struct {
	MomentID uint   `json:"moment_id"` // 被点赞的动态ID
	FromUID  string `json:"from_uid"`  // 点赞者UID
	FromName string `json:"from_name"` // 点赞者昵称
}

// OnlineStatusData 在线状态数据结构
type OnlineStatusData struct {
	UID    string `json:"uid"`    // 用户UID
	Online int    `json:"online"` // 在线状态：1-在线，0-离线
}
