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

func cleanupOAuthData() {
	if config.DB != nil {
		config.DB.Exec("DELETE FROM o_auth_users")
		config.DB.Exec("DELETE FROM users WHERE phone_num LIKE '9%'")
	}
	if config.RDB != nil {
		config.RDB.FlushDB(config.RDB.Context())
	}
}

func TestOAuthLogin_Apple(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	router := gin.New()
	router.POST("/v1/login/oauth", handler.OAuthLogin)

	t.Run("AppleLogin_NewUser", func(t *testing.T) {
		cleanupOAuthData()

		formData := strings.NewReader("provider=apple&id_token=mock_apple_id_token_001")
		req, _ := http.NewRequest("POST", "/v1/login/oauth", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		t.Logf("AppleLogin response status: %d", resp.Code)
		t.Logf("AppleLogin response body: %s", resp.Body.String())

		if resp.Code != http.StatusOK {
			t.Fatalf("AppleLogin failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
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

		accessToken, ok := dataMap["access_token"].(string)
		if !ok || accessToken == "" {
			t.Error("Access token should not be empty")
		}

		refreshToken, ok := dataMap["refresh_token"].(string)
		if !ok || refreshToken == "" {
			t.Error("Refresh token should not be empty")
		}

		newUser, ok := dataMap["new_user"].(bool)
		if !ok || !newUser {
			t.Error("Expected new_user to be true")
		}

		t.Logf("Apple login success - new user")
	})

	t.Run("AppleLogin_ExistingUser", func(t *testing.T) {
		formData := strings.NewReader("provider=apple&id_token=mock_apple_id_token_001")
		req, _ := http.NewRequest("POST", "/v1/login/oauth", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("AppleLogin failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		data := result["data"].(map[string]interface{})
		newUser := data["new_user"].(bool)
		if newUser {
			t.Error("Expected new_user to be false for existing user")
		}

		t.Logf("Apple login success - existing user")
	})
}

func TestOAuthLogin_Google(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	router := gin.New()
	router.POST("/v1/login/oauth", handler.OAuthLogin)

	t.Run("GoogleLogin_NewUser", func(t *testing.T) {
		cleanupOAuthData()

		formData := strings.NewReader("provider=google&id_token=mock_google_id_token_001")
		req, _ := http.NewRequest("POST", "/v1/login/oauth", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		t.Logf("GoogleLogin response status: %d", resp.Code)
		t.Logf("GoogleLogin response body: %s", resp.Body.String())

		if resp.Code != http.StatusOK {
			t.Fatalf("GoogleLogin failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		data := result["data"].(map[string]interface{})
		newUser := data["new_user"].(bool)
		if !newUser {
			t.Error("Expected new_user to be true")
		}

		t.Logf("Google login success - new user")
	})
}

func TestOAuthLogin_Wechat(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	router := gin.New()
	router.POST("/v1/login/oauth", handler.OAuthLogin)

	t.Run("WechatLogin_NewUser", func(t *testing.T) {
		cleanupOAuthData()

		formData := strings.NewReader("provider=wechat&code=mock_wechat_code_001")
		req, _ := http.NewRequest("POST", "/v1/login/oauth", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		t.Logf("WechatLogin response status: %d", resp.Code)
		t.Logf("WechatLogin response body: %s", resp.Body.String())

		if resp.Code != http.StatusOK {
			t.Fatalf("WechatLogin failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		data := result["data"].(map[string]interface{})
		newUser := data["new_user"].(bool)
		if !newUser {
			t.Error("Expected new_user to be true")
		}

		t.Logf("Wechat login success - new user")
	})
}

func TestOAuthLogin_Alipay(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	router := gin.New()
	router.POST("/v1/login/oauth", handler.OAuthLogin)

	t.Run("AlipayLogin_NewUser", func(t *testing.T) {
		cleanupOAuthData()

		formData := strings.NewReader("provider=alipay&code=mock_alipay_code_001")
		req, _ := http.NewRequest("POST", "/v1/login/oauth", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		t.Logf("AlipayLogin response status: %d", resp.Code)
		t.Logf("AlipayLogin response body: %s", resp.Body.String())

		if resp.Code != http.StatusOK {
			t.Fatalf("AlipayLogin failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		data := result["data"].(map[string]interface{})
		newUser := data["new_user"].(bool)
		if !newUser {
			t.Error("Expected new_user to be true")
		}

		t.Logf("Alipay login success - new user")
	})
}

func TestOAuthLogin_Email(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	router := gin.New()
	router.POST("/v1/login/oauth", handler.OAuthLogin)

	t.Run("EmailLogin_NewUser", func(t *testing.T) {
		cleanupOAuthData()

		formData := strings.NewReader("provider=email&email=test@example.com")
		req, _ := http.NewRequest("POST", "/v1/login/oauth", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		t.Logf("EmailLogin response status: %d", resp.Code)
		t.Logf("EmailLogin response body: %s", resp.Body.String())

		if resp.Code != http.StatusOK {
			t.Fatalf("EmailLogin failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		data := result["data"].(map[string]interface{})
		newUser := data["new_user"].(bool)
		if !newUser {
			t.Error("Expected new_user to be true")
		}

		t.Logf("Email login success - new user")
	})

	t.Run("EmailLogin_InvalidEmail", func(t *testing.T) {
		formData := strings.NewReader("provider=email&email=invalid-email")
		req, _ := http.NewRequest("POST", "/v1/login/oauth", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] == float64(0) {
			t.Error("Expected login to fail with invalid email")
		}

		t.Logf("Expected failure with invalid email")
	})
}

func TestOAuthLogin_InvalidProvider(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	router := gin.New()
	router.POST("/v1/login/oauth", handler.OAuthLogin)

	t.Run("InvalidProvider", func(t *testing.T) {
		formData := strings.NewReader("provider=invalid_provider&code=test_code")
		req, _ := http.NewRequest("POST", "/v1/login/oauth", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result["code"] == float64(0) {
			t.Error("Expected login to fail with invalid provider")
		}

		t.Logf("Expected failure with invalid provider")
	})

	t.Run("MissingProvider", func(t *testing.T) {
		formData := strings.NewReader("code=test_code")
		req, _ := http.NewRequest("POST", "/v1/login/oauth", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Error("Expected bad request with missing provider")
		}

		t.Logf("Expected bad request with missing provider")
	})
}