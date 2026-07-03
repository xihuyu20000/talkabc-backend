package middleware

import (
	"backend/pkg/logger"
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := generateRequestID()
		c.Set(RequestIDKey, requestID)

		reqCtx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(reqCtx)

		reqLogger := logger.WithContext(reqCtx)

		reqLogger.Infof("Request start - method: %s, path: %s, client_ip: %s, user_agent: %s",
			c.Request.Method, c.Request.URL.Path, c.ClientIP(), c.Request.UserAgent())

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if status >= 500 {
			reqLogger.Errorf("Request end - method: %s, path: %s, status: %d, latency: %s, bytes_written: %d",
				c.Request.Method, c.Request.URL.Path, status, latency, c.Writer.Size())
		} else if status >= 400 {
			reqLogger.Warnf("Request end - method: %s, path: %s, status: %d, latency: %s, bytes_written: %d",
				c.Request.Method, c.Request.URL.Path, status, latency, c.Writer.Size())
		} else {
			reqLogger.Infof("Request end - method: %s, path: %s, status: %d, latency: %s, bytes_written: %d",
				c.Request.Method, c.Request.URL.Path, status, latency, c.Writer.Size())
		}
	}
}

func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		return requestID.(string)
	}
	return ""
}