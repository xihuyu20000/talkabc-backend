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

func init() {
	InitTest()
}

// TestChangePhone_FullFlow 更换手机号完整流程集成测试
// 【更换手机号安全规则】
// 1. 必须已登录（通过JWT token验证）
// 2. 验证新手机号格式
// 3. 新手机号必须未被注册
// 4. 验证新手机号的短信验证码
// 5. 验证频率限制（24小时内最多更换3次）
// 6. 更新手机号后清空用户所有登录态
// 7. 记录更换手机号操作日志（不可删除）
func TestChangePhone_FullFlow(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)
	router.POST("/v1/login/code", handler.LoginByCode)
	protected := router.Group("/v1/", middleware.JWT())
	protected.POST("/change-phone", handler.ChangePhone)

	oldPhone := "13900139000"
	newPhone := "13900139001"

	// 清理测试环境
	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// Step1: 注册测试用户（旧手机号）
	t.Run("Step1_RegisterUser", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+oldPhone+"&tag=register", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+oldPhone)

		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) != 1 {
			t.Fatalf("Expected 1 sent message, got %d", len(sentMsgs))
		}
		code := sentMsgs[0].Code

		formData := strings.NewReader("phonenum=" + oldPhone + "&code=" + code + "&password=Test@1234")
		registerReq, _ := http.NewRequest("POST", "/v1/register", formData)
		registerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		registerResp := httptest.NewRecorder()
		router.ServeHTTP(registerResp, registerReq)
		if registerResp.Code != http.StatusOK {
			t.Fatalf("Register failed with status %d: %s", registerResp.Code, registerResp.Body.String())
		}

		t.Log("Step1: 注册用户成功")
	})

	// Step2: 登录获取token
	t.Run("Step2_GetToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+oldPhone+"&tag=login", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+oldPhone)

		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) < 2 {
			t.Fatalf("Expected at least 2 sent messages, got %d", len(sentMsgs))
		}
		code := sentMsgs[1].Code

		formData := strings.NewReader("phonenum=" + oldPhone + "&code=" + code)
		loginReq, _ := http.NewRequest("POST", "/v1/login/code", formData)
		loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		loginResp := httptest.NewRecorder()
		router.ServeHTTP(loginResp, loginReq)
		if loginResp.Code != http.StatusOK {
			t.Fatalf("Login failed with status %d: %s", loginResp.Code, loginResp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(loginResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse login response: %v", err)
		}

		tokenData, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected data field in login response")
		}
		token, ok := tokenData["access_token"].(string)
		if !ok || token == "" {
			t.Fatalf("Expected token in login response")
		}

		t.Log("Step2: 获取token成功")

		// Step3: 发送新手机号验证码
		t.Run("Step3_SendNewPhoneCode", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+newPhone+"&tag=change_phone", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != http.StatusOK {
				t.Fatalf("SendSMSCode for new phone failed with status %d: %s", resp.Code, resp.Body.String())
			}

			config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+newPhone)

			sentMsgs := mockSMSGateway.GetSentMessages()
			if len(sentMsgs) < 3 {
				t.Fatalf("Expected at least 3 sent messages, got %d", len(sentMsgs))
			}
			newCode := sentMsgs[2].Code

			t.Log("Step3: 发送新手机号验证码成功")

			// Step4: 更换手机号
			t.Run("Step4_ChangePhone", func(t *testing.T) {
				formData := strings.NewReader("new_phone=" + newPhone + "&code=" + newCode)
				changeReq, _ := http.NewRequest("POST", "/v1/change-phone", formData)
				changeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				changeReq.Header.Set("Authorization", "Bearer "+token)
				changeResp := httptest.NewRecorder()
				router.ServeHTTP(changeResp, changeReq)
				if changeResp.Code != http.StatusOK {
					t.Fatalf("ChangePhone failed with status %d: %s", changeResp.Code, changeResp.Body.String())
				}

				var changeResult map[string]interface{}
				if err := json.Unmarshal(changeResp.Body.Bytes(), &changeResult); err != nil {
					t.Fatalf("Failed to parse change phone response: %v", err)
				}

				if changeResult["code"].(float64) != 0 {
					t.Fatalf("ChangePhone returned error: %s", changeResp.Body.String())
				}

				t.Log("Step4: 更换手机号成功")

				// Step5: 验证旧token失效
				t.Run("Step5_OldTokenInvalid", func(t *testing.T) {
					formData := strings.NewReader("new_phone=13900139002&code=123456")
					testReq, _ := http.NewRequest("POST", "/v1/change-phone", formData)
					testReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					testReq.Header.Set("Authorization", "Bearer "+token)
					testResp := httptest.NewRecorder()
					router.ServeHTTP(testResp, testReq)
					if testResp.Code != http.StatusUnauthorized {
						t.Fatalf("Old token should be invalid, got status %d: %s", testResp.Code, testResp.Body.String())
					}

					t.Log("Step5: 旧token已失效")
				})

				// Step6: 使用新手机号登录
				t.Run("Step6_LoginWithNewPhone", func(t *testing.T) {
					req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+newPhone+"&tag=login", nil)
					resp := httptest.NewRecorder()
					router.ServeHTTP(resp, req)
					if resp.Code != http.StatusOK {
						t.Fatalf("SendSMSCode for new phone login failed with status %d: %s", resp.Code, resp.Body.String())
					}

					config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+newPhone)

					sentMsgs := mockSMSGateway.GetSentMessages()
					if len(sentMsgs) < 4 {
						t.Fatalf("Expected at least 4 sent messages, got %d", len(sentMsgs))
					}
					newLoginCode := sentMsgs[3].Code

					formData := strings.NewReader("phonenum=" + newPhone + "&code=" + newLoginCode)
					loginReq, _ := http.NewRequest("POST", "/v1/login/code", formData)
					loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					loginResp := httptest.NewRecorder()
					router.ServeHTTP(loginResp, loginReq)
					if loginResp.Code != http.StatusOK {
						t.Fatalf("Login with new phone failed with status %d: %s", loginResp.Code, loginResp.Body.String())
					}

					t.Log("Step6: 使用新手机号登录成功")
				})
			})
		})
	})
}

