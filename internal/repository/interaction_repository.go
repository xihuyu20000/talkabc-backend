package repository

import (
	"backend/internal/config"
	"backend/internal/model"
	"time"
)

// GetPraiseMeList 获取谁点赞了我的动态
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.MomentPraise: 点赞记录列表（按创建时间降序）
//   - error: 错误信息
//
// 逻辑：
//   1. 查询用户发布的所有动态ID
//   2. 查询这些动态的所有点赞记录
func GetPraiseMeList(uid uint) ([]model.MomentPraise, error) {
	var praiseIDs []uint
	config.DB.Model(&model.UserMoment{}).Where("user_id = ?", uid).Pluck("id", &praiseIDs)

	var praises []model.MomentPraise
	err := config.DB.Where("moment_id IN (?)", praiseIDs).Order("created_at DESC").Find(&praises).Error
	return praises, err
}

// GetCommentMeList 获取谁评论了我的动态
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.MomentComment: 评论记录列表（按创建时间降序）
//   - error: 错误信息
//
// 逻辑：
//   1. 查询用户发布的所有动态ID
//   2. 查询这些动态的所有评论记录
func GetCommentMeList(uid uint) ([]model.MomentComment, error) {
	var momentIDs []uint
	config.DB.Model(&model.UserMoment{}).Where("user_id = ?", uid).Pluck("id", &momentIDs)

	var comments []model.MomentComment
	err := config.DB.Where("moment_id IN (?)", momentIDs).Order("created_at DESC").Find(&comments).Error
	return comments, err
}

// GetAddMeList 获取好友申请列表
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.AgreeFriend: 好友申请列表（按创建时间降序，状态为待处理）
//   - error: 错误信息
func GetAddMeList(uid uint) ([]model.AgreeFriend, error) {
	var requests []model.AgreeFriend
	err := config.DB.Where("target_id = ? AND status = 0", uid).Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// GetVisitMeList 获取谁访问了我的主页
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.VisitRecord: 访问记录列表（按访问时间降序）
//   - error: 错误信息
func GetVisitMeList(uid uint) ([]model.VisitRecord, error) {
	var visits []model.VisitRecord
	err := config.DB.Where("target_id = ?", uid).Order("visit_time DESC").Find(&visits).Error
	return visits, err
}

// GetLikeMeList 获取谁喜欢了我
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.LikeRecord: 喜欢记录列表（按创建时间降序）
//   - error: 错误信息
func GetLikeMeList(uid uint) ([]model.LikeRecord, error) {
	var likes []model.LikeRecord
	err := config.DB.Where("target_id = ?", uid).Order("created_at DESC").Find(&likes).Error
	return likes, err
}

// LikeUser 喜欢/取消喜欢用户
// 参数说明：
//   - userID: 当前用户数据库ID
//   - targetID: 目标用户数据库ID
//   - flag: 操作标志，1-喜欢，0-取消喜欢
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. flag=1时，检查是否已喜欢，未喜欢则创建记录
//   2. flag=0时，删除已存在的喜欢记录
func LikeUser(userID, targetID uint, flag int) error {
	if flag == 1 {
		var exist model.LikeRecord
		if err := config.DB.Where("user_id = ? AND target_id = ?", userID, targetID).First(&exist).Error; err == nil {
			return nil
		}
		like := model.LikeRecord{
			UserID:   userID,
			TargetID: targetID,
		}
		return config.DB.Create(&like).Error
	} else {
		return config.DB.Where("user_id = ? AND target_id = ?", userID, targetID).Delete(&model.LikeRecord{}).Error
	}
}

// AgreeFriendRequest 处理好友申请
// 参数说明：
//   - userID: 当前用户数据库ID
//   - targetID: 申请者数据库ID
//   - flag: 处理标志，1-同意，2-拒绝
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. 查询对方发起的好友申请
//   2. 存在则更新状态，不存在则创建新记录
func AgreeFriendRequest(userID, targetID uint, flag int) error {
	var req model.AgreeFriend
	config.DB.Where("user_id = ? AND target_id = ?", targetID, userID).First(&req)

	if req.ID != 0 {
		req.Status = flag
		return config.DB.Save(&req).Error
	}

	req = model.AgreeFriend{
		UserID:   userID,
		TargetID: targetID,
		Status:   flag,
	}
	return config.DB.Create(&req).Error
}

// AddVisitRecord 添加访问记录
// 参数说明：
//   - visitorID: 访问者数据库ID
//   - targetID: 被访问者数据库ID
//
// 返回值：
//   - error: 错误信息
func AddVisitRecord(visitorID, targetID uint) error {
	visit := model.VisitRecord{
		VisitorID: visitorID,
		TargetID:  targetID,
		VisitTime: time.Now(),
	}
	return config.DB.Create(&visit).Error
}