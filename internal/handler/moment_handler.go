package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetLatestMoment 获取最新动态列表
// @Summary 获取最新动态列表
// @Description 获取平台最新发布的动态内容列表，按发布时间倒序排列
// @Tags 动态
// @Accept application/json
// @Produce application/json
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /moments [get]
func GetLatestMoment(c *gin.Context) {
	logger.Infof("[Handler] GetLatestMoment")

	moments, err := service.GetLatestMoment()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, moments)
}

// GetMyLatestMoment 获取我的最新动态
// @Summary 获取我的最新动态
// @Description 获取当前用户自己发布的最新动态内容列表
// @Tags 动态
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /moments/me [get]
func GetMyLatestMoment(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetMyLatestMoment - UID: %s", uid)

	moments, err := service.GetMyLatestMoment(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, moments)
}

// GetUserMoment 获取指定用户动态
// @Summary 获取指定用户动态
// @Description 获取指定用户发布的动态内容列表
// @Tags 动态
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "用户UID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /moments/users/{uid} [get]
func GetUserMoment(c *gin.Context) {
	uid := c.Param("uid")

	logger.Infof("[Handler] GetUserMoment - UID: %s", uid)

	moments, err := service.GetUserMoment(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, moments)
}

// GetMomentComments 获取动态评论列表
// @Summary 获取动态评论列表
// @Description 获取指定动态的评论内容列表
// @Tags 动态
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param mid path string true "动态ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /moments/{mid}/comments [get]
func GetMomentComments(c *gin.Context) {
	midStr := c.Param("mid")

	mid, err := strconv.ParseUint(midStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "动态ID参数错误")
		return
	}

	logger.Infof("[Handler] GetMomentComments - MomentID: %d", mid)

	comments, err := service.GetMomentComments(uint(mid))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, comments)
}

// PublishMoment 发布动态
// @Summary 发布动态
// @Description 用户发布新的动态内容，支持文本描述、图片上传和地理位置信息
// @Tags 动态
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param text formData string true "动态文本内容"
// @Param location formData string false "地理位置"
// @Param files formData file false "图片文件列表"
// @Success 200 {object} map[string]interface{} "发布成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "发布失败"
// @Router /moments [post]
func PublishMoment(c *gin.Context) {
	userID := middleware.GetUID(c)

	var req struct {
		Text     string `json:"text"`
		Location string `json:"location"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	var files []string
	formFiles := c.Request.MultipartForm
	if formFiles != nil {
		for _, f := range formFiles.File["files"] {
			fileErr := c.SaveUploadedFile(f, "./uploads/moments/"+f.Filename)
			if fileErr == nil {
				files = append(files, "/uploads/moments/"+f.Filename)
			}
		}
	}

	logger.Infof("[Handler] PublishMoment - UserID: %s, Text: %s, Location: %s, FileCount: %d", userID, req.Text, req.Location, len(files))

	err := service.PublishMoment(userID, req.Text, files, req.Location)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}