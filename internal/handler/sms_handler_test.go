package handler

import (
	"backend/internal/config"
	"backend/internal/infra"
	"backend/internal/repository"
	"backend/pkg/logger"
	"backend/pkg/response"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// ==================== 测试初始化 ====================

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

// ==================== 发送短信验证码测试 ====================

// TestSendSMSCode_MissingPhoneNum 测试发送短信验证码接口-缺少手机号参数
// 测试场景：用户请求发送短信验证码时未传递手机号参数
// 验证点：
//   1. HTTP状态码应为400 Bad Request
//   2. 响应体中的code字段应为400
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

// TestSendSMSCode_SuccessRedisRecord 测试发送短信验证码接口-成功发送并验证Redis记录
// 测试场景：用户请求发送短信验证码，接口成功生成并存储验证码
// 验证点：
//   1. HTTP状态码应为200 OK
//   2. 响应体中的code字段应为0（成功）
//   3. Redis中应存储了对应的验证码
//   4. 验证码长度应为6位数字
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

// ==================== 生成图形验证码测试 ====================

// TestGetAlnumCode_MissingPhoneNum 测试生成图形验证码接口-缺少手机号参数
// 测试场景：用户请求生成图形验证码时未传递手机号参数（旧接口逻辑）
// 验证点：
//   1. HTTP状态码应为400 Bad Request
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

// TestGenerateAlnumCode_Success 测试生成图形验证码接口-成功生成
// 测试场景：用户请求生成图形验证码，接口成功返回验证码图片和ID
// 验证点：
//   1. HTTP状态码应为200 OK
//   2. 响应体中的code字段应为0（成功）
//   3. 响应数据中应包含captcha_id字段
//   4. 响应数据中应包含image字段（base64编码的图片）
func TestGenerateAlnumCode_Success(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/alnum", GenerateAlnumCode)

	req, _ := http.NewRequest("GET", "/v1/code/alnum", nil)
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

	data := result.Data.(map[string]interface{})
	if _, ok := data["captcha_id"]; !ok {
		t.Error("Expected captcha_id in response data")
	}
	if _, ok := data["image"]; !ok {
		t.Error("Expected image in response data")
	}
}

// ==================== 验证图形验证码测试 ====================

// TestVerifyAlnumCode_MissingParams 测试验证图形验证码接口-缺少参数
// 测试场景：用户请求验证图形验证码时缺少必要参数
// 测试用例：
//   1. 空请求体
//   2. 缺少captcha_id参数
//   3. 缺少code参数
// 验证点：
//   1. 所有场景HTTP状态码应为400 Bad Request
func TestVerifyAlnumCode_MissingParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/code/alnum/verify", VerifyAlnumCode)

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

// TestVerifyAlnumCode_InvalidCode 测试验证图形验证码接口-无效验证码
// 测试场景：用户使用不存在的captcha_id进行验证
// 验证点：
//   1. HTTP状态码应为400 Bad Request
func TestVerifyAlnumCode_InvalidCode(t *testing.T) {
	router := gin.New()
	router.POST("/v1/code/alnum/verify", VerifyAlnumCode)

	body := `{"captcha_id": "nonexistent_id", "code": "1234"}`
	req, _ := http.NewRequest("POST", "/v1/code/alnum/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

// TestVerifyAlnumCode_Success 测试验证图形验证码接口-验证成功
// 测试场景：完整的图形验证码流程，先生成验证码再进行验证
// 验证步骤：
//   1. 调用生成验证码接口获取captcha_id和图片
//   2. 从Redis中读取存储的验证码答案
//   3. 调用验证接口，传入正确的captcha_id和验证码
// 验证点：
//   1. 生成接口返回200 OK
//   2. 验证接口返回200 OK
func TestVerifyAlnumCode_Success(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/alnum", GenerateAlnumCode)
	router.POST("/v1/code/alnum/verify", VerifyAlnumCode)

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

// ==================== 验证短信验证码测试 ====================

// TestVerifySMSCode_MissingParams 测试验证短信验证码接口-缺少参数
// 测试场景：用户请求验证短信验证码时缺少必要参数
// 测试用例：
//   1. 空请求体
//   2. 缺少phonenum参数
//   3. 缺少code参数
// 验证点：
//   1. 所有场景HTTP状态码应为400 Bad Request
func TestVerifySMSCode_MissingParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/code/sms/verify", VerifySMSCode)

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

// TestVerifySMSCode_InvalidCode 测试验证短信验证码接口-无效验证码
// 测试场景：用户使用未发送过验证码的手机号进行验证
// 验证点：
//   1. HTTP状态码应为400 Bad Request
func TestVerifySMSCode_InvalidCode(t *testing.T) {
	router := gin.New()
	router.POST("/v1/code/sms/verify", VerifySMSCode)

	body := `{"phonenum": "13800138000", "code": "000000"}`
	req, _ := http.NewRequest("POST", "/v1/code/sms/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

// TestVerifySMSCode_Success 测试验证短信验证码接口-验证成功
// 测试场景：完整的短信验证码流程，先发送验证码再进行验证
// 验证步骤：
//   1. 调用发送验证码接口发送短信验证码
//   2. 从Redis中读取存储的验证码
//   3. 调用验证接口，传入正确的手机号和验证码
// 验证点：
//   1. 发送接口返回200 OK
//   2. 验证接口返回200 OK
func TestVerifySMSCode_Success(t *testing.T) {
	router := gin.New()
	router.GET("/v1/code/sms", SendSMSCode)
	router.POST("/v1/code/sms/verify", VerifySMSCode)

	phoneNum := "13800138001"
	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/v1/code/sms?phonenum=%s", phoneNum), nil)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	if getResp.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, getResp.Code)
	}

	key := fmt.Sprintf("verification_code:%s:%s", phoneNum, repository.VerificationCodeTypeSMS)
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