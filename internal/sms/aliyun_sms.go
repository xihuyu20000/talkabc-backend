package sms

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

type AliyunSMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	RegionID        string
	SignName        string
	TemplateCode    string
}

type AliyunSMSGateway struct {
	config *AliyunSMSConfig
	client *dysmsapi20170525.Client
}

func NewAliyunSMSGateway(config *AliyunSMSConfig) (*AliyunSMSGateway, error) {
	client, err := dysmsapi20170525.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(config.AccessKeyID),
		AccessKeySecret: tea.String(config.AccessKeySecret),
		RegionId:        tea.String(config.RegionID),
	})
	if err != nil {
		return nil, err
	}

	return &AliyunSMSGateway{
		config: config,
		client: client,
	}, nil
}

func (g *AliyunSMSGateway) SendVerificationCode(ctx context.Context, phoneNum, code string) error {
	return g.SendText(ctx, phoneNum, g.config.TemplateCode, map[string]string{
		"code": code,
	})
}

func (g *AliyunSMSGateway) SendText(ctx context.Context, phoneNum, templateID string, params map[string]string) error {
	paramStr := ""
	for k, v := range params {
		if paramStr != "" {
			paramStr += ","
		}
		paramStr += fmt.Sprintf("%s:%s", k, v)
	}

	request := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(phoneNum),
		SignName:     tea.String(g.config.SignName),
		TemplateCode:  tea.String(templateID),
		TemplateParam: tea.String(fmt.Sprintf("{%s}", paramStr)),
	}

	response, err := g.client.SendSms(request)
	if err != nil {
		return err
	}

	if tea.StringValue(response.Body.Code) != "OK" {
		return fmt.Errorf("send sms failed: %s", tea.StringValue(response.Body.Message))
	}

	return nil
}

func (g *AliyunSMSGateway) GetName() string {
	return "aliyun"
}