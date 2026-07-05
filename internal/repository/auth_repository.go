package repository

import (
	"backend/internal/config"
	"backend/internal/model"
	"backend/pkg/logger"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
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

	// 【登录安全规则2】登录IP频率限制key前缀
	LoginIPRateLimitPrefix = "login_ip_rate_limit:"
	// 【登录安全规则4】设备黑名单key前缀
	DeviceBlacklistKeyPrefix = "device_blacklist:"
)

// GetUserByPhone 根据手机号查询用户
func GetUserByPhone(phoneNum string) (*model.User, error) {
	var user model.User
	err := config.DB.Where("phone_num = ?", phoneNum).First(&user).Error
	return &user, err
}

// ClearUserLoginState 清空用户全部登录态
// 【密码存储加密】重置成功后清空该用户全部登录态（Redis token、JWT、设备登录记录全部销毁）
// 【刷新令牌安全规则】同时清除刷新令牌，防止被滥用
func ClearUserLoginState(userID uint) error {
	user, err := GetUserByID(userID)
	if err != nil {
		return err
	}
	
	tokenKey := fmt.Sprintf("user_token:%s", user.Uid)
	refreshTokenKey := fmt.Sprintf("user_refresh_token:%s", user.Uid)
	config.RDB.Del(context.Background(), tokenKey)
	config.RDB.Del(context.Background(), refreshTokenKey)
	
	return nil
}

// SaveUserToken 保存用户token到Redis
// 【安全规则】登录成功后将token保存到Redis，支持主动失效（如更换手机号、修改密码后）
func SaveUserToken(uid, token string) error {
	tokenKey := fmt.Sprintf("user_token:%s", uid)
	return config.RDB.Set(context.Background(), tokenKey, token, 7*24*time.Hour).Err()
}

// SaveRefreshToken 保存刷新令牌到Redis
// 【安全规则】刷新令牌存储在Redis中，key为用户ID，value为刷新令牌，支持主动失效
// 一个用户只允许一个有效的刷新令牌（单设备登录）
func SaveRefreshToken(uid, refreshToken string) error {
	tokenKey := fmt.Sprintf("user_refresh_token:%s", uid)
	return config.RDB.Set(context.Background(), tokenKey, refreshToken, 7*24*time.Hour).Err()
}