// TestChangePhone_NoToken 测试更换手机号未登录场景
// 【更换手机号安全规则1】必须已登录
func TestChangePhone_NoToken(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	router := gin.New()
	protected := router.Group("/v1/", middleware.JWT())
	protected.POST("/change-phone", handler.ChangePhone)

	req, _ := http.NewRequest("POST", "/v1/change-phone", strings.NewReader("new_phone=13900139001&code=654321"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 without token, got %d: %s", resp.Code, resp.Body.String())
	}

	t.Log("Test passed: Unauthorized without token")
}

// TestChangePhone_InvalidPhone 测试更换手机号格式验证
// 【更换手机号安全规则2】验证新手机号格式
func TestChangePhone_InvalidPhone(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)
	router.POST("/v1/login/code", handler.LoginByCode)
	protected := router.Group("/v1/", middleware.JWT())
	protected.POST("/change-phone", handler.ChangePhone)

	oldPhone := "13900139002"

	// 清理测试环境
	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// 注册用户
	req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+oldPhone+"&tag=register", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
	}

	config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+oldPhone)

	sentMsgs := mockSMSGateway.GetSentMessages()
	code := sentMsgs[0].Code

	formData := strings.NewReader("phonenum=" + oldPhone + "&code=" + code + "&password=Test@1234")
	registerReq, _ := http.NewRequest("POST", "/v1/register", formData)
	registerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)
	if registerResp.Code != http.StatusOK {
		t.Fatalf("Register failed with status %d: %s", registerResp.Code, registerResp.Body.String())
	}

	// 登录获取token
	req, _ = http.NewRequest("GET", "/v1/code/sms?phonenum="+oldPhone+"&tag=login", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
	}

	config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+oldPhone)

	sentMsgs = mockSMSGateway.GetSentMessages()
	loginCode := sentMsgs[1].Code

	formData = strings.NewReader("phonenum=" + oldPhone + "&code=" + loginCode)
	loginReq, _ := http.NewRequest("POST", "/v1/login/code", formData)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("Login failed with status %d: %s", loginResp.Code, loginResp.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(loginResp.Body.Bytes(), &result)
	token := result["data"].(map[string]interface{})["access_token"].(string)

	// 测试无效手机号格式
	invalidPhones := []string{"123", "1234567890", "123456789012", "abc12345678", "10900109000"}
	for _, invalidPhone := range invalidPhones {
		changeReq, _ := http.NewRequest("POST", "/v1/change-phone",
			strings.NewReader("new_phone="+invalidPhone+"&code=654321"))
		changeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		changeReq.Header.Set("Authorization", "Bearer "+token)
		changeResp := httptest.NewRecorder()
		router.ServeHTTP(changeResp, changeReq)

		var result map[string]interface{}
		json.Unmarshal(changeResp.Body.Bytes(), &result)
		if result["code"].(float64) == 0 {
			t.Errorf("Invalid phone %s should fail, got success", invalidPhone)
		}
	}

	t.Log("Test passed: Invalid phone formats blocked")
}

