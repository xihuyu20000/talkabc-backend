package handler

import (
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// UploadAvatar 上传头像接口
// 请求方式：POST
// 请求路径：/v1/user/upload-avatar
// 请求参数：file - 头像图片文件（multipart/form-data）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果，包含头像URL
func UploadAvatar(c *gin.Context) {
	uid := middleware.GetUID(c)

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传头像")
		return
	}

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

// UploadImage 上传图片接口
// 请求方式：POST
// 请求路径：/v1/upload/image
// 请求参数：file - 图片文件（multipart/form-data）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：上传成功返回文件URL，失败返回错误信息
func UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传图片")
		return
	}

	err = c.SaveUploadedFile(file, "./uploads/messages/"+file.Filename)
	if err != nil {
		response.InternalError(c, "文件上传失败")
		return
	}

	fileURL := "/uploads/messages/" + file.Filename
	response.Success(c, gin.H{"file_url": fileURL})
}

// UploadAudio 上传音频文件接口
// 请求方式：POST
// 请求路径：/v1/upload/audio
// 请求参数：file - 音频文件（multipart/form-data）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：上传成功返回文件URL，失败返回错误信息
func UploadAudio(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传音频文件")
		return
	}

	err = c.SaveUploadedFile(file, "./uploads/messages/"+file.Filename)
	if err != nil {
		response.InternalError(c, "文件上传失败")
		return
	}

	fileURL := "/uploads/messages/" + file.Filename
	response.Success(c, gin.H{"file_url": fileURL})
}

// UploadVideo 上传视频文件接口
// 请求方式：POST
// 请求路径：/v1/upload/video
// 请求参数：file - 视频文件（multipart/form-data）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：上传成功返回文件URL，失败返回错误信息
func UploadVideo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传视频文件")
		return
	}

	err = c.SaveUploadedFile(file, "./uploads/messages/"+file.Filename)
	if err != nil {
		response.InternalError(c, "文件上传失败")
		return
	}

	fileURL := "/uploads/messages/" + file.Filename
	response.Success(c, gin.H{"file_url": fileURL})
}

// UploadFile 上传文件接口
// 请求方式：POST
// 请求路径：/v1/upload/file
// 请求参数：file - 文件（multipart/form-data）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：上传成功返回文件URL，失败返回错误信息
func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件")
		return
	}

	err = c.SaveUploadedFile(file, "./uploads/messages/"+file.Filename)
	if err != nil {
		response.InternalError(c, "文件上传失败")
		return
	}

	fileURL := "/uploads/messages/" + file.Filename
	response.Success(c, gin.H{"file_url": fileURL})
}