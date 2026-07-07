package handler

import (
	"backend/internal/middleware"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetNotificationList(c *gin.Context) {
	_ = middleware.GetUID(c)

	data := make([]interface{}, 0)
	response.Success(c, gin.H{"data": data, "total": 0, "page": 1, "size": 20})
}

func MarkNotificationsRead(c *gin.Context) {
	_ = middleware.GetUID(c)

	response.Success(c, nil)
}

func GetNotificationCount(c *gin.Context) {
	_ = middleware.GetUID(c)

	response.Success(c, gin.H{"count": 0})
}