package handler

import (
	"backend/internal/service"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

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
