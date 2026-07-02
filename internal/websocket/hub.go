package websocket

import (
	"backend/internal/config"
	"context"
	"encoding/json"
	"strings"
	"time"
)

// ==================== 常量定义 ====================

const (
	onlineKeyPrefix = "online:user:" // Redis在线状态Key前缀
	onlineExpire    = 90 * time.Second // 在线状态过期时间（心跳超时阈值）
)

// ==================== Hub 结构体 ====================

// Hub WebSocket连接管理器
// 负责管理所有WebSocket客户端连接，处理客户端注册、注销和消息广播
type Hub struct {
	clients    map[string][]*Client // 用户连接映射，key为uid，value为客户端列表（支持多端在线）
	register   chan *Client         // 客户端注册通道
	unregister chan *Client         // 客户端注销通道
	broadcast  chan *WSMessage      // 消息广播通道
}

var hub *Hub

// init 初始化Hub单例
// 在包加载时自动创建Hub实例并启动后台运行协程
func init() {
	hub = &Hub{
		clients:    make(map[string][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *WSMessage),
	}
	go hub.Run()
}

// GetHub 获取Hub单例
// 返回全局唯一的Hub实例
func GetHub() *Hub {
	return hub
}

// ==================== Hub 核心方法 ====================

// Register 注册客户端连接
// 参数说明：
//   - client: 要注册的客户端实例
//
// 逻辑：
//   将客户端发送到register通道，由Run协程统一处理
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Run Hub主循环
// 监听register、unregister和broadcast三个通道，处理客户端注册、注销和消息广播
//
// 处理逻辑：
//   1. register通道：添加客户端到连接映射，更新Redis在线状态，广播上线通知
//   2. unregister通道：移除客户端，更新Redis在线状态，若用户离线则广播下线通知
//   3. broadcast通道：将消息发送给目标用户的所有在线客户端
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.uid] = append(h.clients[client.uid], client)
			h.addDeviceOnline(client.uid, client.deviceId)
			h.broadcastOnlineStatus(client.uid, 1)
		case client := <-h.unregister:
			h.removeClient(client)
			h.removeDeviceOnline(client.uid, client.deviceId)
			if !h.IsOnline(client.uid) {
				h.broadcastOnlineStatus(client.uid, 0)
			}
		case message := <-h.broadcast:
			if clients, ok := h.clients[message.ToUID]; ok {
				for _, client := range clients {
					select {
					case client.send <- message:
					default:
						h.removeClient(client)
					}
				}
			}
		}
	}
}

// removeClient 从Hub中移除客户端连接
// 参数说明：
//   - client: 要移除的客户端实例
//
// 逻辑：
//   1. 从clients映射中找到该用户的客户端列表
//   2. 找到目标客户端并关闭其send通道
//   3. 从列表中删除该客户端
//   4. 若用户无在线客户端，从映射中删除该用户
func (h *Hub) removeClient(client *Client) {
	if clients, ok := h.clients[client.uid]; ok {
		for i, c := range clients {
			if c == client {
				close(c.send)
				h.clients[client.uid] = append(clients[:i], clients[i+1:]...)
				if len(h.clients[client.uid]) == 0 {
					delete(h.clients, client.uid)
				}
				break
			}
		}
	}
}

// ==================== Redis 在线状态管理 ====================

// getOnlineKey 生成在线状态Redis键名
// 参数说明：
//   - uid: 用户UID
//
// 返回值：
//   - string: Redis键名，格式为 online:user:{uid}
func getOnlineKey(uid string) string {
	return onlineKeyPrefix + uid
}

// addDeviceOnline 将设备标记为在线
// 参数说明：
//   - uid: 用户UID
//   - deviceId: 设备ID
//
// 逻辑：
//   1. 将设备ID添加到Redis Set中
//   2. 设置Key过期时间为90秒（心跳超时阈值）
func (h *Hub) addDeviceOnline(uid, deviceId string) {
	key := getOnlineKey(uid)
	// SAdd: 将设备ID添加到Redis Set集合中（Set结构支持多设备去重）
	// 业务含义：记录该用户的设备上线，支持同用户多设备同时在线
	config.RDB.SAdd(context.Background(), key, deviceId)
	// Expire: 设置Key的过期时间为90秒
	// 业务含义：若90秒内无心跳续期，Key自动过期，用户被判定为离线
	config.RDB.Expire(context.Background(), key, onlineExpire)
}

// removeDeviceOnline 将设备标记为离线
// 参数说明：
//   - uid: 用户UID
//   - deviceId: 设备ID
//
// 逻辑：
//   从Redis Set中移除该设备ID
func (h *Hub) removeDeviceOnline(uid, deviceId string) {
	key := getOnlineKey(uid)
	// SRem: 从Redis Set集合中移除指定设备ID
	// 业务含义：设备断开连接，从在线设备列表中移除
	config.RDB.SRem(context.Background(), key, deviceId)
}

