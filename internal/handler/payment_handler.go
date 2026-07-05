package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

// BuyDiamond 购买钻石
// @Summary 购买钻石
// @Description 用户使用支付方式购买钻石套餐，钻石将直接充值到用户账户
// @Tags 支付
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param did path string true "钻石套餐ID"
// @Success 200 {object} map[string]interface{} "购买成功"
// @Failure 400 {object} map[string]interface{} "参数错误或余额不足"
// @Failure 500 {object} map[string]interface{} "购买失败"
// @Router /pay/diamond/buy/{did} [post]
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

// GetDiamondStock 获取钻石库存
// @Summary 获取钻石库存
// @Description 获取当前用户账户中的钻石余额
// @Tags 支付
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功，返回钻石数量"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /pay/diamond/stock [get]
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

// GetDiamondHistory 获取钻石购买历史
// @Summary 获取钻石购买历史
// @Description 获取当前用户的钻石购买记录列表
// @Tags 支付
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功，返回购买记录列表"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /pay/diamond/history [get]
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

// BuyMember 购买会员
// @Summary 购买会员
// @Description 用户使用支付方式购买会员套餐，会员权益将立即生效
// @Tags 支付
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param vid path string true "会员套餐ID"
// @Success 200 {object} map[string]interface{} "购买成功"
// @Failure 400 {object} map[string]interface{} "参数错误或余额不足"
// @Failure 500 {object} map[string]interface{} "购买失败"
// @Router /pay/member/buy/{vid} [post]
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

// GetMemberHistory 获取会员购买历史
// @Summary 获取会员购买历史
// @Description 获取当前用户的会员购买记录列表
// @Tags 支付
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功，返回购买记录列表"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /pay/member/history [get]
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
