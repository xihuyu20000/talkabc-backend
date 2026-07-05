package handler

import (
	"backend/pkg/logger"
	"backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SetLogLevelRequest struct {
	Level string `json:"level" binding:"required,oneof=debug info warn error dpanic panic fatal"`
}

// GetLogLevel 获取当前日志级别
// @Summary 获取日志级别
// @Description 获取当前系统的日志级别
// @Tags 系统管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "获取成功，返回当前日志级别"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /system/log-level [get]
func GetLogLevel(c *gin.Context) {
	level := logger.GetLogLevel().String()
	response.Success(c, gin.H{"level": level})
}

// SetLogLevel 设置日志级别
// @Summary 设置日志级别
// @Description 动态调整系统的日志级别，支持 debug、info、warn、error、dpanic、panic、fatal 七种级别
// @Tags 系统管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param body body SetLogLevelRequest true "日志级别参数"
// @Success 200 {object} map[string]interface{} "设置成功，返回新的日志级别"
// @Failure 400 {object} map[string]interface{} "参数错误或无效的日志级别"
// @Failure 500 {object} map[string]interface{} "设置失败"
// @Router /system/log-level [post]
func SetLogLevel(c *gin.Context) {
	var req SetLogLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := logger.SetLogLevel(req.Level); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	logger.Info("Log level updated via API",
		zap.String("level", req.Level),
		zap.String("client_ip", c.ClientIP()),
	)

	response.Success(c, gin.H{"message": "日志级别已更新", "level": req.Level})
}