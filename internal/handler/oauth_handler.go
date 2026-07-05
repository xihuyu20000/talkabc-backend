package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// OAuthLogin 第三方登录
// @Summary 第三方登录
// @Description 使用第三方平台凭证完成用户登录，支持Apple、Google、微信、支付宝、Email五种登录方式
// @Description
// @Description **支持的登录方式及所需参数：**
// @Description - Apple登录：需提供 id_token 参数
// @Description - Google登录：需提供 id_token 参数
// @Description - 微信登录：需提供 code 参数
// @Description - 支付宝登录：需提供 code 参数
// @Description - 邮箱登录：需提供 email 参数
// @Description
// @Description **安全防护机制：**
// @Description - 凭证验证：验证第三方平台返回的凭证有效性
// @Description - 新用户自动注册：首次登录时自动创建用户账号
// @Description - 账号绑定检测：已绑定的第三方账号直接登录
// @Description - 状态校验：验证用户账号状态（正常/封禁/注销）
// @Description - 操作审计：记录完整的登录日志，日志不可删除
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param provider formData string true "登录方式：apple/google/wechat/alipay/email"
// @Param code formData string false "授权码（微信、支付宝登录使用）"
// @Param id_token formData string false "ID令牌（Apple、Google登录使用）"
// @Param email formData string false "邮箱地址（邮箱登录使用）"
// @Param device_id formData string false "设备ID"
// @Success 200 {object} map[string]interface{} "登录成功，返回access_token、refresh_token、new_user"
// @Failure 400 {object} map[string]interface{} "请求参数错误或验证失败"
// @Failure 500 {object} map[string]interface{} "登录失败"
// @Router /auth/login/oauth [post]
func OAuthLogin(c *gin.Context) {
	provider := c.PostForm("provider")
	code := c.PostForm("code")
	idToken := c.PostForm("id_token")
	email := c.PostForm("email")
	deviceID := c.PostForm("device_id")

	if provider == "" {
		response.BadRequest(c, "登录方式不能为空")
		return
	}

	clientIP := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	logger.Infof("[Handler] OAuthLogin - Provider: %s, DeviceID: %s, IP: %s", provider, deviceID, clientIP)

	req := service.OAuthLoginRequest{
		Provider:    provider,
		Code:        code,
		IDToken:     idToken,
		Email:       email,
		IP:          clientIP,
		UA:          ua,
		DeviceID:    deviceID,
	}

	result, err := service.OAuthLogin(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"new_user":      result.NewUser,
	})
}

// OAuthBind 绑定第三方账号
// @Summary 绑定第三方账号
// @Description 将第三方平台账号绑定到当前登录用户，绑定后可使用该第三方平台直接登录
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param provider formData string true "绑定方式：apple/google/wechat/alipay/email"
// @Param code formData string false "授权码（微信、支付宝绑定使用）"
// @Param id_token formData string false "ID令牌（Apple、Google绑定使用）"
// @Param email formData string false "邮箱地址（邮箱绑定使用）"
// @Success 200 {object} map[string]interface{} "绑定成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误或验证失败"
// @Failure 401 {object} map[string]interface{} "未登录"
// @Failure 500 {object} map[string]interface{} "绑定失败"
// @Router /auth/oauth/bind [post]
func OAuthBind(c *gin.Context) {
	uid := middleware.GetUID(c)
	if uid == "" {
		response.Unauthorized(c, "用户未登录")
		return
	}

	provider := c.PostForm("provider")
	code := c.PostForm("code")
	idToken := c.PostForm("id_token")
	email := c.PostForm("email")

	if provider == "" {
		response.BadRequest(c, "绑定方式不能为空")
		return
	}

	clientIP := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	logger.Infof("[Handler] OAuthBind - UID: %s, Provider: %s, IP: %s", uid, provider, clientIP)

	req := service.OAuthBindRequest{
		UID:         uid,
		Provider:    provider,
		Code:        code,
		IDToken:     idToken,
		Email:       email,
		IP:          clientIP,
		UA:          ua,
	}

	err := service.OAuthBind(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// OAuthUnbind 解绑第三方账号
// @Summary 解绑第三方账号
// @Description 解绑当前登录用户的第三方平台账号，解绑后将无法使用该第三方平台登录
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param provider formData string true "解绑方式：apple/google/wechat/alipay/email"
// @Success 200 {object} map[string]interface{} "解绑成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未登录"
// @Failure 500 {object} map[string]interface{} "解绑失败"
// @Router /auth/oauth/unbind [post]
func OAuthUnbind(c *gin.Context) {
	uid := middleware.GetUID(c)
	if uid == "" {
		response.Unauthorized(c, "用户未登录")
		return
	}

	provider := c.PostForm("provider")
	if provider == "" {
		response.BadRequest(c, "解绑方式不能为空")
		return
	}

	clientIP := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	logger.Infof("[Handler] OAuthUnbind - UID: %s, Provider: %s, IP: %s", uid, provider, clientIP)

	req := service.OAuthUnbindRequest{
		UID:      uid,
		Provider: provider,
		IP:       clientIP,
		UA:       ua,
	}

	err := service.OAuthUnbind(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetOAuthBindings 获取已绑定的第三方账号列表
// @Summary 获取已绑定的第三方账号列表
// @Description 获取当前登录用户已绑定的所有第三方平台账号列表
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功，返回已绑定的provider列表"
// @Failure 401 {object} map[string]interface{} "未登录"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /auth/oauth/list [get]
func GetOAuthBindings(c *gin.Context) {
	uid := middleware.GetUID(c)
	if uid == "" {
		response.Unauthorized(c, "用户未登录")
		return
	}

	logger.Infof("[Handler] GetOAuthBindings - UID: %s", uid)

	req := service.OAuthListRequest{
		UID: uid,
	}

	result, err := service.GetOAuthBindings(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"providers": result.Providers})
}