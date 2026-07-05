package handler

import (
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/websocket"
	"backend/pkg/logger"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

/**
登录时的安全性规则如下：
1. 服务器检验通过IP是否位于黑名单中，是则拦截。
2. 服务器检验通过IP检查登录请求的频率，请求频率限制在1分钟10次。
3. 校验手机号手机号是否位于黑名单中，是则拦截。
4. 检查该设备是否在黑名单中，是则拦截。
5. 检查登录失败次数（5分钟内5次失败锁定15分钟）。
6. 检查用户账号状态（正常/封禁/注销）。
7. 登录成功后清理验证码，防止二次复用。
8. 记录登录操作日志（不可删除）。
*/

// LoginByCode 验证码登录
// @Summary 验证码登录
// @Description 使用手机号和验证码登录
// @Description
// @Description **安全规则：**
// @Description - 1. IP黑名单检查
// @Description - 2. IP登录频率限制（1分钟10次）
// @Description - 3. 手机号黑名单检查
// @Description - 4. 设备黑名单检查
// @Description - 5. 登录失败次数限制（5分钟内5次失败锁定15分钟）
// @Description - 6. 用户账号状态检查（正常/封禁/注销）
// @Description - 7. 登录成功后清理验证码，防止二次复用
// @Description - 8. 记录登录操作日志（用户ID、IP、UA、操作类型、是否成功，不可删除）
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum formData string true "手机号"
// @Param code formData string true "验证码"
// @Param device_id formData string false "设备ID"
// @Success 200 {object} map[string]interface{} "登录成功，返回access_token和refresh_token"
// @Failure 400 {object} map[string]interface{} "请求参数错误或安全校验失败"
// @Failure 500 {object} map[string]interface{} "登录失败"
// @Router /auth/login/code [post]
func LoginByCode(c *gin.Context) {
	phoneNum := c.PostForm("phonenum")
	code := c.PostForm("code")
	// 【登录安全规则4】获取设备ID（由客户端传递，用于设备黑名单校验）
	deviceID := c.PostForm("device_id")

	if phoneNum == "" || code == "" {
		response.BadRequest(c, "手机号和验证码不能为空")
		return
	}

	// 获取客户端真实IP和UA（【登录安全规则8】记录操作日志）
	clientIP := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	// 【登录安全规则】组装登录请求，包含IP、设备ID、UA用于安全校验和日志记录
	req := service.LoginRequest{
		PhoneNum: phoneNum,
		Code:     code,
		IP:       clientIP,
		DeviceID: deviceID,
		UA:       ua,
	}

	logger.Infof("[Handler] LoginByCode - PhoneNum: %s, DeviceID: %s, IP: %s", req.PhoneNum, req.DeviceID, req.IP)

	result, err := service.LoginByCode(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"access_token": result.AccessToken, "refresh_token": result.RefreshToken})
}

// LoginByPassword 密码登录
// @Summary 密码登录
// @Description 使用手机号和密码登录
// @Description 
// @Description **安全规则：**
// @Description - 1. IP黑名单检查
// @Description - 2. IP登录频率限制（1分钟10次）
// @Description - 3. 手机号黑名单检查
// @Description - 4. 设备黑名单检查
// @Description - 5. 登录失败次数限制（5分钟内5次失败锁定15分钟）
// @Description - 6. 用户账号状态检查（正常/封禁/注销）
// @Description - 7. 登录成功后重置失败次数
// @Description - 8. 记录登录操作日志（用户ID、IP、UA、操作类型、是否成功，不可删除）
// @Description 
// @Description **密码存储：** 使用bcrypt加密（cost=10，自动内置盐）
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum formData string true "手机号"
// @Param pwd formData string true "密码"
// @Param device_id formData string false "设备ID"
// @Success 200 {object} map[string]interface{} "登录成功，返回access_token和refresh_token"
// @Failure 400 {object} map[string]interface{} "请求参数错误或安全校验失败"
// @Failure 500 {object} map[string]interface{} "登录失败"
// @Router /auth/login/password [post]
func LoginByPassword(c *gin.Context) {
	phoneNum := c.PostForm("phonenum")
	password := c.PostForm("pwd")
	// 【登录安全规则4】获取设备ID（由客户端传递，用于设备黑名单校验）
	deviceID := c.PostForm("device_id")

	if phoneNum == "" || password == "" {
		response.BadRequest(c, "手机号和密码不能为空")
		return
	}

	// 获取客户端真实IP和UA（【登录安全规则8】记录操作日志）
	clientIP := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	// 【登录安全规则】组装登录请求，包含IP、设备ID、UA用于安全校验和日志记录
	req := service.LoginRequest{
		PhoneNum: phoneNum,
		Password: password,
		IP:       clientIP,
		DeviceID: deviceID,
		UA:       ua,
	}

	logger.Infof("[Handler] LoginByPassword - PhoneNum: %s, DeviceID: %s, IP: %s", req.PhoneNum, req.DeviceID, req.IP)

	result, err := service.LoginByPassword(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"access_token": result.AccessToken, "refresh_token": result.RefreshToken})
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌和刷新令牌
// @Description 
// @Description **安全规则：**
// @Description - 1. 验证刷新令牌格式（必须包含随机部分和JWT部分）
// @Description - 2. 验证刷新令牌签名有效性
// @Description - 3. 验证刷新令牌是否在Redis中存在且一致（防止滥用）
// @Description - 4. 验证用户是否存在且账号状态正常
// @Description - 5. 生成新的访问令牌和刷新令牌（刷新令牌轮转）
// @Description - 6. 将新令牌保存到Redis，旧令牌失效
// @Description - 7. 记录刷新操作日志（不可删除）
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param refresh_token formData string true "刷新令牌"
// @Success 200 {object} map[string]interface{} "刷新成功，返回新的access_token和refresh_token"
// @Failure 400 {object} map[string]interface{} "刷新令牌无效或已过期"
// @Failure 500 {object} map[string]interface{} "刷新失败"
// @Router /auth/refresh-token [post]
func RefreshToken(c *gin.Context) {
	refreshToken := c.PostForm("refresh_token")

	if refreshToken == "" {
		response.BadRequest(c, "刷新令牌不能为空")
		return
	}

	// 获取客户端真实IP和UA（【刷新令牌安全规则7】记录操作日志）
	clientIP := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	logger.Infof("[Handler] RefreshToken - IP: %s", clientIP)

	req := service.RefreshTokenRequest{
		RefreshToken: refreshToken,
		IP:           clientIP,
		UA:           ua,
	}

	result, err := service.RefreshToken(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"access_token": result.AccessToken, "refresh_token": result.RefreshToken})
}

// Logout 退出登录
// @Summary 退出登录
// @Description 用户退出登录，强制下线所有设备
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "退出成功"
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	uid := middleware.GetUID(c)

	// 【安全规则】退出登录时清除刷新令牌，防止被滥用
	repository.InvalidateRefreshToken(uid)

	websocket.ForceOffline(uid, "")

	repository.UpdateLastSeenAt(uid)

	c.Header("Authorization", "")
	c.Header("WWW-Authenticate", "Bearer")
	response.Success(c, nil)
}