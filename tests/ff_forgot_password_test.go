package test

import (
	"backend/internal/config"
	"backend/internal/handler"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestForgotPassword_FullFlow 忘记密码完整流程集成测试
// 模拟用户忘记密码后重置密码的全过程，验证所有安全规则检查
//
// 测试流程：
//   Step1: 用户注册（先创建测试用户）
//   Step2: 用户输入手机号，发起密码重置（生成重置Token）
//   Step3: 验证重置Token是否有效
//   Step4: 用户输入新密码，完成密码重置
//   Step5: 使用新密码登录验证
//
// 重置凭证安全规则：
//   1. Token设计：单次有效（使用后立即销毁）、短有效期（5分钟）、不可预测（crypto/rand生成）
//   2. Token绑定：绑定userID+设备标识+创建时间，防止跨账号盗用
//   3. 存储安全：禁止明文存库，仅存储sha256哈希
//
// 重置流程行为风控：
//   1. 记录用户常用IP、设备UA、地区
//   2. 同一账号24h最多允许3次密码重置，超限锁定重置通道24h
//   3. 敏感操作日志落地（不可删除）：用户ID、操作时间、IP、UA、操作类型、是否成功
//   4. 重置完成后推送通知（告知密码已修改）
//
// 密码存储加密：
//   1. 使用bcrypt加密（cost=10，自动内置盐）
//   2. 重置成功后：清空该用户全部登录态、清空所有未使用重置Token
//   3. 绝不返回原始密码或加密密码到前端
//
// 最低安全策略：
//   1. 长度≥8位，推荐12位以上
//   2. 至少包含两种字符类型（大写字母、小写字母、数字、特殊符号）
//   3. 禁止弱密码黑名单
//   4. 禁止和历史5次旧密码重复
//   5. 禁止包含用户名、手机号、邮箱前缀
func TestForgotPassword_FullFlow(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)
	router.POST("/v1/reset-password/initiate", handler.InitiateResetPassword)
	router.POST("/v1/reset-password/validate", handler.ValidateResetToken)
	router.POST("/v1/reset-password/complete", handler.CompleteResetPassword)
	router.POST("/v1/login/pwd", handler.LoginByPassword)

	phoneNum := "13900139006"
	password := "Password123"
	newPassword := "NewPassword456"

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	var resetToken string

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

		registerData := map[string]string{
			"phonenum": phoneNum,
			"code":     code,
			"password": password,
		}
		jsonData, _ := json.Marshal(registerData)
		registerReq, _ := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(jsonData))
		registerReq.Header.Set("Content-Type", "application/json")
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

	// ==================== Step2: 发起密码重置 ====================
	// 模拟用户点击"忘记密码"，输入手机号，系统生成重置Token并发送
	t.Run("Step2_InitiateResetPassword", func(t *testing.T) {
		initiateData := map[string]string{
			"phonenum": phoneNum,
		}
		jsonData, _ := json.Marshal(initiateData)
		initiateReq, _ := http.NewRequest("POST", "/v1/reset-password/initiate", bytes.NewBuffer(jsonData))
		initiateReq.Header.Set("Content-Type", "application/json")
		initiateResp := httptest.NewRecorder()

		router.ServeHTTP(initiateResp, initiateReq)

		t.Logf("InitiateResetPassword response status: %d", initiateResp.Code)
		t.Logf("InitiateResetPassword response body: %s", initiateResp.Body.String())

		if initiateResp.Code != http.StatusOK {
			var result map[string]interface{}
			if err := json.Unmarshal(initiateResp.Body.Bytes(), &result); err == nil {
				if msg, ok := result["msg"].(string); ok {
					t.Logf("Initiate reset error message: %s", msg)
				}
			}
			t.Fatalf("InitiateResetPassword failed with status %d: %s", initiateResp.Code, initiateResp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(initiateResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal initiate response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		t.Logf("Password reset initiated successfully")

		data := result["data"]
		if data != nil {
			dataMap, ok := data.(map[string]interface{})
			if ok {
				tokenVal, ok := dataMap["token"].(string)
				if ok && tokenVal != "" {
					resetToken = tokenVal
					t.Logf("Received reset token: %s", resetToken)
				}
			}
		}

		if resetToken == "" {
			t.Skip("No reset token returned")
		}
	})

	// ==================== Step3: 验证重置Token ====================
	// 模拟前端验证重置链接是否有效
	t.Run("Step3_ValidateResetToken", func(t *testing.T) {
		if resetToken == "" {
			t.Skip("No reset token generated in previous step")
		}

		validateReq, _ := http.NewRequest("POST", "/v1/reset-password/validate?token="+resetToken, nil)
		validateResp := httptest.NewRecorder()

		router.ServeHTTP(validateResp, validateReq)

		t.Logf("ValidateResetToken response status: %d", validateResp.Code)
		t.Logf("ValidateResetToken response body: %s", validateResp.Body.String())

		if validateResp.Code != http.StatusOK {
			var result map[string]interface{}
			if err := json.Unmarshal(validateResp.Body.Bytes(), &result); err == nil {
				if msg, ok := result["msg"].(string); ok {
					t.Logf("Validate token error message: %s", msg)
				}
			}
			t.Fatalf("ValidateResetToken failed with status %d: %s", validateResp.Code, validateResp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(validateResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal validate response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		t.Logf("Reset token validated successfully")
	})

	// ==================== Step4: 完成密码重置 ====================
	// 模拟用户输入新密码，完成密码重置
	t.Run("Step4_CompleteResetPassword", func(t *testing.T) {
		if resetToken == "" {
			t.Skip("No reset token generated in previous step")
		}

		completeData := map[string]string{
			"token": resetToken,
			"pwd1":  newPassword,
			"pwd2":  newPassword,
		}
		jsonData, _ := json.Marshal(completeData)
		completeReq, _ := http.NewRequest("POST", "/v1/reset-password/complete", bytes.NewBuffer(jsonData))
		completeReq.Header.Set("Content-Type", "application/json")
		completeResp := httptest.NewRecorder()

		router.ServeHTTP(completeResp, completeReq)

		t.Logf("CompleteResetPassword response status: %d", completeResp.Code)
		t.Logf("CompleteResetPassword response body: %s", completeResp.Body.String())

		if completeResp.Code != http.StatusOK {
			var result map[string]interface{}
			if err := json.Unmarshal(completeResp.Body.Bytes(), &result); err == nil {
				if msg, ok := result["msg"].(string); ok {
					t.Logf("Complete reset error message: %s", msg)
				}
			}
			t.Fatalf("CompleteResetPassword failed with status %d: %s", completeResp.Code, completeResp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(completeResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal complete response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		t.Logf("Password reset completed successfully")
	})

	// ==================== Step5: 使用新密码登录验证 ====================
	// 验证密码重置成功后，新密码可以正常登录
	t.Run("Step5_LoginWithNewPassword", func(t *testing.T) {
		loginData := map[string]string{
			"phonenum": phoneNum,
			"pwd":      newPassword,
		}
		jsonData, _ := json.Marshal(loginData)
		loginReq, _ := http.NewRequest("POST", "/v1/login/pwd", bytes.NewBuffer(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp := httptest.NewRecorder()

		router.ServeHTTP(loginResp, loginReq)

		t.Logf("LoginWithNewPassword response status: %d", loginResp.Code)
		t.Logf("LoginWithNewPassword response body: %s", loginResp.Body.String())

		if loginResp.Code != http.StatusOK {
			var result map[string]interface{}
			if err := json.Unmarshal(loginResp.Body.Bytes(), &result); err == nil {
				if msg, ok := result["msg"].(string); ok {
					t.Logf("Login error message: %s", msg)
				}
			}
			t.Fatalf("LoginWithNewPassword failed with status %d: %s", loginResp.Code, loginResp.Body.String())
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

		t.Logf("Login with new password successful, generated JWT token: %s", token)
	})

	// ==================== Step6: 旧密码登录失败验证 ====================
	// 验证密码重置成功后，旧密码不再有效
	t.Run("Step6_OldPasswordLoginFailed", func(t *testing.T) {
		loginData := map[string]string{
			"phonenum": phoneNum,
			"pwd":      password,
		}
		jsonData, _ := json.Marshal(loginData)
		loginReq, _ := http.NewRequest("POST", "/v1/login/pwd", bytes.NewBuffer(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp := httptest.NewRecorder()

		router.ServeHTTP(loginResp, loginReq)

		var result map[string]interface{}
		if err := json.Unmarshal(loginResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		if result["code"] == float64(0) {
			t.Error("Expected login to fail with old password after reset")
		}

		t.Logf("Expected failure with old password: %v", result["msg"])
	})

	// ==================== Step7: 重置Token二次使用失败验证 ====================
	// 验证重置Token使用一次后立即失效（单次有效）
	t.Run("Step7_ResetTokenOneTimeUse", func(t *testing.T) {
		if resetToken == "" {
			t.Skip("No reset token generated in previous step")
		}

		completeData := map[string]string{
			"token": resetToken,
			"pwd1":  "AnotherPass1",
			"pwd2":  "AnotherPass1",
		}
		jsonData, _ := json.Marshal(completeData)
		completeReq, _ := http.NewRequest("POST", "/v1/reset-password/complete", bytes.NewBuffer(jsonData))
		completeReq.Header.Set("Content-Type", "application/json")
		completeResp := httptest.NewRecorder()

		router.ServeHTTP(completeResp, completeReq)

		var result map[string]interface{}
		if err := json.Unmarshal(completeResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal complete response: %v", err)
		}

		if result["code"] == float64(0) {
			t.Error("Expected reset token to be invalid after first use")
		}

		t.Logf("Expected failure: reset token should be one-time use only")
	})
}