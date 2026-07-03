package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ==================== 消息模块 ====================

// GetSystemMsgList 获取系统消息列表
// @Summary 获取系统消息列表
// @Description 获取当前用户的系统消息列表
// @Tags 消息
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /messages/system [get]
func GetSystemMsgList(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetSystemMsgList - UID: %s", uid)

	msgs, err := service.GetSystemMsgList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, msgs)
}

// GetLatestUserMsg 获取最新用户消息
// @Summary 获取最新用户消息
// @Description 获取当前用户与其他用户的最新消息列表
// @Tags 消息
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /messages/latest [get]
func GetLatestUserMsg(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetLatestUserMsg - UID: %s", uid)

	msgs, err := service.GetLatestUserMsg(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, msgs)
}

// GetUserMsgHistory 获取用户消息历史
// @Summary 获取用户消息历史
// @Description 获取与指定用户的聊天消息历史
// @Tags 消息
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "目标用户UID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /messages/users/{uid} [get]
func GetUserMsgHistory(c *gin.Context) {
	uid := middleware.GetUID(c)
	targetID := c.Param("uid")

	logger.Infof("[Handler] GetUserMsgHistory - UID: %s, TargetID: %s", uid, targetID)

	msgs, err := service.GetUserMsgHistory(uid, targetID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, msgs)
}

// SetMessageTop 设置消息置顶
// @Summary 设置消息置顶
// @Description 设置或取消与指定用户的消息置顶
// @Tags 消息
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "目标用户UID"
// @Param flag path string true "置顶标志：1-置顶，0-取消置顶"
// @Success 200 {object} map[string]interface{} "设置成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "设置失败"
// @Router /messages/users/{uid}/top/{flag} [post]
func SetMessageTop(c *gin.Context) {
	uid := middleware.GetUID(c)
	targetID := c.Param("uid")
	flagStr := c.Param("flag")

	flag, err := strconv.Atoi(flagStr)
	if err != nil {
		response.BadRequest(c, "flag参数错误")
		return
	}

	logger.Infof("[Handler] SetMessageTop - UID: %s, TargetID: %s, Flag: %d", uid, targetID, flag)

	err = service.SetMessageTop(uid, targetID, flag)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// ClearChatHistory 清除聊天记录
// @Summary 清除聊天记录
// @Description 清除与指定用户的聊天记录
// @Tags 消息
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "目标用户UID"
// @Success 200 {object} map[string]interface{} "清除成功"
// @Failure 500 {object} map[string]interface{} "清除失败"
// @Router /messages/users/{uid}/clear [post]
func ClearChatHistory(c *gin.Context) {
	uid := middleware.GetUID(c)
	targetID := c.Param("uid")

	logger.Infof("[Handler] ClearChatHistory - UID: %s, TargetID: %s", uid, targetID)

	err := service.ClearChatHistory(uid, targetID)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 社交模块 ====================

// AddFriend 添加好友
// @Summary 添加好友
// @Description 添加指定用户为好友
// @Tags 社交
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "目标用户UID"
// @Param flag path string true "操作标志：1-添加好友，0-取消关注"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "操作失败"
// @Router /friends/{uid}/{flag} [post]
func AddFriend(c *gin.Context) {
	uid := middleware.GetUID(c)
	targetID := c.Param("uid")
	flagStr := c.Param("flag")

	flag, err := strconv.Atoi(flagStr)
	if err != nil {
		response.BadRequest(c, "flag参数错误")
		return
	}

	logger.Infof("[Handler] AddFriend - UID: %s, TargetID: %s, Flag: %d", uid, targetID, flag)

	err = service.AddFriend(uid, targetID, flag)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// SendGift 发送礼物
// @Summary 发送礼物
// @Description 向指定用户发送礼物
// @Tags 社交
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "接收用户UID"
// @Param giftid path string true "礼物ID"
// @Success 200 {object} map[string]interface{} "发送成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "发送失败"
// @Router /gifts/send/{uid}/{giftid} [post]
func SendGift(c *gin.Context) {
	senderID := middleware.GetUID(c)
	receiverID := c.Param("uid")
	giftIDStr := c.Param("giftid")

	giftID, err := strconv.ParseUint(giftIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "礼物ID参数错误")
		return
	}

	logger.Infof("[Handler] SendGift - SenderID: %s, ReceiverID: %s, GiftID: %d", senderID, receiverID, giftID)

	err = service.SendGift(senderID, receiverID, uint(giftID))
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 广告模块 ====================

// GetLatestAdBanner 获取最新广告横幅
// @Summary 获取最新广告横幅
// @Description 获取最新的广告横幅列表
// @Tags 广告
// @Accept application/json
// @Produce application/json
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /ads/banners [get]
func GetLatestAdBanner(c *gin.Context) {
	logger.Infof("[Handler] GetLatestAdBanner")

	banners, err := service.GetLatestAdBanner()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, banners)
}