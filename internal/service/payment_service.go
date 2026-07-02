package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"fmt"
)

type DiamondDTO struct {
	PinkDiamond int `json:"pink_diamond"`
	BlueDiamond int `json:"blue_diamond"`
}

func BuyDiamond(uid string, pid uint) error {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	diamondMap := map[uint]struct {
		DiamondType int
		Amount      int
	}{
		1: {1, 10},
		2: {1, 50},
		3: {2, 10},
		4: {2, 50},
	}

	diamond, ok := diamondMap[pid]
	if !ok {
		return fmt.Errorf("产品不存在")
	}

	err = repository.UpdateDiamond(user.ID, 0, 0)
	if err != nil {
		return err
	}

	if diamond.DiamondType == 1 {
		err = repository.UpdateDiamond(user.ID, diamond.Amount, 0)
	} else {
		err = repository.UpdateDiamond(user.ID, 0, diamond.Amount)
	}
	if err != nil {
		return err
	}

	orderID := fmt.Sprintf("diamond_%s_%d", uid, pid)
	return repository.CreateDiamondRecord(user.ID, diamond.DiamondType, diamond.Amount, orderID)
}

func GetDiamondStock(uid string) (*DiamondDTO, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	diamond, err := repository.GetDiamond(user.ID)
	if err != nil {
		return nil, err
	}

	return &DiamondDTO{
		PinkDiamond: diamond.PinkDiamond,
		BlueDiamond: diamond.BlueDiamond,
	}, nil
}

func GetDiamondHistory(uid string) ([]model.DiamondRecord, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	return repository.GetDiamondRecords(user.ID)
}

func BuyMember(uid string, pid uint) error {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	memberMap := map[uint]struct {
		Level      int
		ExpireDays int
	}{
		1: {1, 30},
		2: {1, 90},
		3: {2, 30},
		4: {2, 90},
	}

	member, ok := memberMap[pid]
	if !ok {
		return fmt.Errorf("产品不存在")
	}

	err = repository.UpdateMember(user.ID, member.Level, member.ExpireDays)
	if err != nil {
		return err
	}

	orderID := fmt.Sprintf("member_%s_%d", uid, pid)
	return repository.CreateMemberRecord(user.ID, member.Level, orderID)
}

func GetMemberHistory(uid string) ([]model.MemberRecord, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	return repository.GetMemberRecords(user.ID)
}