package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestGenerateRandomCode 测试生成随机数字验证码
// 验证生成的验证码长度是否正确（6位）
func TestGenerateRandomCode(t *testing.T) {
	code := generateRandomCode(6)
	if len(code) != 6 {
		t.Errorf("Expected code length 6, got %d", len(code))
	}
}

// TestGenerateRandomAlnum 测试生成随机字母数字验证码
// 验证生成的验证码长度是否正确（4位），且只包含数字和大写字母
func TestGenerateRandomAlnum(t *testing.T) {
	code := generateRandomAlnum(4)
	if len(code) != 4 {
		t.Errorf("Expected code length 4, got %d", len(code))
	}

	for _, c := range code {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')) {
			t.Errorf("Invalid character in alphanumeric code: %c", c)
		}
	}
}

// TestValidatePassword 测试密码验证逻辑
// 验证不同密码格式（长度、特殊字符等）的有效性判断是否正确
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"123456", true},
		{"abcdef", true},
		{"Abcdef123", true},
		{"abc", false},
		{"", false},
		{"abc def", false},
		{"abcdefghi1234567890", true},
	}

	for _, tt := range tests {
		result := validatePassword(tt.password)
		if result != tt.valid {
			t.Errorf("validatePassword(%q) = %v, want %v", tt.password, result, tt.valid)
		}
	}
}

// TestBcryptHash 测试 bcrypt 密码哈希功能
// 验证密码哈希生成、验证以及错误密码拒绝的正确性
func TestBcryptHash(t *testing.T) {
	password := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("Failed to generate hash: %v", err)
	}

	if len(hash) == 0 {
		t.Error("Hash should not be empty")
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		t.Error("Password comparison failed")
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte("wrongpassword"))
	if err == nil {
		t.Error("Wrong password should fail comparison")
	}
}

// TestConvertUserToUserInfo 测试用户信息转换
// 验证 UserInfo DTO 的字段赋值和读取是否正确
func TestConvertUserToUserInfo(t *testing.T) {
	user := UserInfo{
		UID:       "12345678901234567890",
		Nickname:  "testuser",
		Gender:    1,
		Country:   "China",
		Language:  "zh",
		BirthYear: 1990,
	}

	if user.UID != "12345678901234567890" {
		t.Errorf("Expected UID '12345678901234567890', got %s", user.UID)
	}

	if user.Nickname != "testuser" {
		t.Errorf("Expected nickname 'testuser', got %s", user.Nickname)
	}
}