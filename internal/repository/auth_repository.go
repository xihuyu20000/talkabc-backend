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

	SMSRateLimitKeyPrefix     = "sms_rate_limit:"
	// 【功能8】1小时内发送次数统计key前缀
	SMSHourlyCountKeyPrefix = "sms_hourly_count:"
	// 【功能9】每日首次发送标记key前缀
	SMSDailyFirstKeyPrefix = "sms_daily_first:"
	// 【功能3】60秒冷却期key前缀
	SMSCooldownKeyPrefix = "sms_cooldown:"

	// 【注册安全规则1】IP黑名单key前缀
	IPBlacklistKeyPrefix = "ip_blacklist:"
	// 【注册安全规则2】注册IP频率限制key前缀
	RegisterIPRateLimitPrefix = "register_ip_rate_limit:"
	// 【注册安全规则3】手机号黑名单key前缀
	PhoneBlacklistKeyPrefix = "phone_blacklist:"
)

// GetUserByPhone 根据手机号查询用户
func GetUserByPhone(phoneNum string) (*model.User, error) {
	var user model.User
	err := config.DB.Where("phone_num = ?", phoneNum).First(&user).Error
	return &user, err
}

// CreateUser 创建用户
func CreateUser(user *model.User) error {
	return config.DB.Create(user).Error
}

// UpdateUser 更新用户信息
func UpdateUser(user *model.User) error {
	return config.DB.Save(user).Error
}

// ==================== 验证码基础操作 ====================

// getVerificationCodeKey 生成验证码Redis键名
// 【功能10】键名格式：verification_code:{phoneNum}:{codeType}:{tag}，支持业务标签隔离
func getVerificationCodeKey(phoneNum, codeType, tag string) string {
	return fmt.Sprintf("verification_code:%s:%s:%s", phoneNum, codeType, tag)
}

// CreateVerificationCode 创建验证码（存储到Redis）
// 【功能1】验证码有效期通过Redis TTL控制（默认5分钟）
// 【功能4】新验证码使用SET操作覆盖旧值，保证同一时间只有一个有效验证码
func CreateVerificationCode(phoneNum, code, codeType, tag string) error {
	key := getVerificationCodeKey(phoneNum, codeType, tag)
	expireMinutes := config.AppConfig.System.SMSValidMinutes
	if expireMinutes <= 0 {
		// 【功能1】默认有效期5分钟
		expireMinutes = 5
	}
	return config.RDB.Set(context.Background(), key, code, time.Duration(expireMinutes)*time.Minute).Err()
}

// VerifyAndDeleteVerificationCode 验证并删除验证码（原子操作，保证一次性使用）
// 【功能5】使用Lua脚本保证验证和删除的原子性，防止并发攻击和暴力破解
// 无论验证是否成功，验证码都会被删除，确保一次性使用
func VerifyAndDeleteVerificationCode(phoneNum, code, codeType, tag string) (bool, error) {
	key := getVerificationCodeKey(phoneNum, codeType, tag)

	luaScript := `
		local storedCode = redis.call('GET', KEYS[1])
		if storedCode == false then
			return 0
		end
		if storedCode == ARGV[1] then
			redis.call('DEL', KEYS[1])
			return 1
		end
		redis.call('DEL', KEYS[1])
		return 0
	`

	result, err := config.RDB.Eval(context.Background(), luaScript, []string{key}, code).Result()
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}

// ==================== 发送频率限制 ====================

// CheckSMSCooldown 检查手机号是否在冷却期（60秒内只能发送一次）
// 【功能3】服务器端校验60秒冷却期
func CheckSMSCooldown(phoneNum string) bool {
	key := fmt.Sprintf("%s%s", SMSCooldownKeyPrefix, phoneNum)
	exists, _ := config.RDB.Exists(context.Background(), key).Result()
	return exists > 0
}

// SetSMSCooldown 设置手机号冷却期（60秒）
// 【功能3】发送成功后设置60秒冷却期，防止频繁发送
func SetSMSCooldown(phoneNum string) error {
	key := fmt.Sprintf("%s%s", SMSCooldownKeyPrefix, phoneNum)
	return config.RDB.Set(context.Background(), key, "1", 60*time.Second).Err()
}

// CheckHourlyLimit 检查1小时内发送次数是否超过限制（10次）
// 【功能8】使用Redis计数器实现1小时内发送次数限制
func CheckHourlyLimit(phoneNum string) (bool, error) {
	key := fmt.Sprintf("%s%s", SMSHourlyCountKeyPrefix, phoneNum)
	
	current, err := config.RDB.Incr(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	if current == 1 {
		config.RDB.Expire(context.Background(), key, 1*time.Hour)
	}

	return current > 10, nil
}

// CheckDailyFirst 获取今日首次发送标记（24小时内首次发送不需要图形验证码）
// 【功能9】使用Redis标记实现每日首次发送免图形验证码
func CheckDailyFirst(phoneNum string) (bool, error) {
	key := fmt.Sprintf("%s%s", SMSDailyFirstKeyPrefix, phoneNum)
	
	exists, err := config.RDB.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	if exists == 0 {
		config.RDB.Set(context.Background(), key, "1", 24*time.Hour)
		return true, nil
	}

	return false, nil
}

// ==================== 图形验证码操作 ====================

func GetCaptchaCode(captchaID string) (string, error) {
	return config.RDB.Get(context.Background(), captchaID).Result()
}

func DeleteCaptchaCode(captchaID string) error {
	return config.RDB.Del(context.Background(), captchaID).Err()
}

// ==================== 注册安全规则 ====================

// CheckIPBlacklist 检查IP是否在黑名单中
// 【注册安全规则1】服务器检验通过IP是否位于黑名单中，是则拦截
func CheckIPBlacklist(ip string) bool {
	key := fmt.Sprintf("%s%s", IPBlacklistKeyPrefix, ip)
	exists, _ := config.RDB.Exists(context.Background(), key).Result()
	return exists > 0
}

// CheckRegisterIPRateLimit 检查IP注册请求频率（1分钟10次）
// 【注册安全规则2】服务器检验通过IP检查注册请求的频率，请求频率限制在1分钟10次
func CheckRegisterIPRateLimit(ip string) (bool, error) {
	key := fmt.Sprintf("%s%s", RegisterIPRateLimitPrefix, ip)

	current, err := config.RDB.Incr(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	if current == 1 {
		config.RDB.Expire(context.Background(), key, 1*time.Minute)
	}

	return current > 10, nil
}

// CheckPhoneBlacklist 检查手机号是否在黑名单中
// 【注册安全规则3】校验手机号是否位于黑名单中，是则拦截
func CheckPhoneBlacklist(phoneNum string) bool {
	key := fmt.Sprintf("%s%s", PhoneBlacklistKeyPrefix, phoneNum)
	exists, _ := config.RDB.Exists(context.Background(), key).Result()
	return exists > 0
}