package sms

import "context"

type SMSGateway interface {
	SendVerificationCode(ctx context.Context, phoneNum, code string) error
	SendText(ctx context.Context, phoneNum, templateID string, params map[string]string) error
	GetName() string
}

type SendResult struct {
	Success bool
	Message string
	RequestID string
}