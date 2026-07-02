package service

import (
	"backend/internal/repository"
	"fmt"
)

func AddFriend(uid, targetUID string, flag int) error {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	target, err := repository.GetUserByUID(targetUID)
	if err != nil {
		return fmt.Errorf("目标用户不存在")
	}

	return repository.AddFriend(user.ID, target.ID, flag)
}

func AgreeFriendRequest(userUID, targetUID string, flag int) error {
	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	target, err := repository.GetUserByUID(targetUID)
	if err != nil {
		return fmt.Errorf("目标用户不存在")
	}

	return repository.AgreeFriendRequest(user.ID, target.ID, flag)
}