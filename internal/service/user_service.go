package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"fmt"
	"strconv"
)

// UserInfo 用户信息DTO
// 用户信息展示数据传输对象，用于API返回用户基本信息和动态统计数据
type UserInfo struct {
	UID            string   `json:"uid"`             // 用户对外唯一标识（雪花ID）
	AvatarURL      string   `json:"avatarurl"`       // 头像URL
	Nickname       string   `json:"nickname"`        // 昵称
	Gender         int      `json:"gender"`          // 性别：0-未知，1-男，2-女
	FocusNum       int      `json:"focusnum"`        // 关注数量（动态统计）
	FansNum        int      `json:"fansnum"`         // 粉丝数量（动态统计）
	Country        string   `json:"myarea"`          // 国家/地区
	Language       string   `json:"mylang"`          // 语言偏好
	BirthYear      int      `json:"birthyear"`       // 出生年份
	StarSign       string   `json:"starsign"`        // 星座
	EduLevel       int      `json:"edulevel"`        // 教育程度：1-初中及以下，2-高中，3-大专，4-本科，5-研究生及以上
	Job            string   `json:"job"`             // 职业
	City           string   `json:"city"`            // 城市
	FreAreas       []string `json:"freareas"`        // 常去地点数组
	Favors         []string `json:"favors"`          // 爱好列表
	DatingPurposes []string `json:"dating_purposes"` // 交友目的列表
	SignText       string   `json:"signtext"`        // 个性签名
}

// ConvertUserToUserInfo 将User模型转换为UserInfo DTO
// 参数说明：
//   - user: 用户模型对象
//
// 返回值：
//   - UserInfo: 用户信息DTO
//
// 转换逻辑：
//   1. 查询用户关注数和粉丝数（动态统计）
//   2. 将User模型字段映射到UserInfo结构体
//   3. 爱好列表暂时返回空数组（后续通过hobby_tags表查询）
func ConvertUserToUserInfo(user model.User) UserInfo {
	focusNum, _ := repository.GetFocusCount(user.ID)
	fansNum, _ := repository.GetFansCount(user.ID)

	return UserInfo{
		UID:            user.Uid,
		AvatarURL:      user.AvatarURL,
		Nickname:       user.Nickname,
		Gender:         user.Gender,
		FocusNum:       focusNum,
		FansNum:        fansNum,
		Country:        user.Country,
		Language:       user.Language,
		BirthYear:      user.BirthYear,
		StarSign:       user.StarSign,
		EduLevel:       user.EduLevel,
		Job:            user.Job,
		City:           user.City,
		FreAreas:       user.FrequentAreas,
		Favors:         []string{},
		DatingPurposes: []string{},
		SignText:       user.SignText,
	}
}

// GetUserList 获取用户列表
// 参数说明：
//   - options: 所有筛选条件（age1, age2, gender, official, real, latest, distance, favor, job, starsign, edulevel, height, dating_purpose）
//
// 返回值：
//   - []UserInfo: 用户信息列表
//   - error: 错误信息
//
// 查询逻辑：
//   1. 构建筛选条件映射
//   2. 将options中的参数转换为筛选条件
//   3. 调用repository层查询用户列表
//   4. 将每个User模型转换为UserInfo DTO
func GetUserList(options map[string]string) ([]UserInfo, error) {
	filters := make(map[string]interface{})

	for key, value := range options {
		if v, err := strconv.Atoi(value); err == nil {
			filters[key] = v
		}
	}

	users, err := repository.GetUserList(filters)
	if err != nil {
		return nil, err
	}

	var result []UserInfo
	for _, user := range users {
		result = append(result, ConvertUserToUserInfo(user))
	}

	return result, nil
}

// GetUserInfo 获取单个用户信息
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//
// 返回值：
//   - *UserInfo: 用户信息DTO指针
//   - error: 错误信息
//
// 查询逻辑：
//   1. 根据uid查询用户模型
//   2. 转换为UserInfo DTO返回
func GetUserInfo(uid string) (*UserInfo, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	result := ConvertUserToUserInfo(*user)
	return &result, nil
}

