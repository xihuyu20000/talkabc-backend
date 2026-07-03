package handler

import (
	"backend/internal/service"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

/**
注册时的安全性规则如下：
1. 服务器检验通过IP是否位于黑名单中，是则拦截。
2. 服务器检验通过IP检查注册请求的频率，请求频率限制在1分钟10次。
3. 校验手机号手机号是否位于黑名单中，是则拦截
4. 校验手机号是否已被占用（用户已存在）。
*/

// Register 用户注册
// @Summary 用户注册
// @Description 使用手机号和验证码进行注册（带IP黑名单、频率限制、手机号黑名单校验）
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum formData string true "手机号"
// @Param code formData string true "验证码"
// @Success 200 {object} map[string]interface{} "注册成功，返回token"
// @Failure 400 {object} map[string]interface{} "请求参数错误或安全校验失败"
// @Failure 500 {object} map[string]interface{} "注册失败"
// @Router /api/v1/auth/register [post]
func Register(c *gin.Context) {
	phoneNum := c.PostForm("phonenum")
	code := c.PostForm("code")

	if phoneNum == "" || code == "" {
		response.BadRequest(c, "手机号和验证码不能为空")
		return
	}

	// 获取客户端真实IP（支持代理）
	clientIP := c.ClientIP()

	// 【注册安全规则】组装注册请求，包含IP信息用于安全校验
	req := service.RegisterRequest{
		PhoneNum: phoneNum,
		Code:     code,
		IP:       clientIP,
	}

	token, err := service.Register(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"token": token})
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 使用手机号和新密码重置密码
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum formData string true "手机号"
// @Param pwd1 formData string true "新密码"
// @Param pwd2 formData string true "确认密码"
// @Success 200 {object} map[string]interface{} "重置成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "重置失败"
// @Router /api/v1/auth/reset-password [post]
func ResetPassword(c *gin.Context) {
	phoneNum := c.PostForm("phonenum")
	pwd1 := c.PostForm("pwd1")
	pwd2 := c.PostForm("pwd2")

	if phoneNum == "" || pwd1 == "" || pwd2 == "" {
		response.BadRequest(c, "手机号和密码不能为空")
		return
	}

	err := service.ResetPassword(phoneNum, pwd1, pwd2)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}