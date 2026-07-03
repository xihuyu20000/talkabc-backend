package sms

import (
	"backend/internal/config"
	"log"
)

var gateway SMSGateway

func InitSMSGateway(cfg *config.SMSProviderConfig) error {
	switch cfg.Default {
	case "aliyun":
		aliyunConfig := &AliyunSMSConfig{
			AccessKeyID:     cfg.Aliyun.AccessKeyID,
			AccessKeySecret: cfg.Aliyun.AccessKeySecret,
			RegionID:        cfg.Aliyun.RegionID,
			SignName:        cfg.Aliyun.SignName,
			TemplateCode:    cfg.Aliyun.TemplateCode,
		}
		client, err := NewAliyunSMSGateway(aliyunConfig)
		if err != nil {
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