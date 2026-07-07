package handler

import (
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetAdList(c *gin.Context) {
	list, err := service.GetAdList()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, list)
}

func TrackAdClick(c *gin.Context) {
	adID := c.PostForm("ad_id")

	logger.Infof("[Handler] TrackAdClick - AdID: %s", adID)

	err := service.TrackAdClick(adID)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

func TrackAdImpression(c *gin.Context) {
	adID := c.PostForm("ad_id")

	logger.Infof("[Handler] TrackAdImpression - AdID: %s", adID)

	err := service.TrackAdImpression(adID)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}