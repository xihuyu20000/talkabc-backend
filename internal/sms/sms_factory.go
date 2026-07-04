package sms

import (
	"backend/internal/config"
	"backend/pkg/logger"
)

var gateway SMSGateway

func InitSMSGateway(cfg *config.SMSProviderConfig) error {
	switch cfg.Default {
	case "aliyun":
		logger.Infof("[SMS] Initializing aliyun gateway - AccessKeyID: %s, RegionID: %s, SignName: %s, TemplateCode: %s",
			cfg.Aliyun.AccessKeyID, cfg.Aliyun.RegionID, cfg.Aliyun.SignName, cfg.Aliyun.TemplateCode)
		
		aliyunConfig := &AliyunSMSConfig{
			AccessKeyID:     cfg.Aliyun.AccessKeyID,
			AccessKeySecret: cfg.Aliyun.AccessKeySecret,
			RegionID:        cfg.Aliyun.RegionID,
			SignName:        cfg.Aliyun.SignName,
			TemplateCode:    cfg.Aliyun.TemplateCode,
			SchemeName:      cfg.Aliyun.SchemeName,
			CountryCode:     cfg.Aliyun.CountryCode,
		}
		
		if aliyunConfig.SchemeName == "" {
			aliyunConfig.SchemeName = "DysmsVerify"
			logger.Infof("[SMS] SchemeName not set, using default: %s", aliyunConfig.SchemeName)
		}
		
		client, err := NewAliyunSMSGateway(aliyunConfig)
		if err != nil {
			logger.Errorf("[SMS] Failed to create aliyun gateway: %v", err)
			return err
		}
		gateway = client
		logger.Infof("SMS gateway initialized: aliyun")
	case "huawei":
		logger.Infof("SMS gateway initialized: huawei (not implemented yet)")
	case "tencent":
		logger.Infof("SMS gateway initialized: tencent (not implemented yet)")
	default:
		return nil
	}
	return nil
}

func GetGateway() SMSGateway {
	return gateway
}