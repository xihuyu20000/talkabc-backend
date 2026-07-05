package test

import (
	"backend/pkg/security"
	"testing"
)

// ==================== 昵称安全检查测试 ====================

// TestValidateNickname_Valid 测试合法昵称
func TestValidateNickname_Valid(t *testing.T) {
	testCases := []string{
		"张三",
		"zhangsan",
		"zhang_san",
		"zhang-san",
		"zhang123",
		"张三123",
		"ZS",
		"test_user_123",
	}

	for _, nickname := range testCases {
		t.Run(nickname, func(t *testing.T) {
			err := security.ValidateNickname(nickname)
			if err != nil {
				t.Errorf("合法昵称 '%s' 验证失败: %v", nickname, err)
			}
		})
	}
}

// TestValidateNickname_InvalidLength 测试昵称长度不合法
func TestValidateNickname_InvalidLength(t *testing.T) {
	testCases := []struct {
		nickname string
		reason   string
	}{
		{"a", "长度不足2个字符"},
		{"abcdefghijklmnopqrstu", "长度超过20个字符"},
	}

	for _, tc := range testCases {
		t.Run(tc.reason, func(t *testing.T) {
			err := security.ValidateNickname(tc.nickname)
			if err == nil {
				t.Errorf("昵称 '%s' 应该验证失败（%s），但未报错", tc.nickname, tc.reason)
			}
		})
	}
}

// TestValidateNickname_SpecialCharacters 测试昵称包含特殊字符
func TestValidateNickname_SpecialCharacters(t *testing.T) {
	testCases := []string{
		"张三@",
		"zhangsan#",
		"zhang$san",
		"test*user",
		"hello&world",
		"你好！",
		"测试+用户",
	}

	for _, nickname := range testCases {
		t.Run(nickname, func(t *testing.T) {
			err := security.ValidateNickname(nickname)
			if err == nil {
				t.Errorf("昵称 '%s' 包含特殊字符，应该验证失败", nickname)
			}
		})
	}
}

// TestValidateNickname_SensitiveWords 测试昵称包含敏感词
func TestValidateNickname_SensitiveWords(t *testing.T) {
	testCases := []string{
		"暴力王",
		"色情狂",
		"赌博达人",
		"吸毒哥",
		"傻逼",
		"操你妈",
		"去死吧",
		"黑客攻击",
		"诈骗专家",
		"邪教徒",
	}

	for _, nickname := range testCases {
		t.Run(nickname, func(t *testing.T) {
			err := security.ValidateNickname(nickname)
			if err == nil {
				t.Errorf("昵称 '%s' 包含敏感词，应该验证失败", nickname)
			}
		})
	}
}

// TestValidateNickname_URL 测试昵称包含URL
func TestValidateNickname_URL(t *testing.T) {
	testCases := []string{
		"http://example.com",
		"https://www.google.com",
		"www.baidu.com",
		"test www.example.com",
		"myblog http://test.com",
	}

	for _, nickname := range testCases {
		t.Run(nickname, func(t *testing.T) {
			err := security.ValidateNickname(nickname)
			if err == nil {
				t.Errorf("昵称 '%s' 包含URL，应该验证失败", nickname)
			}
		})
	}
}

// TestValidateNickname_HTML 测试昵称包含HTML标签
func TestValidateNickname_HTML(t *testing.T) {
	testCases := []string{
		"<script>alert(1)</script>",
		"<b>test</b>",
		"<img src=x onerror=alert(1)>",
		"<div>hello</div>",
	}

	for _, nickname := range testCases {
		t.Run(nickname, func(t *testing.T) {
			err := security.ValidateNickname(nickname)
			if err == nil {
				t.Errorf("昵称 '%s' 包含HTML标签，应该验证失败", nickname)
			}
		})
	}
}

// TestValidateNickname_XSS 测试昵称包含XSS攻击代码
func TestValidateNickname_XSS(t *testing.T) {
	testCases := []string{
		"javascript:alert(1)",
		"onclick=alert(1)",
		"onerror=alert(1)",
		"eval('alert(1)')",
	}

	for _, nickname := range testCases {
		t.Run(nickname, func(t *testing.T) {
			err := security.ValidateNickname(nickname)
			if err == nil {
				t.Errorf("昵称 '%s' 包含XSS代码，应该验证失败", nickname)
			}
		})
	}
}

// ==================== 签名安全检查测试 ====================

// TestValidateSignText_Valid 测试合法签名
func TestValidateSignText_Valid(t *testing.T) {
	testCases := []string{
		"",
		"这是我的个性签名",
		"Hello World!",
		"喜欢音乐、电影和旅行",
		"生活不止眼前的苟且，还有诗和远方",
	}

	for _, sign := range testCases {
		t.Run(sign, func(t *testing.T) {
			err := security.ValidateSignText(sign)
			if err != nil {
				t.Errorf("合法签名 '%s' 验证失败: %v", sign, err)
			}
		})
	}
}

