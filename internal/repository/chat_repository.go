package repository

import (
	"backend/internal/config"
	"backend/internal/model"
	"time"
)

// GetLatestAdBanner 获取最新广告横幅列表
// 返回值：
//   - []model.AdBanner: 广告横幅列表（按优先级降序）
//   - error: 错误信息
//
// 逻辑：
//   1. 查询未过期的广告（end_time > 当前时间）
//   2. 按优先级降序排列
func GetLatestAdBanner() ([]model.AdBanner, error) {
	var banners []model.AdBanner
	err := config.DB.Where("end_time > ?", time.Now()).Order("priority DESC").Find(&banners).Error
	return banners, err
}

// GetSystemMsgList 获取系统消息列表
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.SystemMsg: 系统消息列表（按创建时间降序）
//   - error: 错误信息
func GetSystemMsgList(uid uint) ([]model.SystemMsg, error) {
	var msgs []model.SystemMsg
	err := config.DB.Where("user_id = ?", uid).Order("created_at DESC").Find(&msgs).Error
	return msgs, err
}

// GetLatestUserMsg 获取用户最新聊天消息
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.ChatMessage: 最新20条聊天消息（按发送时间降序）
//   - error: 错误信息
func GetLatestUserMsg(uid uint) ([]model.ChatMessage, error) {
	var msgs []model.ChatMessage
	err := config.DB.Where("sender_id = ? OR receiver_id = ?", uid, uid).
		Order("send_time DESC").Limit(20).Find(&msgs).Error
	return msgs, err
}

// GetUserMsgHistory 获取用户聊天历史记录
// 参数说明：
//   - uid: 当前用户数据库ID
//   - targetID: 聊天对象数据库ID
//
// 返回值：
//   - []model.ChatMessage: 聊天记录列表（按发送时间升序）
//   - error: 错误信息
func GetUserMsgHistory(uid, targetID uint) ([]model.ChatMessage, error) {
	var msgs []model.ChatMessage
	err := config.DB.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		uid, targetID, targetID, uid).Order("send_time ASC").Find(&msgs).Error
	return msgs, err
}

// SetMessageTop 设置消息置顶
// 参数说明：
//   - uid: 用户数据库ID
//   - targetID: 聊天对象数据库ID
//   - flag: 置顶标志，0-取消置顶，1-置顶
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. 查询已有置顶记录
//   2. 不存在则创建，存在则更新
func SetMessageTop(uid, targetID uint, flag int) error {
	var top model.UserMessageTop
	config.DB.Where("user_id = ? AND target_id = ?", uid, targetID).First(&top)

	if top.ID == 0 {
		top = model.UserMessageTop{
			UserID:   uid,
			TargetID: targetID,
			Top:      flag,
		}
		return config.DB.Create(&top).Error
	}

	top.Top = flag
	return config.DB.Save(&top).Error
}

// AddFriend 添加好友
// 参数说明：
//   - uid: 用户数据库ID
//   - targetID: 目标用户数据库ID
//   - flag: 状态标志，0-待确认，1-已添加
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. 查询已有好友关系
//   2. 不存在则创建，存在则更新状态
func AddFriend(uid, targetID uint, flag int) error {
	var friend model.UserFriend
	config.DB.Where("user_id = ? AND target_id = ?", uid, targetID).First(&friend)

	if friend.ID == 0 {
		friend = model.UserFriend{
			UserID:   uid,
			TargetID: targetID,
			Status:   flag,
		}
		return config.DB.Create(&friend).Error
	}

	friend.Status = flag
	return config.DB.Save(&friend).Error
}

// ClearChatHistory 清空聊天记录
// 参数说明：
//   - uid: 当前用户数据库ID
//   - targetID: 聊天对象数据库ID
//
// 返回值：
//   - error: 错误信息
func ClearChatHistory(uid, targetID uint) error {
	return config.DB.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		uid, targetID, targetID, uid).Delete(&model.ChatMessage{}).Error
}

