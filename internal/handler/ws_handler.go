package handler

import (
	"backend/internal/config"
	"backend/internal/middleware"
	ws "backend/internal/websocket"
	"backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// wsUpgrader WebSocket连接升级器
// 配置WebSocket连接参数，包括读写缓冲区大小和跨域检查
var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler WebSocket连接处理接口
// 请求方式：GET（WebSocket握手请求）
// 请求路径：/ws
// 请求参数：token - JWT令牌（Query参数）；deviceId - 设备标识（Query参数，支持多端同时在线）
// 返回值：WebSocket连接升级成功或错误信息
//
// 业务流程：
//   1. 从查询参数获取JWT令牌和设备标识deviceId
//   2. 验证令牌有效性并解析出用户ID
//   3. 将HTTP连接升级为WebSocket连接
//   4. 创建WebSocket客户端实例（包含deviceId）并注册到Hub
//   5. Hub将设备ID添加到Redis Set（online:user:{uid}）并设置90s过期
//   6. 广播用户上线状态通知
//   7. 启动消息读取和写入协程
//   8. 心跳机制：客户端每30s发ping，服务端收到后执行EXPIRE续期90s
func WebSocketHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "缺少token"})
		return
	}

	deviceId := c.Query("deviceId")
	if deviceId == "" {
		deviceId = c.ClientIP()
	}

	claims := &jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "token无效"})
		return
	}

	uid, ok := (*claims)["uid"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "用户ID无效"})
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "连接升级失败"})
		return
	}

	client := ws.NewClient(ws.GetHub(), conn, uid, deviceId)
	ws.GetHub().Register(client)

	go client.WritePump()
	client.ReadPump()
}

// GetOnlineStatus 获取在线状态接口
// 请求方式：GET
// 请求路径：/v1/onlinestatus
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：当前用户的在线状态
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 websocket.GetOnlineStatus 查询在线状态
//   3. 返回在线状态数据（1=在线，0=离线）
func GetOnlineStatus(c *gin.Context) {
	uid := middleware.GetUID(c)

	status := ws.GetOnlineStatus(uid)
	response.Success(c, gin.H{"uid": uid, "online": status})
}