// TestValidateSignText_TooLong 测试签名过长
func TestValidateSignText_TooLong(t *testing.T) {
	longSign := ""
	for i := 0; i < 201; i++ {
		longSign += "a"
	}

	err := security.ValidateSignText(longSign)
	if err == nil {
		t.Errorf("签名长度超过200字符，应该验证失败")
	}
}

// TestValidateSignText_SensitiveWords 测试签名包含敏感词
func TestValidateSignText_SensitiveWords(t *testing.T) {
	testCases := []string{
		"我是暴力分子",
		"欢迎访问色情网站",
		"赌博赢钱技巧",
		"吸毒有害健康",
		"傻逼滚蛋",
		"操你妈",
		"黑客入侵教程",
		"诈骗电话",
	}

	for _, sign := range testCases {
		t.Run(sign, func(t *testing.T) {
			err := security.ValidateSignText(sign)
			if err == nil {
				t.Errorf("签名 '%s' 包含敏感词，应该验证失败", sign)
			}
		})
	}
}

// TestValidateSignText_URL 测试签名包含URL
func TestValidateSignText_URL(t *testing.T) {
	testCases := []string{
		"我的网站：http://example.com",
		"访问 https://www.google.com",
		"更多信息请访问 www.baidu.com",
		"点击 http://test.com 查看详情",
	}

	for _, sign := range testCases {
		t.Run(sign, func(t *testing.T) {
			err := security.ValidateSignText(sign)
			if err == nil {
				t.Errorf("签名 '%s' 包含URL，应该验证失败", sign)
			}
		})
	}
}

// TestValidateSignText_HTML 测试签名包含HTML标签
func TestValidateSignText_HTML(t *testing.T) {
	testCases := []string{
		"<b>加粗文字</b>",
		"<div>内容</div>",
		"<span style='color:red'>红色</span>",
		"<p>段落</p>",
	}

	for _, sign := range testCases {
		t.Run(sign, func(t *testing.T) {
			err := security.ValidateSignText(sign)
			if err == nil {
				t.Errorf("签名 '%s' 包含HTML标签，应该验证失败", sign)
			}
		})
	}
}

// TestValidateSignText_Script 测试签名包含JavaScript代码
func TestValidateSignText_Script(t *testing.T) {
	testCases := []string{
		"<script>alert(1)</script>",
		"javascript:alert(1)",
		"eval('alert(1)')",
		"alert('hello')",
		"confirm('test')",
		"prompt('input')",
	}

	for _, sign := range testCases {
		t.Run(sign, func(t *testing.T) {
			err := security.ValidateSignText(sign)
			if err == nil {
				t.Errorf("签名 '%s' 包含JavaScript代码，应该验证失败", sign)
			}
		})
	}
}

// TestValidateSignText_SQLInjection 测试签名包含SQL注入代码
func TestValidateSignText_SQLInjection(t *testing.T) {
	testCases := []string{
		"SELECT * FROM users",
		"INSERT INTO users VALUES (1)",
		"UPDATE users SET name='test'",
		"DELETE FROM users",
		"DROP TABLE users",
		"1 OR 1=1",
		"-- comment",
		"/* injection */",
	}

	for _, sign := range testCases {
		t.Run(sign, func(t *testing.T) {
			err := security.ValidateSignText(sign)
			if err == nil {
				t.Errorf("签名 '%s' 包含SQL注入代码，应该验证失败", sign)
			}
		})
	}
}

// TestValidateSignText_XSS 测试签名包含XSS攻击代码
func TestValidateSignText_XSS(t *testing.T) {
	testCases := []string{
		"<img src=x onerror=alert(1)>",
		"<input onfocus=alert(1)>",
		"<a href=javascript:alert(1)>click</a>",
		"onclick=alert(1)",
		"onload=alert(1)",
	}

	for _, sign := range testCases {
		t.Run(sign, func(t *testing.T) {
			err := security.ValidateSignText(sign)
			if err == nil {
				t.Errorf("签名 '%s' 包含XSS代码，应该验证失败", sign)
			}
		})
	}
}

// ==================== 单个安全检查函数测试 ====================

// TestContainsSensitiveWord 测试敏感词检测
func TestContainsSensitiveWord(t *testing.T) {
	tests := []struct {
		content string
		want    bool
	}{
		{"暴力游戏", true},
		{"色情电影", true},
		{"正常内容", false},
		{"hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			got, _ := security.ContainsSensitiveWord(tt.content)
			if got != tt.want {
				t.Errorf("ContainsSensitiveWord(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}

// TestContainsURL 测试URL检测
func TestContainsURL(t *testing.T) {
	tests := []struct {
		content string
		want    bool
	}{
		{"http://example.com", true},
		{"https://google.com", true},
		{"www.baidu.com", true},
		{"normal text", false},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			if got := security.ContainsURL(tt.content); got != tt.want {
				t.Errorf("ContainsURL(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}

// TestContainsHTML 测试HTML检测
func TestContainsHTML(t *testing.T) {
	tests := []struct {
		content string
		want    bool
	}{
		{"<div>test</div>", true},
		{"<script></script>", true},
		{"normal text", false},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			if got := security.ContainsHTML(tt.content); got != tt.want {
				t.Errorf("ContainsHTML(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}