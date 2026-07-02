package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

// GetLatestMoment 获取最新动态列表接口
// 请求方式：GET
// 请求路径：/v1/moment/latest
// 返回值：最新动态列表
//
// 业务流程：
//   1. 调用 service.GetLatestMoment 查询最新动态
//   2. 返回动态列表数据
func GetLatestMoment(c *gin.Context) {
	moments, err := service.GetLatestMoment()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, moments)
}

// GetMyLatestMoment 获取我的最新动态接口
// 请求方式：GET
// 请求路径：/v1/moment/mylatest
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：当前用户发布的最新动态列表
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetMyLatestMoment 查询我的动态
//   3. 返回动态列表数据
func GetMyLatestMoment(c *gin.Context) {
	uid := middleware.GetUID(c)

	moments, err := service.GetMyLatestMoment(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, moments)
}

// GetUserMoment 获取指定用户动态接口
// 请求方式：GET
// 请求路径：/v1/moment/user/:uid
// 请求参数：uid - 用户ID（路径参数）
// 返回值：指定用户发布的动态列表
//
// 业务流程：
//   1. 从路径参数获取目标用户ID
//   2. 调用 service.GetUserMoment 查询用户动态
//   3. 返回动态列表数据
func GetUserMoment(c *gin.Context) {
	uid := c.Param("uid")

	moments, err := service.GetUserMoment(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, moments)
}

// GetMomentComments 获取动态评论列表接口
// 请求方式：GET
// 请求路径：/v1/moment/comments/:mid
// 请求参数：mid - 动态ID（路径参数）
// 返回值：动态的评论列表
//
// 业务流程：
//   1. 从路径参数获取动态ID并转换为数字
//   2. 调用 service.GetMomentComments 查询评论列表
//   3. 返回评论列表数据
func GetMomentComments(c *gin.Context) {
	midStr := c.Param("mid")

	mid, err := strconv.ParseUint(midStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "动态ID参数错误")
		return
	}

	comments, err := service.GetMomentComments(uint(mid))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, comments)
}



// PublishMoment 发布动态接口
// 请求方式：POST
// 请求路径：/v1/moment/publish
// 请求参数：
//   - text: 动态文本内容（JSON请求体）
//   - location: 地理位置（JSON请求体）
//   - files: 图片文件列表（multipart/form-data）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 解析 JSON 请求体获取文本内容和地理位置
//   3. 解析 multipart/form-data 获取图片文件列表
//   4. 保存图片到服务器 uploads/moments/ 目录
//   5. 调用 service.PublishMoment 保存动态记录
//   6. 返回操作结果
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

	err := service.PublishMoment(userID, req.Text, files, req.Location)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}