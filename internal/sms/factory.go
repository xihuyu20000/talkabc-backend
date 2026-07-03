package sms

import (
	"backend/internal/config"
	"log"
)

var gateway SMSGateway

func InitSMSGateway(cfg *config.SMSProviderConfig) error {
	switch cfg.Default {
	case "aliyun":
		log.Printf("[SMS] Initializing aliyun gateway - AccessKeyID: %s, RegionID: %s, SignName: %s, TemplateCode: %s",
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
			log.Printf("[SMS] SchemeName not set, using default: %s", aliyunConfig.SchemeName)
		}
		
		client, err := NewAliyunSMSGateway(aliyunConfig)
		if err != nil {
			log.Printf("[SMS] Failed to create aliyun gateway: %v", err)
			return err
		}
		gateway = client
		log.Printf("SMS gateway initialized: aliyun")
	case "huawei":
		log.Printf("SMS gateway initialized: huawei (not implemented yet)")
	case "tencent":
		log.Printf("SMS gateway initialized: tencent (not implemented yet)")
	default:
		return nil
	}
	return nil
}

func GetGateway() SMSGateway {
	return gateway
}