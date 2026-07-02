package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

// GetPraiseMeList 获取赞我的列表接口
// 请求方式：GET
// 请求路径：/v1/notify/praiseme
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：赞我的用户列表
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetPraiseMeList 查询赞我的用户
//   3. 返回用户列表数据
func GetPraiseMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	list, err := service.GetPraiseMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetCommentMeList 获取评论我的列表接口
// 请求方式：GET
// 请求路径：/v1/notify/commentme
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：评论我的用户列表
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetCommentMeList 查询评论我的用户
//   3. 返回用户列表数据
func GetCommentMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	list, err := service.GetCommentMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetAddMeList 获取添加我的列表接口
// 请求方式：GET
// 请求路径：/v1/notify/addme
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：添加我为好友的用户列表
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetAddMeList 查询添加我的用户
//   3. 返回用户列表数据
func GetAddMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	list, err := service.GetAddMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetVisitMeList 获取访问我的列表接口
// 请求方式：GET
// 请求路径：/v1/notify/visitme
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：访问我主页的用户列表
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetVisitMeList 查询访问我的用户
//   3. 返回用户列表数据
func GetVisitMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	list, err := service.GetVisitMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetLikeMeList 获取喜欢我的列表接口
// 请求方式：GET
// 请求路径：/v1/notify/likeme
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：喜欢我的用户列表
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetLikeMeList 查询喜欢我的用户
//   3. 返回用户列表数据
func GetLikeMeList(c *gin.Context) {
	uid := middleware.GetUID(c)

	list, err := service.GetLikeMeList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// AgreeFriendRequest 同意好友请求接口
// 请求方式：POST
// 请求路径：/v1/agreefriend/:uid/:flag
// 请求参数：
//   - uid: 请求好友的用户ID（路径参数）
//   - flag: 操作标志，1=同意，0=拒绝（路径参数）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 从路径参数获取请求用户ID和操作标志
//   3. 调用 service.AgreeFriendRequest 执行同意或拒绝操作
//   4. 返回操作结果
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