package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"fmt"
)

func GetPraiseMeList(uid string) ([]model.MomentPraise, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return repository.GetPraiseMeList(user.ID)
}

func GetCommentMeList(uid string) ([]model.MomentComment, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return repository.GetCommentMeList(user.ID)
}

func GetAddMeList(uid string) ([]model.AgreeFriend, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return repository.GetAddMeList(user.ID)
}

func GetVisitMeList(uid string) ([]model.VisitRecord, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return repository.GetVisitMeList(user.ID)
}

func GetLikeMeList(uid string) ([]model.LikeRecord, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return repository.GetLikeMeList(user.ID)
}

func LikeUser(userUID, targetUID string, flag int) error {
	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	target, err := repository.GetUserByUID(targetUID)
	if err != nil {
		return fmt.Errorf("目标用户不存在")
	}

	return repository.LikeUser(user.ID, target.ID, flag)
}