package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPraiseMeList 获取赞我的列表
// @Summary 获取赞我的列表
// @Description 获取点赞我动态的用户列表
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /notifications/praise-me [get]
func GetPraiseMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetPraiseMeList - UID: %s", uid)

	list, err := service.GetPraiseMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetCommentMeList 获取评论我的列表
// @Summary 获取评论我的列表
// @Description 获取评论我动态的用户列表
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /notifications/comment-me [get]
func GetCommentMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetCommentMeList - UID: %s", uid)

	list, err := service.GetCommentMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetAddMeList 获取添加我的列表
// @Summary 获取添加我的列表
// @Description 获取发送好友请求给我的用户列表
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /notifications/add-me [get]
func GetAddMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetAddMeList - UID: %s", uid)

	list, err := service.GetAddMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetVisitMeList 获取访问我的列表
// @Summary 获取访问我的列表
// @Description 获取访问我个人主页的用户列表
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /notifications/visit-me [get]
func GetVisitMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetVisitMeList - UID: %s", uid)

	list, err := service.GetVisitMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetLikeMeList 获取喜欢我的列表
// @Summary 获取喜欢我的列表
// @Description 获取表示喜欢我的用户列表
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /notifications/like-me [get]
func GetLikeMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetLikeMeList - UID: %s", uid)

	list, err := service.GetLikeMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// AgreeFriendRequest 同意好友请求
// @Summary 同意好友请求
// @Description 处理好友请求，flag=1同意添加好友，flag=0拒绝好友请求
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "请求好友的用户UID"
// @Param flag path string true "操作标志：1-同意，0-拒绝"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "操作失败"
// @Router /notifications/friend-request/{uid}/{flag} [post]
func AgreeFriendRequest(c *gin.Context) {
	userID := middleware.GetUID(c)
	targetID := c.Param("uid")
	flagStr := c.Param("flag")

	flag, err := strconv.Atoi(flagStr)
	if err != nil {
		response.BadRequest(c, "flag参数错误")
		return
	}

	logger.Infof("[Handler] AgreeFriendRequest - UserID: %s, TargetID: %s, Flag: %d", userID, targetID, flag)

	err = service.AgreeFriendRequest(userID, targetID, flag)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}