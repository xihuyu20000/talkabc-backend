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