// broadcastOnlineStatus 广播用户在线状态变化
// 参数说明：
//   - uid: 用户UID
//   - online: 在线状态（1-在线，0-离线）
//
// 逻辑：
//   1. 构造在线状态消息
//   2. 遍历所有在线客户端，向非本人客户端发送状态变化通知
func (h *Hub) broadcastOnlineStatus(uid string, online int) {
	data := OnlineStatusData{
		UID:    uid,
		Online: online,
	}
	msg := &WSMessage{
		Type:    MessageTypeOnline,
		FromUID: uid,
		Data:    data,
	}

	for _, clients := range h.clients {
		for _, client := range clients {
			if client.uid != uid {
				select {
				case client.send <- msg:
				default:
					h.removeClient(client)
				}
			}
		}
	}
}

// IsOnline 检查用户是否在线
// 参数说明：
//   - uid: 用户UID
//
// 返回值：
//   - bool: 用户是否在线
//
// 逻辑：
//   通过Redis Set的大小判断用户是否有在线设备
func (h *Hub) IsOnline(uid string) bool {
	key := getOnlineKey(uid)
	// SCard: 获取Redis Set集合的元素数量
	// 业务含义：判断用户是否有在线设备，Set长度>0表示至少有一个设备在线
	count, err := config.RDB.SCard(context.Background(), key).Result()
	return err == nil && count > 0
}

// RefreshOnlineStatus 刷新用户在线状态过期时间
// 参数说明：
//   - uid: 用户UID
//
// 逻辑：
//   更新Redis Key的过期时间，用于心跳续期
func (h *Hub) RefreshOnlineStatus(uid string) {
	key := getOnlineKey(uid)
	// Expire: 重置Key的过期时间为90秒
	// 业务含义：心跳续期，用户发送消息或响应ping时刷新在线状态，防止误判离线
	config.RDB.Expire(context.Background(), key, onlineExpire)
}

// GetOnlineDevices 获取用户的在线设备列表
// 参数说明：
//   - uid: 用户UID
//
// 返回值：
//   - []string: 设备ID列表
//
// 逻辑：
//   从Redis Set中获取用户所有在线设备ID
func (h *Hub) GetOnlineDevices(uid string) []string {
	key := getOnlineKey(uid)
	// SMembers: 获取Redis Set集合中的所有元素
	// 业务含义：获取用户所有在线设备ID列表，用于多设备管理和消息路由
	result, err := config.RDB.SMembers(context.Background(), key).Result()
	if err != nil {
		return []string{}
	}
	return result
}

// ==================== 消息发送接口 ====================

// SendToUser 向指定用户发送消息
// 参数说明：
//   - uid: 目标用户UID
//   - msg: 要发送的消息
//
// 逻辑：
//   将消息发送到broadcast通道，由Run协程处理分发
func (h *Hub) SendToUser(uid string, msg *WSMessage) {
	h.broadcast <- msg
}

// SendChatMessage 发送聊天消息
// 参数说明：
//   - toUID: 接收者UID
//   - fromUID: 发送者UID
//   - data: 消息数据
//
// 逻辑：
//   构造聊天消息并发送给接收者
func SendChatMessage(toUID, fromUID string, data ChatMessageData) {
	msg := &WSMessage{
		Type:    MessageTypeChat,
		FromUID: fromUID,
		ToUID:   toUID,
		Data:    data,
	}
	GetHub().SendToUser(toUID, msg)
}

// SendWithdrawMessage 发送消息撤回通知
// 参数说明：
//   - toUID: 接收者UID
//   - fromUID: 发送者UID
//   - msgID: 被撤回的消息ID
//
// 逻辑：
//   构造撤回消息并发送给接收者
func SendWithdrawMessage(toUID, fromUID string, msgID uint) {
	data := WithdrawMessageData{MsgID: msgID}
	msg := &WSMessage{
		Type:    MessageTypeWithdraw,
		FromUID: fromUID,
		ToUID:   toUID,
		Data:    data,
	}
	GetHub().SendToUser(toUID, msg)
}

// SendFriendRequest 发送好友请求通知
// 参数说明：
//   - toUID: 接收者UID
//   - fromUID: 发送者UID
//   - fromName: 发送者昵称
//   - fromAvatar: 发送者头像URL
//
// 逻辑：
//   构造好友请求消息并发送给接收者
func SendFriendRequest(toUID, fromUID string, fromName, fromAvatar string) {
	data := FriendRequestData{
		FromUID:    fromUID,
		FromName:   fromName,
		FromAvatar: fromAvatar,
	}
	msg := &WSMessage{
		Type:    MessageTypeFriendReq,
		FromUID: fromUID,
		ToUID:   toUID,
		Data:    data,
	}
	GetHub().SendToUser(toUID, msg)
}

