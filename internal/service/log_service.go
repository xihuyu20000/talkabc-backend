package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/pkg/logger"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// generateLogHash 生成日志哈希（用于去重）
func generateLogHash(message, errorType, stackTrace string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s|%s|%s", message, errorType, stackTrace)))
	return hex.EncodeToString(hash.Sum(nil))
}

// ============================================
// 操作审计日志 Service
// ============================================

// AuditLogParams 操作审计日志参数
type AuditLogParams struct {
	RequestID   string
	TraceID     string
	UserID      uint
	Uid         string
	PhoneNum    string
	IP          string
	UA          string
	Method      string
	Path        string
	Operation   string
	ResourceType string
	ResourceID  string
	Action      string
	BeforeData  map[string]interface{}
	AfterData   map[string]interface{}
	Result      int
	ErrorMessage string
	DurationMs  int
	StatusCode  int
	Extra       map[string]interface{}
}

// RecordAuditLog 记录操作审计日志
// 【等保合规】记录用户操作行为，支持关联业务数据排查问题
func RecordAuditLog(params AuditLogParams) {
	log := &model.AuditLog{
		RequestID:   params.RequestID,
		TraceID:     params.TraceID,
		UserID:      params.UserID,
		Uid:         params.Uid,
		PhoneNum:    params.PhoneNum,
		IP:          params.IP,
		UA:          params.UA,
		Method:      params.Method,
		Path:        params.Path,
		Operation:   params.Operation,
		ResourceType: params.ResourceType,
		ResourceID:  params.ResourceID,
		Action:      params.Action,
		BeforeData:  params.BeforeData,
		AfterData:   params.AfterData,
		Result:      params.Result,
		ErrorMessage: params.ErrorMessage,
		DurationMs:  params.DurationMs,
		StatusCode:  params.StatusCode,
		Extra:       params.Extra,
	}

	go func() {
		if err := repository.CreateAuditLog(log); err != nil {
			logger.Error("Failed to write audit log", zap.Error(err))
		}
	}()
}

// RecordAuditLogSync 同步记录操作审计日志（事务场景）
func RecordAuditLogSync(params AuditLogParams) error {
	log := &model.AuditLog{
		RequestID:   params.RequestID,
		TraceID:     params.TraceID,
		UserID:      params.UserID,
		Uid:         params.Uid,
		PhoneNum:    params.PhoneNum,
		IP:          params.IP,
		UA:          params.UA,
		Method:      params.Method,
		Path:        params.Path,
		Operation:   params.Operation,
		ResourceType: params.ResourceType,
		ResourceID:  params.ResourceID,
		Action:      params.Action,
		BeforeData:  params.BeforeData,
		AfterData:   params.AfterData,
		Result:      params.Result,
		ErrorMessage: params.ErrorMessage,
		DurationMs:  params.DurationMs,
		StatusCode:  params.StatusCode,
		Extra:       params.Extra,
	}

	return repository.CreateAuditLog(log)
}

// ============================================
// 后台变更日志 Service
// ============================================

// AdminChangeLogParams 后台变更日志参数
type AdminChangeLogParams struct {
	RequestID     string
	TraceID       string
	OperatorID    uint
	OperatorName  string
	OperatorRole  string
	IP            string
	UA            string
	Module        string
	Action        string
	TargetType    string
	TargetID      string
	TargetName    string
	ChangeContent map[string]interface{}
	BeforeData    map[string]interface{}
	AfterData     map[string]interface{}
	Reason        string
	Result        int
	Extra         map[string]interface{}
}

// RecordAdminChangeLog 记录后台变更日志
// 【等保合规】记录管理员/系统后台操作，支持审计追踪
func RecordAdminChangeLog(params AdminChangeLogParams) {
	log := &model.AdminChangeLog{
		RequestID:     params.RequestID,
		TraceID:       params.TraceID,
		OperatorID:    params.OperatorID,
		OperatorName:  params.OperatorName,
		OperatorRole:  params.OperatorRole,
		IP:            params.IP,
		UA:            params.UA,
		Module:        params.Module,
		Action:        params.Action,
		TargetType:    params.TargetType,
		TargetID:      params.TargetID,
		TargetName:    params.TargetName,
		ChangeContent: params.ChangeContent,
		BeforeData:    params.BeforeData,
		AfterData:     params.AfterData,
		Reason:        params.Reason,
		Result:        params.Result,
		Extra:         params.Extra,
	}

	go func() {
		if err := repository.CreateAdminChangeLog(log); err != nil {
			logger.Error("Failed to write admin change log", zap.Error(err))
		}
	}()
}

// RecordAdminChangeLogSync 同步记录后台变更日志（事务场景）
func RecordAdminChangeLogSync(params AdminChangeLogParams) error {
	log := &model.AdminChangeLog{
		RequestID:     params.RequestID,
		TraceID:       params.TraceID,
		OperatorID:    params.OperatorID,
		OperatorName:  params.OperatorName,
		OperatorRole:  params.OperatorRole,
		IP:            params.IP,
		UA:            params.UA,
		Module:        params.Module,
		Action:        params.Action,
		TargetType:    params.TargetType,
		TargetID:      params.TargetID,
		TargetName:    params.TargetName,
		ChangeContent: params.ChangeContent,
		BeforeData:    params.BeforeData,
		AfterData:     params.AfterData,
		Reason:        params.Reason,
		Result:        params.Result,
		Extra:         params.Extra,
	}

	return repository.CreateAdminChangeLog(log)
}

