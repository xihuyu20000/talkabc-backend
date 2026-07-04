package test

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/pkg/response"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	InitTest()
}

// ==================== 发送短信验证码测试 ====================

func TestSendSMSCode_MissingPhoneNum(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)

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

// ==================== 生成图形验证码测试 ====================



func TestGenerateAlnumCode_Success(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/alnum", handler.GenerateAlnumCode)

	req, _ := http.NewRequest("GET", "/v1/code/alnum", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		var result response.Response
		json.Unmarshal(resp.Body.Bytes(), &result)
		t.Errorf("Expected status %d, got %d. Error: %s", http.StatusOK, resp.Code, result.Msg)
		return
	}

	var result response.Response
	json.Unmarshal(resp.Body.Bytes(), &result)

	if result.Code != response.Code0 {
		t.Errorf("Expected code %d, got %d. Error: %s", response.Code0, result.Code, result.Msg)
	}

	data := result.Data.(map[string]interface{})
	if _, ok := data["captcha_id"]; !ok {
		t.Error("Expected captcha_id in response data")
	}
	if _, ok := data["image"]; !ok {
		t.Error("Expected image in response data")
	}
}

// ==================== 验证图形验证码测试 ====================

func TestVerifyAlnumCode_MissingParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/code/alnum/verify", handler.VerifyAlnumCode)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing captcha_id", `{"code": "1234"}`},
		{"missing code", `{"captcha_id": "test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/v1/code/alnum/verify", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}

func TestVerifyAlnumCode_InvalidCode(t *testing.T) {
	router := gin.New()
	router.POST("/v1/code/alnum/verify", handler.VerifyAlnumCode)

	body := `{"captcha_id": "nonexistent_id", "code": "1234"}`
	req, _ := http.NewRequest("POST", "/v1/code/alnum/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

// ==================== 验证短信验证码测试 ====================

func TestVerifySMSCode_MissingParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/code/sms/verify", handler.VerifySMSCode)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing phonenum", `{"code": "123456"}`},
		{"missing code", `{"phonenum": "13800138000"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/v1/code/sms/verify", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}

func TestVerifySMSCode_InvalidCode(t *testing.T) {
	router := gin.New()
	router.POST("/v1/code/sms/verify", handler.VerifySMSCode)

	body := `{"phonenum": "13800138000", "code": "000000"}`
	req, _ := http.NewRequest("POST", "/v1/code/sms/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

// ==================== 短信验证码集成测试 ====================

func TestSendSMSCode_SuccessRedisRecord(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)

	phoneNum := "13800138000"
	
	config.RDB.FlushDB(config.RDB.Context())
	
	req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/code/sms?phonenum=%s", phoneNum), nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		var result response.Response
		json.Unmarshal(resp.Body.Bytes(), &result)
		t.Errorf("Expected status %d, got %d. Error: %s", http.StatusOK, resp.Code, result.Msg)
		return
	}

	var result response.Response
	json.Unmarshal(resp.Body.Bytes(), &result)

	if result.Code != response.Code0 {
		t.Errorf("Expected code %d, got %d. Error: %s", response.Code0, result.Code, result.Msg)
	}

	key := fmt.Sprintf("verification_code:%s:%s:%s", phoneNum, repository.VerificationCodeTypeSMS, "default")
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

func TestVerifyAlnumCode_Success(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/alnum", handler.GenerateAlnumCode)
	router.POST("/v1/code/alnum/verify", handler.VerifyAlnumCode)

	getReq, _ := http.NewRequest("GET", "/v1/code/alnum", nil)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	if getResp.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, getResp.Code)
	}

	var getResult response.Response
	json.Unmarshal(getResp.Body.Bytes(), &getResult)

	data := getResult.Data.(map[string]interface{})
	captchaID := data["captcha_id"].(string)

	storedCode, err := config.RDB.Get(config.RDB.Context(), captchaID).Result()
	if err != nil {
		t.Fatalf("Failed to get captcha code from Redis: %v", err)
	}

	body := fmt.Sprintf(`{"captcha_id": "%s", "code": "%s"}`, captchaID, storedCode)
	verifyReq, _ := http.NewRequest("POST", "/v1/code/alnum/verify", bytes.NewBufferString(body))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyResp := httptest.NewRecorder()

	router.ServeHTTP(verifyResp, verifyReq)

	if verifyResp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, verifyResp.Code)
	}
}

func TestVerifySMSCode_Success(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/code/sms/verify", handler.VerifySMSCode)

	phoneNum := "13800138001"
	
	config.RDB.FlushDB(config.RDB.Context())
	
	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/v1/code/sms?phonenum=%s", phoneNum), nil)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	if getResp.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, getResp.Code)
	}

	key := fmt.Sprintf("verification_code:%s:%s:%s", phoneNum, repository.VerificationCodeTypeSMS, "default")
	storedCode, err := config.RDB.Get(config.RDB.Context(), key).Result()
	if err != nil {
		t.Fatalf("Failed to get verification code from Redis: %v", err)
	}

	body := fmt.Sprintf(`{"phonenum": "%s", "code": "%s"}`, phoneNum, storedCode)
	verifyReq, _ := http.NewRequest("POST", "/v1/code/sms/verify", bytes.NewBufferString(body))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyResp := httptest.NewRecorder()

	router.ServeHTTP(verifyResp, verifyReq)

	if verifyResp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, verifyResp.Code)
	}
}

