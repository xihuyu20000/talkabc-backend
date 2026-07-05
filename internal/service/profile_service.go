package service

import (
	"backend/internal/repository"
	"backend/pkg/security"
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

// CheckProfileStatus 检查用户资料收集状态
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//
// 返回值：
//   - bool: 资料是否已收集完成
//   - error: 错误信息
//
// 判断逻辑：
//   1. 查询用户ProfileCompleted字段
//   2. 如果ProfileCompleted为1，表示资料已收集完成
//   3. 如果ProfileCompleted为0，表示需要收集资料
func CheckProfileStatus(uid string) (bool, error) {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return false, fmt.Errorf("用户不存在")
	}

	return user.ProfileCompleted == 1, nil
}

// SetProfileCompleted 设置资料收集完成状态
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. 查询用户记录
//   2. 将ProfileCompleted字段设置为1
//   3. 更新用户记录
func SetProfileCompleted(uid string) error {
	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	user.ProfileCompleted = 1

	return repository.UpdateUser(user)
}

// SetSignText 设置用户个性签名
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//   - signText: 个性签名内容
//
// 返回值：
//   - error: 错误信息
//
// 安全规则（调用security包进行全面安全检查）：
//   1. 签名长度限制：最大200字符
//   2. 敏感词过滤：禁止包含敏感词汇
//   3. URL过滤：禁止包含超链接
//   4. HTML过滤：禁止包含HTML标签
//   5. JavaScript过滤：禁止包含脚本代码
//   6. SQL注入过滤：禁止包含SQL注入代码
//   7. XSS过滤：禁止包含XSS攻击代码
func SetSignText(uid string, signText string) error {
	if err := security.ValidateSignText(signText); err != nil {
		return err
	}

	user, err := repository.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	user.SignText = signText

	return repository.UpdateUser(user)
}