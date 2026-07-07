package handler

import (
	"backend/internal/middleware"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetBlacklist(c *gin.Context) {
	_ = middleware.GetUID(c)

	data := make([]interface{}, 0)
	response.Success(c, gin.H{"data": data, "total": 0, "page": 1, "size": 20})
}

func ToggleBlacklist(c *gin.Context) {
	_ = middleware.GetUID(c)
	_ = c.Param("uid")

	response.Success(c, nil)
}

func GetLanguage(c *gin.Context) {
	_ = middleware.GetUID(c)

	response.Success(c, gin.H{"language": "zh"})
}

func UpdateLanguage(c *gin.Context) {
	_ = middleware.GetUID(c)

	var req struct {
		Language string `json:"language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	response.Success(c, nil)
}

func GetPrivacySettings(c *gin.Context) {
	_ = middleware.GetUID(c)

	response.Success(c, gin.H{})
}

func UpdatePrivacySettings(c *gin.Context) {
	_ = middleware.GetUID(c)

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	response.Success(c, nil)
}