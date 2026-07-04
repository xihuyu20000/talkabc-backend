package sms

import "context"

type SMSGateway interface {
	SendVerificationCode(ctx context.Context, phoneNum, code, minutes string) error
}

var DefaultGateway SMSGateway