// GetFocusList 获取用户关注列表
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//
// 返回值：
//   - []UserInfo: 关注用户信息列表
//   - error: 错误信息
//
// 查询逻辑：
//   1. 根据uid查询用户模型
//   2. 查询该用户的关注列表
//   3. 将每个关注用户转换为UserInfo DTO
func GetFocusList(uid string) ([]UserInfo, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, err
	}

	users, err := repository.GetFocusList(user.ID)
	if err != nil {
		return nil, err
	}

	var result []UserInfo
	for _, user := range users {
		result = append(result, ConvertUserToUserInfo(user))
	}

	return result, nil
}

// GetFansList 获取用户粉丝列表
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//
// 返回值：
//   - []UserInfo: 粉丝用户信息列表
//   - error: 错误信息
//
// 查询逻辑：
//   1. 根据uid查询用户模型
//   2. 查询该用户的粉丝列表
//   3. 将每个粉丝用户转换为UserInfo DTO
func GetFansList(uid string) ([]UserInfo, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, err
	}

	users, err := repository.GetFansList(user.ID)
	if err != nil {
		return nil, err
	}

	var result []UserInfo
	for _, user := range users {
		result = append(result, ConvertUserToUserInfo(user))
	}

	return result, nil
}

// FocusUser 关注/取消关注用户
// 参数说明：
//   - userUID: 当前用户对外唯一标识（雪花ID）
//   - targetUID: 目标用户对外唯一标识（雪花ID）
//   - flag: 操作标志，1=关注，0=取消关注
//
// 返回值：
//   - error: 错误信息
//
// 操作逻辑：
//   1. 校验不能关注自己
//   2. 查询当前用户和目标用户模型
//   3. 根据flag执行关注或取消关注操作
func FocusUser(userUID, targetUID string, flag int) error {
	if userUID == targetUID {
		return fmt.Errorf("不能关注自己")
	}

	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	target, err := repository.GetUserByUID(targetUID)
	if err != nil {
		return fmt.Errorf("目标用户不存在")
	}

	if flag == 1 {
		return repository.AddFocus(user.ID, target.ID)
	} else {
		return repository.RemoveFocus(user.ID, target.ID)
	}
}

// BlockUser 拉黑/取消拉黑用户
// 参数说明：
//   - userUID: 当前用户对外唯一标识（雪花ID）
//   - targetUID: 目标用户对外唯一标识（雪花ID）
//   - flag: 操作标志，1=拉黑，0=取消拉黑
//
// 返回值：
//   - error: 错误信息
//
// 操作逻辑：
//   1. 校验不能拉黑自己
//   2. 查询当前用户和目标用户模型
//   3. 根据flag执行拉黑或取消拉黑操作
func BlockUser(userUID, targetUID string, flag int) error {
	if userUID == targetUID {
		return fmt.Errorf("不能拉黑自己")
	}

	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	target, err := repository.GetUserByUID(targetUID)
	if err != nil {
		return fmt.Errorf("目标用户不存在")
	}

	if flag == 1 {
		return repository.AddBlock(user.ID, target.ID)
	} else {
		return repository.RemoveBlock(user.ID, target.ID)
	}
}

// SetUserNotify 设置用户通知开关
// 参数说明：
//   - userUID: 当前用户对外唯一标识（雪花ID）
//   - targetUID: 目标用户对外唯一标识（雪花ID）
//   - flag: 通知标志，1=开启，0=关闭
//
// 返回值：
//   - error: 错误信息
//
// 操作逻辑：
//   1. 查询当前用户和目标用户模型
//   2. 设置对目标用户的通知开关状态
func SetUserNotify(userUID, targetUID string, flag int) error {
	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	target, err := repository.GetUserByUID(targetUID)
	if err != nil {
		return fmt.Errorf("目标用户不存在")
	}

	return repository.SetNotify(user.ID, target.ID, flag)
}

