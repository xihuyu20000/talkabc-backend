package test

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRefreshToken_FullFlow 刷新令牌完整流程集成测试
// 【刷新令牌安全规则】
//   1. 验证刷新令牌格式（必须包含随机部分和JWT部分）
//   2. 验证刷新令牌签名有效性
//   3. 验证刷新令牌是否在Redis中存在且一致（防止滥用）
//   4. 验证用户是否存在且账号状态正常
//   5. 生成新的访问令牌和刷新令牌（刷新令牌轮转）
//   6. 将新令牌保存到Redis，旧令牌失效
//   7. 记录刷新操作日志（不可删除）
func TestRefreshToken_FullFlow(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)
	router.POST("/v1/login/code", handler.LoginByCode)
	router.POST("/v1/refresh-token", handler.RefreshToken)
	protected := router.Group("/v1/", middleware.JWT())
	protected.GET("/users/me", func(c *gin.Context) {
		uid := middleware.GetUID(c)
		c.JSON(http.StatusOK, gin.H{"uid": uid})
	})
	protected.POST("/logout", handler.Logout)

	phoneNum := "13900139020"
	password := "Test@1234"

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// ==================== Step1: 注册测试用户 ====================
	t.Run("Step1_RegisterUser", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneNum+"&tag=register", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+phoneNum)

		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) != 1 {
			t.Fatalf("Expected 1 sent message, got %d", len(sentMsgs))
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

		t.Log("Step1: 注册用户成功")
	})

	// ==================== Step2: 登录获取令牌 ====================
	t.Run("Step2_GetToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneNum+"&tag=login", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+phoneNum)

		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) != 2 {
			t.Fatalf("Expected 2 sent messages, got %d", len(sentMsgs))
		}
		code := sentMsgs[1].Code

		loginData := map[string]string{
			"phonenum": phoneNum,
			"code":     code,
		}
		jsonData, _ := json.Marshal(loginData)
		loginReq, _ := http.NewRequest("POST", "/v1/login/code", bytes.NewBuffer(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp := httptest.NewRecorder()
		router.ServeHTTP(loginResp, loginReq)
		if loginResp.Code != http.StatusOK {
			t.Fatalf("Login failed with status %d: %s", loginResp.Code, loginResp.Body.String())
		}

		var loginResult map[string]interface{}
		if err := json.NewDecoder(loginResp.Body).Decode(&loginResult); err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}

		data, ok := loginResult["data"].(map[string]interface{})
		if !ok {
			t.Fatal("Login response does not contain data")
		}

		accessToken, ok := data["access_token"].(string)
		if !ok || accessToken == "" {
			t.Fatal("Login response does not contain access_token")
		}

		refreshToken, ok := data["refresh_token"].(string)
		if !ok || refreshToken == "" {
			t.Fatal("Login response does not contain refresh_token")
		}

		t.Logf("Step2: 登录成功 - access_token: %s, refresh_token: %s", accessToken, refreshToken)

		// ==================== Step3: 使用刷新令牌获取新令牌 ====================
		t.Run("Step3_RefreshToken", func(t *testing.T) {
			formData := strings.NewReader("refresh_token=" + refreshToken)
			refreshReq, _ := http.NewRequest("POST", "/v1/refresh-token", formData)
			refreshReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			refreshResp := httptest.NewRecorder()
			router.ServeHTTP(refreshResp, refreshReq)
			if refreshResp.Code != http.StatusOK {
				t.Fatalf("RefreshToken failed with status %d: %s", refreshResp.Code, refreshResp.Body.String())
			}

			var refreshResult map[string]interface{}
			if err := json.NewDecoder(refreshResp.Body).Decode(&refreshResult); err != nil {
				t.Fatalf("Failed to decode refresh response: %v", err)
			}

			data, ok := refreshResult["data"].(map[string]interface{})
			if !ok {
				t.Fatal("Refresh response does not contain data")
			}

			newAccessToken, ok := data["access_token"].(string)
			if !ok || newAccessToken == "" {
				t.Fatal("Refresh response does not contain access_token")
			}

			newRefreshToken, ok := data["refresh_token"].(string)
			if !ok || newRefreshToken == "" {
				t.Fatal("Refresh response does not contain refresh_token")
			}

			if newAccessToken == accessToken {
				t.Error("New access_token should be different from old one")
			}

			if newRefreshToken == refreshToken {
				t.Error("New refresh_token should be different from old one")
			}

			t.Log("Step3: 刷新令牌成功")

			// ==================== Step4: 旧刷新令牌失效 ====================
			t.Run("Step4_OldRefreshTokenInvalid", func(t *testing.T) {
				formData := strings.NewReader("refresh_token=" + refreshToken)
				refreshReq, _ := http.NewRequest("POST", "/v1/refresh-token", formData)
				refreshReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				refreshResp := httptest.NewRecorder()
				router.ServeHTTP(refreshResp, refreshReq)

				var result map[string]interface{}
				if err := json.NewDecoder(refreshResp.Body).Decode(&result); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if result["code"] != float64(1) {
					t.Errorf("Expected code 1, got %v", result["code"])
				}

				t.Log("Step4: 旧刷新令牌已失效")
			})

			// ==================== Step5: 新访问令牌有效 ====================
			t.Run("Step5_NewAccessTokenValid", func(t *testing.T) {
				req, _ := http.NewRequest("GET", "/v1/users/me", nil)
				req.Header.Set("Authorization", "Bearer "+newAccessToken)
				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req)
				if resp.Code != http.StatusOK {
					t.Errorf("New access token should be valid, got status %d", resp.Code)
				}

				t.Log("Step5: 新访问令牌有效")
			})

			// ==================== Step6: 验证新令牌可用于刷新 ====================
			t.Run("Step6_RefreshAgain", func(t *testing.T) {
				formData := strings.NewReader("refresh_token=" + newRefreshToken)
				refreshReq, _ := http.NewRequest("POST", "/v1/refresh-token", formData)
				refreshReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				refreshResp := httptest.NewRecorder()
				router.ServeHTTP(refreshResp, refreshReq)

				if refreshResp.Code != http.StatusOK {
					t.Errorf("Refresh should succeed, got status %d", refreshResp.Code)
					return
				}

				var result map[string]interface{}
				json.NewDecoder(refreshResp.Body).Decode(&result)
				data := result["data"].(map[string]interface{})

				if data["access_token"] == nil || data["access_token"] == "" {
					t.Error("Refresh response should contain access_token")
				}

				if data["refresh_token"] == nil || data["refresh_token"] == "" {
					t.Error("Refresh response should contain refresh_token")
				}

				t.Log("Step6: 新刷新令牌可用于再次刷新")
			})
		})
	})
}

