package test

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	InitTest()
}

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

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

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

		registerData := map[string]string{
			"phonenum": oldPhone,
			"code":     code,
			"password": "Test@1234",
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

		loginData := map[string]string{
			"phonenum": oldPhone,
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

		var result map[string]interface{}
		json.Unmarshal(loginResp.Body.Bytes(), &result)
		data := result["data"].(map[string]interface{})
		accessToken = data["access_token"].(string)

		t.Log("Step2: 登录获取token成功")
	})

	t.Run("Step3_SendVerificationCode", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+newPhone+"&tag=change_phone", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		t.Log("Step3: 发送新手机号验证码成功")
	})

	t.Run("Step4_ChangePhone", func(t *testing.T) {
		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) < 3 {
			t.Fatalf("Expected at least 3 sent messages, got %d", len(sentMsgs))
		}
		code := sentMsgs[2].Code

		changeData := map[string]string{
			"new_phone": newPhone,
			"code":      code,
		}
		jsonData, _ := json.Marshal(changeData)
		req, _ := http.NewRequest("POST", "/v1/change-phone", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("ChangePhone failed with status %d: %s", resp.Code, resp.Body.String())
		}

		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
		}

		t.Log("Step4: 更换手机号成功")
	})

	t.Run("Step5_VerifyNewPhone", func(t *testing.T) {
		config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+newPhone)
		
		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+newPhone+"&tag=login", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) < 4 {
			t.Fatalf("Expected at least 4 sent messages, got %d", len(sentMsgs))
		}
		code := sentMsgs[3].Code

		loginData := map[string]string{
			"phonenum": newPhone,
			"code":     code,
		}
		jsonData, _ := json.Marshal(loginData)
		loginReq, _ := http.NewRequest("POST", "/v1/login/code", bytes.NewBuffer(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp := httptest.NewRecorder()
		router.ServeHTTP(loginResp, loginReq)

		if loginResp.Code != http.StatusOK {
			t.Fatalf("Login with new phone failed with status %d: %s", loginResp.Code, loginResp.Body.String())
		}

		t.Log("Step5: 使用新手机号登录成功")
	})
}

var accessToken string

func TestChangePhone_NoToken(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	router := gin.New()
	protected := router.Group("/v1/", middleware.JWT())
	protected.POST("/change-phone", handler.ChangePhone)

	changeData := map[string]string{
		"new_phone": "13900139001",
		"code":      "123456",
	}
	jsonData, _ := json.Marshal(changeData)
	req, _ := http.NewRequest("POST", "/v1/change-phone", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.Code)
	}

	t.Log("Test passed: Unauthorized without token")
}

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

	phone := "13900139002"

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phone+"&tag=register", nil)
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
		"phonenum": phone,
		"code":     code,
		"password": "Test@1234",
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
	data := result["data"].(map[string]interface{})
	token := data["access_token"].(string)

	changeData := map[string]string{
		"new_phone": "invalid_phone",
		"code":      "123456",
	}
	jsonData, _ = json.Marshal(changeData)
	req, _ = http.NewRequest("POST", "/v1/change-phone", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	changeResp := httptest.NewRecorder()
	router.ServeHTTP(changeResp, req)

	var invalidResult map[string]interface{}
	json.Unmarshal(changeResp.Body.Bytes(), &invalidResult)
	if invalidResult["code"] == float64(0) {
		t.Error("Expected change phone to fail with invalid phone")
	}

	t.Log("Test passed: Change phone failed with invalid phone")
}

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

	phoneA := "13900139010"
	phoneB := "13900139011"

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneA+"&tag=register", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
	}

	sentMsgs := mockSMSGateway.GetSentMessages()
	if len(sentMsgs) == 0 {
		t.Fatal("No SMS code sent")
	}
	codeA := sentMsgs[0].Code

	registerData := map[string]string{
		"phonenum": phoneA,
		"code":     codeA,
		"password": "Test@1234",
	}
	jsonData, _ := json.Marshal(registerData)
	registerReq, _ := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(jsonData))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)
	if registerResp.Code != http.StatusOK {
		t.Fatalf("Register user A failed with status %d: %s", registerResp.Code, registerResp.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(registerResp.Body.Bytes(), &result)
	data := result["data"].(map[string]interface{})
	token := data["access_token"].(string)

	req2, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneB+"&tag=register", nil)
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)
	if resp2.Code != http.StatusOK {
		t.Fatalf("SendSMSCode failed with status %d: %s", resp2.Code, resp2.Body.String())
	}

	sentMsgs2 := mockSMSGateway.GetSentMessages()
	if len(sentMsgs2) < 2 {
		t.Fatalf("Expected at least 2 sent messages, got %d", len(sentMsgs2))
	}
	codeB := sentMsgs2[1].Code

	registerData2 := map[string]string{
		"phonenum": phoneB,
		"code":     codeB,
		"password": "Test@1234",
	}
	jsonData2, _ := json.Marshal(registerData2)
	registerReq2, _ := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(jsonData2))
	registerReq2.Header.Set("Content-Type", "application/json")
	registerResp2 := httptest.NewRecorder()
	router.ServeHTTP(registerResp2, registerReq2)
	if registerResp2.Code != http.StatusOK {
		t.Fatalf("Register user B failed with status %d: %s", registerResp2.Code, registerResp2.Body.String())
	}

	changeData := map[string]string{
		"new_phone": phoneB,
		"code":      "123456",
	}
	jsonData, _ = json.Marshal(changeData)
	changeReq, _ := http.NewRequest("POST", "/v1/change-phone", bytes.NewBuffer(jsonData))
	changeReq.Header.Set("Content-Type", "application/json")
	changeReq.Header.Set("Authorization", "Bearer "+token)
	changeResp := httptest.NewRecorder()
	router.ServeHTTP(changeResp, changeReq)

	var existsResult map[string]interface{}
	json.Unmarshal(changeResp.Body.Bytes(), &existsResult)
	if existsResult["code"] == float64(0) {
		t.Error("Expected change phone to fail when new phone already exists")
	}

	t.Log("Test passed: Change phone failed when phone already exists")
}

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

	oldPhone := "13900139012"
	newPhone := "13900139013"

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+oldPhone+"&tag=register", nil)
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
		"phonenum": oldPhone,
		"code":     code,
		"password": "Test@1234",
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
	data := result["data"].(map[string]interface{})
	token := data["access_token"].(string)

	req2, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+newPhone+"&tag=change_phone", nil)
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)
	if resp2.Code != http.StatusOK {
		t.Fatalf("SendSMSCode failed with status %d: %s", resp2.Code, resp2.Body.String())
	}

	changeData := map[string]string{
		"new_phone": newPhone,
		"code":      "invalid_code",
	}
	jsonData, _ = json.Marshal(changeData)
	changeReq, _ := http.NewRequest("POST", "/v1/change-phone", bytes.NewBuffer(jsonData))
	changeReq.Header.Set("Content-Type", "application/json")
	changeReq.Header.Set("Authorization", "Bearer "+token)
	changeResp := httptest.NewRecorder()
	router.ServeHTTP(changeResp, changeReq)

	var codeResult map[string]interface{}
	json.Unmarshal(changeResp.Body.Bytes(), &codeResult)
	if codeResult["code"] == float64(0) {
		t.Error("Expected change phone to fail with invalid code")
	}

	t.Log("Test passed: Change phone failed with invalid code")
}