// SendTextMessage 发送文本消息
// 参数说明：
//   - senderID: 发送者数据库ID
//   - receiverID: 接收者数据库ID
//   - text: 消息内容
//
// 返回值：
//   - error: 错误信息
func SendTextMessage(senderID, receiverID uint, text string) error {
	msg := model.ChatMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Text:       text,
		MsgType:    1,
		ReadStatus: 0,
		SendTime:   time.Now(),
	}
	return config.DB.Create(&msg).Error
}

// SendBinaryMessage 发送二进制消息（图片/音视频，已废弃）
// 参数说明：
//   - senderID: 发送者数据库ID
//   - receiverID: 接收者数据库ID
//   - fileURL: 文件URL
//
// 返回值：
//   - error: 错误信息
//
// 说明：该方法已废弃，推荐使用SendImageMessage、SendVoiceMessage、SendVideoMessage或SendFileMessage
func SendBinaryMessage(senderID, receiverID uint, fileURL string) error {
	msg := model.ChatMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		FileURL:    fileURL,
		MsgType:    2,
		ReadStatus: 0,
		SendTime:   time.Now(),
	}
	return config.DB.Create(&msg).Error
}

// WithdrawMessage 撤回消息
// 参数说明：
//   - senderID: 发送者数据库ID
//   - receiverID: 接收者数据库ID
//   - msgID: 消息ID
//
// 返回值：
//   - error: 错误信息
func WithdrawMessage(senderID, receiverID, msgID uint) error {
	return config.DB.Where("id = ? AND sender_id = ?", msgID, senderID).Delete(&model.ChatMessage{}).Error
}

// SendImageMessage 发送图片消息
// 参数说明：
//   - senderID: 发送者数据库ID
//   - receiverID: 接收者数据库ID
//   - fileURL: 图片URL
//
// 返回值：
//   - error: 错误信息
func SendImageMessage(senderID, receiverID uint, fileURL string) error {
	msg := model.ChatMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		FileURL:    fileURL,
		MsgType:    2,
		ReadStatus: 0,
		SendTime:   time.Now(),
	}
	return config.DB.Create(&msg).Error
}

// SendVoiceMessage 发送语音消息
// 参数说明：
//   - senderID: 发送者数据库ID
//   - receiverID: 接收者数据库ID
//   - fileURL: 视频文件URL
//
// 返回值：
//   - error: 错误信息
func SendVoiceMessage(senderID, receiverID uint, fileURL string) error {
	msg := model.ChatMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		FileURL:    fileURL,
		MsgType:    3,
		ReadStatus: 0,
		SendTime:   time.Now(),
	}
	return config.DB.Create(&msg).Error
}

// SendVideoMessage 发送视频消息
// 参数说明：
//   - senderID: 发送者数据库ID
//   - receiverID: 接收者数据库ID
//   - fileURL: 视频文件URL
//
// 返回值：
//   - error: 错误信息
func SendVideoMessage(senderID, receiverID uint, fileURL string) error {
	msg := model.ChatMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		FileURL:    fileURL,
		MsgType:    4,
		ReadStatus: 0,
		SendTime:   time.Now(),
	}
	return config.DB.Create(&msg).Error
}

// SendFileMessage 发送文件消息
// 参数说明：
//   - senderID: 发送者数据库ID
//   - receiverID: 接收者数据库ID
//   - fileURL: 文件URL
//
// 返回值：
//   - error: 错误信息
func SendFileMessage(senderID, receiverID uint, fileURL string) error {
	msg := model.ChatMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		FileURL:    fileURL,
		MsgType:    5,
		ReadStatus: 0,
		SendTime:   time.Now(),
	}
	return config.DB.Create(&msg).Error
}

// SendGift 发送礼物（预留方法）
// 参数说明：
//   - senderID: 发送者数据库ID
//   - receiverID: 接收者数据库ID
//   - giftID: 礼物ID
//
// 返回值：
//   - error: 错误信息（当前返回nil）
func SendGift(senderID, receiverID, giftID uint) error {
	return nil
}