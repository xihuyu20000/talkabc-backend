package repository

import (
	"backend/internal/config"
	"backend/internal/model"
	"time"
)

// GetUserByID 根据数据库ID查询用户
// 参数说明：
//   - id: 用户数据库ID
//
// 返回值：
//   - *model.User: 用户模型指针
//   - error: 错误信息
func GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := config.DB.Where("id = ?", id).First(&user).Error
	return &user, err
}

// GetUserByUID 根据雪花ID查询用户
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//
// 返回值：
//   - *model.User: 用户模型指针
//   - error: 错误信息
func GetUserByUID(uid string) (*model.User, error) {
	var user model.User
	err := config.DB.Where("uid = ?", uid).First(&user).Error
	return &user, err
}

// GetUserList 获取用户列表（支持筛选条件和分页）
// 参数说明：
//   - filters: 筛选条件映射
//     - gender: 性别（0-未知，1-男，2-女），2表示不限
//     - age: 年龄范围数组 [minAge, maxAge]
//     - official: 是否官方认证，1表示官方账号
//     - latest: 是否按最新排序，1表示按创建时间降序
//     - page: 页码，默认1
//     - size: 每页数量，默认20
//
// 返回值：
//   - []model.User: 用户列表
//   - int: 总记录数
//   - error: 错误信息
func GetUserList(filters map[string]interface{}) ([]model.User, int, error) {
	var users []model.User
	query := config.DB

	if gender, ok := filters["gender"].(int); ok {
		if gender != 2 {
			query = query.Where("gender = ?", gender)
		}
	}
	var currentYear = time.Now().Year()
	if age1, ok := filters["age1"].(int); ok && age1 > 0 {
		query = query.Where("birth_year <= ?", currentYear-age1)
	}
	if age2, ok := filters["age2"].(int); ok && age2 > 0 {
		query = query.Where("birth_year >= ?", currentYear-age2)
	}

	if official, ok := filters["official"].(int); ok && official == 1 {
		query = query.Order("id ASC")
	}

	if latest, ok := filters["latest"].(int); ok && latest == 1 {
		query = query.Order("created_at DESC")
	}

	var total int
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	page := 1
	size := 20
	if p, ok := filters["page"].(int); ok && p > 0 {
		page = p
	}
	if s, ok := filters["size"].(int); ok && s > 0 {
		size = s
	}

	offset := (page - 1) * size
	err = query.Offset(offset).Limit(size).Find(&users).Error
	return users, total, err
}

// GetFocusList 获取用户关注列表
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.User: 关注用户列表
//   - error: 错误信息
//
// 查询逻辑：
//   1. 查询用户关注的目标ID列表
//   2. 根据ID列表查询用户信息
func GetFocusList(uid uint) ([]model.User, error) {
	var focusIDs []uint
	config.DB.Model(&model.UserFocus{}).Where("user_id = ?", uid).Pluck("target_id", &focusIDs)

	var users []model.User
	err := config.DB.Where("id IN (?)", focusIDs).Find(&users).Error
	return users, err
}

// GetFansList 获取用户粉丝列表
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.User: 粉丝用户列表
//   - error: 错误信息
//
// 查询逻辑：
//   1. 查询关注当前用户的用户ID列表
//   2. 根据ID列表查询用户信息
func GetFansList(uid uint) ([]model.User, error) {
	var fanIDs []uint
	config.DB.Model(&model.UserFocus{}).Where("target_id = ?", uid).Pluck("user_id", &fanIDs)

	var users []model.User
	err := config.DB.Where("id IN (?)", fanIDs).Find(&users).Error
	return users, err
}

// AddFocus 添加关注
// 参数说明：
//   - userID: 关注者数据库ID
//   - targetID: 被关注者数据库ID
//
// 返回值：
//   - error: 错误信息
func AddFocus(userID, targetID uint) error {
	focus := model.UserFocus{
		UserID:   userID,
		TargetID: targetID,
	}
	return config.DB.Create(&focus).Error
}

// RemoveFocus 取消关注
// 参数说明：
//   - userID: 关注者数据库ID
//   - targetID: 被关注者数据库ID
//
// 返回值：
//   - error: 错误信息
func RemoveFocus(userID, targetID uint) error {
	return config.DB.Where("user_id = ? AND target_id = ?", userID, targetID).Delete(&model.UserFocus{}).Error
}

// AddBlock 添加拉黑
// 参数说明：
//   - userID: 用户数据库ID
//   - targetID: 被拉黑用户数据库ID
//
// 返回值：
//   - error: 错误信息
func AddBlock(userID, targetID uint) error {
	block := model.UserBlock{
		UserID:   userID,
		TargetID: targetID,
	}
	return config.DB.Create(&block).Error
}

// RemoveBlock 取消拉黑
// 参数说明：
//   - userID: 用户数据库ID
//   - targetID: 被拉黑用户数据库ID
//
// 返回值：
//   - error: 错误信息
func RemoveBlock(userID, targetID uint) error {
	return config.DB.Where("user_id = ? AND target_id = ?", userID, targetID).Delete(&model.UserBlock{}).Error
}

