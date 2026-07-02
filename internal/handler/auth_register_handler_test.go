package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRegister_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/register", Register)

	tests := []struct {
		name string
		body string
	}{
		{"Empty body", "{}"},
		{"Missing phonenum", `{"code": "123456"}`},
		{"Missing code", `{"phonenum": "13800138000"}`},
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

func TestResetPassword_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/resetpwd", ResetPassword)

	tests := []struct {
		name string
		body string
	}{
		{"Empty body", "{}"},
		{"Missing fields", `{"phonenum": "13800138000"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/v1/resetpwd", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}
