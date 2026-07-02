package websocket

import (
	"backend/internal/service"
	"backend/pkg/logger"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second  // 写入超时时间
	pongWait       = 90 * time.Second  // 接收pong响应超时时间（心跳超时阈值）
	pingPeriod     = 30 * time.Second  // 客户端心跳发送周期，每30s发送一次ping
	maxMessageSize = 512               // 最大消息大小限制（字节）
)

// upgrader WebSocket连接升级器
// 将HTTP连接升级为WebSocket连接，配置读写缓冲区大小
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源的连接（生产环境应根据实际需求配置）
	},
}

// Client WebSocket客户端连接结构体
// 每个客户端连接对应一个Client实例，管理连接生命周期和消息处理
type Client struct {
	hub      *Hub          // 所属的Hub，用于管理所有客户端连接
	conn     *websocket.Conn // WebSocket连接对象
	uid      string        // 用户ID，标识此连接所属的用户
	deviceId string        // 设备标识，支持多端同时在线
	send     chan *WSMessage // 消息发送通道，用于异步发送消息给客户端
}

// NewClient 创建新的WebSocket客户端实例
// 参数说明：
//   - hub: 所属的Hub实例
//   - conn: WebSocket连接对象
//   - uid: 用户UID
//   - deviceId: 设备ID
//
// 返回值：
//   - *Client: 客户端实例
func NewClient(hub *Hub, conn *websocket.Conn, uid string, deviceId string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		uid:      uid,
		deviceId: deviceId,
		send:     make(chan *WSMessage, 256), // 消息通道缓冲区大小为256
	}
}

// ReadPump 读取消息协程
// 持续从WebSocket连接读取客户端发送的消息
//
// 处理逻辑：
//   1. 设置读取限制和超时时间
//   2. 注册pong处理器（心跳响应），收到pong时刷新在线状态
//   3. 循环读取消息，每次读取成功后刷新在线状态并处理消息
//   4. 连接断开时自动注销客户端
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.hub.RefreshOnlineStatus(c.uid)
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("websocket read error: %v", err)
			}
			break
		}

		c.hub.RefreshOnlineStatus(c.uid)
		c.handleMessage(message)
	}
}

// handleMessage 处理收到的消息
// 参数说明：
//   - data: 消息数据（JSON格式字节数组）
//
// 处理逻辑：
//   1. 将JSON数据解析为WSMessage结构体
//   2. 根据消息类型分发到对应的处理函数
func (c *Client) handleMessage(data []byte) {
	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Error("websocket message parse error: %v", err)
		return
	}

	switch msg.Type {
	case MessageTypeSendText:
		c.handleSendText(msg)
	case MessageTypeSendImage:
		c.handleSendImage(msg)
	case MessageTypeSendVoice:
		c.handleSendVoice(msg)
	case MessageTypeSendVideo:
		c.handleSendVideo(msg)
	case MessageTypeSendFile:
		c.handleSendFile(msg)
	case MessageTypeSendWithdraw:
		c.handleSendWithdraw(msg)
	case MessageTypeFocusUser:
		c.handleFocusUser(msg)
	case MessageTypeBlockUser:
		c.handleBlockUser(msg)
	case MessageTypeLikeUser:
		c.handleLikeUser(msg)
	case MessageTypePraiseMoment:
		c.handlePraiseMoment(msg)
	case MessageTypeCommentMoment:
		c.handleCommentMoment(msg)
	case MessageTypeReportMoment:
		c.handleReportMoment(msg)
	}
}

// handleSendText 处理发送文本消息
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取接收者UID和文本内容
//   2. 调用service层保存消息
//   3. 向接收者发送聊天消息
func (c *Client) handleSendText(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket send text data format error")
		return
	}

	toUID, _ := dataMap["to_uid"].(string)
	text, _ := dataMap["text"].(string)

	if toUID == "" || text == "" {
		return
	}

	err := service.SendTextMessage(c.uid, toUID, text)
	if err != nil {
		logger.Error("websocket send text error: %v", err)
		return
	}

	SendChatMessage(toUID, c.uid, ChatMessageData{
		Text:     text,
		MsgType:  1,
		SendTime: time.Now().Unix(),
	})
}