// TestRefreshToken_InvalidToken 测试无效刷新令牌
func TestRefreshToken_InvalidToken(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.POST("/v1/refresh-token", handler.RefreshToken)

	config.RDB.FlushDB(config.RDB.Context())

	tests := []struct {
		name       string
		token      string
		expectCode int
	}{
		{"Empty token", "", http.StatusBadRequest},
		{"Invalid format", "invalid_token", http.StatusOK},
		{"Random string", "abc123", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formData := strings.NewReader("refresh_token=" + tt.token)
			req, _ := http.NewRequest("POST", "/v1/refresh-token", formData)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != tt.expectCode {
				t.Errorf("Expected status %d for %s, got %d", tt.expectCode, tt.name, resp.Code)
				return
			}

			if tt.expectCode == http.StatusOK {
				var result map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if result["code"] != float64(1) {
					t.Errorf("Expected code 1 for %s, got %v", tt.name, result["code"])
				}
			}
		})
	}
}

// TestRefreshToken_AfterLogout 测试退出登录后刷新令牌失效
func TestRefreshToken_AfterLogout(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)
	router.POST("/v1/login/code", handler.LoginByCode)
	router.POST("/v1/refresh-token", handler.RefreshToken)
	protected := router.Group("/v1/", middleware.JWT())
	protected.POST("/logout", handler.Logout)

	phoneNum := "13900139021"

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// 注册用户
	req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneNum+"&tag=register", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+phoneNum)
	code := mockSMSGateway.GetSentMessages()[0].Code

	registerData := map[string]string{
		"phonenum": phoneNum,
		"code":     code,
		"password": "Test@1234",
	}
	jsonData, _ := json.Marshal(registerData)
	registerReq, _ := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(jsonData))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)

	var registerResult map[string]interface{}
	json.NewDecoder(registerResp.Body).Decode(&registerResult)
	data := registerResult["data"].(map[string]interface{})
	accessToken := data["access_token"].(string)
	refreshToken := data["refresh_token"].(string)

	// 退出登录
	logoutReq, _ := http.NewRequest("POST", "/v1/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+accessToken)
	logoutResp := httptest.NewRecorder()
	router.ServeHTTP(logoutResp, logoutReq)

	// 尝试使用刷新令牌
	refreshReq, _ := http.NewRequest("POST", "/v1/refresh-token", strings.NewReader("refresh_token="+refreshToken))
	refreshReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	refreshResp := httptest.NewRecorder()
	router.ServeHTTP(refreshResp, refreshReq)

	var result map[string]interface{}
	json.NewDecoder(refreshResp.Body).Decode(&result)

	if result["code"] != float64(1) {
		t.Errorf("Expected code 1 after logout, got %v", result["code"])
	}

	t.Log("Test passed: Refresh token invalid after logout")
}