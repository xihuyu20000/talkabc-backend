package test

import (
	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"testing"
	"time"
)

func TestAuditLog_WriteAndQuery(t *testing.T) {
	InitTest()

	t.Run("Create and query audit log", func(t *testing.T) {
		log := &model.AuditLog{
			RequestID:    "test-request-id-1",
			TraceID:      "test-trace-id-1",
			UserID:       1,
			Uid:          "test-uid-1",
			PhoneNum:     "13800138001",
			IP:           "192.168.1.1",
			UA:           "Mozilla/5.0 (Test)",
			Method:       "POST",
			Path:         "/v1/login",
			Operation:    "user_login",
			ResourceType: "user",
			ResourceID:   "1",
			Action:       "create",
			Result:       1,
			StatusCode:   200,
			DurationMs:   50,
		}

		err := repository.CreateAuditLog(log)
		if err != nil {
			t.Fatalf("Failed to create audit log: %v", err)
		}

		if log.ID == 0 {
			t.Error("Audit log ID should be set after creation")
		}

		logs, err := repository.GetAuditLogsByRequestID("test-request-id-1")
		if err != nil {
			t.Fatalf("Failed to query audit logs: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 audit log, got %d", len(logs))
		}

		if logs[0].Operation != "user_login" {
			t.Errorf("Expected operation 'user_login', got '%s'", logs[0].Operation)
		}

		if logs[0].UserID != 1 {
			t.Errorf("Expected user_id 1, got %d", logs[0].UserID)
		}
	})

	t.Run("Query audit logs by user ID", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			log := &model.AuditLog{
				RequestID:    "test-request-id-" + string(rune('0'+i)),
				UserID:       2,
				Uid:          "test-uid-2",
				Operation:    "user_update",
				ResourceType: "user",
				ResourceID:   "2",
				Action:       "update",
				Result:       1,
			}
			_ = repository.CreateAuditLog(log)
		}

		logs, total, err := repository.GetAuditLogsByUserID(2, 1, 10)
		if err != nil {
			t.Fatalf("Failed to query audit logs by user ID: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}

		if len(logs) != 3 {
			t.Errorf("Expected 3 audit logs, got %d", len(logs))
		}
	})

	t.Run("Delete old audit logs", func(t *testing.T) {
		log := &model.AuditLog{
			RequestID:    "test-old-request",
			UserID:       1,
			Operation:    "user_login",
			ResourceType: "user",
			ResourceID:   "1",
			Action:       "create",
			Result:       1,
		}
		_ = repository.CreateAuditLog(log)

		_, _ = repository.DeleteOldAuditLogs(time.Now().Add(1 * time.Second))

		logs, total, err := repository.GetAuditLogsByUserID(1, 1, 10)
		if err != nil {
			t.Fatalf("Failed to query audit logs after delete: %v", err)
		}

		if total > 0 && logs[0].RequestID == "test-old-request" {
			t.Error("Old audit log should have been deleted")
		}
	})
}

func TestAdminChangeLog_WriteAndQuery(t *testing.T) {
	InitTest()

	t.Run("Create and query admin change log", func(t *testing.T) {
		log := &model.AdminChangeLog{
			OperatorID:     1,
			OperatorName:   "admin",
			Module:         "user",
			Action:         "update",
			TargetType:     "user",
			TargetID:       "1",
			ChangeContent:  model.JSONMap{"field": "name", "old": "old", "new": "new"},
			BeforeData:     model.JSONMap{"name": "old"},
			AfterData:      model.JSONMap{"name": "new"},
			Result:         1,
			IP:             "192.168.1.100",
			UA:             "Admin/1.0",
		}

		err := repository.CreateAdminChangeLog(log)
		if err != nil {
			t.Fatalf("Failed to create admin change log: %v", err)
		}

		if log.ID == 0 {
			t.Error("Admin change log ID should be set after creation")
		}

		logs, total, err := repository.GetAdminChangeLogsByOperator(1, 1, 10)
		if err != nil {
			t.Fatalf("Failed to query admin change logs: %v", err)
		}

		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}

		if logs[0].Module != "user" {
			t.Errorf("Expected module 'user', got '%s'", logs[0].Module)
		}
	})

	t.Run("Query admin change logs by module", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			log := &model.AdminChangeLog{
				OperatorID:   1,
				OperatorName: "admin",
				Module:       "config",
				Action:       "update",
				TargetType:   "config",
				TargetID:     string(rune('0' + i)),
				Result:       1,
			}
			_ = repository.CreateAdminChangeLog(log)
		}

		logs, total, err := repository.GetAdminChangeLogsByModule("config", 1, 10)
		if err != nil {
			t.Fatalf("Failed to query admin change logs by module: %v", err)
		}

		if total != 2 {
			t.Errorf("Expected total 2, got %d", total)
		}

		if len(logs) != 2 {
			t.Errorf("Expected 2 admin change logs, got %d", len(logs))
		}
	})
}

