package handler

import (
	"backend/internal/service"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// SendSMSCode 发送短信验证码
// @Summary 发送短信验证码
// @Description 向指定手机号发送短信验证码
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum query string true "手机号"
// @Success 200 {object} map[string]interface{} "发送成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "发送失败"
// @Router /api/v1/auth/code-sms [get]
func SendSMSCode(c *gin.Context) {
	phoneNum := c.Query("phonenum")

	if phoneNum == "" {
		response.BadRequest(c, "手机号不能为空")
		return
	}

	err := service.GenerateSMSCode(phoneNum)
	if err != nil {
		response.InternalError(c, "发送验证码失败")
		return
	}

	response.Success(c, nil)
}

// GenerateAlnumCode 生成字母数字验证码
// @Summary 生成字母数字验证码
// @Description 向指定手机号发送字母数字组合的验证码
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum query string true "手机号"
// @Success 200 {object} map[string]interface{} "生成成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "生成失败"
// @Router /api/v1/auth/code-alnum [get]
func GenerateAlnumCode(c *gin.Context) {
	phoneNum := c.Query("phonenum")
	if phoneNum == "" {
		response.BadRequest(c, "手机号不能为空")
		return
	}

	err := service.GenerateAlnumCode(phoneNum)
	if err != nil {
		response.InternalError(c, "生成验证码失败")
		return
	}

	response.Success(c, nil)
}