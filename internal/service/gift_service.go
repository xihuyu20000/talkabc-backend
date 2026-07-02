package service

import (
	"backend/internal/repository"
	"fmt"
)

func SendGift(senderUID, receiverUID string, giftID uint) error {
	sender, err := repository.GetUserByUID(senderUID)
	if err != nil {
		return fmt.Errorf("发送者不存在")
	}

	receiver, err := repository.GetUserByUID(receiverUID)
	if err != nil {
		return fmt.Errorf("接收者不存在")
	}

	return repository.SendGift(sender.ID, receiver.ID, giftID)
}