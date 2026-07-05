package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

/**
昵称的安全性检查包括（调用 security 包进行全面验证）：
- 长度限制：2-20个字符
- 字符类型：仅允许中文、英文、数字、下划线、连字符
- 敏感词过滤：禁止包含暴力、色情、政治敏感、诈骗等违规词汇
- URL过滤：禁止包含 http/https/www 等超链接
- HTML过滤：禁止包含 HTML 标签
- XSS过滤：禁止包含 XSS 攻击代码

个性签名的安全性检查包括（调用 security 包进行全面验证）：
- 长度限制：最大200字符
- 敏感词过滤：禁止包含暴力、色情、政治敏感、诈骗等违规词汇
- URL过滤：禁止包含 http/https/www 等超链接
- HTML过滤：禁止包含 HTML 标签
- JavaScript过滤：禁止包含脚本代码
- SQL注入过滤：禁止包含 SQL 注入代码
- XSS过滤：禁止包含 XSS 攻击代码
*/
// CollectMyInfo 完善个人信息
// @Summary 完善个人信息
// @Description 用户完善个人资料信息
// @Description 收集内容包括：昵称、性别、出生年份、身高、体重、城市、学校、职业、教育程度、星座、爱好、交友目的等
// @Description 【安全规则】昵称需通过以下校验：
// @Description   1. 长度：2-20个字符
// @Description   2. 字符类型：仅允许中文、英文、数字、下划线、连字符
// @Description   3. 敏感词过滤：禁止包含暴力、色情、政治敏感、诈骗等违规词汇
// @Description   4. URL过滤：禁止包含超链接
// @Description   5. HTML过滤：禁止包含HTML标签
// @Description   6. XSS过滤：禁止包含XSS攻击代码
// @Tags 个人资料
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param body body map[string]interface{} true "个人信息"
// @Success 200 {object} map[string]interface{} "保存成功"
// @Failure 400 {object} map[string]interface{} "参数错误或昵称包含违规内容"
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
// @Description 设置内容包括：年龄范围、性别、身高范围、体重范围、教育程度、星座、爱好等
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

// CheckProfileStatus 检查资料收集状态
// @Summary 检查资料收集状态
// @Description 检查用户是否已完成资料收集，用于首次登录时判断是否需要跳转到资料收集页面
// @Tags 个人资料
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "返回资料收集状态"
// @Success 200 {object} map[string]interface{} "data.profile_completed": true表示已完成，false表示需要收集
// @Failure 401 {object} map[string]interface{} "未登录"
// @Failure 500 {object} map[string]interface{} "查询失败"
// @Router /profile/status [get]
func CheckProfileStatus(c *gin.Context) {
	userID := middleware.GetUID(c)

	logger.Infof("[Handler] CheckProfileStatus - UserID: %s", userID)

	completed, err := service.CheckProfileStatus(userID)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"profile_completed": completed})
}

// SetSignText 设置个性签名
// @Summary 设置个性签名
// @Description 用户设置个人主页的个性签名
// @Description 【安全规则】签名需通过以下校验：
// @Description   1. 长度：最大200字符
// @Description   2. 敏感词过滤：禁止包含暴力、色情、政治敏感、诈骗等违规词汇
// @Description   3. URL过滤：禁止包含 http/https/www 等超链接
// @Description   4. HTML过滤：禁止包含 HTML 标签
// @Description   5. JavaScript过滤：禁止包含脚本代码（如 <script>、eval()、alert()）
// @Description   6. SQL注入过滤：禁止包含 SQL 注入代码（如 SELECT、INSERT、OR 1=1）
// @Description   7. XSS过滤：禁止包含 XSS 攻击代码（如 onclick=、onerror=）
// @Tags 个人资料
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param body body map[string]string true "签名内容"
// @Success 200 {object} map[string]interface{} "保存成功"
// @Failure 400 {object} map[string]interface{} "参数错误或签名包含违规内容"
// @Failure 500 {object} map[string]interface{} "保存失败"
// @Router /profile/sign [post]
func SetSignText(c *gin.Context) {
	userID := middleware.GetUID(c)

	var req struct {
		SignText string `json:"sign_text"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	logger.Infof("[Handler] SetSignText - UserID: %s, SignText: %s", userID, req.SignText)

	err := service.SetSignText(userID, req.SignText)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// CompleteProfile 完成资料收集
// @Summary 完成资料收集
// @Description 用户完成所有资料收集后调用此接口，标记资料收集状态为已完成，允许进入首页
// @Tags 个人资料
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "标记成功"
// @Failure 401 {object} map[string]interface{} "未登录"
// @Failure 500 {object} map[string]interface{} "操作失败"
// @Router /profile/complete [post]
func CompleteProfile(c *gin.Context) {
	userID := middleware.GetUID(c)

	logger.Infof("[Handler] CompleteProfile - UserID: %s", userID)

	err := service.SetProfileCompleted(userID)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "资料收集完成"})
}