package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"fmt"
)

type ChatMessageDTO struct {
	ID         uint      `json:"id"`
	SenderID   uint      `json:"sender_id"`
	ReceiverID uint      `json:"receiver_id"`
	Text       string    `json:"text"`
	FileURL    string    `json:"file_url"`
	MsgType    int       `json:"msg_type"`
	ReadStatus int       `json:"read_status"`
	SendTime   string    `json:"send_time"`
}

func ConvertChatMessage(msg model.ChatMessage) ChatMessageDTO {
	return ChatMessageDTO{
		ID:         msg.ID,
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
		Text:       msg.Text,
		FileURL:    msg.FileURL,
		MsgType:    msg.MsgType,
		ReadStatus: msg.ReadStatus,
		SendTime:   msg.SendTime.Format("2006-01-02 15:04:05"),
	}
}

func GetSystemMsgList(uid string) ([]model.SystemMsg, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return repository.GetSystemMsgList(user.ID)
}

func GetLatestUserMsg(uid string) ([]ChatMessageDTO, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	msgs, err := repository.GetLatestUserMsg(user.ID)
	if err != nil {
		return nil, err
	}

	var result []ChatMessageDTO
	for _, msg := range msgs {
		result = append(result, ConvertChatMessage(msg))
	}

	return result, nil
}

func GetUserMsgHistory(uid, targetUID string) ([]ChatMessageDTO, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	target, err := repository.GetUserByUID(targetUID)
	if err != nil {
		return nil, fmt.Errorf("目标用户不存在")
	}

	msgs, err := repository.GetUserMsgHistory(user.ID, target.ID)
	if err != nil {
		return nil, err
	}

	var result []ChatMessageDTO
	for _, msg := range msgs {
		result = append(result, ConvertChatMessage(msg))
	}

	return result, nil
}

func SetMessageTop(uid, targetUID string, flag int) error {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	target, err := repository.GetUserByUID(targetUID)
	if err != nil {
		return fmt.Errorf("目标用户不存在")
	}

	return repository.SetMessageTop(user.ID, target.ID, flag)
}

func ClearChatHistory(uid, targetUID string) error {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	target, err := repository.GetUserByUID(targetUID)
	if err != nil {
		return fmt.Errorf("目标用户不存在")
	}

	return repository.ClearChatHistory(user.ID, target.ID)
}

func SendTextMessage(senderUID, receiverUID string, text string) error {
	sender, err := repository.GetUserByUID(senderUID)
	if err != nil {
		return fmt.Errorf("发送者不存在")
	}

	receiver, err := repository.GetUserByUID(receiverUID)
	if err != nil {
		return fmt.Errorf("接收者不存在")
	}

	return repository.SendTextMessage(sender.ID, receiver.ID, text)
}

func SendImageMessage(senderUID, receiverUID string, fileURL string) error {
	sender, err := repository.GetUserByUID(senderUID)
	if err != nil {
		return fmt.Errorf("发送者不存在")
	}

	receiver, err := repository.GetUserByUID(receiverUID)
	if err != nil {
		return fmt.Errorf("接收者不存在")
	}

	return repository.SendImageMessage(sender.ID, receiver.ID, fileURL)	
}

func WithdrawMessage(senderUID, receiverUID string, msgID uint) error {
	sender, err := repository.GetUserByUID(senderUID)
	if err != nil {
		return fmt.Errorf("发送者不存在")
	}

	receiver, err := repository.GetUserByUID(receiverUID)
	if err != nil {
		return fmt.Errorf("接收者不存在")
	}

	return repository.WithdrawMessage(sender.ID, receiver.ID, msgID)
}

func SendVideoMessage(senderUID, receiverUID string, fileURL string) error {
	sender, err := repository.GetUserByUID(senderUID)
	if err != nil {
		return fmt.Errorf("发送者不存在")
	}

	receiver, err := repository.GetUserByUID(receiverUID)
	if err != nil {
		return fmt.Errorf("接收者不存在")
	}

	return repository.SendVideoMessage(sender.ID, receiver.ID, fileURL)	
}

func SendVoiceMessage(senderUID, receiverUID string, fileURL string) error {
	sender, err := repository.GetUserByUID(senderUID)
	if err != nil {
		return fmt.Errorf("发送者不存在")
	}

	receiver, err := repository.GetUserByUID(receiverUID)
	if err != nil {
		return fmt.Errorf("接收者不存在")
	}

	return repository.SendVoiceMessage(sender.ID, receiver.ID, fileURL)	
}

func SendFileMessage(senderUID, receiverUID string, fileURL string) error {
	sender, err := repository.GetUserByUID(senderUID)
	if err != nil {
		return fmt.Errorf("发送者不存在")
	}

	receiver, err := repository.GetUserByUID(receiverUID)
	if err != nil {
		return fmt.Errorf("接收者不存在")
	}

	return repository.SendFileMessage(sender.ID, receiver.ID, fileURL)
}