// ============================================
// 等保合规日志 Service
// ============================================

// ComplianceLogParams 等保合规日志参数
type ComplianceLogParams struct {
	RequestID string
	TraceID   string
	UserID    uint
	Uid       string
	PhoneNum  string
	IP        string
	UserAgent string
	Action    string
	Resource  string
	Permission string
	Result    string
	Detail    map[string]interface{}
	RawLog    string
}

// LogType 等保合规日志类型
const (
	ComplianceLogTypeLoginSuccess      = "login_success"
	ComplianceLogTypeLoginFailure      = "login_failure"
	ComplianceLogTypeLogout            = "logout"
	ComplianceLogTypePrivilegeEscalation = "privilege_escalation"
	ComplianceLogTypeDataAccess        = "data_access"
	ComplianceLogTypeDataExport        = "data_export"
	ComplianceLogTypeConfigurationChange = "configuration_change"
	ComplianceLogTypeSecurityIncident = "security_incident"
	ComplianceLogTypeAccessDenied      = "access_denied"
)

// Severity 严重级别
const (
	SeverityInfo    = "INFO"
	SeverityWarn    = "WARN"
	SeverityError   = "ERROR"
	SeverityCritical = "CRITICAL"
)

// RecordComplianceLog 记录等保合规日志
// 【等保合规】记录关键安全事件，事务保障，不可删除
// 此函数会自动生成唯一的event_id
func RecordComplianceLog(logType, severity string, params ComplianceLogParams) {
	log := &model.ComplianceLog{
		LogType:   logType,
		Severity:  severity,
		EventID:   uuid.New().String(),
		RequestID: params.RequestID,
		TraceID:   params.TraceID,
		UserID:    params.UserID,
		Uid:       params.Uid,
		PhoneNum:  params.PhoneNum,
		IP:        params.IP,
		UserAgent: params.UserAgent,
		Action:    params.Action,
		Resource:  params.Resource,
		Permission: params.Permission,
		Result:    params.Result,
		Detail:    params.Detail,
		RawLog:    params.RawLog,
	}

	go func() {
		if err := repository.CreateComplianceLog(log); err != nil {
			logger.Error("Failed to write compliance log", zap.Error(err))
		}
	}()
}

// RecordComplianceLogSync 同步记录等保合规日志（事务场景）
func RecordComplianceLogSync(logType, severity string, params ComplianceLogParams) error {
	log := &model.ComplianceLog{
		LogType:   logType,
		Severity:  severity,
		EventID:   uuid.New().String(),
		RequestID: params.RequestID,
		TraceID:   params.TraceID,
		UserID:    params.UserID,
		Uid:       params.Uid,
		PhoneNum:  params.PhoneNum,
		IP:        params.IP,
		UserAgent: params.UserAgent,
		Action:    params.Action,
		Resource:  params.Resource,
		Permission: params.Permission,
		Result:    params.Result,
		Detail:    params.Detail,
		RawLog:    params.RawLog,
	}

	return repository.CreateComplianceLog(log)
}

// RecordLoginSuccess 记录登录成功事件
func RecordLoginSuccess(requestID, traceID string, userID uint, uid, phoneNum, ip, userAgent string) {
	RecordComplianceLog(
		ComplianceLogTypeLoginSuccess,
		SeverityInfo,
		ComplianceLogParams{
			RequestID: requestID,
			TraceID:   traceID,
			UserID:    userID,
			Uid:       uid,
			PhoneNum:  phoneNum,
			IP:        ip,
			UserAgent: userAgent,
			Action:    "用户登录成功",
			Resource:  "/api/v1/auth/login",
			Permission: "auth:login",
			Result:    "SUCCESS",
		},
	)
}

// RecordLoginFailure 记录登录失败事件
func RecordLoginFailure(requestID, traceID string, phoneNum, ip, userAgent, reason string) {
	RecordComplianceLog(
		ComplianceLogTypeLoginFailure,
		SeverityWarn,
		ComplianceLogParams{
			RequestID: requestID,
			TraceID:   traceID,
			PhoneNum:  phoneNum,
			IP:        ip,
			UserAgent: userAgent,
			Action:    "用户登录失败",
			Resource:  "/api/v1/auth/login",
			Permission: "auth:login",
			Result:    "FAILURE",
			Detail: map[string]interface{}{
				"reason": reason,
			},
		},
	)
}

// RecordAccessDenied 记录访问拒绝事件
func RecordAccessDenied(requestID, traceID string, userID uint, ip, resource, permission string) {
	RecordComplianceLog(
		ComplianceLogTypeAccessDenied,
		SeverityWarn,
		ComplianceLogParams{
			RequestID: requestID,
			TraceID:   traceID,
			UserID:    userID,
			IP:        ip,
			Action:    "访问被拒绝",
			Resource:  resource,
			Permission: permission,
			Result:    "DENY",
		},
	)
}

