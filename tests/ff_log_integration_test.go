package test

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestLoginOperationLog_WriteToPG 验证登录操作日志写入PG表
// 测试登录成功和失败时，operation_logs表是否正确记录日志
func TestLoginOperationLog_WriteToPG(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)
	router.POST("/v1/login/code", handler.LoginByCode)
	router.POST("/v1/login/pwd", handler.LoginByPassword)

	phoneNum := "13900139006"
	password := "Password123"

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

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

		mockSMSGateway.ClearSentMessages()
	})

	t.Run("Step2_LoginSuccess_OperationLogWritten", func(t *testing.T) {
		config.RDB.Del(config.RDB.Context(), "sms_cooldown:"+phoneNum)

		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneNum+"&tag=login", nil)
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

		formData := strings.NewReader("phonenum=" + phoneNum + "&code=" + code)
		loginReq, _ := http.NewRequest("POST", "/v1/login/code", formData)
		loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		loginReq.Header.Set("User-Agent", "TestAgent/1.0")
		loginResp := httptest.NewRecorder()

		router.ServeHTTP(loginResp, loginReq)

		if loginResp.Code != http.StatusOK {
			t.Fatalf("LoginByCode failed with status %d: %s", loginResp.Code, loginResp.Body.String())
		}

		time.Sleep(100 * time.Millisecond)

		user, err := repository.GetUserByPhone(phoneNum)
		if err != nil || user.ID == 0 {
			t.Fatalf("Failed to get user: %v", err)
		}

		logs, err := repository.GetOperationLogsByUserID(user.ID)
		if err != nil {
			t.Fatalf("Failed to query operation logs: %v", err)
		}

		var loginSuccessLog *model.OperationLog
		for i := range logs {
			if logs[i].Operation == "login_code" && logs[i].Success == 1 {
				loginSuccessLog = &logs[i]
				break
			}
		}

		if loginSuccessLog == nil {
			t.Error("Expected login_code success log not found in operation_logs table")
			return
		}

		if loginSuccessLog.UserID != user.ID {
			t.Errorf("Expected user_id %d, got %d", user.ID, loginSuccessLog.UserID)
		}

		if loginSuccessLog.Operation != "login_code" {
			t.Errorf("Expected operation 'login_code', got '%s'", loginSuccessLog.Operation)
		}

		if loginSuccessLog.Success != 1 {
			t.Errorf("Expected success 1, got %d", loginSuccessLog.Success)
		}

		if loginSuccessLog.Detail != "验证码登录成功" {
			t.Errorf("Expected detail '验证码登录成功', got '%s'", loginSuccessLog.Detail)
		}

		t.Logf("Login success operation log verified: ID=%d, Operation=%s, Success=%d",
			loginSuccessLog.ID, loginSuccessLog.Operation, loginSuccessLog.Success)
	})

	t.Run("Step3_LoginFailed_OperationLogWritten", func(t *testing.T) {
		formData := strings.NewReader("phonenum=" + phoneNum + "&pwd=WrongPassword123")
		loginReq, _ := http.NewRequest("POST", "/v1/login/pwd", formData)
		loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		loginResp := httptest.NewRecorder()

		router.ServeHTTP(loginResp, loginReq)

		time.Sleep(100 * time.Millisecond)

		user, err := repository.GetUserByPhone(phoneNum)
		if err != nil || user.ID == 0 {
			t.Fatalf("Failed to get user: %v", err)
		}

		logs, err := repository.GetOperationLogsByUserID(user.ID)
		if err != nil {
			t.Fatalf("Failed to query operation logs: %v", err)
		}

		var loginFailedLog *model.OperationLog
		for i := range logs {
			if logs[i].Operation == "login_password" && logs[i].Success == 0 {
				loginFailedLog = &logs[i]
				break
			}
		}

		if loginFailedLog == nil {
			t.Error("Expected login_password failed log not found in operation_logs table")
			return
		}

		if loginFailedLog.Success != 0 {
			t.Errorf("Expected success 0, got %d", loginFailedLog.Success)
		}

		if loginFailedLog.Detail != "密码错误" {
			t.Errorf("Expected detail '密码错误', got '%s'", loginFailedLog.Detail)
		}

		t.Logf("Login failed operation log verified: ID=%d, Operation=%s, Success=%d, Detail=%s",
			loginFailedLog.ID, loginFailedLog.Operation, loginFailedLog.Success, loginFailedLog.Detail)
	})
}

// TestRegisterOperationLog_WriteToPG 验证注册操作日志写入PG表
func TestRegisterOperationLog_WriteToPG(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)

	phoneNum := "13900139007"
	password := "Password123"

	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	t.Run("RegisterSuccess_OperationLogWritten", func(t *testing.T) {
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

		time.Sleep(100 * time.Millisecond)

		logs, err := repository.GetOperationLogsByOperation("register")
		if err != nil {
			t.Fatalf("Failed to query operation logs: %v", err)
		}

		var registerSuccessLog *model.OperationLog
		for i := range logs {
			if logs[i].Success == 1 && logs[i].Detail == "注册成功" {
				registerSuccessLog = &logs[i]
				break
			}
		}

		if registerSuccessLog == nil {
			t.Error("Expected register success log not found in operation_logs table")
			return
		}

		if registerSuccessLog.Operation != "register" {
			t.Errorf("Expected operation 'register', got '%s'", registerSuccessLog.Operation)
		}

		if registerSuccessLog.Success != 1 {
			t.Errorf("Expected success 1, got %d", registerSuccessLog.Success)
		}

		if registerSuccessLog.Detail != "注册成功" {
			t.Errorf("Expected detail '注册成功', got '%s'", registerSuccessLog.Detail)
		}

		t.Logf("Register success operation log verified: ID=%d, Operation=%s, Success=%d",
			registerSuccessLog.ID, registerSuccessLog.Operation, registerSuccessLog.Success)
	})
}

