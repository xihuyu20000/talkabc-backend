package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserList 获取用户列表接口
// 请求方式：GET
// 请求参数：
//   - age1: 年龄下限（可选）
//   - age2: 年龄上限（可选）
//   - gender: 性别（可选）
//   - official: 是否官方认证（可选）
//   - real: 是否实名认证（可选）
//   - latest: 是否最新注册（可选）
//   - distance: 距离范围（可选）
//   - favor: 爱好标签（可选）
//   - job: 职业（可选）
//   - starsign: 星座（可选）
//   - edulevel: 学历（可选）
//   - height: 身高范围（可选）
//   - dating_purpose: 交友目的（可选）
// 返回值：用户列表数据
//
// 业务流程：
//   1. 解析查询参数（均为可选）
//   2. 将参数整理为 map 传递给服务层
//   3. 调用 service.GetUserList 获取符合条件的用户列表
//   4. 返回用户列表数据
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

// GetUserInfo 获取用户信息接口
// 请求方式：GET
// 请求路径：/v1/userinfo/:uid
// 请求参数：uid - 用户ID（路径参数，字符串类型）
// 返回值：用户详细信息
//
// 业务流程：
//   1. 从路径参数获取目标用户ID
//   2. 调用 service.GetUserInfo 查询用户信息
//   3. 返回用户信息数据
func GetUserInfo(c *gin.Context) {
	uid := c.Param("uid")

	user, err := service.GetUserInfo(uid)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, user)
}

// GetFocusList 获取关注列表接口
// 请求方式：GET
// 请求路径：/v1/focuslist/:uid
// 请求参数：uid - 用户ID（路径参数）
// 返回值：该用户关注的用户列表
//
// 业务流程：
//   1. 从路径参数获取用户ID
//   2. 调用 service.GetFocusList 查询关注列表
//   3. 返回关注用户列表
func GetFocusList(c *gin.Context) {
	uid := c.Param("uid")

	users, err := service.GetFocusList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, users)
}

// GetFansList 获取粉丝列表接口
// 请求方式：GET
// 请求路径：/v1/fanslist/:uid
// 请求参数：uid - 用户ID（路径参数）
// 返回值：该用户的粉丝列表
//
// 业务流程：
//   1. 从路径参数获取用户ID
//   2. 调用 service.GetFansList 查询粉丝列表
//   3. 返回粉丝用户列表
func GetFansList(c *gin.Context) {
	uid := c.Param("uid")

	users, err := service.GetFansList(uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, users)
}



// SetUserNotify 设置用户通知开关接口
// 请求方式：POST
// 请求路径：/v1/aimuser/notify/:uid/:flag
// 请求参数：
//   - uid: 目标用户ID（路径参数）
//   - flag: 通知开关，1=开启，0=关闭（路径参数）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 从路径参数获取目标用户ID和通知标志
//   3. 调用 service.SetUserNotify 设置通知状态
//   4. 返回操作结果
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

// GreetUser 打招呼接口
// 请求方式：POST
// 请求路径：/v1/aimuser/greet/:uid
// 请求参数：
//   - uid: 目标用户ID（路径参数）
//   - text: 打招呼内容（POST表单参数）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 从路径参数获取目标用户ID
//   3. 从表单参数获取打招呼内容
//   4. 调用 service.GreetUser 发送打招呼消息
//   5. 返回操作结果
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

// CollectMyInfo 完善个人信息接口
// 请求方式：POST
// 请求路径：/v1/collect/myinfo
// 请求参数：JSON格式
//   - regcountry: 注册国家
//   - mylang: 使用语言
//   - nickname: 昵称
//   - birthyear: 出生年份
//   - gender: 性别
//   - height: 身高
//   - weight: 体重
//   - city: 城市
//   - school: 学校
//   - job: 职业
//   - edulevel: 学历
//   - starsign: 星座
//   - favors: 爱好列表
//   - dating_purposes: 交友目的列表（多选）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 解析 JSON 请求体获取用户信息
//   3. 将信息整理为 map 格式
//   4. 调用 service.CollectMyInfo 更新用户信息
//   5. 返回操作结果
func CollectMyInfo(c *gin.Context) {
	userID := middleware.GetUID(c)

	var req struct {
		RegCountry      string   `json:"regcountry"`
		MyLang          string   `json:"mylang"`
		Nickname        string   `json:"nickname"`
		BirthYear       int      `json:"birthyear"`
		Gender          int      `json:"gender"`
		Height          int      `json:"height"`
		Weight          int      `json:"weight"`
		City            string   `json:"city"`
		School          string   `json:"school"`
		Job             string   `json:"job"`
		EduLevel        int      `json:"edulevel"`
		StarSign        int      `json:"starsign"`
		Favors          []string `json:"favors"`
		DatingPurposes  []string `json:"dating_purposes"`
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

	err := service.CollectMyInfo(userID, info)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// CollectAimInfo 设置理想对象条件接口
// 请求方式：POST
// 请求路径：/v1/collect/aiminfo
// 请求参数：JSON格式
//   - birthyear: 出生年份范围（数组）
//   - gender: 性别
//   - height: 身高范围（数组）
//   - weight: 体重范围
//   - edulevel: 学历范围（数组）
//   - starsign: 星座范围（数组）
//   - favors: 共同爱好列表（数组）
// 身份验证：通过 JWT token 获取当前用户ID
// 返回值：操作结果
//
// 业务流程：
//   1. 从 JWT token 获取当前用户ID
//   2. 解析 JSON 请求体获取理想对象条件
//   3. 将条件整理为 map 格式
//   4. 调用 service.CollectAimInfo 保存条件设置
//   5. 返回操作结果
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

	err := service.CollectAimInfo(userID, info)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}