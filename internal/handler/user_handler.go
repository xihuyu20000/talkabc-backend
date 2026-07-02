package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

	users, err := service.GetUserList(options)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, users)
}

func GetUserInfo(c *gin.Context) {
	uid := c.Param("uid")

	user, err := service.GetUserInfo(uid)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, user)
}

func GetFocusList(c *gin.Context) {
	uid := c.Param("uid")

	users, err := service.GetFocusList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, users)
}

func GetFansList(c *gin.Context) {
	uid := c.Param("uid")

	users, err := service.GetFansList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, users)
}

func SetUserNotify(c *gin.Context) {
	userID := middleware.GetUID(c)
	targetID := c.Param("uid")
	flagStr := c.Param("flag")

	flag, err := strconv.Atoi(flagStr)
	if err != nil {
		response.BadRequest(c, "flag参数错误")
		return
	}

	err = service.SetUserNotify(userID, targetID, flag)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

func GreetUser(c *gin.Context) {
	userID := middleware.GetUID(c)
	targetID := c.Param("uid")

	text := c.PostForm("text")

	err := service.GreetUser(userID, targetID, text)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}