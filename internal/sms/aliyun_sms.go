package sms

import (
	"backend/internal/config"
	"backend/pkg/logger"
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
	credential "github.com/aliyun/credentials-go/credentials"
)

type AliyunSMSConfig struct {
	AccessKeyId     string
	AccessKeySecret string
	RegionID        string
	SignName        string
	TemplateCode    string
	SchemeName      string
	CountryCode     string
}

type AliyunSMSGateway struct {
	config *AliyunSMSConfig
	client *openapi.Client
}

func createClient () (_result *openapi.Client, _err error) {
  credentialsConfig := new(credentials.Config).
		SetType("access_key").
		SetAccessKeyId(config.AppConfig.SMSProvider.Aliyun.AccessKeyId).
		SetAccessKeySecret(config.AppConfig.SMSProvider.Aliyun.AccessKeySecret)
  credential, _err := credential.NewCredential(credentialsConfig)
  if _err != nil {
    return _result, _err
  }

  clientConfig := &openapi.Config{
    Credential: credential,
  }
  // Endpoint 请参考 https://api.aliyun.com/product/Dypnsapi
  clientConfig.Endpoint = tea.String("dypnsapi.aliyuncs.com")
  _result = &openapi.Client{}
  _result, _err = openapi.NewClient(clientConfig)
  return _result, _err
}

func createApiInfo () (_result *openapi.Params) {
  params := &openapi.Params{
    // 接口名称
    Action: tea.String("SendSmsVerifyCode"),
    // 接口版本
    Version: tea.String("2017-05-25"),
    // 接口协议
    Protocol: tea.String("HTTPS"),
    // 接口 HTTP 方法
    Method: tea.String("POST"),
    AuthType: tea.String("AK"),
    Style: tea.String("RPC"),
    // 接口 PATH
    Pathname: tea.String("/"),
    // 接口请求体内容格式
    ReqBodyType: tea.String("json"),
    // 接口响应体内容格式
    BodyType: tea.String("json"),
  }
  _result = params
  return _result
}


func NewAliyunSMSGateway() *AliyunSMSGateway {
	return &AliyunSMSGateway{
		config: &AliyunSMSConfig{
			AccessKeyId:     config.AppConfig.SMSProvider.Aliyun.AccessKeyId,
			AccessKeySecret: config.AppConfig.SMSProvider.Aliyun.AccessKeySecret,
			RegionID:        config.AppConfig.SMSProvider.Aliyun.RegionID,
			SignName:        config.AppConfig.SMSProvider.Aliyun.SignName,
			TemplateCode:    config.AppConfig.SMSProvider.Aliyun.TemplateCode,
			SchemeName:      config.AppConfig.SMSProvider.Aliyun.SchemeName,
			CountryCode:     config.AppConfig.SMSProvider.Aliyun.CountryCode,
		},
	}
}

func (a *AliyunSMSGateway) SendVerificationCode(ctx context.Context, phoneNum, code, minutes string) error {
  client, _err := createClient()
  if _err != nil {
    return _err
  }

  params := createApiInfo()
  // query params
  queries := map[string]interface{}{}
  queries["SchemeName"] = tea.String(a.config.SchemeName)
  queries["CountryCode"] = tea.String(a.config.CountryCode)
  queries["PhoneNumber"] = tea.String(phoneNum)
  queries["SignName"] = tea.String(a.config.SignName)
  queries["TemplateCode"] = tea.String(a.config.TemplateCode)
  queries["TemplateParam"] = tea.String(fmt.Sprintf("{\"code\":\"%s\",\"min\":\"%s\"}", code, minutes))
  // runtime options
  runtime := &util.RuntimeOptions{}
  request := &openapi.OpenApiRequest{
    Query: openapiutil.Query(queries),
  }
  // 返回值实际为 Map 类型，可从 Map 中获得三类数据：响应体 body、响应头 headers、HTTP 返回的状态码 statusCode。
  resp, _err := client.CallApi(params, request, runtime)
  if _err != nil {
	logger.Errorf("[SMS] SendVerificationCode error: %s", _err.Error())
    return _err
  }

	logger.Infof("[LOG] SMS response: %v", resp)
  return nil

}

func AliyunSendVerificationCode(ctx context.Context, phoneNum, code, minutes string) error {
	if DefaultGateway != nil {
		return DefaultGateway.SendVerificationCode(ctx, phoneNum, code, minutes)
	}
	gateway := NewAliyunSMSGateway()
	return gateway.SendVerificationCode(ctx, phoneNum, code, minutes)
}