// ============================================
// 异常日志 Service
// ============================================

// ExceptionLogParams 异常日志参数
type ExceptionLogParams struct {
	RequestID  string
	TraceID    string
	Level      string
	LoggerName string
	Message    string
	ErrorType  string
	StackTrace string
	UserID     uint
	Uid        string
	IP         string
	Path       string
	Method     string
	OrderID    string
	BusinessID string
	DurationMs int
	RetryCount int
	Extra      map[string]interface{}
}

// RecordExceptionLog 记录异常日志
// 【系统运维】记录ERROR/WARN级别异常，过滤DEBUG/INFO
// 支持自动去重，相同的异常不会重复记录
func RecordExceptionLog(params ExceptionLogParams) {
	hash := generateLogHash(params.Message, params.ErrorType, params.StackTrace)

	existingLog, err := repository.GetExceptionLogByHash(hash)
	if err == nil && existingLog.ID > 0 {
		go func() {
			_ = repository.UpdateExceptionLogRetry(hash, existingLog.RetryCount+1)
		}()
		return
	}

	log := &model.ExceptionLog{
		RequestID:  params.RequestID,
		TraceID:    params.TraceID,
		Level:      params.Level,
		LoggerName: params.LoggerName,
		Message:    params.Message,
		ErrorType:  params.ErrorType,
		StackTrace: params.StackTrace,
		UserID:     params.UserID,
		Uid:        params.Uid,
		IP:         params.IP,
		Path:       params.Path,
		Method:     params.Method,
		OrderID:    params.OrderID,
		BusinessID: params.BusinessID,
		DurationMs: params.DurationMs,
		RetryCount: params.RetryCount,
		Extra:      params.Extra,
		Hash:       hash,
	}

	go func() {
		if err := repository.CreateExceptionLog(log); err != nil {
			logger.Error("Failed to write exception log", zap.Error(err))
		}
	}()
}

// RecordExceptionLogSync 同步记录异常日志（事务场景）
func RecordExceptionLogSync(params ExceptionLogParams) error {
	hash := generateLogHash(params.Message, params.ErrorType, params.StackTrace)

	existingLog, err := repository.GetExceptionLogByHash(hash)
	if err == nil && existingLog.ID > 0 {
		return repository.UpdateExceptionLogRetry(hash, existingLog.RetryCount+1)
	}

	log := &model.ExceptionLog{
		RequestID:  params.RequestID,
		TraceID:    params.TraceID,
		Level:      params.Level,
		LoggerName: params.LoggerName,
		Message:    params.Message,
		ErrorType:  params.ErrorType,
		StackTrace: params.StackTrace,
		UserID:     params.UserID,
		Uid:        params.Uid,
		IP:         params.IP,
		Path:       params.Path,
		Method:     params.Method,
		OrderID:    params.OrderID,
		BusinessID: params.BusinessID,
		DurationMs: params.DurationMs,
		RetryCount: params.RetryCount,
		Extra:      params.Extra,
		Hash:       hash,
	}

	return repository.CreateExceptionLog(log)
}

// RecordError 记录错误日志（便捷方法）
func RecordError(requestID, traceID, message, errorType, stackTrace string, userID uint, orderID string) {
	RecordExceptionLog(ExceptionLogParams{
		RequestID:  requestID,
		TraceID:    traceID,
		Level:      "ERROR",
		LoggerName: "system",
		Message:    message,
		ErrorType:  errorType,
		StackTrace: stackTrace,
		UserID:     userID,
		OrderID:    orderID,
	})
}

// RecordWarn 记录警告日志（便捷方法）
func RecordWarn(requestID, traceID, message, errorType string, userID uint) {
	RecordExceptionLog(ExceptionLogParams{
		RequestID:  requestID,
		TraceID:    traceID,
		Level:      "WARN",
		LoggerName: "system",
		Message:    message,
		ErrorType:  errorType,
		UserID:     userID,
	})
}

// ============================================
// 日志清理任务
// ============================================

// CleanupOldLogs 清理过期日志
// 按数据保留策略自动清理：
// - audit_logs: 90天
// - exception_logs: 30天
// - admin_change_logs: 永久保留
// - compliance_logs: 永久保留（等保要求）
func CleanupOldLogs() {
	auditBefore := time.Now().Add(-90 * 24 * time.Hour)
	exceptionBefore := time.Now().Add(-30 * 24 * time.Hour)

	auditDeleted, auditErr := repository.DeleteOldAuditLogs(auditBefore)
	if auditErr != nil {
		logger.Error("Failed to clean up old audit logs", zap.Error(auditErr))
	} else {
		logger.Info("Cleaned up old audit logs", zap.Int64("deleted", auditDeleted))
	}

	exceptionDeleted, exceptionErr := repository.DeleteOldExceptionLogs(exceptionBefore)
	if exceptionErr != nil {
		logger.Error("Failed to clean up old exception logs", zap.Error(exceptionErr))
	} else {
		logger.Info("Cleaned up old exception logs", zap.Int64("deleted", exceptionDeleted))
	}
}