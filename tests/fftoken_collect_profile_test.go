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

// TestCollectProfile_FullFlow 资料收集完整流程集成测试
// 模拟用户注册成功后首次访问，需要收集资料信息的全过程
//
// 测试流程：
//   Step1: 用户注册（使用手机号和验证码）
//     - 发送短信验证码
//     - 使用验证码和密码完成注册
//     - 获取access_token和refresh_token
//
//   Step2: 检查资料收集状态（首次登录）
//     - 注册后ProfileCompleted默认为0（未完成）
//     - 返回profile_completed: false，表示需要收集资料
//
//   Step3: 收集个人资料
//     - 设置昵称、性别、出生年份、身高、体重、城市、学校、职业等
//     - 设置爱好和交友目的
//
//   Step4: 设置个性签名
//     - 设置个人主页的个性签名（最大200字符）
//
//   Step5: 设置理想对象条件
//     - 设置理想对象的年龄范围、性别、身高范围等筛选条件
//
//   Step6: 标记资料收集完成
//     - 将ProfileCompleted设置为1（已完成）
//     - 允许用户进入首页
//
//   Step7: 验证资料收集状态已更新
//     - 返回profile_completed: true，表示资料已收集完成
//
// 安全规则：
//   1. 所有接口需要JWT Token认证
//   2. 个性签名长度限制：最大200字符
//   3. 昵称有效性校验（敏感词过滤、长度检查）
func TestCollectProfile_FullFlow(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	// 清理测试环境
	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// 初始化路由（公开路由和需要认证的路由分开）
	router := gin.New()

	// 公开路由（无需认证）
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)

	// 需要认证的路由（带JWT中间件）
	private := router.Group("/", middleware.JWT())
	{
		private.POST("/v1/collect/myinfo", handler.CollectMyInfo)
		private.POST("/v1/collect/aiminfo", handler.CollectAimInfo)
		private.GET("/v1/profile/status", handler.CheckProfileStatus)
		private.POST("/v1/profile/sign", handler.SetSignText)
		private.POST("/v1/profile/complete", handler.CompleteProfile)
	}

	// 测试数据（使用不会与其他测试冲突的手机号）
	phoneNum := "13900139999"
	password := "Password123"
	var accessToken string

	// ==================== Step1: 用户注册 ====================
	t.Run("Step1_Register", func(t *testing.T) {
		// 发送验证码
		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneNum+"&tag=register", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		// 获取验证码
		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) == 0 {
			t.Fatal("No SMS code sent")
		}
		code := sentMsgs[0].Code

		// 注册
		formData := strings.NewReader("phonenum=" + phoneNum + "&code=" + code + "&password=" + password)
		registerReq, _ := http.NewRequest("POST", "/v1/register", formData)
		registerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		registerResp := httptest.NewRecorder()

		router.ServeHTTP(registerResp, registerReq)

		if registerResp.Code != http.StatusOK {
			t.Fatalf("Register failed with status %d: %s", registerResp.Code, registerResp.Body.String())
		}

		// 解析注册响应
		var result map[string]interface{}
		if err := json.Unmarshal(registerResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal register response: %v", err)
		}

		data := result["data"].(map[string]interface{})
		accessToken = data["access_token"].(string)

		t.Logf("Registered user with access_token: %s", accessToken)
	})

	// ==================== Step2: 检查资料收集状态（首次登录） ====================
	t.Run("Step2_CheckProfileStatus_NotCompleted", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/profile/status", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("CheckProfileStatus failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		data := result["data"].(map[string]interface{})
		profileCompleted := data["profile_completed"].(bool)

		if profileCompleted != false {
			t.Errorf("Expected profile_completed to be false, got %v", profileCompleted)
		}

		t.Logf("Profile status: profile_completed=%v", profileCompleted)
	})

	// ==================== Step3: 收集个人资料 ====================
	t.Run("Step3_CollectMyInfo", func(t *testing.T) {
		body := `{
			"nickname": "测试用户",
			"gender": 1,
			"birthyear": 1995,
			"height": 175,
			"weight": 70,
			"city": "北京",
			"school": "清华大学",
			"job": "工程师",
			"edulevel": 4,
			"starsign": 1,
			"favors": ["1", "2"],
			"dating_purposes": ["1"]
		}`

		req, _ := http.NewRequest("POST", "/v1/collect/myinfo", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("CollectMyInfo failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
		}

		t.Log("Personal info collected successfully")
	})

	// ==================== Step4: 设置个性签名 ====================
	t.Run("Step4_SetSignText", func(t *testing.T) {
		body := `{"sign_text": "这是我的个性签名"}`

		req, _ := http.NewRequest("POST", "/v1/profile/sign", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("SetSignText failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
		}

		t.Log("Signature set successfully")
	})

	// ==================== Step5: 设置理想对象条件 ====================
	t.Run("Step5_CollectAimInfo", func(t *testing.T) {
		body := `{
			"birthyear": ["1990", "2000"],
			"gender": 2,
			"height": ["160", "180"],
			"weight": "50-60",
			"edulevel": ["3", "4"],
			"starsign": ["1", "2"],
			"favors": ["1", "2"]
		}`

		req, _ := http.NewRequest("POST", "/v1/collect/aiminfo", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("CollectAimInfo failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
		}

		t.Log("Aim info collected successfully")
	})

	// ==================== Step6: 标记资料收集完成 ====================
	t.Run("Step6_CompleteProfile", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/profile/complete", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("CompleteProfile failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
		}

		data := result["data"].(map[string]interface{})
		message := data["message"].(string)
		if message != "资料收集完成" {
			t.Errorf("Expected message '资料收集完成', got '%s'", message)
		}

		t.Log("Profile completed successfully")
	})

	// ==================== Step7: 验证资料收集状态已更新 ====================
	t.Run("Step7_CheckProfileStatus_Completed", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/profile/status", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("CheckProfileStatus failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		data := result["data"].(map[string]interface{})
		profileCompleted := data["profile_completed"].(bool)

		if profileCompleted != true {
			t.Errorf("Expected profile_completed to be true, got %v", profileCompleted)
		}

		t.Logf("Profile status updated: profile_completed=%v", profileCompleted)
	})
}