func TestComplianceLog_WriteAndQuery(t *testing.T) {
	InitTest()

	t.Run("Create and query compliance log", func(t *testing.T) {
		log := &model.ComplianceLog{
			EventID:    "event-20240101-001",
			LogType:    "login_success",
			UserID:     1,
			Uid:        "uid-001",
			PhoneNum:   "13800138001",
			IP:         "192.168.1.1",
			UserAgent:  "Mozilla/5.0 (Test)",
			Severity:   "info",
			Action:     "用户登录成功",
			Result:     "success",
		}

		err := repository.CreateComplianceLog(log)
		if err != nil {
			t.Fatalf("Failed to create compliance log: %v", err)
		}

		if log.ID == 0 {
			t.Error("Compliance log ID should be set after creation")
		}

		retrievedLog, err := repository.GetComplianceLogsByEventID("event-20240101-001")
		if err != nil {
			t.Fatalf("Failed to query compliance log by event ID: %v", err)
		}

		if retrievedLog.LogType != "login_success" {
			t.Errorf("Expected log_type 'login_success', got '%s'", retrievedLog.LogType)
		}

		if retrievedLog.UserID != 1 {
			t.Errorf("Expected user_id 1, got %d", retrievedLog.UserID)
		}
	})

	t.Run("Query compliance logs by severity", func(t *testing.T) {
		log1 := &model.ComplianceLog{
			EventID:   "event-sec-1",
			LogType:   "login_failure",
			PhoneNum:  "13800138002",
			IP:        "192.168.1.2",
			Severity:  "warning",
			Action:    "登录失败",
			Result:    "failed",
		}
		_ = repository.CreateComplianceLog(log1)

		log2 := &model.ComplianceLog{
			EventID:   "event-sec-2",
			LogType:   "access_denied",
			IP:        "192.168.1.3",
			Severity:  "critical",
			Action:    "访问被拒绝",
			Result:    "failed",
		}
		_ = repository.CreateComplianceLog(log2)

		logs, err := repository.GetComplianceLogsBySeverity("warning", time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
		if err != nil {
			t.Fatalf("Failed to query compliance logs by severity: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 warning log, got %d", len(logs))
		}

		if logs[0].Severity != "warning" {
			t.Errorf("Expected severity 'warning', got '%s'", logs[0].Severity)
		}
	})

	t.Run("Mark compliance log as verified", func(t *testing.T) {
		log := &model.ComplianceLog{
			EventID:   "event-verify-1",
			LogType:   "privilege_escalation",
			Severity:  "critical",
			Action:    "权限变更",
			Result:    "success",
			Verified:  0,
		}
		_ = repository.CreateComplianceLog(log)

		err := repository.MarkComplianceLogVerified("event-verify-1")
		if err != nil {
			t.Fatalf("Failed to mark compliance log as verified: %v", err)
		}

		retrievedLog, err := repository.GetComplianceLogsByEventID("event-verify-1")
		if err != nil {
			t.Fatalf("Failed to query compliance log: %v", err)
		}

		if retrievedLog.Verified != 1 {
			t.Error("Compliance log should be verified")
		}
	})
}

func TestExceptionLog_WriteAndQuery(t *testing.T) {
	InitTest()

	config.DB.Delete(&model.ExceptionLog{})

	t.Run("Create and query exception log", func(t *testing.T) {
		log := &model.ExceptionLog{
			RequestID:   "req-exception-1",
			TraceID:     "trace-exception-1",
			Level:       "ERROR",
			LoggerName:  "service.payment",
			Message:     "支付失败：余额不足",
			ErrorType:   "payment_error",
			StackTrace:  "stack trace here",
			UserID:      1,
			Uid:         "uid-exc-1",
			IP:          "192.168.1.1",
			Path:        "/v1/payment",
			Method:      "POST",
			OrderID:     "ORDER-20240101-001",
			BusinessID:  "biz-001",
			RetryCount:  1,
		}

		err := repository.CreateExceptionLog(log)
		if err != nil {
			t.Fatalf("Failed to create exception log: %v", err)
		}

		if log.ID == 0 {
			t.Error("Exception log ID should be set after creation")
		}

		logs, err := repository.GetExceptionLogsByRequestID("req-exception-1")
		if err != nil {
			t.Fatalf("Failed to query exception logs: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 exception log, got %d", len(logs))
		}

		if logs[0].Level != "ERROR" {
			t.Errorf("Expected level 'ERROR', got '%s'", logs[0].Level)
		}

		if logs[0].OrderID != "ORDER-20240101-001" {
			t.Errorf("Expected order_id 'ORDER-20240101-001', got '%s'", logs[0].OrderID)
		}
	})

	t.Run("Query exception logs by order ID", func(t *testing.T) {
		log1 := &model.ExceptionLog{
			RequestID:   "req-order-1",
			Level:       "ERROR",
			Message:     "订单处理失败",
			ErrorType:   "order_error",
			OrderID:     "ORDER-20240101-002",
			UserID:      2,
		}
		_ = repository.CreateExceptionLog(log1)

		log2 := &model.ExceptionLog{
			RequestID:   "req-order-2",
			Level:       "WARN",
			Message:     "订单超时警告",
			ErrorType:   "order_warn",
			OrderID:     "ORDER-20240101-002",
			UserID:      2,
		}
		_ = repository.CreateExceptionLog(log2)

		logs, err := repository.GetExceptionLogsByOrderID("ORDER-20240101-002")
		if err != nil {
			t.Fatalf("Failed to query exception logs by order ID: %v", err)
		}

		if len(logs) != 2 {
			t.Errorf("Expected 2 exception logs for order, got %d", len(logs))
		}
	})

	t.Run("Query exception logs by level", func(t *testing.T) {
		log1 := &model.ExceptionLog{
			RequestID:   "req-level-1",
			Level:       "ERROR",
			Message:     "严重错误",
			ErrorType:   "system_error",
		}
		_ = repository.CreateExceptionLog(log1)

		log2 := &model.ExceptionLog{
			RequestID:   "req-level-2",
			Level:       "WARN",
			Message:     "警告信息",
			ErrorType:   "system_warn",
		}
		_ = repository.CreateExceptionLog(log2)

		logs, _, err := repository.GetExceptionLogsByLevel("ERROR", time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
		if err != nil {
			t.Fatalf("Failed to query exception logs by level: %v", err)
		}

		found := false
		for _, log := range logs {
			if log.RequestID == "req-level-1" && log.Level == "ERROR" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected ERROR log with request_id 'req-level-1' not found")
		}
	})

	t.Run("Delete old exception logs", func(t *testing.T) {
		log := &model.ExceptionLog{
			RequestID:   "req-old-exc",
			Level:       "ERROR",
			Message:     "旧错误日志",
			ErrorType:   "system_error",
		}
		_ = repository.CreateExceptionLog(log)

		_, err := repository.DeleteOldExceptionLogs(time.Now().Add(1 * time.Second))
		if err != nil {
			t.Fatalf("Failed to delete old exception logs: %v", err)
		}

		logs, err := repository.GetExceptionLogsByRequestID("req-old-exc")
		if err != nil {
			t.Fatalf("Failed to query exception logs after delete: %v", err)
		}

		if len(logs) != 0 {
			t.Error("Expected exception log to be deleted")
		}
	})

	t.Run("Exception log deduplication", func(t *testing.T) {
		log1 := &model.ExceptionLog{
			RequestID:   "req-dedup-1",
			Level:       "ERROR",
			Message:     "重复错误",
			ErrorType:   "duplicate_error",
			StackTrace:  "same stack trace",
			Hash:        "test-hash-1",
			RetryCount:  1,
		}
		_ = repository.CreateExceptionLog(log1)

		retrieved, err := repository.GetExceptionLogByHash("test-hash-1")
		if err != nil {
			t.Fatalf("Failed to get exception log by hash: %v", err)
		}

		if retrieved.ID == 0 {
			t.Error("Exception log should be found by hash")
		}

		err = repository.UpdateExceptionLogRetry("test-hash-1", 2)
		if err != nil {
			t.Fatalf("Failed to update retry count: %v", err)
		}

		retrieved, _ = repository.GetExceptionLogByHash("test-hash-1")
		if retrieved.RetryCount != 2 {
			t.Errorf("Expected retry_count 2, got %d", retrieved.RetryCount)
		}
	})
}

func TestServiceLayer_LogRecording(t *testing.T) {
	InitTest()

	t.Run("Record audit log via service", func(t *testing.T) {
		service.RecordAuditLog(service.AuditLogParams{
			RequestID:    "service-audit-1",
			TraceID:      "service-trace-1",
			UserID:       1,
			Uid:          "service-uid-1",
			PhoneNum:     "13800138003",
			IP:           "192.168.1.5",
			UA:           "Service/1.0",
			Method:       "POST",
			Path:         "/v1/test",
			Operation:    "service_test",
			ResourceType: "test",
			ResourceID:   "1",
			Action:       "create",
			Result:       1,
			StatusCode:   200,
			DurationMs:   100,
		})

		time.Sleep(100 * time.Millisecond)

		logs, err := repository.GetAuditLogsByRequestID("service-audit-1")
		if err != nil {
			t.Fatalf("Failed to query audit log: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 audit log, got %d", len(logs))
		}

		if logs[0].Operation != "service_test" {
			t.Errorf("Expected operation 'service_test', got '%s'", logs[0].Operation)
		}
	})

	t.Run("Record compliance log via service", func(t *testing.T) {
		service.RecordLoginSuccess("service-compliance-1", "service-trace-2", 1, "compliance-uid-1", "13800138004", "192.168.1.6", "Mozilla/5.0")

		time.Sleep(100 * time.Millisecond)

		logs, total, err := repository.GetComplianceLogsByLogType("login_success", 1, 10)
		if err != nil {
			t.Fatalf("Failed to query compliance log: %v", err)
		}

		if total < 1 {
			t.Error("Expected at least 1 login_success compliance log")
		}

		found := false
		for _, log := range logs {
			if log.PhoneNum == "13800138004" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Login success log should be recorded")
		}
	})

	t.Run("Record exception log via service", func(t *testing.T) {
		service.RecordError("service-exc-1", "service-trace-3", "服务层错误", "service_error", "stack trace", 1, "ORDER-SERVICE-001")

		time.Sleep(100 * time.Millisecond)

		logs, err := repository.GetExceptionLogsByOrderID("ORDER-SERVICE-001")
		if err != nil {
			t.Fatalf("Failed to query exception log: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 exception log, got %d", len(logs))
		}

		if logs[0].Level != "ERROR" {
			t.Errorf("Expected level 'ERROR', got '%s'", logs[0].Level)
		}
	})

	t.Run("Record warn log via service", func(t *testing.T) {
		service.RecordWarn("service-warn-1", "service-trace-4", "服务层警告", "service_warn", 1)

		time.Sleep(100 * time.Millisecond)

		_, total, err := repository.GetExceptionLogsByLevel("WARN", time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
		if err != nil {
			t.Fatalf("Failed to query warn log: %v", err)
		}

		if total < 1 {
			t.Error("Expected at least 1 WARN log")
		}
	})
}

func TestTransactionSupport(t *testing.T) {
	InitTest()

	t.Run("Audit log with transaction", func(t *testing.T) {
		tx := config.DB.Begin()
		if tx.Error != nil {
			t.Fatalf("Failed to begin transaction: %v", tx.Error)
		}

		log := &model.AuditLog{
			RequestID:    "tx-audit-1",
			Operation:    "tx_test",
			ResourceType: "test",
			ResourceID:   "tx-1",
			Action:       "create",
			Result:       1,
		}

		err := repository.CreateAuditLogWithTx(tx, log)
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to create audit log with transaction: %v", err)
		}

		err = tx.Commit().Error
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		logs, err := repository.GetAuditLogsByRequestID("tx-audit-1")
		if err != nil {
			t.Fatalf("Failed to query audit log: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 audit log after transaction commit, got %d", len(logs))
		}
	})

	t.Run("Compliance log with transaction rollback", func(t *testing.T) {
		tx := config.DB.Begin()
		if tx.Error != nil {
			t.Fatalf("Failed to begin transaction: %v", tx.Error)
		}

		log := &model.ComplianceLog{
			EventID:  "tx-compliance-rollback",
			LogType:  "test_event",
			Severity: "info",
			Action:   "测试事务回滚",
			Result:   "success",
		}

		err := repository.CreateComplianceLogWithTx(tx, log)
		if err != nil {
			t.Fatalf("Failed to create compliance log with transaction: %v", err)
		}

		err = tx.Rollback().Error
		if err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		retrievedLog, err := repository.GetComplianceLogsByEventID("tx-compliance-rollback")
		if err != nil && err.Error() != "record not found" {
			t.Fatalf("Unexpected error: %v", err)
		}

		if retrievedLog.ID != 0 {
			t.Error("Compliance log should not exist after transaction rollback")
		}
	})
}