// TestComplianceLog_WriteToPG 验证等保合规日志写入PG表
// 测试登录成功后，compliance_logs表是否正确记录登录成功事件
func TestComplianceLog_WriteToPG(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	InitTest()

	t.Run("LoginSuccess_ComplianceLogWritten", func(t *testing.T) {
		service.RecordLoginSuccess("compliance-test-1", "trace-1", 1, "uid-compliance-1", "13800138008", "192.168.1.100", "Mozilla/5.0")

		time.Sleep(100 * time.Millisecond)

		logs, total, err := repository.GetComplianceLogsByLogType("login_success", 1, 10)
		if err != nil {
			t.Fatalf("Failed to query compliance logs: %v", err)
		}

		if total < 1 {
			t.Error("Expected at least 1 login_success compliance log")
			return
		}

		var found bool
		for i := range logs {
			if logs[i].PhoneNum == "13800138008" && logs[i].LogType == "login_success" {
				found = true
				t.Logf("Compliance log verified: ID=%d, LogType=%s, Severity=%s, Action=%s",
					logs[i].ID, logs[i].LogType, logs[i].Severity, logs[i].Action)
				break
			}
		}

		if !found {
			t.Error("Expected login_success compliance log for phone 13800138008 not found")
		}
	})

	t.Run("LoginFailure_ComplianceLogWritten", func(t *testing.T) {
		service.RecordLoginFailure("compliance-test-2", "trace-2", "13800138009", "192.168.1.101", "Mozilla/5.0", "验证码错误")

		time.Sleep(100 * time.Millisecond)

		logs, total, err := repository.GetComplianceLogsByLogType("login_failure", 1, 10)
		if err != nil {
			t.Fatalf("Failed to query compliance logs: %v", err)
		}

		if total < 1 {
			t.Error("Expected at least 1 login_failure compliance log")
			return
		}

		var found bool
		for i := range logs {
			if logs[i].PhoneNum == "13800138009" && logs[i].LogType == "login_failure" {
				found = true
				if logs[i].Severity != "WARN" {
					t.Errorf("Expected severity 'WARN', got '%s'", logs[i].Severity)
				}
				t.Logf("Compliance log verified: ID=%d, LogType=%s, Severity=%s, Action=%s",
					logs[i].ID, logs[i].LogType, logs[i].Severity, logs[i].Action)
				break
			}
		}

		if !found {
			t.Error("Expected login_failure compliance log for phone 13800138009 not found")
		}
	})
}

// TestAuditLog_WriteToPG 验证操作审计日志写入PG表
func TestAuditLog_WriteToPG(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	InitTest()

	t.Run("RecordAuditLogSync_LogWritten", func(t *testing.T) {
		err := service.RecordAuditLogSync(service.AuditLogParams{
			RequestID:    "audit-test-1",
			UserID:       1,
			Uid:          "uid-audit-1",
			PhoneNum:     "13800138010",
			IP:           "192.168.1.102",
			Method:       "POST",
			Path:         "/v1/test/audit",
			Operation:    "test_audit",
			ResourceType: "test",
			ResourceID:   "1",
			Action:       "create",
			Result:       1,
			StatusCode:   200,
		})

		if err != nil {
			t.Fatalf("Failed to record audit log: %v", err)
		}

		logs, err := repository.GetAuditLogsByRequestID("audit-test-1")
		if err != nil {
			t.Fatalf("Failed to query audit logs: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 audit log, got %d", len(logs))
			return
		}

		if logs[0].Operation != "test_audit" {
			t.Errorf("Expected operation 'test_audit', got '%s'", logs[0].Operation)
		}

		if logs[0].Result != 1 {
			t.Errorf("Expected result 1, got %d", logs[0].Result)
		}

		t.Logf("Audit log verified: ID=%d, Operation=%s, Result=%d",
			logs[0].ID, logs[0].Operation, logs[0].Result)
	})
}

// TestExceptionLog_WriteToPG 验证异常日志写入PG表
func TestExceptionLog_WriteToPG(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	InitTest()

	t.Run("RecordError_ExceptionLogWritten", func(t *testing.T) {
		service.RecordError("exception-test-1", "trace-exc-1", "测试错误日志", "test_error", "stack trace here", 1, "ORDER-TEST-001")

		time.Sleep(100 * time.Millisecond)

		logs, err := repository.GetExceptionLogsByOrderID("ORDER-TEST-001")
		if err != nil {
			t.Fatalf("Failed to query exception logs: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 exception log, got %d", len(logs))
			return
		}

		if logs[0].Level != "ERROR" {
			t.Errorf("Expected level 'ERROR', got '%s'", logs[0].Level)
		}

		if logs[0].Message != "测试错误日志" {
			t.Errorf("Expected message '测试错误日志', got '%s'", logs[0].Message)
		}

		t.Logf("Exception log verified: ID=%d, Level=%s, Message=%s",
			logs[0].ID, logs[0].Level, logs[0].Message)
	})
}