// ValidateRefreshToken 验证刷新令牌是否有效
// 【安全规则】检查Redis中存储的刷新令牌是否与请求中的一致
func ValidateRefreshToken(uid, refreshToken string) (bool, error) {
	tokenKey := fmt.Sprintf("user_refresh_token:%s", uid)
	storedToken, err := config.RDB.Get(context.Background(), tokenKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return storedToken == refreshToken, nil
}

// InvalidateRefreshToken 使刷新令牌失效
// 【安全规则】退出登录时清除刷新令牌，防止被滥用
func InvalidateRefreshToken(uid string) error {
	tokenKey := fmt.Sprintf("user_refresh_token:%s", uid)
	return config.RDB.Del(context.Background(), tokenKey).Err()
}

// CreateUser 创建用户
func CreateUser(user *model.User) error {
	return config.DB.Create(user).Error
}

// UpdateUser 更新用户信息
func UpdateUser(user *model.User) error {
	return config.DB.Save(user).Error
}

// CheckPhoneExists 检查手机号是否已存在
// 【更换手机号安全规则3】新手机号必须未被注册
func CheckPhoneExists(phoneNum string) (bool, error) {
	var count int
	err := config.DB.Model(&model.User{}).Where("phone_num = ?", phoneNum).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateUserPhone 更新用户手机号
// 【更换手机号】更新用户手机号，并清空用户登录态
func UpdateUserPhone(userID uint, newPhoneNum string) error {
	user, err := GetUserByID(userID)
	if err != nil {
		return err
	}

	user.PhoneNum = newPhoneNum
	if err := config.DB.Save(user).Error; err != nil {
		return err
	}

	return ClearUserLoginState(userID)
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
	expireMinutes := config.AppConfig.Security.SMSValidMinutes
	if expireMinutes <= 0 {
		expireMinutes = 5
	}
	err := config.RDB.Set(context.Background(), key, code, time.Duration(expireMinutes)*time.Minute).Err()
	logger.Debugf("[Redis] SET key=%s, ttl=%dmin, err=%v", key, expireMinutes, err)
	return err
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

// VerifyVerificationCode 验证验证码（不删除，用于更换手机号等场景）
// 【更换手机号安全规则4】仅验证验证码有效性，不删除，由后续流程清理
func VerifyVerificationCode(phoneNum, code, codeType, tag string) (bool, error) {
	key := getVerificationCodeKey(phoneNum, codeType, tag)

	storedCode, err := config.RDB.Get(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	return storedCode == code, nil
}

// ClearVerificationCode 清理验证码
// 【更换手机号安全规则6】验证通过后清理验证码，防止二次复用
func ClearVerificationCode(phoneNum, codeType, tag string) error {
	key := getVerificationCodeKey(phoneNum, codeType, tag)
	return config.RDB.Del(context.Background(), key).Err()
}

// CheckChangePhoneRateLimit 检查更换手机号频率限制（24小时内最多3次）
// 【更换手机号安全规则5】同一账号24小时最多允许更换3次手机号
func CheckChangePhoneRateLimit(userID uint) (bool, error) {
	key := fmt.Sprintf("change_phone_rate_limit:%d", userID)

	current, err := config.RDB.Incr(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	if current == 1 {
		config.RDB.Expire(context.Background(), key, 24*time.Hour)
	}

	return current > 3, nil
}

// ==================== 发送频率限制 ====================

// CheckSMSCooldown 检查手机号是否在冷却期（60秒内只能发送一次）
// 【功能3】服务器端校验60秒冷却期
func CheckSMSCooldown(phoneNum string) bool {
	key := fmt.Sprintf("%s%s", SMSCooldownKeyPrefix, phoneNum)
	exists, err := config.RDB.Exists(context.Background(), key).Result()
	logger.Debugf("[Redis] EXISTS key=%s, result=%d, err=%v", key, exists, err)
	return exists > 0
}

// SetSMSCooldown 设置手机号冷却期（60秒）
// 【功能3】发送成功后设置60秒冷却期，防止频繁发送
func SetSMSCooldown(phoneNum string) error {
	key := fmt.Sprintf("%s%s", SMSCooldownKeyPrefix, phoneNum)
	ttl := time.Duration(config.AppConfig.Security.SMSCooldownSeconds) * time.Second
	err := config.RDB.Set(context.Background(), key, "1", ttl).Err()
	logger.Debugf("[Redis] SET key=%s, value=1, ttl=%ds, err=%v", key, config.AppConfig.Security.SMSCooldownSeconds, err)
	return err
}

// CheckHourlyLimit 检查1小时内发送次数是否超过限制（10次）
// 【功能8】使用Redis计数器实现1小时内发送次数限制
func CheckHourlyLimit(phoneNum string) (bool, error) {
	key := fmt.Sprintf("%s%s", SMSHourlyCountKeyPrefix, phoneNum)
	
	_, err := config.RDB.SetNX(context.Background(), key, "0", 1*time.Hour).Result()
	if err != nil {
		return false, err
	}

	current, err := config.RDB.Incr(context.Background(), key).Result()
	logger.Debugf("[Redis] INCR key=%s, result=%d, err=%v", key, current, err)
	if err != nil {
		return false, err
	}

	expireErr := config.RDB.Expire(context.Background(), key, 1*time.Hour).Err()
	logger.Debugf("[Redis] EXPIRE key=%s, ttl=1h, err=%v", key, expireErr)

	limit := config.AppConfig.Security.SMSHourlyLimit
	exceeded := current > int64(limit)
	logger.Debugf("[SMS] Hourly limit check - key=%s, current=%d, limit=%d, exceeded=%v", key, current, limit, exceeded)
	return exceeded, nil
}

// CheckDailyFirst 获取今日首次发送标记（24小时内首次发送不需要图形验证码）
// 【功能9】使用Redis标记实现每日首次发送免图形验证码
func CheckDailyFirst(phoneNum string) (bool, error) {
	key := fmt.Sprintf("%s%s", SMSDailyFirstKeyPrefix, phoneNum)
	
	exists, err := config.RDB.Exists(context.Background(), key).Result()
	logger.Debugf("[Redis] EXISTS key=%s, result=%d, err=%v", key, exists, err)
	if err != nil {
		return false, err
	}

	if exists == 0 {
		setErr := config.RDB.Set(context.Background(), key, "1", 24*time.Hour).Err()
		logger.Debugf("[Redis] SET key=%s, value=1, ttl=24h, err=%v", key, setErr)
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
// 【注册安全规则1】【登录安全规则1】服务器检验通过IP是否位于黑名单中，是则拦截
func CheckIPBlacklist(ip string) bool {
	key := fmt.Sprintf("%s%s", IPBlacklistKeyPrefix, ip)
	exists, _ := config.RDB.Exists(context.Background(), key).Result()
	return exists > 0
}

// CheckRegisterIPRateLimit 检查IP注册请求频率（1小时内每个IP发送次数限制，单位次）
// 【注册安全规则2】服务器检验通过IP检查注册请求的频率，请求频率限制在1小时内发送次数限制，单位次
func CheckRegisterIPRateLimit(ip string) (bool, error) {
	key := fmt.Sprintf("%s%s", RegisterIPRateLimitPrefix, ip)

	current, err := config.RDB.Incr(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	if current == 1 {
		config.RDB.Expire(context.Background(), key, 1*time.Hour)
	}

	return current > int64(config.AppConfig.Security.IPRegisterHourlyLimit), nil
}

// CheckPhoneBlacklist 检查手机号是否在黑名单中
// 【注册安全规则3】【登录安全规则3】校验手机号是否位于黑名单中，是则拦截
func CheckPhoneBlacklist(phoneNum string) bool {
	key := fmt.Sprintf("%s%s", PhoneBlacklistKeyPrefix, phoneNum)
	exists, _ := config.RDB.Exists(context.Background(), key).Result()
	return exists > 0
}

// ==================== 登录安全规则 ====================

// CheckLoginIPRateLimit 检查IP登录请求频率（1小时内每个IP发送次数限制，单位次）
// 【登录安全规则2】服务器检验通过IP检查登录请求的频率，请求频率限制在1小时内发送次数限制，单位次
func CheckLoginIPRateLimit(ip string) (bool, error) {
	key := fmt.Sprintf("%s%s", LoginIPRateLimitPrefix, ip)

	current, err := config.RDB.Incr(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	if current == 1 {
		config.RDB.Expire(context.Background(), key, 1*time.Minute)
	}

	return current > int64(config.AppConfig.Security.IPLoginMinuteLimit), nil
}

// CheckDeviceBlacklist 检查设备是否在黑名单中
// 【登录安全规则4】检查该设备是否在黑名单中，是则拦截
func CheckDeviceBlacklist(deviceID string) bool {
	if deviceID == "" {
		return false
	}
	key := fmt.Sprintf("%s%s", DeviceBlacklistKeyPrefix, deviceID)
	exists, _ := config.RDB.Exists(context.Background(), key).Result()
	return exists > 0
}

// CheckLoginFailedAttempt 检查登录失败次数（5分钟内5次失败锁定5分钟）
// 【登录安全规则5】登录失败次数限制，防止暴力破解
func CheckLoginFailedAttempt(phoneNum string) (bool, error) {
	key := fmt.Sprintf("%s%s", LoginFailedAttemptPrefix, phoneNum)
	
	current, err := config.RDB.Incr(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	
	if current == 1 {
		config.RDB.Expire(context.Background(), key, 5*time.Minute)
	}
	
	if current > 5 {
		config.RDB.Expire(context.Background(), key, time.Duration(config.AppConfig.Security.LoginFailureLockMinutes)*time.Minute)
		return true, nil
	}
	
	return false, nil
}

// ResetLoginFailedAttempt 重置登录失败次数（登录成功后调用）
func ResetLoginFailedAttempt(phoneNum string) error {
	key := fmt.Sprintf("%s%s", LoginFailedAttemptPrefix, phoneNum)
	return config.RDB.Del(context.Background(), key).Err()
}

// DeleteSMSVerificationCode 删除短信验证码（注册/登录成功后调用）
// 【密码存储加密】重置成功后清空所有未使用验证码，防止二次复用
func DeleteSMSVerificationCode(phoneNum, tag string) error {
	key := getVerificationCodeKey(phoneNum, VerificationCodeTypeSMS, tag)
	return config.RDB.Del(context.Background(), key).Err()
}

// ==================== 重置密码相关 ====================

// ResetPasswordRateLimitPrefix 重置密码频率限制key前缀
const ResetPasswordRateLimitPrefix = "reset_password_rate_limit:"

// LoginFailedAttemptPrefix 登录失败次数限制key前缀
const LoginFailedAttemptPrefix = "login_failed_attempt:"

// HashToken 对Token进行sha256哈希
// 【重置凭证】禁止明文存库，数据库只存Token哈希
func HashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

// CreateResetToken 创建重置密码Token（存储哈希值）
// 【重置凭证】单次有效、短有效期、绑定userID+设备标识、禁止明文存库
func CreateResetToken(token, deviceID string, userID uint, expireMinutes int) error {
	tokenHash := HashToken(token)
	
	// 删除该用户之前未使用的重置Token，防止多Token攻击
	config.DB.Where("user_id = ? AND used = 0", userID).Delete(&model.ResetToken{})
	
	resetToken := &model.ResetToken{
		TokenHash: tokenHash,
		UserID:    userID,
		DeviceID:  deviceID,
		ExpireAt:  time.Now().Add(time.Duration(expireMinutes) * time.Minute),
		Used:      0,
	}
	return config.DB.Create(resetToken).Error
}

// VerifyResetToken 验证重置Token是否有效（不标记为已使用）
// 【重置凭证】仅验证有效性，不改变状态
func VerifyResetToken(token string) (*model.ResetToken, error) {
	tokenHash := HashToken(token)
	
	var resetToken model.ResetToken
	err := config.DB.Where("token_hash = ? AND used = 0 AND expire_at > ?", tokenHash, time.Now()).First(&resetToken).Error
	if err != nil {
		return nil, err
	}
	
	return &resetToken, nil
}

// DeleteResetToken 删除指定Token
func DeleteResetToken(token string) error {
	tokenHash := HashToken(token)
	return config.DB.Where("token_hash = ?", tokenHash).Delete(&model.ResetToken{}).Error
}

// DeleteAllResetTokensByUser 删除用户所有重置Token
// 【密码存储加密】重置成功后清空所有未使用重置Token，防止二次复用
func DeleteAllResetTokensByUser(userID uint) error {
	return config.DB.Where("user_id = ?", userID).Delete(&model.ResetToken{}).Error
}

// CheckResetPasswordRateLimit 检查同一账号24h内重置次数是否超过限制（3次）
// 【重置流程行为风控】同一账号24h最多允许3次密码重置，超限锁定重置通道24h
func CheckResetPasswordRateLimit(userID uint) (bool, error) {
	key := fmt.Sprintf("%s%d", ResetPasswordRateLimitPrefix, userID)
	
	current, err := config.RDB.Incr(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	
	if current == 1 {
		config.RDB.Expire(context.Background(), key, 24*time.Hour)
	}
	
	return current > 3, nil
}

// LogOperation 记录敏感操作日志（不可删除）
// 【重置流程行为风控】记录用户ID、操作时间、IP、UA、操作类型、是否成功
func LogOperation(userID uint, ip, ua, operation string, success bool, detail string) error {
	successInt := 0
	if success {
		successInt = 1
	}
	
	operationLog := &model.OperationLog{
		UserID:    userID,
		IP:        ip,
		UA:        ua,
		Operation: operation,
		Success:   successInt,
		Detail:    detail,
	}
	
	return config.DB.Create(operationLog).Error
}

// GetOperationLogsByUserID 根据用户ID查询操作日志
func GetOperationLogsByUserID(userID uint) ([]model.OperationLog, error) {
	var logs []model.OperationLog
	err := config.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetOperationLogsByOperation 根据操作类型查询操作日志
func GetOperationLogsByOperation(operation string) ([]model.OperationLog, error) {
	var logs []model.OperationLog
	err := config.DB.Where("operation = ?", operation).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// SavePasswordHistory 保存密码历史记录
// 【最低安全策略】记录用户历史密码，禁止和历史5次旧密码重复
func SavePasswordHistory(userID uint, passwordHash string) error {
	passwordHistory := &model.PasswordHistory{
		UserID:       userID,
		PasswordHash: passwordHash,
	}
	
	if err := config.DB.Create(passwordHistory).Error; err != nil {
		return err
	}
	
	// 只保留最近5次密码记录
	var count int
	config.DB.Model(&model.PasswordHistory{}).Where("user_id = ?", userID).Count(&count)
	if count > 5 {
		var oldestID uint
		config.DB.Model(&model.PasswordHistory{}).Where("user_id = ?", userID).
			Order("created_at ASC").Limit(1).Pluck("id", &oldestID)
		config.DB.Where("user_id = ? AND id < ?", userID, oldestID).Delete(&model.PasswordHistory{})
	}
	
	return nil
}

// CheckPasswordHistory 检查新密码是否与历史密码重复
// 【最低安全策略】禁止和历史5次旧密码重复
func CheckPasswordHistory(userID uint, password string) (bool, error) {
	var histories []model.PasswordHistory
	err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(5).Find(&histories).Error
	if err != nil {
		return false, err
	}
	
	for _, history := range histories {
		if err := bcrypt.CompareHashAndPassword([]byte(history.PasswordHash), []byte(password)); err == nil {
			return true, nil
		}
	}
	
	return false, nil
}