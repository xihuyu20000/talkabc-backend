package test

import (
	"backend/internal/config"
	"backend/internal/handler"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	InitTest()
}

// ==================== 注册接口测试 ====================

func TestRegister_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/register", handler.Register)

	tests := []struct {
		name string
		body string
	}{
		{"Empty body", "{}"},
		{"Missing phonenum", `{"code": "123456", "password": "Password123"}`},
		{"Missing code", `{"phonenum": "13800138000", "password": "Password123"}`},
		{"Missing password", `{"phonenum": "13800138000", "code": "123456"}`},
		{"Invalid JSON", "invalid json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/v1/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}

// ==================== 注册接口集成测试 ====================

func TestRegister_FormData(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	router := gin.New()
	router.POST("/v1/register", handler.Register)

	formData := strings.NewReader("phonenum=13800138000&code=123456&password=Password123")
	req, _ := http.NewRequest("POST", "/v1/register", formData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest && resp.Code != http.StatusOK {
		t.Logf("Response body: %s", resp.Body.String())
	}
}

// ==================== 重置密码接口测试 ====================

func TestInitiateResetPassword_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/reset-password/initiate", handler.InitiateResetPassword)

	tests := []struct {
		name string
		body string
	}{
		{"Empty body", "{}"},
		{"Missing phonenum", `{"device_id": "test_device"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/v1/reset-password/initiate", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}

func TestValidateResetToken_InvalidParams(t *testing.T) {
	router := gin.New()
	router.GET("/v1/reset-password/validate", handler.ValidateResetToken)

	req, _ := http.NewRequest("GET", "/v1/reset-password/validate", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestCompleteResetPassword_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/reset-password/complete", handler.CompleteResetPassword)

	tests := []struct {
		name string
		body string
	}{
		{"Empty body", "{}"},
		{"Missing token", `{"pwd1": "Password123", "pwd2": "Password123"}`},
		{"Missing pwd1", `{"token": "test_token", "pwd2": "Password123"}`},
		{"Missing pwd2", `{"token": "test_token", "pwd1": "Password123"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/v1/reset-password/complete", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}