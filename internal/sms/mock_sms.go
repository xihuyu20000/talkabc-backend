package sms

import (
	"backend/pkg/logger"
	"context"
	"fmt"
	"sync"
)

type MockSMSGateway struct {
	sync.Mutex
	sentMessages []SMSSentMessage
	shouldFail   bool
	failCount    int
}

type SMSSentMessage struct {
	PhoneNum string
	Code     string
	Minutes  string
}

func NewMockSMSGateway() *MockSMSGateway {
	return &MockSMSGateway{
		sentMessages: make([]SMSSentMessage, 0),
	}
}

func (m *MockSMSGateway) SendVerificationCode(ctx context.Context, phoneNum, code, minutes string) error {
	m.Lock()
	defer m.Unlock()

	if m.shouldFail {
		m.failCount++
		return fmt.Errorf("mock sms gateway configured to fail")
	}

	m.sentMessages = append(m.sentMessages, SMSSentMessage{
		PhoneNum: phoneNum,
		Code:     code,
		Minutes:  minutes,
	})

	logger.Infof("[MockSMS] Sent verification code to %s - code: %s, expires in %s minutes", phoneNum, code, minutes)
	return nil
}

func (m *MockSMSGateway) GetSentMessages() []SMSSentMessage {
	m.Lock()
	defer m.Unlock()
	return append([]SMSSentMessage{}, m.sentMessages...)
}

func (m *MockSMSGateway) GetLastSentMessage() *SMSSentMessage {
	m.Lock()
	defer m.Unlock()
	if len(m.sentMessages) == 0 {
		return nil
	}
	return &m.sentMessages[len(m.sentMessages)-1]
}

func (m *MockSMSGateway) ClearSentMessages() {
	m.Lock()
	defer m.Unlock()
	m.sentMessages = make([]SMSSentMessage, 0)
}

func (m *MockSMSGateway) SetShouldFail(fail bool) {
	m.Lock()
	defer m.Unlock()
	m.shouldFail = fail
}

func (m *MockSMSGateway) GetFailCount() int {
	m.Lock()
	defer m.Unlock()
	return m.failCount
}