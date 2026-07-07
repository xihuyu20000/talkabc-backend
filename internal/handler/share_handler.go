package handler

import (
	"backend/internal/middleware"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GenerateShareLink(c *gin.Context) {
	_ = middleware.GetUID(c)

	response.Success(c, gin.H{"link": ""})
}

func GetInviteReward(c *gin.Context) {
	_ = middleware.GetUID(c)

	response.Success(c, gin.H{"reward": 0})
}