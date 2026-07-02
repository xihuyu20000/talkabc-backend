package handler

import (
	"backend/internal/service"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Register 用户注册
// @Summary 用户注册
// @Description 使用手机号和验证码进行注册
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum formData string true "手机号"
// @Param code formData string true "验证码"
// @Success 200 {object} map[string]interface{} "注册成功，返回token"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "注册失败"
// @Router /api/v1/auth/register [post]
func Register(c *gin.Context) {
	phoneNum := c.PostForm("phonenum")
	code := c.PostForm("code")

	if phoneNum == "" || code == "" {
		response.BadRequest(c, "手机号和验证码不能为空")
		return
	}

	token, err := service.Register(phoneNum, code)
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