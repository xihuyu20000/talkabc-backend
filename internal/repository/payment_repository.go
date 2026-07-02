package repository

import (
	"backend/internal/config"
	"backend/internal/model"
	"github.com/jinzhu/gorm"
	"time"
)

// GetDiamond 获取用户钻石账户
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - *model.Diamond: 钻石账户模型指针
//   - error: 错误信息
//
// 逻辑：
//   1. 查询用户钻石账户
//   2. 不存在则创建默认账户（粉钻0，蓝钻0）
func GetDiamond(uid uint) (*model.Diamond, error) {
	var diamond model.Diamond
	err := config.DB.Where("user_id = ?", uid).First(&diamond).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if diamond.ID == 0 {
		diamond = model.Diamond{
			UserID:      uid,
			PinkDiamond: 0,
			BlueDiamond: 0,
		}
		config.DB.Create(&diamond)
	}
	return &diamond, nil
}

// UpdateDiamond 更新用户钻石数量
// 参数说明：
//   - uid: 用户数据库ID
//   - pink: 粉钻变动数量（正数增加，负数减少）
//   - blue: 蓝钻变动数量（正数增加，负数减少）
//
// 返回值：
//   - error: 错误信息
func UpdateDiamond(uid uint, pink, blue int) error {
	diamond, err := GetDiamond(uid)
	if err != nil {
		return err
	}

	diamond.PinkDiamond += pink
	diamond.BlueDiamond += blue
	return config.DB.Save(diamond).Error
}

// CreateDiamondRecord 创建钻石交易记录
// 参数说明：
//   - uid: 用户数据库ID
//   - diamondType: 钻石类型，1-粉钻，2-蓝钻
//   - amount: 交易数量
//   - orderID: 订单号
//
// 返回值：
//   - error: 错误信息
func CreateDiamondRecord(uid uint, diamondType, amount int, orderID string) error {
	record := model.DiamondRecord{
		UserID:  uid,
		Type:    diamondType,
		Amount:  amount,
		OrderID: orderID,
	}
	return config.DB.Create(&record).Error
}

// GetDiamondRecords 获取钻石交易记录列表
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.DiamondRecord: 交易记录列表（按创建时间降序）
//   - error: 错误信息
func GetDiamondRecords(uid uint) ([]model.DiamondRecord, error) {
	var records []model.DiamondRecord
	err := config.DB.Where("user_id = ?", uid).Order("created_at DESC").Find(&records).Error
	return records, err
}

// GetMember 获取用户会员信息
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - *model.Member: 会员模型指针
//   - error: 错误信息
//
// 逻辑：
//   1. 查询用户会员信息
//   2. 不存在则创建默认会员（等级0，到期时间为当前时间）
func GetMember(uid uint) (*model.Member, error) {
	var member model.Member
	err := config.DB.Where("user_id = ?", uid).First(&member).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if member.ID == 0 {
		member = model.Member{
			UserID:   uid,
			Level:    0,
			ExpireAt: time.Now(),
		}
		config.DB.Create(&member)
	}
	return &member, nil
}

// UpdateMember 更新用户会员等级和到期时间
// 参数说明：
//   - uid: 用户数据库ID
//   - level: 会员等级，1-月度，2-季度，3-年度，99-永久
//   - expireDays: 有效期天数
//
// 返回值：
//   - error: 错误信息
func UpdateMember(uid uint, level int, expireDays int) error {
	member, err := GetMember(uid)
	if err != nil {
		return err
	}

	member.Level = level
	member.ExpireAt = time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)
	return config.DB.Save(member).Error
}

// CreateMemberRecord 创建会员购买记录
// 参数说明：
//   - uid: 用户数据库ID
//   - level: 会员等级
//   - orderID: 订单号
//
// 返回值：
//   - error: 错误信息
func CreateMemberRecord(uid uint, level int, orderID string) error {
	record := model.MemberRecord{
		UserID:  uid,
		Level:   level,
		OrderID: orderID,
	}
	return config.DB.Create(&record).Error
}

// GetMemberRecords 获取会员购买记录列表
// 参数说明：
//   - uid: 用户数据库ID
//
// 返回值：
//   - []model.MemberRecord: 购买记录列表（按创建时间降序）
//   - error: 错误信息
func GetMemberRecords(uid uint) ([]model.MemberRecord, error) {
	var records []model.MemberRecord
	err := config.DB.Where("user_id = ?", uid).Order("created_at DESC").Find(&records).Error
	return records, err
}