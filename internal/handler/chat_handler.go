package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetLatestAdBanner 获取最新广告横幅接口
// 请求方式：GET
// 请求路径：/v1/ad/latest
// 返回值：最新广告横幅列表
//
// 业务流程：
//   1. 调用 service.GetLatestAdBanner 查询最新广告
//   2. 返回广告列表数据
func GetLatestAdBanner(c *gin.Context) {
	banners, err := service.GetLatestAdBanner()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, banners)
}

// GetSystemMsgList 获取系统消息列表接口
// 请求方式：GET
// 请求路径：/v1/sysmsglist
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：系统消息列表
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetSystemMsgList 查询系统消息
//   3. 返回系统消息列表
func GetSystemMsgList(c *gin.Context) {
	uid := middleware.GetUID(c)

	msgs, err := service.GetSystemMsgList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, msgs)
}

// GetLatestUserMsg 获取最新用户消息接口
// 请求方式：GET
// 请求路径：/v1/latestusermsg
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：最新的用户聊天消息列表
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetLatestUserMsg 查询最新用户消息
//   3. 返回最新消息列表（每个对话的最后一条消息）
func GetLatestUserMsg(c *gin.Context) {
	uid := middleware.GetUID(c)

	msgs, err := service.GetLatestUserMsg(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, msgs)
}

// GetUserMsgHistory 获取用户聊天历史记录接口
// 请求方式：GET
// 请求路径：/v1/usermsg/:uid
// 请求参数：uid - 聊天对象ID（路径参数）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：与指定用户的聊天历史记录
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 从路径参数获取聊天对象ID
//   3. 调用 service.GetUserMsgHistory 查询聊天历史
//   4. 返回聊天记录列表
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

// SetMessageTop 设置消息置顶接口
// 请求方式：POST
// 请求路径：/v1/setmsg/top/:uid/:flag
// 请求参数：
//   - uid: 聊天对象ID（路径参数）
//   - flag: 置顶标志，1=置顶，0=取消置顶（路径参数）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 从路径参数获取聊天对象ID和置顶标志
//   3. 调用 service.SetMessageTop 设置置顶状态
//   4. 返回操作结果
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

// AddFriend 添加好友接口
// 请求方式：POST
// 请求路径：/v1/addfriend/:uid/:flag
// 请求参数：
//   - uid: 目标用户ID（路径参数）
//   - flag: 操作标志，1=发送好友请求，0=删除好友（路径参数）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 从路径参数获取目标用户ID和操作标志
//   3. 调用 service.AddFriend 执行添加好友或删除好友操作
//   4. 返回操作结果
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

// ClearChatHistory 清空聊天记录接口
// 请求方式：POST
// 请求路径：/v1/clearchathistory/:uid
// 请求参数：uid - 聊天对象ID（路径参数）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 从路径参数获取聊天对象ID
//   3. 调用 service.ClearChatHistory 清空聊天记录
//   4. 返回操作结果
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

// SendGift 发送礼物接口
// 请求方式：POST
// 请求路径：/v1/sendmsg/gift/:uid/:giftid
// 请求参数：
//   - uid: 接收者ID（路径参数）
//   - giftid: 礼物ID（路径参数）
// 身份验证：通过 JWT token 获取发送者ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取发送者ID
//   2. 从路径参数获取接收者ID和礼物ID
//   3. 调用 service.SendGift 执行礼物赠送逻辑
//   4. 返回操作结果
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