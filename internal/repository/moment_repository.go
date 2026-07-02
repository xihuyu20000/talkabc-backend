package repository

import (
	"backend/internal/config"
	"backend/internal/model"
	"time"
)

// GetLatestMoment 获取最新动态列表
// 返回值：
//   - []model.UserMoment: 动态列表（按发布时间降序，最多20条）
//   - error: 错误信息
func GetLatestMoment() ([]model.UserMoment, error) {
	var moments []model.UserMoment
	err := config.DB.Order("pub_ts DESC").Limit(20).Find(&moments).Error
	return moments, err
}

// GetMyLatestMoment 获取当前用户最新动态列表
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.UserMoment: 动态列表（按发布时间降序，最多20条）
//   - error: 错误信息
func GetMyLatestMoment(uid uint) ([]model.UserMoment, error) {
	var moments []model.UserMoment
	err := config.DB.Where("user_id = ?", uid).Order("pub_ts DESC").Limit(20).Find(&moments).Error
	return moments, err
}

// GetUserMoment 获取指定用户动态列表
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.UserMoment: 动态列表（按发布时间降序，最多20条）
//   - error: 错误信息
func GetUserMoment(uid uint) ([]model.UserMoment, error) {
	var moments []model.UserMoment
	err := config.DB.Where("user_id = ?", uid).Order("pub_ts DESC").Limit(20).Find(&moments).Error
	return moments, err
}

// GetMomentByID 根据ID获取单个动态
// 参数说明：
//   - mid: 动态ID
//
// 返回值：
//   - *model.UserMoment: 动态模型指针
//   - error: 错误信息
func GetMomentByID(mid uint) (*model.UserMoment, error) {
	var moment model.UserMoment
	err := config.DB.Where("id = ?", mid).First(&moment).Error
	return &moment, err
}

// AddMomentPraise 点赞动态
// 参数说明：
//   - userID: 用户数据库ID
//   - momentID: 动态ID
//
// 返回值：
//   - error: 错误信息
func AddMomentPraise(userID, momentID uint) error {
	praise := model.MomentPraise{
		UserID:   userID,
		MomentID: momentID,
	}
	return config.DB.Create(&praise).Error
}

// RemoveMomentPraise 取消点赞动态
// 参数说明：
//   - userID: 用户数据库ID
//   - momentID: 动态ID
//
// 返回值：
//   - error: 错误信息
func RemoveMomentPraise(userID, momentID uint) error {
	return config.DB.Where("user_id = ? AND moment_id = ?", userID, momentID).Delete(&model.MomentPraise{}).Error
}

// AddMomentComment 评论动态
// 参数说明：
//   - userID: 用户数据库ID
//   - momentID: 动态ID
//   - text: 评论内容
//
// 返回值：
//   - error: 错误信息
func AddMomentComment(userID, momentID uint, text string) error {
	comment := model.MomentComment{
		UserID:   userID,
		MomentID: momentID,
		Text:     text,
	}
	return config.DB.Create(&comment).Error
}

// CreateMoment 创建动态
// 参数说明：
//   - userID: 用户数据库ID
//   - text: 动态文字内容
//   - files: 图片/视频URL数组
//   - location: 地理位置
//
// 返回值：
//   - error: 错误信息
func CreateMoment(userID uint, text string, files []string, location string) error {
	moment := model.UserMoment{
		UserID:    userID,
		Text:      text,
		Files:     files,
		Location:  location,
		PraiseNum: 0,
		PubTS:     time.Now().Unix(),
	}
	return config.DB.Create(&moment).Error
}

// UpdateMomentPraiseNum 更新动态点赞数
// 参数说明：
//   - momentID: 动态ID
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. 统计该动态的点赞数量
//   2. 更新动态的点赞数字段
func UpdateMomentPraiseNum(momentID uint) error {
	var count int
	config.DB.Model(&model.MomentPraise{}).Where("moment_id = ?", momentID).Count(&count)
	return config.DB.Model(&model.UserMoment{}).Where("id = ?", momentID).Update("praise_num", count).Error
}

// GetMomentComments 获取动态评论列表
// 参数说明：
//   - momentID: 动态ID
//
// 返回值：
//   - []model.MomentComment: 评论列表（按创建时间降序）
//   - error: 错误信息
func GetMomentComments(momentID uint) ([]model.MomentComment, error) {
	var comments []model.MomentComment
	err := config.DB.Where("moment_id = ?", momentID).Order("created_at DESC").Find(&comments).Error
	return comments, err
}