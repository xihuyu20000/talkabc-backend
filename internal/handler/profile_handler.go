package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

/**
昵称的有效性检查包括：
- 是否含有特殊字符：可通过正则判断
- 是否长度合法
- 是否含有暴力色情等违规字符：可通过敏感字库过滤
*/
// CollectMyInfo 完善个人信息
// @Summary 完善个人信息
// @Description 用户完善个人资料信息
// @Tags 个人资料
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param body body map[string]interface{} true "个人信息"
// @Success 200 {object} map[string]interface{} "保存成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "保存失败"
// @Router /profile/me [post]
func CollectMyInfo(c *gin.Context) {
	userID := middleware.GetUID(c)

	var req struct {
		RegCountry     string   `json:"regcountry"`
		MyLang         string   `json:"mylang"`
		Nickname       string   `json:"nickname"`
		BirthYear      int      `json:"birthyear"`
		Gender         int      `json:"gender"`
		Height         int      `json:"height"`
		Weight         int      `json:"weight"`
		City           string   `json:"city"`
		School         string   `json:"school"`
		Job            string   `json:"job"`
		EduLevel       int      `json:"edulevel"`
		StarSign       int      `json:"starsign"`
		Favors         []string `json:"favors"`
		DatingPurposes []string `json:"dating_purposes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	info := make(map[string]interface{})
	info["regcountry"] = req.RegCountry
	info["mylang"] = req.MyLang
	info["nickname"] = req.Nickname
	info["birthyear"] = req.BirthYear
	info["gender"] = req.Gender
	info["height"] = req.Height
	info["weight"] = req.Weight
	info["city"] = req.City
	info["school"] = req.School
	info["job"] = req.Job
	info["edulevel"] = req.EduLevel
	info["starsign"] = req.StarSign
	info["favors"] = req.Favors
	info["dating_purposes"] = req.DatingPurposes

	logger.Infof("[Handler] CollectMyInfo - UserID: %s, Nickname: %s, Gender: %d, BirthYear: %d", userID, req.Nickname, req.Gender, req.BirthYear)

	err := service.CollectMyInfo(userID, info)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// CollectAimInfo 设置理想对象条件
// @Summary 设置理想对象条件
// @Description 用户设置理想对象的筛选条件
// @Tags 个人资料
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param body body map[string]interface{} true "理想对象条件"
// @Success 200 {object} map[string]interface{} "保存成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "保存失败"
// @Router /profile/preferences [post]
func CollectAimInfo(c *gin.Context) {
	userID := middleware.GetUID(c)

	var req struct {
		BirthYear []string `json:"birthyear"`
		Gender    int      `json:"gender"`
		Height    []string `json:"height"`
		Weight    string   `json:"weight"`
		EduLevel  []string `json:"edulevel"`
		StarSign  []string `json:"starsign"`
		Favors    []string `json:"favors"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	info := make(map[string]interface{})
	info["birthyear"] = req.BirthYear
	info["gender"] = req.Gender
	info["height"] = req.Height
	info["weight"] = req.Weight
	info["edulevel"] = req.EduLevel
	info["starsign"] = req.StarSign
	info["favors"] = req.Favors

	logger.Infof("[Handler] CollectAimInfo - UserID: %s, Gender: %d, Height: %v", userID, req.Gender, req.Height)

	err := service.CollectAimInfo(userID, info)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}