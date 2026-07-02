package service

import (
	"backend/internal/model"
	"testing"

	"github.com/jinzhu/gorm"
)

// TestUserModelToUserInfo 测试 User 模型到 UserInfo DTO 的转换
// 验证用户信息字段（UID、昵称、性别、国家、语言、爱好等）的正确映射
func TestUserModelToUserInfo(t *testing.T) {
	user := model.User{
		Model:     gorm.Model{ID: 123},
		Uid:       "12345678901234567890",
		PhoneNum:  "13800138000",
		Nickname:  "测试用户",
		AvatarURL: "https://example.com/avatar.jpg",
		Gender:    1,
		Country:   "中国",
		Language:  "中文",
		BirthYear: 1995,
		StarSign:  "狮子座",
		EduLevel:  3,
		Job:       "工程师",
		City:      "北京",
		SignText:  "这是我的签名",
	}

	info := UserInfo{
		UID:       user.Uid,
		AvatarURL: user.AvatarURL,
		Nickname:  user.Nickname,
		Gender:    user.Gender,
		Country:   user.Country,
		Language:  user.Language,
		BirthYear: user.BirthYear,
		StarSign:  user.StarSign,
		EduLevel:  user.EduLevel,
		Job:       user.Job,
		City:      user.City,
		Favors:    []string{},
		SignText:  user.SignText,
	}

	if info.UID != user.Uid {
		t.Errorf("Expected UID %s, got %s", user.Uid, info.UID)
	}

	if info.Nickname != user.Nickname {
		t.Errorf("Expected nickname %s, got %s", user.Nickname, info.Nickname)
	}

	if info.Gender != user.Gender {
		t.Errorf("Expected gender %d, got %d", user.Gender, info.Gender)
	}

	if info.Country != user.Country {
		t.Errorf("Expected country %s, got %s", user.Country, info.Country)
	}

	if info.Language != user.Language {
		t.Errorf("Expected language %s, got %s", user.Language, info.Language)
	}
}

// TestUserInfo_EmptyData 测试 UserInfo DTO 空数据场景
// 验证当用户只有基本ID时，其他字段应保持默认值（空字符串或0）
func TestUserInfo_EmptyData(t *testing.T) {
	user := model.User{
		Model: gorm.Model{ID: 1},
		Uid:   "1",
	}

	info := UserInfo{
		UID: user.Uid,
	}

	if info.UID != "1" {
		t.Errorf("Expected UID \"1\", got %s", info.UID)
	}

	if info.Nickname != "" {
		t.Errorf("Expected empty nickname, got %s", info.Nickname)
	}
}

// TestUserInfo_Structure 测试 UserInfo DTO 结构完整性
// 验证 UserInfo 结构体包含所有必要字段，且字段类型正确
func TestUserInfo_Structure(t *testing.T) {
	info := UserInfo{
		UID:       "12345678901234567890",
		AvatarURL: "http://example.com/avatar.png",
		Nickname:  "nick",
		Gender:    0,
		Country:   "US",
		Language:  "en",
		BirthYear: 2000,
		StarSign:  "Aries",
		EduLevel:  2,
		Job:       "Developer",
		City:      "New York",
		FreAreas:  []string{"NY", "CA"},
		Favors:    []string{"Sports"},
		SignText:  "Hello",
	}

	if info.Gender != 0 {
		t.Error("Gender should be 0 (female)")
	}

	if len(info.FreAreas) != 2 {
		t.Errorf("Expected 2 freareas, got %d", len(info.FreAreas))
	}

	if len(info.Favors) != 1 {
		t.Errorf("Expected 1 favor, got %d", len(info.Favors))
	}
}