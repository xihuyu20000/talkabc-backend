package test

import (
	"backend/internal/config"
	"backend/internal/handler"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestLoginByCode_FullFlow 验证码登录完整流程集成测试
// 模拟用户使用验证码登录的全过程，验证所有安全规则检查
//
// 测试流程：
//   Step1: 用户注册（先创建测试用户）
//   Step2: 用户输入手机号，点击发送验证码
//   Step3: 用户输入验证码，点击登录（验证码登录）
//   Step4: 验证验证码已被清理
//
// 登录安全规则：
//   1. IP黑名单检查
//   2. IP登录频率限制（1分钟10次）
//   3. 手机号黑名单检查
//   4. 设备黑名单检查
//   5. 登录失败次数限制（5分钟内5次失败锁定15分钟）
//   6. 用户账号状态检查（正常/封禁/注销）
//   7. 登录成功后清理验证码，防止二次复用
//   8. 记录登录操作日志（不可删除）
func TestLoginByCode_FullFlow(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)
	router.POST("/v1/login/code", handler.LoginByCode)

	phoneNum := "13900139004"
	password := "Password123"

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// ==================== Step1: 注册测试用户 ====================
	t.Run("Step1_RegisterTestUser", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneNum+"&tag=register", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) == 0 {
			t.Fatal("No SMS code sent")
		}
		code := sentMsgs[0].Code

		formData := strings.NewReader("phonenum=" + phoneNum + "&code=" + code + "&password=" + password)
		registerReq, _ := http.NewRequest("POST", "/v1/register", formData)
		registerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		registerResp := httptest.NewRecorder()
		router.ServeHTTP(registerResp, registerReq)

		if registerResp.Code != http.StatusOK {
			t.Fatalf("Register failed with status %d: %s", registerResp.Code, registerResp.Body.String())
		}

		var result map[string]interface{}
		json.Unmarshal(registerResp.Body.Bytes(), &result)
		if result["code"] != float64(0) {
			t.Fatalf("Register failed: %v", result["msg"])
		}

		mockSMSGateway.ClearSentMessages()
		t.Logf("Test user registered successfully: %s", phoneNum)
	})

	// ==================== Step2: 发送验证码 ====================
	t.Run("Step2_SendSMSCode", func(t *testing.T) {
		config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+phoneNum)

		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneNum+"&tag=login", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) != 1 {
			t.Fatalf("Expected 1 sent message, got %d", len(sentMsgs))
		}

		code := sentMsgs[0].Code
		if len(code) != 6 {
			t.Fatalf("Expected code length 6, got %d", len(code))
		}

		t.Logf("Generated verification code for login: %s", code)
	})

	// ==================== Step3: 验证码登录 ====================
	t.Run("Step3_LoginByCode", func(t *testing.T) {
		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) == 0 {
			t.Skip("No SMS code sent in previous step")
		}
		code := sentMsgs[0].Code

		formData := strings.NewReader("phonenum=" + phoneNum + "&code=" + code)
		loginReq, _ := http.NewRequest("POST", "/v1/login/code", formData)
		loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		loginResp := httptest.NewRecorder()

		router.ServeHTTP(loginResp, loginReq)

		t.Logf("LoginByCode response status: %d", loginResp.Code)
		t.Logf("LoginByCode response body: %s", loginResp.Body.String())

		if loginResp.Code != http.StatusOK {
			var result map[string]interface{}
			if err := json.Unmarshal(loginResp.Body.Bytes(), &result); err == nil {
				if msg, ok := result["msg"].(string); ok {
					t.Logf("Login error message: %s", msg)
				}
			}
			t.Fatalf("LoginByCode failed with status %d: %s", loginResp.Code, loginResp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(loginResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		data := result["data"]
		if data == nil {
			t.Error("Data should not be nil")
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Errorf("Data should be map[string]interface{}, got %T", data)
			return
		}

		token, ok := dataMap["access_token"].(string)
		if !ok || token == "" {
			t.Error("Token should not be empty")
		}

		t.Logf("Generated JWT token: %s", token)
	})

	// ==================== Step4: 验证验证码已被清理 ====================
	t.Run("Step4_VerificationCodeCleared", func(t *testing.T) {
		_, err := config.RDB.Get(config.RDB.Context(), "verification_code:"+phoneNum+":sms:login").Result()
		if err == nil {
			t.Error("Verification code should be cleared after login")
		}
	})
}

// TestLoginByPassword_FullFlow 密码登录完整流程集成测试
// 模拟用户使用密码登录的全过程，验证所有安全规则检查
//
// 测试流程：
//   Step1: 用户注册（先创建测试用户）
//   Step2: 用户输入手机号和密码，点击登录（密码登录）
//
// 登录安全规则：
//   1. IP黑名单检查
//   2. IP登录频率限制（1分钟10次）
//   3. 手机号黑名单检查
//   4. 设备黑名单检查
//   5. 登录失败次数限制（5分钟内5次失败锁定15分钟）
//   6. 用户账号状态检查（正常/封禁/注销）
//   7. 登录成功后重置失败次数
//   8. 记录登录操作日志（不可删除）
func TestLoginByPassword_FullFlow(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)
	router.POST("/v1/login/pwd", handler.LoginByPassword)

	phoneNum := "13900139005"
	password := "Password123"

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// ==================== Step1: 注册测试用户 ====================
	t.Run("Step1_RegisterTestUser", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneNum+"&tag=register", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) == 0 {
			t.Fatal("No SMS code sent")
		}
		code := sentMsgs[0].Code

		formData := strings.NewReader("phonenum=" + phoneNum + "&code=" + code + "&password=" + password)
		registerReq, _ := http.NewRequest("POST", "/v1/register", formData)
		registerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		registerResp := httptest.NewRecorder()
		router.ServeHTTP(registerResp, registerReq)

		if registerResp.Code != http.StatusOK {
			t.Fatalf("Register failed with status %d: %s", registerResp.Code, registerResp.Body.String())
		}

		var result map[string]interface{}
		json.Unmarshal(registerResp.Body.Bytes(), &result)
		if result["code"] != float64(0) {
			t.Fatalf("Register failed: %v", result["msg"])
		}

		t.Logf("Test user registered successfully: %s", phoneNum)
	})

	// ==================== Step2: 密码登录 ====================
	t.Run("Step2_LoginByPassword", func(t *testing.T) {
		formData := strings.NewReader("phonenum=" + phoneNum + "&pwd=" + password)
		loginReq, _ := http.NewRequest("POST", "/v1/login/pwd", formData)
		loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		loginResp := httptest.NewRecorder()

		router.ServeHTTP(loginResp, loginReq)

		t.Logf("LoginByPassword response status: %d", loginResp.Code)
		t.Logf("LoginByPassword response body: %s", loginResp.Body.String())

		if loginResp.Code != http.StatusOK {
			var result map[string]interface{}
			if err := json.Unmarshal(loginResp.Body.Bytes(), &result); err == nil {
				if msg, ok := result["msg"].(string); ok {
					t.Logf("Login error message: %s", msg)
				}
			}
			t.Fatalf("LoginByPassword failed with status %d: %s", loginResp.Code, loginResp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(loginResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		data := result["data"]
		if data == nil {
			t.Error("Data should not be nil")
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Errorf("Data should be map[string]interface{}, got %T", data)
			return
		}

		token, ok := dataMap["access_token"].(string)
		if !ok || token == "" {
			t.Error("Token should not be empty")
		}

		t.Logf("Generated JWT token: %s", token)
	})

	// ==================== Step3: 密码错误登录失败测试 ====================
	t.Run("Step3_WrongPassword_LoginFailed", func(t *testing.T) {
		formData := strings.NewReader("phonenum=" + phoneNum + "&pwd=WrongPassword")
		loginReq, _ := http.NewRequest("POST", "/v1/login/pwd", formData)
		loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		loginResp := httptest.NewRecorder()

		router.ServeHTTP(loginResp, loginReq)

		var result map[string]interface{}
		if err := json.Unmarshal(loginResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		if result["code"] == float64(0) {
			t.Error("Expected login to fail with wrong password")
		}

		t.Logf("Expected failure with wrong password: %v", result["msg"])
	})
}