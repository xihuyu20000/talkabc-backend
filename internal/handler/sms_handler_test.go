package handler

import (
	"backend/internal/config"
	"backend/internal/infra"
	"backend/internal/repository"
	"backend/pkg/logger"
	"backend/pkg/response"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
	logger.InitLogger("info")
	initTestRedis()
}

func initTestRedis() {
	if config.RDB == nil {
		config.RDB = infra.NewRedis(infra.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		})
	}
}

func TestSendSMSCode_MissingPhoneNum(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/sms", SendSMSCode)

	req, _ := http.NewRequest("GET", "/v1/code/sms", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}

	var result response.Response
	json.Unmarshal(resp.Body.Bytes(), &result)

	if result.Code != response.Code400 {
		t.Errorf("Expected code %d, got %d", response.Code400, result.Code)
	}
}

func TestGetAlnumCode_MissingPhoneNum(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/alnum", GenerateAlnumCode)

	req, _ := http.NewRequest("GET", "/v1/code/alnum", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestSendSMSCode_SuccessRedisRecord(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/sms", SendSMSCode)

	phoneNum := "13800138000"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/code/sms?phonenum=%s", phoneNum), nil)
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

	key := fmt.Sprintf("verification_code:%s:%s", phoneNum, repository.VerificationCodeTypeSMS)
	code, err := config.RDB.Get(config.RDB.Context(), key).Result()
	if err != nil {
		t.Errorf("Failed to get verification code from Redis: %v", err)
	}

	if code == "" {
		t.Error("Verification code should not be empty in Redis")
	}

	if len(code) != 6 {
		t.Errorf("Expected verification code length 6, got %d", len(code))
	}
}
