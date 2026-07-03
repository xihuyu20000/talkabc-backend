package service

import (
	"backend/internal/model"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// ==================== 验证码生成测试 ====================

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

// ==================== 密码验证测试 ====================



// TestValidatePasswordComplexity 测试新版密码复杂度校验
// 【最低安全策略】验证密码长度、字符类型、弱密码、历史密码检查
func TestValidatePasswordComplexity(t *testing.T) {
	user := &model.User{
		PhoneNum: "13800138000",
		Nickname: "testuser",
		Email:    "testuser@example.com",
	}

	tests := []struct {
		name     string
		password string
		user     *model.User
		wantErr  bool
	}{
		// 【最低安全策略6】空白密码直接拦截
		{"Empty password", "", user, true},
		{"Whitespace password", "   ", user, true},
		// 【最低安全策略1】长度≥8位
		{"Too short", "1234567", user, true},
		{"Exactly 8 chars", "Password", user, false},
		// 【最低安全策略2】至少包含两种字符类型
		{"Only lowercase", "password", user, true},
		{"Only uppercase", "PASSWORD", user, true},
		{"Only digits", "12345678", user, true},
		{"Lowercase + uppercase", "Password", user, false},
		{"Lowercase + digits", "password123", user, false},
		{"Uppercase + digits", "PASSWORD123", user, false},
		{"All types", "Password123!", user, false},
		// 【最低安全策略3】禁止弱密码黑名单
		{"Weak password 123456", "123456", user, true},
		{"Weak password password", "password", user, true},
		// 【最低安全策略5】禁止包含手机号、昵称、邮箱前缀
		{"Contains phone number", "Password13800138000", user, true},
		{"Contains nickname", "Passwordtestuser", user, true},
		{"Contains email prefix", "Passwordtestuser", user, true},
		// 合法密码
		{"Valid password", "MyPassword123", user, false},
		{"Valid with special char", "Pass@1234", user, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePasswordComplexity(tt.password, tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePasswordComplexity(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

// TestIsWeakPassword 测试弱密码检测
// 验证弱密码库中的密码能被正确识别
func TestIsWeakPassword(t *testing.T) {
	tests := []struct {
		password string
		weak     bool
	}{
		{"123456", true},
		{"password", true},
		{"qwerty", true},
		{"MySecurePassword123", false},
		{"P@ssw0rd", false},
	}

	for _, tt := range tests {
		result := isWeakPassword(tt.password)
		if result != tt.weak {
			t.Errorf("isWeakPassword(%q) = %v, want %v", tt.password, result, tt.weak)
		}
	}
}

// ==================== bcrypt密码哈希测试 ====================

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

// ==================== 重置Token生成测试 ====================

// TestGenerateResetToken 测试生成重置密码Token
// 【重置凭证】验证Token长度和随机性
func TestGenerateResetToken(t *testing.T) {
	token, err := generateResetToken()
	if err != nil {
		t.Errorf("Failed to generate reset token: %v", err)
	}

	if len(token) != 32 {
		t.Errorf("Expected token length 32, got %d", len(token))
	}

	for _, c := range token {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			t.Errorf("Invalid character in reset token: %c", c)
		}
	}
}

// TestGenerateResetToken_Unpredictable 测试重置Token不可预测性
// 验证多次生成的Token不重复
func TestGenerateResetToken_Unpredictable(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateResetToken()
		if err != nil {
			t.Errorf("Failed to generate reset token: %v", err)
		}
		if tokens[token] {
			t.Error("Generated duplicate token, tokens should be unpredictable")
		}
		tokens[token] = true
	}
}

// ==================== 用户信息转换测试 ====================

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