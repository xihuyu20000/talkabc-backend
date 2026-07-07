package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"fmt"
)

type UserMomentDTO struct {
	UID       string    `json:"uid"`
	MID       uint      `json:"mid"`
	PraiseNum int       `json:"praisenum"`
	PubTS     int64     `json:"pubts"`
	Text      string    `json:"text"`
	Files     []string  `json:"files"`
	Location  string    `json:"location"`
}

func ConvertMomentToDTO(moment model.UserMoment) UserMomentDTO {
	user, _ := repository.GetUserByID(moment.UserID)
	return UserMomentDTO{
		UID:       user.Uid,
		MID:       moment.ID,
		PraiseNum: moment.PraiseNum,
		PubTS:     moment.PubTS,
		Text:      moment.Text,
		Files:     moment.Files,
		Location:  moment.Location,
	}
}

func GetLatestMoment() ([]UserMomentDTO, error) {
	moments, err := repository.GetLatestMoment()
	if err != nil {
		return nil, err
	}

	var result []UserMomentDTO
	for _, moment := range moments {
		result = append(result, ConvertMomentToDTO(moment))
	}

	return result, nil
}

func GetMyLatestMoment(uid string) ([]UserMomentDTO, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	moments, err := repository.GetMyLatestMoment(user.ID)
	if err != nil {
		return nil, err
	}

	var result []UserMomentDTO
	for _, moment := range moments {
		result = append(result, ConvertMomentToDTO(moment))
	}

	return result, nil
}

func GetUserMoment(uid string) ([]UserMomentDTO, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	moments, err := repository.GetUserMoment(user.ID)
	if err != nil {
		return nil, err
	}

	var result []UserMomentDTO
	for _, moment := range moments {
		result = append(result, ConvertMomentToDTO(moment))
	}

	return result, nil
}

func PraiseMoment(userUID string, momentID uint) error {
	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	err = repository.AddMomentPraise(user.ID, momentID)
	if err != nil {
		return err
	}

	return repository.UpdateMomentPraiseNum(momentID)
}

func ReportMoment(userUID string, momentID uint) error {
	_, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	return nil
}

func CommentMoment(userUID string, momentID uint, text string) error {
	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	if text == "" {
		return nil
	}
	return repository.AddMomentComment(user.ID, momentID, text)
}

func PublishMoment(userUID string, text string, files []string, location string) error {
	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	return repository.CreateMoment(user.ID, text, files, location)
}

func GetMomentComments(momentID uint) ([]interface{}, error) {
	comments, err := repository.GetMomentComments(momentID)
	if err != nil {
		return nil, err
	}

	var result []interface{}
	for _, comment := range comments {
		result = append(result, comment)
	}

	return result, nil
}

func GetFollowingMoment(uid string) ([]UserMomentDTO, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	moments, err := repository.GetFollowingMoment(user.ID)
	if err != nil {
		return nil, err
	}

	var result []UserMomentDTO
	for _, moment := range moments {
		result = append(result, ConvertMomentToDTO(moment))
	}

	return result, nil
}

func GetMomentDetail(momentID uint) (*UserMomentDTO, error) {
	moment, err := repository.GetMomentByID(momentID)
	if err != nil {
		return nil, fmt.Errorf("动态不存在")
	}

	result := ConvertMomentToDTO(*moment)
	return &result, nil
}

func CancelPraiseMoment(userUID string, momentID uint) error {
	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	err = repository.RemoveMomentPraise(user.ID, momentID)
	if err != nil {
		return err
	}

	return repository.UpdateMomentPraiseNum(momentID)
}

func DeleteMoment(userUID string, momentID uint) error {
	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	return repository.DeleteMoment(user.ID, momentID)
}

func DeleteComment(userUID string, commentID uint) error {
	user, err := repository.GetUserByUID(userUID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	return repository.DeleteMomentComment(user.ID, commentID)
}