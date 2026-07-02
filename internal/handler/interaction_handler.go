package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

// GetPraiseMeList 获取赞我的列表
// @Summary 获取赞我的列表
// @Description 获取赞我动态的用户列表
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /api/v1/notifications/praise-me [get]
func GetPraiseMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

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
// @Router /api/v1/notifications/comment-me [get]
func GetCommentMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	list, err := service.GetCommentMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetAddMeList 获取添加我的列表
// @Summary 获取添加我的列表
// @Description 获取添加我为好友的用户列表
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /api/v1/notifications/add-me [get]
func GetAddMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	list, err := service.GetAddMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetVisitMeList 获取访问我的列表
// @Summary 获取访问我的列表
// @Description 获取访问我主页的用户列表
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /api/v1/notifications/visit-me [get]
func GetVisitMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	list, err := service.GetVisitMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetLikeMeList 获取喜欢我的列表
// @Summary 获取喜欢我的列表
// @Description 获取喜欢我的用户列表
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /api/v1/notifications/like-me [get]
func GetLikeMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	list, err := service.GetLikeMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// AgreeFriendRequest 同意好友请求
// @Summary 同意好友请求
// @Description 同意或拒绝好友请求
// @Tags 互动通知
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "请求好友的用户UID"
// @Param flag path string true "操作标志：1-同意，0-拒绝"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "操作失败"
// @Router /api/v1/notifications/friend-request/{uid}/{flag} [post]
func AgreeFriendRequest(c *gin.Context) {
	userID := middleware.GetUID(c)
	targetID := c.Param("uid")
	flagStr := c.Param("flag")

	flag, err := strconv.Atoi(flagStr)
	if err != nil {
		response.BadRequest(c, "flag参数错误")
		return
	}

	err = service.AgreeFriendRequest(userID, targetID, flag)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}