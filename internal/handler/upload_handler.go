package handler

import (
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/pkg/logger"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// UploadAvatar 上传头像
// @Summary 上传头像
// @Description 用户上传个人头像图片，上传成功后自动更新用户头像信息
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "头像图片文件"
// @Success 200 {object} map[string]interface{} "上传成功，返回头像URL"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "上传失败"
// @Router /users/avatar [post]
func UploadAvatar(c *gin.Context) {
	uid := middleware.GetUID(c)

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传头像")
		return
	}

	logger.Infof("[Handler] UploadAvatar - UID: %s, Filename: %s, Size: %d", uid, file.Filename, file.Size)

	err = c.SaveUploadedFile(file, "./uploads/avatars/"+file.Filename)
	if err != nil {
		response.InternalError(c, "文件上传失败")
		return
	}

	avatarURL := "/uploads/avatars/" + file.Filename

	user, err := repository.GetUserByUID(uid)
	if err != nil {
		response.InternalError(c, "用户不存在")
		return
	}

	user.AvatarURL = avatarURL
	err = repository.UpdateUser(user)
	if err != nil {
		response.InternalError(c, "更新头像失败")
		return
	}

	response.Success(c, gin.H{"avatar_url": avatarURL})
}

// UploadImage 上传图片
// @Summary 上传图片
// @Description 上传图片文件，支持消息、动态等场景使用
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "图片文件"
// @Success 200 {object} map[string]interface{} "上传成功，返回文件URL"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "上传失败"
// @Router /uploads/image [post]
func UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传图片")
		return
	}

	logger.Infof("[Handler] UploadImage - Filename: %s, Size: %d", file.Filename, file.Size)

	err = c.SaveUploadedFile(file, "./uploads/messages/"+file.Filename)
	if err != nil {
		response.InternalError(c, "文件上传失败")
		return
	}

	fileURL := "/uploads/messages/" + file.Filename
	response.Success(c, gin.H{"file_url": fileURL})
}

// UploadAudio 上传音频
// @Summary 上传音频
// @Description 上传音频文件，支持语音消息等场景使用
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "音频文件"
// @Success 200 {object} map[string]interface{} "上传成功，返回文件URL"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "上传失败"
// @Router /uploads/audio [post]
func UploadAudio(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传音频文件")
		return
	}

	logger.Infof("[Handler] UploadAudio - Filename: %s, Size: %d", file.Filename, file.Size)

	err = c.SaveUploadedFile(file, "./uploads/messages/"+file.Filename)
	if err != nil {
		response.InternalError(c, "文件上传失败")
		return
	}

	fileURL := "/uploads/messages/" + file.Filename
	response.Success(c, gin.H{"file_url": fileURL})
}

// UploadVideo 上传视频
// @Summary 上传视频
// @Description 上传视频文件，支持动态视频、消息视频等场景使用
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "视频文件"
// @Success 200 {object} map[string]interface{} "上传成功，返回文件URL"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "上传失败"
// @Router /uploads/video [post]
func UploadVideo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传视频文件")
		return
	}

	logger.Infof("[Handler] UploadVideo - Filename: %s, Size: %d", file.Filename, file.Size)

	err = c.SaveUploadedFile(file, "./uploads/messages/"+file.Filename)
	if err != nil {
		response.InternalError(c, "文件上传失败")
		return
	}

	fileURL := "/uploads/messages/" + file.Filename
	response.Success(c, gin.H{"file_url": fileURL})
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传通用文件，支持各种类型的文件上传
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "文件"
// @Success 200 {object} map[string]interface{} "上传成功，返回文件URL"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "上传失败"
// @Router /uploads/file [post]
func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件")
		return
	}

	logger.Infof("[Handler] UploadFile - Filename: %s, Size: %d", file.Filename, file.Size)

	err = c.SaveUploadedFile(file, "./uploads/messages/"+file.Filename)
	if err != nil {
		response.InternalError(c, "文件上传失败")
		return
	}

	fileURL := "/uploads/messages/" + file.Filename
	response.Success(c, gin.H{"file_url": fileURL})
}