// SetNotify 设置通知开关
// 参数说明：
//   - userID: 用户数据库ID
//   - targetID: 目标用户数据库ID
//   - flag: 通知标志，0-关闭，1-开启
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. 查询已有记录
//   2. 不存在则创建，存在则更新
func SetNotify(userID, targetID uint, flag int) error {
	var notify model.UserNotify
	config.DB.Where("user_id = ? AND target_id = ?", userID, targetID).First(&notify)

	if notify.ID == 0 {
		notify = model.UserNotify{
			UserID:   userID,
			TargetID: targetID,
			Notify:   flag,
		}
		return config.DB.Create(&notify).Error
	}

	notify.Notify = flag
	return config.DB.Save(&notify).Error
}

// GetUserCount 获取用户总数
// 返回值：
//   - int: 用户总数
//   - error: 错误信息
func GetUserCount() (int, error) {
	var count int
	err := config.DB.Model(&model.User{}).Count(&count).Error
	return count, err
}

// GetFocusCount 获取用户关注数量
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - int: 关注数量
//   - error: 错误信息
func GetFocusCount(uid uint) (int, error) {
	var count int
	err := config.DB.Model(&model.UserFocus{}).Where("user_id = ?", uid).Count(&count).Error
	return count, err
}

// GetFansCount 获取用户粉丝数量
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - int: 粉丝数量
//   - error: 错误信息
func GetFansCount(uid uint) (int, error) {
	var count int
	err := config.DB.Model(&model.UserFocus{}).Where("target_id = ?", uid).Count(&count).Error
	return count, err
}

// GreetUser 打招呼（预留方法）
// 参数说明：
//   - userID: 当前用户数据库ID
//   - targetID: 目标用户数据库ID
//   - text: 打招呼内容
//
// 返回值：
//   - error: 错误信息（当前返回nil）
func GreetUser(userID, targetID uint, text string) error {
	return nil
}

// UpdateLastSeenAt 更新用户最后活跃时间
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//
// 返回值：
//   - error: 错误信息
//
// 说明：直接执行UPDATE语句，不查询用户对象，提高性能
func UpdateLastSeenAt(uid string) error {
	return config.DB.Model(&model.User{}).Where("uid = ?", uid).Update("last_seen_at", time.Now()).Error
}

// GetHobbyTags 获取所有爱好标签
// 返回值：
//   - []model.HobbyTag: 爱好标签列表（按排序字段升序）
//   - error: 错误信息
func GetHobbyTags() ([]model.HobbyTag, error) {
	var tags []model.HobbyTag
	err := config.DB.Order("sort ASC").Find(&tags).Error
	return tags, err
}

// GetUserHobbies 获取用户的爱好列表
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//
// 返回值：
//   - []model.HobbyTag: 用户爱好标签列表
//   - error: 错误信息
//
// 逻辑：
//   1. 查询用户-爱好关联记录
//   2. 根据标签ID查询爱好标签信息
func GetUserHobbies(uid string) ([]model.HobbyTag, error) {
	var rels []model.UserHobbyRel
	config.DB.Where("uid = ?", uid).Find(&rels)

	var tagIDs []uint
	for _, rel := range rels {
		tagIDs = append(tagIDs, rel.TagID)
	}

	var tags []model.HobbyTag
	err := config.DB.Where("id IN (?)", tagIDs).Find(&tags).Error
	return tags, err
}

// SaveUserHobbies 保存用户爱好（先删除再插入）
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//   - tagIDs: 爱好标签ID列表
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. 删除用户现有爱好关联记录
//   2. 批量插入新的爱好关联记录
func SaveUserHobbies(uid string, tagIDs []uint) error {
	tx := config.DB.Begin()

	if err := tx.Where("uid = ?", uid).Delete(&model.UserHobbyRel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, tagID := range tagIDs {
		rel := model.UserHobbyRel{
			Uid:   uid,
			TagID: tagID,
		}
		if err := tx.Create(&rel).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetDatingPurposes 获取所有交友目的标签
// 返回值：
//   - []model.DatingPurpose: 交友目的标签列表（按排序字段升序）
//   - error: 错误信息
func GetDatingPurposes() ([]model.DatingPurpose, error) {
	var purposes []model.DatingPurpose
	err := config.DB.Order("sort ASC").Find(&purposes).Error
	return purposes, err
}

// GetUserDatingPurposes 获取用户的交友目的列表
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//
// 返回值：
//   - []model.DatingPurpose: 用户交友目的标签列表
//   - error: 错误信息
//
// 逻辑：
//   1. 查询用户-交友目的关联记录
//   2. 根据目的ID查询交友目的信息
func GetUserDatingPurposes(uid string) ([]model.DatingPurpose, error) {
	var rels []model.UserDatingPurposeRel
	config.DB.Where("uid = ?", uid).Find(&rels)

	var purposeIDs []uint
	for _, rel := range rels {
		purposeIDs = append(purposeIDs, rel.PurposeID)
	}

	var purposes []model.DatingPurpose
	err := config.DB.Where("id IN (?)", purposeIDs).Find(&purposes).Error
	return purposes, err
}

// SaveUserDatingPurposes 保存用户交友目的（先删除再插入）
// 参数说明：
//   - uid: 用户对外唯一标识（雪花ID）
//   - purposeIDs: 交友目的标签ID列表
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. 删除用户现有交友目的关联记录
//   2. 批量插入新的交友目的关联记录
func SaveUserDatingPurposes(uid string, purposeIDs []uint) error {
	tx := config.DB.Begin()

	if err := tx.Where("uid = ?", uid).Delete(&model.UserDatingPurposeRel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, purposeID := range purposeIDs {
		rel := model.UserDatingPurposeRel{
			Uid:       uid,
			PurposeID: purposeID,
		}
		if err := tx.Create(&rel).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}