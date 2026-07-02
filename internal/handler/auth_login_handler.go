package handler

import (
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/websocket"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// LoginByCode 验证码登录
// @Summary 验证码登录
// @Description 使用手机号和验证码登录
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum formData string true "手机号"
// @Param code formData string true "验证码"
// @Success 200 {object} map[string]interface{} "登录成功，返回token"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "登录失败"
// @Router /api/v1/auth/login/code [post]
func LoginByCode(c *gin.Context) {
	phoneNum := c.PostForm("phonenum")
	code := c.PostForm("code")

	if phoneNum == "" || code == "" {
		response.BadRequest(c, "手机号和验证码不能为空")
		return
	}

	token, err := service.LoginByCode(phoneNum, code)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"token": token})
}

// LoginByPassword 密码登录
// @Summary 密码登录
// @Description 使用手机号和密码登录
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum formData string true "手机号"
// @Param pwd formData string true "密码"
// @Success 200 {object} map[string]interface{} "登录成功，返回token"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "登录失败"
// @Router /api/v1/auth/login/password [post]
func LoginByPassword(c *gin.Context) {
	phoneNum := c.PostForm("phonenum")
	password := c.PostForm("pwd")

	if phoneNum == "" || password == "" {
		response.BadRequest(c, "手机号和密码不能为空")
		return
	}

	token, err := service.LoginByPassword(phoneNum, password)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"token": token})
}

// Logout 退出登录
// @Summary 退出登录
// @Description 用户退出登录，强制下线所有设备
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "退出成功"
// @Router /api/v1/auth/logout [post]
func Logout(c *gin.Context) {
	uid := middleware.GetUID(c)

	websocket.ForceOffline(uid, "")

	repository.UpdateLastSeenAt(uid)

	c.Header("Authorization", "")
	c.Header("WWW-Authenticate", "Bearer")
	response.Success(c, nil)
}