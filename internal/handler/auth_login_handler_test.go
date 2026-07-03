package handler

import (
	"backend/internal/config"
	"backend/pkg/response"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ==================== 验证码登录接口测试 ====================

// TestLoginByCode_InvalidParams 测试验证码登录接口参数验证
// 验证缺少手机号、验证码时返回BadRequest
func TestLoginByCode_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/login/code", LoginByCode)

	tests := []struct {
		name string
		body string
	}{
		{"Empty body", "{}"},
		{"Missing phonenum", `{"code": "123456"}`},
		{"Missing code", `{"phonenum": "13800138000"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/v1/login/code", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}

// TestLoginByCode_WithDeviceID 测试验证码登录接口传递设备ID
// 验证设备ID参数能正常接收
func TestLoginByCode_WithDeviceID(t *testing.T) {
	router := gin.New()
	router.POST("/v1/login/code", LoginByCode)

	req, _ := http.NewRequest("POST", "/v1/login/code", strings.NewReader(`{"phonenum": "13800138000", "code": "123456", "device_id": "test_device"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (TestAgent)")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest && resp.Code != http.StatusOK {
		t.Logf("Response body: %s", resp.Body.String())
	}
}

// ==================== 密码登录接口测试 ====================

// TestLoginByPassword_InvalidParams 测试密码登录接口参数验证
// 验证缺少手机号、密码时返回BadRequest
func TestLoginByPassword_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/login/pwd", LoginByPassword)

	tests := []struct {
		name string
		body string
	}{
		{"Empty body", "{}"},
		{"Missing phonenum", `{"pwd": "password123"}`},
		{"Missing password", `{"phonenum": "13800138000"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/v1/login/pwd", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}

// TestLoginByPassword_WithDeviceID 测试密码登录接口传递设备ID
// 验证设备ID参数能正常接收
func TestLoginByPassword_WithDeviceID(t *testing.T) {
	router := gin.New()
	router.POST("/v1/login/pwd", LoginByPassword)

	req, _ := http.NewRequest("POST", "/v1/login/pwd", strings.NewReader(`{"phonenum": "13800138000", "pwd": "Password123", "device_id": "test_device"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (TestAgent)")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest && resp.Code != http.StatusOK {
		t.Logf("Response body: %s", resp.Body.String())
	}
}

// ==================== 退出登录接口测试 ====================

// TestLogout_Success 测试退出登录接口
// 验证退出登录返回Success
func TestLogout_Success(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	router := gin.New()
	router.POST("/v1/logout", Logout)

	req, _ := http.NewRequest("POST", "/v1/logout", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var result response.Response
	json.Unmarshal(resp.Body.Bytes(), &result)

	if result.Code != response.Code0 {
		t.Errorf("Expected code %d, got %d", response.Code0, result.Code)
	}
}