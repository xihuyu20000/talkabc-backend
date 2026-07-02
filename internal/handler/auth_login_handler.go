package handler

import (
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/websocket"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

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

func Logout(c *gin.Context) {
	uid := middleware.GetUID(c)

	websocket.ForceOffline(uid, "")

	repository.UpdateLastSeenAt(uid)

	c.Header("Authorization", "")
	c.Header("WWW-Authenticate", "Bearer")
	response.Success(c, nil)
}