// TestChangePhone_PhoneAlreadyExists 测试更换手机号已被注册场景
// 【更换手机号安全规则3】新手机号必须未被注册
func TestChangePhone_PhoneAlreadyExists(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)
	router.POST("/v1/login/code", handler.LoginByCode)
	protected := router.Group("/v1/", middleware.JWT())
	protected.POST("/change-phone", handler.ChangePhone)

	oldPhone := "13900139003"
	existingPhone := "13900139004"

	// 清理测试环境
	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// 注册用户A
	req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+oldPhone+"&tag=register", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("SendSMSCode for user A failed with status %d: %s", resp.Code, resp.Body.String())
	}

	config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+oldPhone)

	sentMsgs := mockSMSGateway.GetSentMessages()
	codeA := sentMsgs[0].Code

	formData := strings.NewReader("phonenum=" + oldPhone + "&code=" + codeA + "&password=Test@1234")
	registerReq, _ := http.NewRequest("POST", "/v1/register", formData)
	registerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)
	if registerResp.Code != http.StatusOK {
		t.Fatalf("Register user A failed with status %d: %s", registerResp.Code, registerResp.Body.String())
	}

	// 注册用户B（占用新手机号）
	req, _ = http.NewRequest("GET", "/v1/code/sms?phonenum="+existingPhone+"&tag=register", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("SendSMSCode for user B failed with status %d: %s", resp.Code, resp.Body.String())
	}

	config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+existingPhone)

	sentMsgs = mockSMSGateway.GetSentMessages()
	codeB := sentMsgs[1].Code

	formData = strings.NewReader("phonenum=" + existingPhone + "&code=" + codeB + "&password=Test@1234")
	registerReq, _ = http.NewRequest("POST", "/v1/register", formData)
	registerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	registerResp = httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)
	if registerResp.Code != http.StatusOK {
		t.Fatalf("Register user B failed with status %d: %s", registerResp.Code, registerResp.Body.String())
	}

	// 用户A登录获取token
	req, _ = http.NewRequest("GET", "/v1/code/sms?phonenum="+oldPhone+"&tag=login", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("SendSMSCode for user A login failed with status %d: %s", resp.Code, resp.Body.String())
	}

	config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+oldPhone)

	sentMsgs = mockSMSGateway.GetSentMessages()
	loginCode := sentMsgs[2].Code

	formData = strings.NewReader("phonenum=" + oldPhone + "&code=" + loginCode)
	loginReq, _ := http.NewRequest("POST", "/v1/login/code", formData)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("Login user A failed with status %d: %s", loginResp.Code, loginResp.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(loginResp.Body.Bytes(), &result)
	token := result["data"].(map[string]interface{})["access_token"].(string)

	// 用户A尝试更换到已注册的手机号
	req, _ = http.NewRequest("GET", "/v1/code/sms?phonenum="+existingPhone+"&tag=change_phone", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("SendSMSCode for change phone failed with status %d: %s", resp.Code, resp.Body.String())
	}

	config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+existingPhone)

	sentMsgs = mockSMSGateway.GetSentMessages()
	changeCode := sentMsgs[3].Code

	changeReq, _ := http.NewRequest("POST", "/v1/change-phone",
		strings.NewReader("new_phone="+existingPhone+"&code="+changeCode))
	changeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	changeReq.Header.Set("Authorization", "Bearer "+token)
	changeResp := httptest.NewRecorder()
	router.ServeHTTP(changeResp, changeReq)

	var changeResult map[string]interface{}
	json.Unmarshal(changeResp.Body.Bytes(), &changeResult)
	if changeResult["code"].(float64) == 0 {
		t.Fatalf("Change to existing phone should fail, got success: %s", changeResp.Body.String())
	}

	t.Log("Test passed: Changing to existing phone blocked")
}

