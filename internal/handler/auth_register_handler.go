package handler

import (
	"backend/internal/service"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

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
