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

func GetLogLevel(c *gin.Context) {
	level := logger.GetLogLevel().String()
	response.Success(c, gin.H{"level": level})
}

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