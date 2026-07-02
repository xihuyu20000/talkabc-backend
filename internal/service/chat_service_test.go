package service

import (
	"backend/internal/model"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

// TestConvertChatMessage 测试聊天消息模型到DTO的转换
// 验证消息字段（ID、发送者ID、接收者ID、文本、消息类型、阅读状态）的正确映射
func TestConvertChatMessage(t *testing.T) {
	msg := model.ChatMessage{
		Model:      gorm.Model{ID: 1},
		SenderID:   100,
		ReceiverID: 200,
		Text:       "Hello World",
		FileURL:    "",
		MsgType:    1,
		ReadStatus: 0,
		SendTime:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	dto := ConvertChatMessage(msg)

	if dto.ID != msg.ID {
		t.Errorf("Expected ID %d, got %d", msg.ID, dto.ID)
	}

	if dto.SenderID != msg.SenderID {
		t.Errorf("Expected SenderID %d, got %d", msg.SenderID, dto.SenderID)
	}

	if dto.ReceiverID != msg.ReceiverID {
		t.Errorf("Expected ReceiverID %d, got %d", msg.ReceiverID, dto.ReceiverID)
	}

	if dto.Text != msg.Text {
		t.Errorf("Expected Text %s, got %s", msg.Text, dto.Text)
	}

	if dto.MsgType != msg.MsgType {
		t.Errorf("Expected MsgType %d, got %d", msg.MsgType, dto.MsgType)
	}

	if dto.ReadStatus != msg.ReadStatus {
		t.Errorf("Expected ReadStatus %d, got %d", msg.ReadStatus, dto.ReadStatus)
	}
}

// TestConvertChatMessage_SendTimeFormat 测试聊天消息发送时间格式化
// 验证消息发送时间转换为字符串后的格式是否正确（YYYY-MM-DD HH:MM:SS）
func TestConvertChatMessage_SendTimeFormat(t *testing.T) {
	msg := model.ChatMessage{
		Model:      gorm.Model{ID: 1},
		SenderID:   1,
		ReceiverID: 2,
		Text:       "Test",
		SendTime:   time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC),
	}

	dto := ConvertChatMessage(msg)

	expectedTime := "2024-06-15 14:30:45"
	if dto.SendTime != expectedTime {
		t.Errorf("Expected SendTime %s, got %s", expectedTime, dto.SendTime)
	}
}

// TestConvertChatMessage_WithFile 测试带文件的聊天消息转换
// 验证文件消息的URL字段正确映射，且文本字段为空
func TestConvertChatMessage_WithFile(t *testing.T) {
	msg := model.ChatMessage{
		Model:      gorm.Model{ID: 2},
		SenderID:   10,
		ReceiverID: 20,
		Text:       "",
		FileURL:    "https://cdn.example.com/image.jpg",
		MsgType:    2,
		ReadStatus: 1,
		SendTime:   time.Now(),
	}

	dto := ConvertChatMessage(msg)

	if dto.FileURL != msg.FileURL {
		t.Errorf("Expected FileURL %s, got %s", msg.FileURL, dto.FileURL)
	}

	if dto.Text != "" {
		t.Error("Text should be empty for file messages")
	}
}

// TestChatMessageDTO_Structure 测试 ChatMessageDTO 结构完整性
// 验证消息DTO包含所有必要字段，且字段类型正确
func TestChatMessageDTO_Structure(t *testing.T) {
	dto := ChatMessageDTO{
		ID:         1,
		SenderID:   100,
		ReceiverID: 200,
		Text:       "Test message",
		FileURL:    "",
		MsgType:    1,
		ReadStatus: 0,
		SendTime:   "2024-01-01 00:00:00",
	}

	if dto.ID != 1 {
		t.Error("ID should be 1")
	}

	if dto.SenderID != 100 {
		t.Error("SenderID should be 100")
	}

	if dto.MsgType != 1 {
		t.Error("MsgType should be 1 (text)")
	}

	if dto.ReadStatus != 0 {
		t.Error("ReadStatus should be 0 (unread)")
	}
}