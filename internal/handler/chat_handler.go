package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ==================== 消息模块 ====================

func GetSystemMsgList(c *gin.Context) {
	uid := middleware.GetUID(c)

	msgs, err := service.GetSystemMsgList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, msgs)
}

func GetLatestUserMsg(c *gin.Context) {
	uid := middleware.GetUID(c)

	msgs, err := service.GetLatestUserMsg(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, msgs)
}

func GetUserMsgHistory(c *gin.Context) {
	uid := middleware.GetUID(c)
	targetID := c.Param("uid")

	msgs, err := service.GetUserMsgHistory(uid, targetID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, msgs)
}

func SetMessageTop(c *gin.Context) {
	uid := middleware.GetUID(c)
	targetID := c.Param("uid")
	flagStr := c.Param("flag")

	flag, err := strconv.Atoi(flagStr)
	if err != nil {
		response.BadRequest(c, "flag参数错误")
		return
	}

	err = service.SetMessageTop(uid, targetID, flag)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

func ClearChatHistory(c *gin.Context) {
	uid := middleware.GetUID(c)
	targetID := c.Param("uid")

	err := service.ClearChatHistory(uid, targetID)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 社交模块 ====================

func AddFriend(c *gin.Context) {
	uid := middleware.GetUID(c)
	targetID := c.Param("uid")
	flagStr := c.Param("flag")

	flag, err := strconv.Atoi(flagStr)
	if err != nil {
		response.BadRequest(c, "flag参数错误")
		return
	}

	err = service.AddFriend(uid, targetID, flag)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

func SendGift(c *gin.Context) {
	senderID := middleware.GetUID(c)
	receiverID := c.Param("uid")
	giftIDStr := c.Param("giftid")

	giftID, err := strconv.ParseUint(giftIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "礼物ID参数错误")
		return
	}

	err = service.SendGift(senderID, receiverID, uint(giftID))
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 广告模块 ====================

func GetLatestAdBanner(c *gin.Context) {
	banners, err := service.GetLatestAdBanner()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, banners)
}