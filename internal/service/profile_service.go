package service

import (
	"backend/internal/repository"
	"fmt"
	"strconv"
)

func CollectMyInfo(uid string, info map[string]interface{}) error {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	if v, ok := info["regcountry"].(string); ok {
		user.Country = v
	}
	if v, ok := info["mylang"].(string); ok {
		user.Language = v
	}
	if v, ok := info["nickname"].(string); ok {
		// 【昵称校验规则】保存昵称前先进行有效性校验
		if err := ValidateNickname(v); err != nil {
			return err
		}
		user.Nickname = v
	}
	if v, ok := info["birthyear"].(int); ok {
		user.BirthYear = v
	}
	if v, ok := info["gender"].(int); ok {
		user.Gender = v
	}
	if v, ok := info["height"].(int); ok {
		user.Height = v
	}
	if v, ok := info["weight"].(int); ok {
		user.Weight = v
	}
	if v, ok := info["city"].(string); ok {
		user.City = v
	}
	if v, ok := info["school"].(string); ok {
		user.School = v
	}
	if v, ok := info["job"].(string); ok {
		user.Job = v
	}
	if v, ok := info["edulevel"].(int); ok {
		user.EduLevel = v
	}
	if v, ok := info["starsign"].(int); ok {
		user.StarSign = strconv.Itoa(v)
	}

	if err := repository.UpdateUser(user); err != nil {
		return err
	}

	if favors, ok := info["favors"].([]string); ok && len(favors) > 0 {
		var tagIDs []uint
		for _, id := range favors {
			if v, err := strconv.ParseUint(id, 10, 32); err == nil {
				tagIDs = append(tagIDs, uint(v))
			}
		}
		if err := repository.SaveUserHobbies(uid, tagIDs); err != nil {
			return err
		}
	}

	if datingPurposes, ok := info["dating_purposes"].([]string); ok && len(datingPurposes) > 0 {
		var purposeIDs []uint
		for _, id := range datingPurposes {
			if v, err := strconv.ParseUint(id, 10, 32); err == nil {
				purposeIDs = append(purposeIDs, uint(v))
			}
		}
		if err := repository.SaveUserDatingPurposes(uid, purposeIDs); err != nil {
			return err
		}
	}

	return nil
}

func CollectAimInfo(uid string, info map[string]interface{}) error {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	user.Aim = info

	return repository.UpdateUser(user)
}