// TestCollectProfile_WithoutToken 测试未携带Token访问资料收集接口
func TestCollectProfile_WithoutToken(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	// 使用带JWT中间件的路由
	router := gin.New()
	private := router.Group("/", middleware.JWT())
	{
		private.GET("/v1/profile/status", handler.CheckProfileStatus)
		private.POST("/v1/profile/sign", handler.SetSignText)
		private.POST("/v1/profile/complete", handler.CompleteProfile)
	}

	t.Run("CheckProfileStatus_WithoutToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/profile/status", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.Code)
		}
	})

	t.Run("SetSignText_WithoutToken", func(t *testing.T) {
		body := `{"sign_text": "测试签名"}`
		req, _ := http.NewRequest("POST", "/v1/profile/sign", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.Code)
		}
	})

	t.Run("CompleteProfile_WithoutToken", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/profile/complete", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.Code)
		}
	})
}

// TestSetSignText_TooLong 测试个性签名过长
func TestSetSignText_TooLong(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	// 清理测试环境
	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// 初始化路由（公开路由和需要认证的路由分开）
	router := gin.New()

	// 公开路由（无需认证）
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)

	// 需要认证的路由（带JWT中间件）
	private := router.Group("/", middleware.JWT())
	{
		private.POST("/v1/profile/sign", handler.SetSignText)
	}

	// 注册用户获取token（使用不会与其他测试冲突的手机号）
	phoneNum := "13900139998"
	password := "Password123"

	// 发送验证码
	req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneNum+"&tag=register", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	sentMsgs := mockSMSGateway.GetSentMessages()
	code := sentMsgs[0].Code

	// 注册
	formData := strings.NewReader("phonenum=" + phoneNum + "&code=" + code + "&password=" + password)
	registerReq, _ := http.NewRequest("POST", "/v1/register", formData)
	registerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)

	var result map[string]interface{}
	json.Unmarshal(registerResp.Body.Bytes(), &result)
	data := result["data"].(map[string]interface{})
	accessToken := data["access_token"].(string)

	// 测试超长签名（超过200字符）
	t.Run("SignTextTooLong", func(t *testing.T) {
		// 生成250个字符的签名
		longSign := strings.Repeat("测试", 125)

		body := `{"sign_text": "` + longSign + `"}`
		req, _ := http.NewRequest("POST", "/v1/profile/sign", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("SetSignText failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)

		if result["code"] != float64(1) {
			t.Errorf("Expected code 1 for error, got %v", result["code"])
		}

		msg := result["msg"].(string)
		if msg != "签名长度不能超过200字符" {
			t.Errorf("Expected error message '签名长度不能超过200字符', got '%s'", msg)
		}

		t.Log("Sign text too long test passed")
	})
}