// handleSendImage 处理发送图片消息
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取接收者UID和图片URL
//   2. 调用service层保存消息
//   3. 向接收者发送聊天消息
func (c *Client) handleSendImage(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket send image data format error")
		return
	}

	toUID, _ := dataMap["to_uid"].(string)
	fileURL, _ := dataMap["file_url"].(string)

	if toUID == "" || fileURL == "" {
		return
	}

	err := service.SendImageMessage(c.uid, toUID, fileURL)
	if err != nil {
		logger.Error("websocket send image error: %v", err)
		return
	}

	SendChatMessage(toUID, c.uid, ChatMessageData{
		FileURL:  fileURL,
		MsgType:  2,
		SendTime: time.Now().Unix(),
	})
}

// handleSendVoice 处理发送语音消息
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取接收者UID和语音URL
//   2. 调用service层保存消息
//   3. 向接收者发送聊天消息
func (c *Client) handleSendVoice(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket send voice data format error")
		return
	}

	toUID, _ := dataMap["to_uid"].(string)
	fileURL, _ := dataMap["file_url"].(string)

	if toUID == "" || fileURL == "" {
		return
	}

	err := service.SendVoiceMessage(c.uid, toUID, fileURL)
	if err != nil {
		logger.Error("websocket send voice error: %v", err)
		return
	}

	SendChatMessage(toUID, c.uid, ChatMessageData{
		FileURL:  fileURL,
		MsgType:  3,
		SendTime: time.Now().Unix(),
	})
}

// handleSendVideo 处理发送视频消息
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取接收者UID和视频URL
//   2. 调用service层保存消息
//   3. 向接收者发送聊天消息
func (c *Client) handleSendVideo(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket send video data format error")
		return
	}

	toUID, _ := dataMap["to_uid"].(string)
	fileURL, _ := dataMap["file_url"].(string)

	if toUID == "" || fileURL == "" {
		return
	}

	err := service.SendVideoMessage(c.uid, toUID, fileURL)
	if err != nil {
		logger.Error("websocket send video error: %v", err)
		return
	}

	SendChatMessage(toUID, c.uid, ChatMessageData{
		FileURL:  fileURL,
		MsgType:  4,
		SendTime: time.Now().Unix(),
	})
}

// handleSendFile 处理发送文件消息
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取接收者UID和文件URL
//   2. 调用service层保存消息
//   3. 向接收者发送聊天消息
func (c *Client) handleSendFile(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket send file data format error")
		return
	}

	toUID, _ := dataMap["to_uid"].(string)
	fileURL, _ := dataMap["file_url"].(string)

	if toUID == "" || fileURL == "" {
		return
	}

	err := service.SendFileMessage(c.uid, toUID, fileURL)
	if err != nil {
		logger.Error("websocket send file error: %v", err)
		return
	}

	SendChatMessage(toUID, c.uid, ChatMessageData{
		FileURL:  fileURL,
		MsgType:  5,
		SendTime: time.Now().Unix(),
	})
}

// handleSendWithdraw 处理撤回消息
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取接收者UID和消息ID
//   2. 调用service层执行撤回操作
//   3. 向接收者发送撤回通知
func (c *Client) handleSendWithdraw(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket withdraw data format error")
		return
	}

	toUID, _ := dataMap["to_uid"].(string)
	msgIDStr, _ := dataMap["msg_id"].(string)

	if toUID == "" || msgIDStr == "" {
		return
	}

	msgID, err := strconv.ParseUint(msgIDStr, 10, 32)
	if err != nil {
		logger.Error("websocket withdraw msg_id parse error: %v", err)
		return
	}

	err = service.WithdrawMessage(c.uid, toUID, uint(msgID))
	if err != nil {
		logger.Error("websocket withdraw error: %v", err)
		return
	}

	SendWithdrawMessage(toUID, c.uid, uint(msgID))
}