// CollectMyInfo 更新用户个人信息
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//   - info: 更新的用户信息（键值对形式）
//
// 返回值：
//   - error: 错误信息
//
// 更新字段支持：
//   - regcountry: 国家/地区（string）
//   - mylang: 语言偏好（string）
//   - nickname: 昵称（string）
//   - birthyear: 出生年份（int）
//   - gender: 性别（int）
//   - height: 身高（int，cm）
//   - weight: 体重（int，kg）
//   - city: 城市（string）
//   - school: 学校（string）
//   - job: 职业（string）
//   - edulevel: 教育程度（int）
//   - starsign: 星座（int，转换为字符串存储）
//
// 更新逻辑：
//   1. 根据uid查询用户模型
//   2. 遍历info中的键值对，更新对应的用户字段
//   3. 保存更新到数据库
func CollectMyInfo(uid string, info map[string]interface{}) error {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	if v, ok := info["regcountry"].(string); ok {
		user.Country = v
	}
	if v, ok := info["mylang"].(string); ok {
		user.Language = v
	}
	if v, ok := info["nickname"].(string); ok {
		user.Nickname = v
	}
	if v, ok := info["birthyear"].(int); ok {
		user.BirthYear = v
	}
	if v, ok := info["gender"].(int); ok {
		user.Gender = v
	}
	if v, ok := info["height"].(int); ok {
		user.Height = v
	}
	if v, ok := info["weight"].(int); ok {
		user.Weight = v
	}
	if v, ok := info["city"].(string); ok {
		user.City = v
	}
	if v, ok := info["school"].(string); ok {
		user.School = v
	}
	if v, ok := info["job"].(string); ok {
		user.Job = v
	}
	if v, ok := info["edulevel"].(int); ok {
		user.EduLevel = v
	}
	if v, ok := info["starsign"].(int); ok {
		user.StarSign = strconv.Itoa(v)
	}

	if err := repository.UpdateUser(user); err != nil {
		return err
	}

	if favors, ok := info["favors"].([]string); ok && len(favors) > 0 {
		var tagIDs []uint
		for _, id := range favors {
			if v, err := strconv.ParseUint(id, 10, 32); err == nil {
				tagIDs = append(tagIDs, uint(v))
			}
		}
		if err := repository.SaveUserHobbies(uid, tagIDs); err != nil {
			return err
		}
	}

	if datingPurposes, ok := info["dating_purposes"].([]string); ok && len(datingPurposes) > 0 {
		var purposeIDs []uint
		for _, id := range datingPurposes {
			if v, err := strconv.ParseUint(id, 10, 32); err == nil {
				purposeIDs = append(purposeIDs, uint(v))
			}
		}
		if err := repository.SaveUserDatingPurposes(uid, purposeIDs); err != nil {
			return err
		}
	}

	return nil
}

// CollectAimInfo 更新用户理想对象条件
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//   - info: 理想对象条件（键值对形式）
//
// 返回值：
//   - error: 错误信息
//
// 操作逻辑：
//   1. 根据uid查询用户模型
//   2. 将info直接作为JSON数据保存到用户的aim字段
//   3. 更新用户记录到数据库
func CollectAimInfo(uid string, info map[string]interface{}) error {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	user.Aim = info

	return repository.UpdateUser(user)
}

// GreetUser 打招呼（发送问候消息）
// 参数说明：
//   - userUID: 当前用户对外唯一标识（雪花ID）
//   - targetUID: 目标用户对外唯一标识（雪花ID）
//   - text: 问候消息内容
//
// 返回值：
//   - error: 错误信息
//
// 操作逻辑：
//   1. 查询当前用户和目标用户模型
//   2. 调用repository层发送问候消息
func GreetUser(userUID, targetUID string, text string) error {
	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	target, err := repository.GetUserByUID(targetUID)
	if err != nil {
		return fmt.Errorf("目标用户不存在")
	}

	return repository.GreetUser(user.ID, target.ID, text)
}