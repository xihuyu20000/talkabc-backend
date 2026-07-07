package handler

import (
	"backend/internal/middleware"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetRealVerifyStatus(c *gin.Context) {
	_ = middleware.GetUID(c)

	response.Success(c, gin.H{"status": 0})
}

func ApplyRealVerify(c *gin.Context) {
	_ = middleware.GetUID(c)

	var req struct {
		RealName string `json:"real_name"`
		IDCard   string `json:"id_card"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	response.Success(c, nil)
}

func GetOfficialVerifyStatus(c *gin.Context) {
	_ = middleware.GetUID(c)

	response.Success(c, gin.H{"status": 0})
}

func ApplyOfficialVerify(c *gin.Context) {
	_ = middleware.GetUID(c)

	response.Success(c, nil)
}