// TestChangePhone_InvalidCode 测试更换手机号无效验证码场景
// 【更换手机号安全规则4】验证新手机号的短信验证码
func TestChangePhone_InvalidCode(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)
	router.POST("/v1/login/code", handler.LoginByCode)
	protected := router.Group("/v1/", middleware.JWT())
	protected.POST("/change-phone", handler.ChangePhone)

	oldPhone := "13900139005"
	newPhone := "13900139006"

	// 清理测试环境
	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// 注册用户
	req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+oldPhone+"&tag=register", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
	}

	config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+oldPhone)

	sentMsgs := mockSMSGateway.GetSentMessages()
	code := sentMsgs[0].Code

	formData := strings.NewReader("phonenum=" + oldPhone + "&code=" + code + "&password=Test@1234")
	registerReq, _ := http.NewRequest("POST", "/v1/register", formData)
	registerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)
	if registerResp.Code != http.StatusOK {
		t.Fatalf("Register failed with status %d: %s", registerResp.Code, registerResp.Body.String())
	}

	// 登录获取token
	req, _ = http.NewRequest("GET", "/v1/code/sms?phonenum="+oldPhone+"&tag=login", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
	}

	config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+oldPhone)

	sentMsgs = mockSMSGateway.GetSentMessages()
	loginCode := sentMsgs[1].Code

	formData = strings.NewReader("phonenum=" + oldPhone + "&code=" + loginCode)
	loginReq, _ := http.NewRequest("POST", "/v1/login/code", formData)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("Login failed with status %d: %s", loginResp.Code, loginResp.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(loginResp.Body.Bytes(), &result)
	token := result["data"].(map[string]interface{})["access_token"].(string)

	// 发送新手机号验证码
	req, _ = http.NewRequest("GET", "/v1/code/sms?phonenum="+newPhone+"&tag=change_phone", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("SendSMSCode for new phone failed with status %d: %s", resp.Code, resp.Body.String())
	}

	config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+newPhone)

	sentMsgs = mockSMSGateway.GetSentMessages()
	_ = sentMsgs[2].Code

	// 使用错误验证码
	invalidCodes := []string{"123456", "654320", "", "123"}
	for _, invalidCode := range invalidCodes {
		changeReq, _ := http.NewRequest("POST", "/v1/change-phone",
			strings.NewReader("new_phone="+newPhone+"&code="+invalidCode))
		changeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		changeReq.Header.Set("Authorization", "Bearer "+token)
		changeResp := httptest.NewRecorder()
		router.ServeHTTP(changeResp, changeReq)

		var result map[string]interface{}
		json.Unmarshal(changeResp.Body.Bytes(), &result)
		if result["code"].(float64) == 0 {
			t.Errorf("Invalid code %s should fail, got success", invalidCode)
		}
	}

	t.Log("Test passed: Invalid codes blocked")
}
