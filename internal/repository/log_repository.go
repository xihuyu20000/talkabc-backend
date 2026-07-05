package repository

import (
	"backend/internal/config"
	"backend/internal/model"
	"time"

	"github.com/jinzhu/gorm"
)

// ============================================
// 操作审计日志 Repository
// ============================================

// CreateAuditLog 创建操作审计日志
// 【等保合规】记录用户操作行为，支持关联业务数据排查问题
func CreateAuditLog(log *model.AuditLog) error {
	return config.DB.Create(log).Error
}

// CreateAuditLogWithTx 创建操作审计日志（事务支持）
func CreateAuditLogWithTx(tx *gorm.DB, log *model.AuditLog) error {
	return tx.Create(log).Error
}

// GetAuditLogsByUserID 根据用户ID查询审计日志
func GetAuditLogsByUserID(userID uint, page, pageSize int) ([]model.AuditLog, int, error) {
	var logs []model.AuditLog
	var total int

	err := config.DB.Model(&model.AuditLog{}).
		Where("user_id = ?", userID).
		Count(&total).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// GetAuditLogsByRequestID 根据请求ID查询审计日志
func GetAuditLogsByRequestID(requestID string) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := config.DB.Where("request_id = ?", requestID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetAuditLogsByResource 根据资源类型和ID查询审计日志
func GetAuditLogsByResource(resourceType, resourceID string, page, pageSize int) ([]model.AuditLog, int, error) {
	var logs []model.AuditLog
	var total int

	err := config.DB.Model(&model.AuditLog{}).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Count(&total).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// GetAuditLogsByOperation 根据操作类型查询审计日志
func GetAuditLogsByOperation(operation string, startTime, endTime time.Time) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := config.DB.Where("operation = ? AND created_at BETWEEN ? AND ?",
		operation, startTime, endTime).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// DeleteOldAuditLogs 删除指定时间之前的审计日志（数据保留策略：90天）
func DeleteOldAuditLogs(beforeTime time.Time) (int64, error) {
	result := config.DB.Where("created_at < ?", beforeTime).
		Delete(&model.AuditLog{})
	return result.RowsAffected, result.Error
}

// ============================================
// 后台变更日志 Repository
// ============================================

// CreateAdminChangeLog 创建后台变更日志
// 【等保合规】记录管理员/系统后台操作，支持审计追踪
func CreateAdminChangeLog(log *model.AdminChangeLog) error {
	return config.DB.Create(log).Error
}

// CreateAdminChangeLogWithTx 创建后台变更日志（事务支持）
func CreateAdminChangeLogWithTx(tx *gorm.DB, log *model.AdminChangeLog) error {
	return tx.Create(log).Error
}

// GetAdminChangeLogsByOperator 根据操作人ID查询变更日志
func GetAdminChangeLogsByOperator(operatorID uint, page, pageSize int) ([]model.AdminChangeLog, int, error) {
	var logs []model.AdminChangeLog
	var total int

	err := config.DB.Model(&model.AdminChangeLog{}).
		Where("operator_id = ?", operatorID).
		Count(&total).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// GetAdminChangeLogsByModule 根据模块查询变更日志
func GetAdminChangeLogsByModule(module string, page, pageSize int) ([]model.AdminChangeLog, int, error) {
	var logs []model.AdminChangeLog
	var total int

	err := config.DB.Model(&model.AdminChangeLog{}).
		Where("module = ?", module).
		Count(&total).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// GetAdminChangeLogsByTarget 根据目标类型和ID查询变更日志
func GetAdminChangeLogsByTarget(targetType, targetID string, page, pageSize int) ([]model.AdminChangeLog, int, error) {
	var logs []model.AdminChangeLog
	var total int

	err := config.DB.Model(&model.AdminChangeLog{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Count(&total).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// ============================================
// 等保合规日志 Repository
// ============================================

// CreateComplianceLog 创建等保合规日志
// 【等保合规】记录关键安全事件，事务保障，不可删除
func CreateComplianceLog(log *model.ComplianceLog) error {
	return config.DB.Create(log).Error
}

// CreateComplianceLogWithTx 创建等保合规日志（事务支持）
func CreateComplianceLogWithTx(tx *gorm.DB, log *model.ComplianceLog) error {
	return tx.Create(log).Error
}

// GetComplianceLogsByEventID 根据事件ID查询合规日志
func GetComplianceLogsByEventID(eventID string) (*model.ComplianceLog, error) {
	var log model.ComplianceLog
	err := config.DB.Where("event_id = ?", eventID).First(&log).Error
	return &log, err
}

// GetComplianceLogsByLogType 根据日志类型查询合规日志
func GetComplianceLogsByLogType(logType string, page, pageSize int) ([]model.ComplianceLog, int, error) {
	var logs []model.ComplianceLog
	var total int

	err := config.DB.Model(&model.ComplianceLog{}).
		Where("log_type = ?", logType).
		Count(&total).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// GetComplianceLogsBySeverity 根据严重级别查询合规日志
func GetComplianceLogsBySeverity(severity string, startTime, endTime time.Time) ([]model.ComplianceLog, error) {
	var logs []model.ComplianceLog
	err := config.DB.Where("severity = ? AND created_at BETWEEN ? AND ?",
		severity, startTime, endTime).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetComplianceLogsByUserID 根据用户ID查询合规日志
func GetComplianceLogsByUserID(userID uint, page, pageSize int) ([]model.ComplianceLog, int, error) {
	var logs []model.ComplianceLog
	var total int

	err := config.DB.Model(&model.ComplianceLog{}).
		Where("user_id = ?", userID).
		Count(&total).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// MarkComplianceLogVerified 标记合规日志已审核
func MarkComplianceLogVerified(eventID string) error {
	return config.DB.Model(&model.ComplianceLog{}).
		Where("event_id = ?", eventID).
		Update("verified", 1).Error
}

// ============================================
// 异常日志 Repository
// ============================================

// CreateExceptionLog 创建异常日志
// 【系统运维】记录ERROR/WARN级别异常，过滤DEBUG/INFO
func CreateExceptionLog(log *model.ExceptionLog) error {
	return config.DB.Create(log).Error
}

// CreateExceptionLogWithTx 创建异常日志（事务支持）
func CreateExceptionLogWithTx(tx *gorm.DB, log *model.ExceptionLog) error {
	return tx.Create(log).Error
}

// GetExceptionLogsByRequestID 根据请求ID查询异常日志
func GetExceptionLogsByRequestID(requestID string) ([]model.ExceptionLog, error) {
	var logs []model.ExceptionLog
	err := config.DB.Where("request_id = ?", requestID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetExceptionLogsByLevel 根据日志级别查询异常日志
func GetExceptionLogsByLevel(level string, startTime, endTime time.Time) ([]model.ExceptionLog, int, error) {
	var logs []model.ExceptionLog
	var total int

	err := config.DB.Model(&model.ExceptionLog{}).
		Where("level = ? AND created_at BETWEEN ? AND ?",
			level, startTime, endTime).
		Count(&total).
		Order("created_at DESC").
		Find(&logs).Error

	return logs, total, err
}

// GetExceptionLogsByOrderID 根据订单号查询异常日志
func GetExceptionLogsByOrderID(orderID string) ([]model.ExceptionLog, error) {
	var logs []model.ExceptionLog
	err := config.DB.Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetExceptionLogsByUserID 根据用户ID查询异常日志
func GetExceptionLogsByUserID(userID uint, page, pageSize int) ([]model.ExceptionLog, int, error) {
	var logs []model.ExceptionLog
	var total int

	err := config.DB.Model(&model.ExceptionLog{}).
		Where("user_id = ?", userID).
		Count(&total).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// GetExceptionLogsByErrorType 根据错误类型查询异常日志
func GetExceptionLogsByErrorType(errorType string, page, pageSize int) ([]model.ExceptionLog, int, error) {
	var logs []model.ExceptionLog
	var total int

	err := config.DB.Model(&model.ExceptionLog{}).
		Where("error_type = ?", errorType).
		Count(&total).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// DeleteOldExceptionLogs 删除指定时间之前的异常日志（数据保留策略：30天）
func DeleteOldExceptionLogs(beforeTime time.Time) (int64, error) {
	result := config.DB.Where("created_at < ?", beforeTime).
		Delete(&model.ExceptionLog{})
	return result.RowsAffected, result.Error
}

// GetExceptionLogByHash 根据哈希查询异常日志（用于去重检查）
func GetExceptionLogByHash(hash string) (*model.ExceptionLog, error) {
	var log model.ExceptionLog
	err := config.DB.Where("hash = ?", hash).First(&log).Error
	return &log, err
}

// UpdateExceptionLogRetry 更新异常日志重试次数
func UpdateExceptionLogRetry(hash string, retryCount int) error {
	return config.DB.Model(&model.ExceptionLog{}).
		Where("hash = ?", hash).
		Update("retry_count", retryCount).Error
}