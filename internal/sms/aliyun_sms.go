package sms

import (
	"backend/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi20170525 "github.com/alibabacloud-go/dypnsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type AliyunSMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	RegionID        string
	SignName        string
	TemplateCode    string
	SchemeName      string
	CountryCode     string
}

type AliyunSMSGateway struct {
	config *AliyunSMSConfig
	client *dypnsapi20170525.Client
}

func NewAliyunSMSGateway(config *AliyunSMSConfig) (*AliyunSMSGateway, error) {
	if config.AccessKeyID == "" || config.AccessKeySecret == "" {
		logger.Fatalf("aliyun sms credentials are empty")
		os.Exit(1)
	}

	cfg := &openapi.Config{
		AccessKeyId:     tea.String(config.AccessKeyID),
		AccessKeySecret: tea.String(config.AccessKeySecret),
		RegionId:        tea.String(config.RegionID),
	}

	if config.RegionID == "" {
		cfg.RegionId = tea.String("cn-hangzhou")
	}

	client, err := dypnsapi20170525.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	if config.CountryCode == "" {
		config.CountryCode = "86"
	}

	return &AliyunSMSGateway{
		config: config,
		client: client,
	}, nil
}

func (g *AliyunSMSGateway) SendVerificationCode(ctx context.Context, phoneNum, code string) error {
	sendSmsVerifyCodeRequest := &dypnsapi20170525.SendSmsVerifyCodeRequest{
		SchemeName:    tea.String(g.config.SchemeName),
		CountryCode:   tea.String(g.config.CountryCode),
		PhoneNumber:   tea.String(phoneNum),
		SignName:      tea.String(g.config.SignName),
		TemplateParam: tea.String(fmt.Sprintf("{\"code\":\"%s\"}", code)),
		TemplateCode:  tea.String(g.config.TemplateCode),
	}

	runtime := &util.RuntimeOptions{}

	resp, err := g.client.SendSmsVerifyCodeWithOptions(sendSmsVerifyCodeRequest, runtime)
	if err != nil {
		return g.handleError(err)
	}

	logger.Infof("[LOG] SMS response: %v", resp)
	return nil
}

func (g *AliyunSMSGateway) SendText(ctx context.Context, phoneNum, templateID string, params map[string]string) error {
	paramStr := ""
	for k, v := range params {
		if paramStr != "" {
			paramStr += ","
		}
		paramStr += fmt.Sprintf("\"%s\":\"%s\"", k, v)
	}

	sendSmsVerifyCodeRequest := &dypnsapi20170525.SendSmsVerifyCodeRequest{
		SchemeName:    tea.String(g.config.SchemeName),
		CountryCode:   tea.String(g.config.CountryCode),
		PhoneNumber:   tea.String(phoneNum),
		SignName:      tea.String(g.config.SignName),
		TemplateParam: tea.String(fmt.Sprintf("{%s}", paramStr)),
		TemplateCode:  tea.String(templateID),
	}

	runtime := &util.RuntimeOptions{}

	resp, err := g.client.SendSmsVerifyCodeWithOptions(sendSmsVerifyCodeRequest, runtime)
	if err != nil {
		return g.handleError(err)
	}

	logger.Infof("[LOG] SMS response: %v", resp)
	return nil
}

func (g *AliyunSMSGateway) handleError(err error) error {
	var error = &tea.SDKError{}
	if _t, ok := err.(*tea.SDKError); ok {
		error = _t
	} else {
		error.Message = tea.String(err.Error())
	}

	var data interface{}
	d := json.NewDecoder(strings.NewReader(tea.StringValue(error.Data)))
	d.Decode(&data)
	if m, ok := data.(map[string]interface{}); ok {
		recommend, _ := m["Recommend"]
		return fmt.Errorf("send sms failed: %s, recommend: %v", tea.StringValue(error.Message), recommend)
	}

	return fmt.Errorf("send sms failed: %s", tea.StringValue(error.Message))
}

func (g *AliyunSMSGateway) GetName() string {
	return "aliyun"
}