// SendCommentNotification 发送评论通知
// 参数说明：
//   - toUID: 动态作者UID
//   - fromUID: 评论者UID
//   - momentID: 动态ID
//   - fromName: 评论者昵称
//   - text: 评论内容
//
// 逻辑：
//   构造评论通知消息并发送给动态作者
func SendCommentNotification(toUID, fromUID string, momentID uint, fromName, text string) {
	data := CommentData{
		MomentID: momentID,
		FromUID:  fromUID,
		FromName: fromName,
		Text:     text,
	}
	msg := &WSMessage{
		Type:    MessageTypeComment,
		FromUID: fromUID,
		ToUID:   toUID,
		Data:    data,
	}
	GetHub().SendToUser(toUID, msg)
}

// SendPraiseNotification 发送点赞通知
// 参数说明：
//   - toUID: 动态作者UID
//   - fromUID: 点赞者UID
//   - momentID: 动态ID
//   - fromName: 点赞者昵称
//
// 逻辑：
//   构造点赞通知消息并发送给动态作者
func SendPraiseNotification(toUID, fromUID string, momentID uint, fromName string) {
	data := PraiseData{
		MomentID: momentID,
		FromUID:  fromUID,
		FromName: fromName,
	}
	msg := &WSMessage{
		Type:    MessageTypePraise,
		FromUID: fromUID,
		ToUID:   toUID,
		Data:    data,
	}
	GetHub().SendToUser(toUID, msg)
}

// SendSystemMessage 发送系统消息
// 参数说明：
//   - toUID: 接收者UID
//   - content: 消息内容
//
// 逻辑：
//   构造系统消息并发送给指定用户
func SendSystemMessage(toUID string, content string) {
	msg := &WSMessage{
		Type:    MessageTypeSystem,
		ToUID:   toUID,
		Data:    content,
	}
	GetHub().SendToUser(toUID, msg)
}

// ==================== 在线状态查询接口 ====================

// GetOnlineStatus 获取用户在线状态
// 参数说明：
//   - uid: 用户UID
//
// 返回值：
//   - int: 在线状态（1-在线，0-离线）
func GetOnlineStatus(uid string) int {
	if GetHub().IsOnline(uid) {
		return 1
	}
	return 0
}

// GetOnlineDeviceCount 获取用户在线设备数量
// 参数说明：
//   - uid: 用户UID
//
// 返回值：
//   - int: 在线设备数量
func GetOnlineDeviceCount(uid string) int {
	key := getOnlineKey(uid)
	// SCard: 获取Redis Set集合的元素数量
	// 业务含义：获取用户当前在线设备数量，用于判断是否多端登录
	count, err := config.RDB.SCard(context.Background(), key).Result()
	if err != nil {
		return 0
	}
	return int(count)
}

// ForceOffline 强制用户离线
// 参数说明：
//   - uid: 用户UID
//   - deviceId: 设备ID（为空时强制所有设备离线）
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. 若deviceId为空，删除整个Redis Key（强制所有设备离线）
//   2. 若deviceId不为空，仅从Set中移除该设备
func ForceOffline(uid, deviceId string) error {
	key := getOnlineKey(uid)
	if deviceId == "" {
		// Del: 删除整个Redis Key
		// 业务含义：强制用户所有设备离线（退出登录场景）
		return config.RDB.Del(context.Background(), key).Err()
	}
	// SRem: 从Redis Set集合中移除指定设备ID
	// 业务含义：强制单个设备离线（单点登录或踢人场景）
	return config.RDB.SRem(context.Background(), key, deviceId).Err()
}

// GetOnlineUsers 获取所有在线用户列表
// 返回值：
//   - []string: 在线用户UID列表
//
// 逻辑：
//   1. 使用Redis Keys命令匹配所有在线状态Key
//   2. 从Key中提取用户UID
func GetOnlineUsers() []string {
	var users []string
	pattern := onlineKeyPrefix + "*"
	// Keys: 根据通配符模式匹配所有Redis Key
	// 业务含义：获取所有在线用户的Key，用于统计在线人数或广播消息
	keys, err := config.RDB.Keys(context.Background(), pattern).Result()
	if err != nil {
		return users
	}
	for _, key := range keys {
		users = append(users, strings.TrimPrefix(key, onlineKeyPrefix))
	}
	return users
}

// ==================== 消息工具函数 ====================

// MarshalMessage 将消息序列化为JSON字节数组
// 参数说明：
//   - msg: 要序列化的消息
//
// 返回值：
//   - []byte: JSON字节数组
func MarshalMessage(msg *WSMessage) []byte {
	data, _ := json.Marshal(msg)
	return data
}