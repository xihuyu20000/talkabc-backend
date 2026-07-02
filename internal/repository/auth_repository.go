package repository

import (
	"backend/internal/config"
	"backend/internal/model"
	"context"
	"fmt"
	"time"
)

const (
	VerificationCodeTypeSMS   = "sms"
	VerificationCodeTypeAlnum = "alnum"
)

// GetUserByPhone 根据手机号查询用户
// 参数说明：
//   - phoneNum: 手机号
//
// 返回值：
//   - *model.User: 用户模型指针
//   - error: 错误信息
func GetUserByPhone(phoneNum string) (*model.User, error) {
	var user model.User
	err := config.DB.Where("phone_num = ?", phoneNum).First(&user).Error
	return &user, err
}

// CreateUser 创建用户
// 参数说明：
//   - user: 用户模型指针
//
// 返回值：
//   - error: 错误信息
func CreateUser(user *model.User) error {
	return config.DB.Create(user).Error
}

// UpdateUser 更新用户信息
// 参数说明：
//   - user: 用户模型指针（包含更新后的字段）
//
// 返回值：
//   - error: 错误信息
func UpdateUser(user *model.User) error {
	return config.DB.Save(user).Error
}

// getVerificationCodeKey 生成验证码Redis键名
// 参数说明：
//   - phoneNum: 手机号
//   - codeType: 验证码类型（如sms、register等）
//
// 返回值：
//   - string: Redis键名，格式为 verification_code:{phoneNum}:{codeType}
func getVerificationCodeKey(phoneNum, codeType string) string {
	return fmt.Sprintf("verification_code:%s:%s", phoneNum, codeType)
}

// CreateVerificationCode 创建验证码（存储到Redis）
// 参数说明：
//   - phoneNum: 手机号
//   - code: 验证码
//   - codeType: 验证码类型
//
// 返回值：
//   - error: 错误信息
//
// 逻辑：
//   1. 生成Redis键名
//   2. 获取配置的验证码有效期（默认为5分钟）
//   3. 将验证码存入Redis并设置过期时间
func CreateVerificationCode(phoneNum, code, codeType string) error {
	key := getVerificationCodeKey(phoneNum, codeType)
	expireMinutes := config.AppConfig.System.SMSValidMinutes
	if expireMinutes <= 0 {
		expireMinutes = 5
	}
	// Set: 设置Redis字符串键值对，并指定过期时间
	// 业务含义：存储验证码，过期自动删除（默认5分钟），防止暴力破解
	return config.RDB.Set(context.Background(), key, code, time.Duration(expireMinutes)*time.Minute).Err()
}

// GetValidVerificationCode 验证验证码是否有效
// 参数说明：
//   - phoneNum: 手机号
//   - code: 用户输入的验证码
//   - codeType: 验证码类型
//
// 返回值：
//   - error: 错误信息（验证码不匹配或已过期）
//
// 逻辑：
//   1. 从Redis获取存储的验证码
//   2. 比较验证码是否匹配
//   3. 验证通过不删除验证码，允许有限次数重试
func GetValidVerificationCode(phoneNum, code, codeType string) error {
	key := getVerificationCodeKey(phoneNum, codeType)
	// Get: 获取Redis字符串键的值
	// 业务含义：从Redis读取存储的验证码，用于验证用户输入是否正确
	storedCode, err := config.RDB.Get(context.Background(), key).Result()
	if err != nil {
		return err
	}
	if storedCode != code {
		return fmt.Errorf("验证码不匹配")
	}
	return nil
}

// DeleteVerificationCode 删除验证码（验证通过后清理）
// 参数说明：
//   - phoneNum: 手机号
//   - codeType: 验证码类型
//
// 返回值：
//   - error: 错误信息
func DeleteVerificationCode(phoneNum, codeType string) error {
	key := getVerificationCodeKey(phoneNum, codeType)
	return config.RDB.Del(context.Background(), key).Err()
}