package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Description 根据筛选条件获取用户列表
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param age1 query string false "最小年龄"
// @Param age2 query string false "最大年龄"
// @Param gender query string false "性别"
// @Param official query string false "官方认证"
// @Param real query string false "真实认证"
// @Param latest query string false "最新注册"
// @Param distance query string false "距离"
// @Param favor query string false "爱好"
// @Param job query string false "职业"
// @Param starsign query string false "星座"
// @Param edulevel query string false "教育程度"
// @Param height query string false "身高"
// @Param dating_purpose query string false "交友目的"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /users [get]
func GetUserList(c *gin.Context) {
	options := make(map[string]string)

	if age1 := c.Query("age1"); age1 != "" {
		options["age1"] = age1
	}
	if age2 := c.Query("age2"); age2 != "" {
		options["age2"] = age2
	}
	if gender := c.Query("gender"); gender != "" {
		options["gender"] = gender
	}
	if official := c.Query("official"); official != "" {
		options["official"] = official
	}
	if real := c.Query("real"); real != "" {
		options["real"] = real
	}
	if latest := c.Query("latest"); latest != "" {
		options["latest"] = latest
	}
	if distance := c.Query("distance"); distance != "" {
		options["distance"] = distance
	}
	if favor := c.Query("favor"); favor != "" {
		options["favor"] = favor
	}
	if job := c.Query("job"); job != "" {
		options["job"] = job
	}
	if starsign := c.Query("starsign"); starsign != "" {
		options["starsign"] = starsign
	}
	if edulevel := c.Query("edulevel"); edulevel != "" {
		options["edulevel"] = edulevel
	}
	if height := c.Query("height"); height != "" {
		options["height"] = height
	}
	if datingPurpose := c.Query("dating_purpose"); datingPurpose != "" {
		options["dating_purpose"] = datingPurpose
	}

	logger.Infof("[Handler] GetUserList - Options: %v", options)

	users, err := service.GetUserList(options)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, users)
}

// GetUserInfo 获取用户信息
// @Summary 获取用户信息
// @Description 获取指定用户的详细信息
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "用户UID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /users/{uid} [get]
func GetUserInfo(c *gin.Context) {
	uid := c.Param("uid")

	logger.Infof("[Handler] GetUserInfo - UID: %s", uid)

	user, err := service.GetUserInfo(uid)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, user)
}

// GetFocusList 获取关注列表
// @Summary 获取关注列表
// @Description 获取指定用户的关注列表
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "用户UID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /users/{uid}/following [get]
func GetFocusList(c *gin.Context) {
	uid := c.Param("uid")

	logger.Infof("[Handler] GetFocusList - UID: %s", uid)

	users, err := service.GetFocusList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, users)
}

// GetFansList 获取粉丝列表
// @Summary 获取粉丝列表
// @Description 获取指定用户的粉丝列表
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "用户UID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /users/{uid}/fans [get]
func GetFansList(c *gin.Context) {
	uid := c.Param("uid")

	logger.Infof("[Handler] GetFansList - UID: %s", uid)

	users, err := service.GetFansList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, users)
}

// SetUserNotify 设置用户通知
// @Summary 设置用户通知
// @Description 设置是否接收指定用户的消息通知
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "目标用户UID"
// @Param flag path string true "通知标志：1-开启，0-关闭"
// @Success 200 {object} map[string]interface{} "设置成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "设置失败"
// @Router /users/{uid}/notification/{flag} [post]
func SetUserNotify(c *gin.Context) {
	userID := middleware.GetUID(c)
	targetID := c.Param("uid")
	flagStr := c.Param("flag")

	flag, err := strconv.Atoi(flagStr)
	if err != nil {
		response.BadRequest(c, "flag参数错误")
		return
	}

	logger.Infof("[Handler] SetUserNotify - UserID: %s, TargetID: %s, Flag: %d", userID, targetID, flag)

	err = service.SetUserNotify(userID, targetID, flag)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// GreetUser 打招呼
// @Summary 打招呼
// @Description 向指定用户发送打招呼消息
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param uid path string true "目标用户UID"
// @Param text formData string true "消息内容"
// @Success 200 {object} map[string]interface{} "发送成功"
// @Failure 500 {object} map[string]interface{} "发送失败"
// @Router /users/{uid}/greet [post]
func GreetUser(c *gin.Context) {
	userID := middleware.GetUID(c)
	targetID := c.Param("uid")

	text := c.PostForm("text")

	logger.Infof("[Handler] GreetUser - UserID: %s, TargetID: %s, Text: %s", userID, targetID, text)

	err := service.GreetUser(userID, targetID, text)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}