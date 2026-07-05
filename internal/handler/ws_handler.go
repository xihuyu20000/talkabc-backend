package handler

import (
	"backend/internal/config"
	"backend/internal/middleware"
	ws "backend/internal/websocket"
	"backend/pkg/logger"
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

// @Summary WebSocket连接握手
// @Description 建立WebSocket连接，用于实时消息通信。客户端需携带JWT令牌进行身份认证。连接成功后，服务端会自动管理用户在线状态，并支持多端同时在线。
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param token query string true "JWT令牌"
// @Param deviceId query string false "设备标识，不传则使用客户端IP"
// @Success 101 "WebSocket连接升级成功"
// @Failure 400 {object} response.Response "缺少token或用户ID无效"
// @Failure 401 {object} response.Response "token无效"
// @Failure 500 {object} response.Response "连接升级失败"
// @Router /ws [get]
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
		response.BadRequest(c, "缺少token")
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
		response.Unauthorized(c, "token无效")
		return
	}

	uid, ok := (*claims)["uid"].(string)
	if !ok {
		response.BadRequest(c, "用户ID无效")
		return
	}

	logger.Infof("[Handler] WebSocketHandler - UID: %s, DeviceID: %s", uid, deviceId)

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.InternalError(c, "连接升级失败")
		return
	}

	client := ws.NewClient(ws.GetHub(), conn, uid, deviceId)
	ws.GetHub().Register(client)

	go client.WritePump()
	client.ReadPump()
}

// @Summary 获取用户在线状态
// @Description 查询指定用户的在线状态。通过JWT令牌获取当前用户ID，然后从Redis中查询该用户的在线状态。支持多端在线状态管理。
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response "在线状态数据"
// @Failure 500 {object} response.Response "查询失败"
// @Router /v1/onlinestatus [get]
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 websocket.GetOnlineStatus 查询在线状态
//   3. 返回在线状态数据（1=在线，0=离线）
func GetOnlineStatus(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetOnlineStatus - UID: %s", uid)

	status := ws.GetOnlineStatus(uid)
	response.Success(c, gin.H{"uid": uid, "online": status})
}