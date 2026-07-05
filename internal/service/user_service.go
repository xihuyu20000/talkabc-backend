package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/pkg/security"
	"fmt"
	"strconv"
)

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

// ValidateNickname 验证昵称的有效性
// 校验规则（调用security包进行全面安全检查）：
//   1. 长度限制：2-20个字符
//   2. 字符限制：仅允许中文、英文、数字、下划线、连字符
//   3. 敏感词过滤：禁止包含敏感词汇
//   4. URL过滤：禁止包含超链接
//   5. HTML过滤：禁止包含HTML标签
//   6. XSS过滤：禁止包含XSS攻击代码
func ValidateNickname(nickname string) error {
	return security.ValidateNickname(nickname)
}

// ==================== 用户信息转换与查询 ====================

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

func GetUserInfo(uid string) (*UserInfo, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	result := ConvertUserToUserInfo(*user)
	return &result, nil
}

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