// handleFocusUser 处理关注/取消关注用户
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取目标用户UID和操作标志
//   2. 调用service层执行关注/取消关注操作
func (c *Client) handleFocusUser(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket focus user data format error")
		return
	}

	toUID, _ := dataMap["to_uid"].(string)
	flag, _ := dataMap["flag"].(float64)

	if toUID == "" {
		return
	}

	err := service.FocusUser(c.uid, toUID, int(flag))
	if err != nil {
		logger.Error("websocket focus user error: %v", err)
		return
	}
}

// handleBlockUser 处理拉黑/取消拉黑用户
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取目标用户UID和操作标志
//   2. 调用service层执行拉黑/取消拉黑操作
func (c *Client) handleBlockUser(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket block user data format error")
		return
	}

	toUID, _ := dataMap["to_uid"].(string)
	flag, _ := dataMap["flag"].(float64)

	if toUID == "" {
		return
	}

	err := service.BlockUser(c.uid, toUID, int(flag))
	if err != nil {
		logger.Error("websocket block user error: %v", err)
		return
	}
}

// handleLikeUser 处理喜欢/取消喜欢用户
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取目标用户UID和操作标志
//   2. 调用service层执行喜欢/取消喜欢操作
func (c *Client) handleLikeUser(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket like user data format error")
		return
	}

	toUID, _ := dataMap["to_uid"].(string)
	flag, _ := dataMap["flag"].(float64)

	if toUID == "" {
		return
	}

	err := service.LikeUser(c.uid, toUID, int(flag))
	if err != nil {
		logger.Error("websocket like user error: %v", err)
		return
	}
}

// handlePraiseMoment 处理点赞/取消点赞动态
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取动态ID
//   2. 调用service层执行点赞/取消点赞操作
func (c *Client) handlePraiseMoment(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket praise moment data format error")
		return
	}

	midStr, _ := dataMap["moment_id"].(string)
	mid, err := strconv.ParseUint(midStr, 10, 32)
	if err != nil {
		logger.Error("websocket praise moment moment_id parse error: %v", err)
		return
	}

	err = service.PraiseMoment(c.uid, uint(mid))
	if err != nil {
		logger.Error("websocket praise moment error: %v", err)
		return
	}
}

// handleCommentMoment 处理评论动态
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取动态ID和评论内容
//   2. 调用service层执行评论操作
func (c *Client) handleCommentMoment(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket comment moment data format error")
		return
	}

	midStr, _ := dataMap["moment_id"].(string)
	text, _ := dataMap["text"].(string)
	mid, err := strconv.ParseUint(midStr, 10, 32)
	if err != nil {
		logger.Error("websocket comment moment moment_id parse error: %v", err)
		return
	}

	if text == "" {
		return
	}

	err = service.CommentMoment(c.uid, uint(mid), text)
	if err != nil {
		logger.Error("websocket comment moment error: %v", err)
		return
	}
}

// handleReportMoment 处理举报动态
// 参数说明：
//   - msg: WebSocket消息
//
// 处理逻辑：
//   1. 解析消息数据，获取动态ID
//   2. 调用service层执行举报操作
func (c *Client) handleReportMoment(msg WSMessage) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		logger.Error("websocket report moment data format error")
		return
	}

	midStr, _ := dataMap["moment_id"].(string)
	mid, err := strconv.ParseUint(midStr, 10, 32)
	if err != nil {
		logger.Error("websocket report moment moment_id parse error: %v", err)
		return
	}

	err = service.ReportMoment(c.uid, uint(mid))
	if err != nil {
		logger.Error("websocket report moment error: %v", err)
		return
	}
}

// WritePump 写入消息协程
// 持续从send通道读取消息并发送给客户端，同时定期发送ping心跳
//
// 处理逻辑：
//   1. 监听send通道，发送消息给客户端
//   2. 每30秒发送一次ping消息，检测客户端是否在线
//   3. 发送失败时关闭连接
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(MarshalMessage(message))

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Close 关闭客户端连接
// 关闭消息通道和WebSocket连接
func (c *Client) Close() {
	close(c.send)
	c.conn.Close()
}