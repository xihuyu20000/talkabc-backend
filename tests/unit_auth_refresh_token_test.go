package test

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	InitTest()
}

// ==================== 刷新令牌单元测试 ====================

// TestGenerateRefreshToken 测试刷新令牌生成
// 【刷新令牌安全规则】验证刷新令牌生成格式是否正确
func TestGenerateRefreshToken(t *testing.T) {
	uid := "test_uid_12345"
	
	token, err := middleware.GenerateRefreshToken(uid)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}
	
	if token == "" {
		t.Fatal("GenerateRefreshToken returned empty token")
	}
	
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		t.Fatalf("Refresh token should have 2 parts separated by colon, got %d", len(parts))
	}
	
	if len(parts[0]) != 32 {
		t.Errorf("Random part should be 32 characters, got %d", len(parts[0]))
	}
	
	if len(parts[1]) == 0 {
		t.Fatal("JWT part should not be empty")
	}
}

// TestParseRefreshToken_ValidToken 测试解析有效的刷新令牌
func TestParseRefreshToken_ValidToken(t *testing.T) {
	uid := "test_uid_67890"
	
	token, err := middleware.GenerateRefreshToken(uid)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}
	
	parsedUID, err := middleware.ParseRefreshToken(token)
	if err != nil {
		t.Fatalf("ParseRefreshToken failed: %v", err)
	}
	
	if parsedUID != uid {
		t.Errorf("Expected UID %s, got %s", uid, parsedUID)
	}
}

// TestParseRefreshToken_InvalidFormat 测试解析格式错误的刷新令牌
func TestParseRefreshToken_InvalidFormat(t *testing.T) {
	tests := []struct {
		name string
		token string
	}{
		{"Empty token", ""},
		{"No colon", "abc123"},
		{"Missing JWT", "random_part:"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := middleware.ParseRefreshToken(tt.token)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
		})
	}
}

// TestParseRefreshToken_InvalidSignature 测试解析签名无效的刷新令牌
func TestParseRefreshToken_InvalidSignature(t *testing.T) {
	invalidToken := "random:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJ0ZXN0X3VpZF8xMjM0NSJ9.invalid_signature"
	
	_, err := middleware.ParseRefreshToken(invalidToken)
	if err == nil {
		t.Error("Expected error for invalid signature token, got nil")
	}
}

// ==================== 刷新令牌接口测试 ====================

// TestRefreshToken_EmptyToken 测试空刷新令牌
func TestRefreshToken_EmptyToken(t *testing.T) {
	router := gin.New()
	router.POST("/v1/refresh-token", handler.RefreshToken)
	
	req, _ := http.NewRequest("POST", "/v1/refresh-token", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

// TestRefreshToken_UnitInvalidToken 测试无效刷新令牌（单元测试）
func TestRefreshToken_UnitInvalidToken(t *testing.T) {
	if config.RDB == nil {
		t.Skip("Redis not initialized, skipping test")
	}
	
	router := gin.New()
	router.POST("/v1/refresh-token", handler.RefreshToken)
	
	config.RDB.FlushDB(config.RDB.Context())
	
	req, _ := http.NewRequest("POST", "/v1/refresh-token", strings.NewReader("refresh_token=invalid_token"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if result["code"] != float64(1) {
		t.Errorf("Expected code 1, got %v", result["code"])
	}
}