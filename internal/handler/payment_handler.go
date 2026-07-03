package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

// BuyDiamond 购买钻石接口
// 请求方式：POST
// 请求路径：/v1/pay/buy/diamond/:did
// 请求参数：did - 钻石套餐ID（路径参数）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 从路径参数获取钻石套餐ID并转换为数字
//   3. 调用 service.BuyDiamond 执行钻石购买逻辑
//   4. 返回操作结果
func BuyDiamond(c *gin.Context) {
	uid := middleware.GetUID(c)
	didStr := c.Param("did")

	did, err := strconv.ParseUint(didStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "钻石ID参数错误")
		return
	}

	logger.Infof("[Handler] BuyDiamond - UID: %s, DiamondID: %d", uid, did)

	err = service.BuyDiamond(uid, uint(did))
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetDiamondStock 获取钻石库存接口
// 请求方式：GET
// 请求路径：/v1/pay/diamond/stock
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：用户当前钻石数量
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetDiamondStock 查询钻石余额
//   3. 返回钻石数量数据
func GetDiamondStock(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetDiamondStock - UID: %s", uid)

	diamond, err := service.GetDiamondStock(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, diamond)
}

// GetDiamondHistory 获取钻石购买历史接口
// 请求方式：GET
// 请求路径：/v1/pay/diamond/history
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：钻石购买记录列表
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetDiamondHistory 查询购买记录
//   3. 返回记录列表数据
func GetDiamondHistory(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetDiamondHistory - UID: %s", uid)

	records, err := service.GetDiamondHistory(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, records)
}

// BuyMember 购买会员接口
// 请求方式：POST
// 请求路径：/v1/pay/buy/member/:vid
// 请求参数：vid - 会员套餐ID（路径参数）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 从路径参数获取会员套餐ID并转换为数字
//   3. 调用 service.BuyMember 执行会员购买逻辑
//   4. 返回操作结果
func BuyMember(c *gin.Context) {
	uid := middleware.GetUID(c)
	vidStr := c.Param("vid")

	vid, err := strconv.ParseUint(vidStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "会员ID参数错误")
		return
	}

	logger.Infof("[Handler] BuyMember - UID: %s, MemberID: %d", uid, vid)

	err = service.BuyMember(uid, uint(vid))
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetMemberHistory 获取会员购买历史接口
// 请求方式：GET
// 请求路径：/v1/pay/member/history
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：会员购买记录列表
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 调用 service.GetMemberHistory 查询购买记录
//   3. 返回记录列表数据
func GetMemberHistory(c *gin.Context) {
	uid := middleware.GetUID(c)

	logger.Infof("[Handler] GetMemberHistory - UID: %s", uid)

	records, err := service.GetMemberHistory(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, records)
}