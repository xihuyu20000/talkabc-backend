package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

// ==================== 昵称校验规则 ====================

// 敏感词库：包含暴力、色情、政治敏感等违规词汇
var sensitiveWords = []string{
	"暴力", "色情", "赌博", "毒品", "违法", "犯罪", "枪支", "爆炸",
	"恐怖", "邪教", "反动", "淫秽", "低俗", "辱骂", "侮辱", "诽谤",
	"诈骗", "传销", "黑客", "病毒", "木马", "破解", "盗版", "色情",
	"妓女", "卖淫", "嫖娼", "强奸", "乱伦", "变态", "裸露", "性器官",
	"傻逼", "操你", "妈逼", "狗日", "去死", "滚蛋", "垃圾", "废物",
}

// ValidateNickname 验证昵称的有效性
// 校验规则：
//   1. 【昵称校验规则1】是否含有特殊字符：可通过正则判断
//   2. 【昵称校验规则2】是否长度合法（2-20个字符）
//   3. 【昵称校验规则3】是否含有暴力色情等违规字符：可通过敏感字库过滤
func ValidateNickname(nickname string) error {
	if nickname == "" {
		return nil
	}

	// 【昵称校验规则2】检查长度是否合法（2-20个字符）
	if len(nickname) < 2 || len(nickname) > 20 {
		return fmt.Errorf("昵称长度必须在2-20个字符之间")
	}

	// 【昵称校验规则1】检查是否含有特殊字符
	// 允许：中文、英文、数字、下划线、连字符
	regex := regexp.MustCompile(`^[\u4e00-\u9fa5a-zA-Z0-9_-]+$`)
	if !regex.MatchString(nickname) {
		return fmt.Errorf("昵称只能包含中文、英文、数字、下划线和连字符")
	}

	// 【昵称校验规则3】检查是否含有敏感词
	if containsSensitiveWord(nickname) {
		return fmt.Errorf("昵称包含违规内容，请修改后重试")
	}

	return nil
}

// containsSensitiveWord 检查昵称是否包含敏感词
// 【昵称校验规则3】敏感词库过滤
func containsSensitiveWord(nickname string) bool {
	lowerNickname := strings.ToLower(nickname)
	for _, word := range sensitiveWords {
		if strings.Contains(lowerNickname, strings.ToLower(word)) {
			return true
		}
	}
	return false
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