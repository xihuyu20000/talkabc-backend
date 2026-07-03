package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/pkg/utils"
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// GenerateSMSCodeRequest 发送短信验证码请求参数
type GenerateSMSCodeRequest struct {
	PhoneNum    string
	CaptchaID   string
	CaptchaCode string
	// 【功能10】业务标签，用于区分不同业务场景（如register、login）
	Tag string
}

// LoginRequest 登录请求参数（包含安全校验所需信息）
type LoginRequest struct {
	PhoneNum string
	Code     string
	Password string
	IP       string
	DeviceID string
}

// GenerateSMSCode 生成并发送短信验证码（带完整的频率限制和安全校验）
// 参数说明：
//   - req: 发送验证码请求，包含手机号、图形验证码ID、图形验证码、业务标签
//
// 返回值：
//   - error: 错误信息
//
// 验证流程：
//   1. 检查手机号格式
//   2. 检查60秒冷却期
//   3. 检查1小时发送次数限制（10次）
//   4. 判断是否需要验证图形验证码（每日首次不需要）
//   5. 如果需要，验证图形验证码
//   6. 生成6位数字验证码
//   7. 设置冷却期
//   8. 存储验证码到Redis（带过期时间）
func GenerateSMSCode(req GenerateSMSCodeRequest) error {
	if req.PhoneNum == "" {
		return fmt.Errorf("手机号不能为空")
	}

	// 【功能10】默认业务标签为default
	if req.Tag == "" {
		req.Tag = "default"
	}

	// 【功能3】检查60秒冷却期（服务器端校验）
	if repository.CheckSMSCooldown(req.PhoneNum) {
		return fmt.Errorf("发送过于频繁，请稍后再试")
	}

	// 【功能8】检查1小时内发送次数限制（10次）
	hourlyLimitExceeded, err := repository.CheckHourlyLimit(req.PhoneNum)
	if err != nil {
		return fmt.Errorf("验证失败")
	}
	if hourlyLimitExceeded {
		return fmt.Errorf("发送次数过多，请1小时后再试")
	}

	// 【功能9】判断是否为每日首次发送（首次不需要图形验证码）
	dailyFirst, err := repository.CheckDailyFirst(req.PhoneNum)
	if err != nil {
		return fmt.Errorf("验证失败")
	}

	// 【功能7】非每日首次发送时，必须先验证图形验证码
	if !dailyFirst {
		if req.CaptchaID == "" || req.CaptchaCode == "" {
			return fmt.Errorf("请先获取并验证图形验证码")
		}

		err = verifyCaptcha(req.CaptchaID, req.CaptchaCode)
		if err != nil {
			return err
		}
	}

	// 【功能2】生成6位纯数字验证码
	code := generateRandomCode(6)

	// 【功能3】设置60秒冷却期
	err = repository.SetSMSCooldown(req.PhoneNum)
	if err != nil {
		return fmt.Errorf("发送失败")
	}

	// 【功能1】存储验证码到Redis（带TTL，默认5分钟）
	// 【功能4】新验证码会覆盖旧的，保证同一时间只有一个有效验证码
	// 【功能10】使用业务标签隔离不同场景的验证码
	err = repository.CreateVerificationCode(req.PhoneNum, code, repository.VerificationCodeTypeSMS, req.Tag)
	if err != nil {
		return fmt.Errorf("发送失败")
	}

	return nil
}

// verifyCaptcha 验证图形验证码
func verifyCaptcha(captchaID, code string) error {
	storedCode, err := repository.GetCaptchaCode(captchaID)
	if err != nil {
		return fmt.Errorf("图形验证码已过期或不存在")
	}

	if storedCode != code {
		return fmt.Errorf("图形验证码不正确")
	}

	// 【功能5】验证通过后立即删除图形验证码，防止重复使用
	err = repository.DeleteCaptchaCode(captchaID)
	if err != nil {
		return fmt.Errorf("验证失败")
	}

	return nil
}

// VerifySMSCode 验证短信验证码
// 参数说明：
//   - phoneNum: 用户手机号
//   - code: 用户输入的验证码
//   - tag: 业务标签（与发送时一致）
//
// 返回值：
//   - error: 错误信息
//
// 验证逻辑：
//   1. 使用Lua脚本原子验证并删除验证码
//   2. 无论验证是否成功，验证码都会被删除（防止暴力攻击）
func VerifySMSCode(phoneNum, code, tag string) error {
	// 【功能10】默认业务标签为default
	if tag == "" {
		tag = "default"
	}

	// 【功能5】使用Lua脚本实现验证码的原子验证和删除，保证一次性使用
	// 无论验证成功与否，验证码都会被删除，防止暴力攻击
	ok, err := repository.VerifyAndDeleteVerificationCode(phoneNum, code, repository.VerificationCodeTypeSMS, tag)
	if err != nil {
		return fmt.Errorf("验证码无效或已过期")
	}

	if !ok {
		return fmt.Errorf("验证码不正确")
	}

	return nil
}

// GenerateAlnumCode 生成4位字母数字混合验证码
func GenerateAlnumCode(phoneNum string) error {
	code := generateRandomAlnum(4)
	return repository.CreateVerificationCode(phoneNum, code, repository.VerificationCodeTypeAlnum, "default")
}

// generateRandomCode 生成指定长度的数字验证码
// 【功能2】生成6位纯数字验证码
func generateRandomCode(length int) string {
	var result string
	for i := 0; i < length; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		result += fmt.Sprintf("%d", num)
	}
	return result
}

// generateRandomAlnum 生成指定长度的字母数字混合验证码
func generateRandomAlnum(length int) string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var result string
	for i := 0; i < length; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result += string(chars[num.Int64()])
	}
	return result
}

// RegisterRequest 用户注册请求参数（包含IP和设备ID）
type RegisterRequest struct {
	PhoneNum string
	Code     string
	IP       string
	DeviceID string
}

// Register 用户注册函数（带安全校验）
// 注册安全规则：
//   1. 【注册安全规则1】服务器检验通过IP是否位于黑名单中，是则拦截
//   2. 【注册安全规则2】服务器检验通过IP检查注册请求的频率，请求频率限制在1分钟10次
//   3. 【注册安全规则3】校验手机号是否位于黑名单中，是则拦截
//   4. 【注册安全规则4】校验手机号是否已被占用（用户已存在）
//   5. 【注册安全规则5】检查该设备是否在黑名单中，是则拦截
func Register(req RegisterRequest) (string, error) {
	if req.PhoneNum == "" || req.Code == "" {
		return "", fmt.Errorf("手机号和验证码不能为空")
	}

	// 【注册安全规则1】检查IP是否在黑名单中
	if req.IP != "" && repository.CheckIPBlacklist(req.IP) {
		return "", fmt.Errorf("当前IP已被限制注册")
	}

	// 【注册安全规则2】检查IP注册请求频率（1分钟10次）
	if req.IP != "" {
		rateLimitExceeded, err := repository.CheckRegisterIPRateLimit(req.IP)
		if err != nil {
			return "", fmt.Errorf("验证失败")
		}
		if rateLimitExceeded {
			return "", fmt.Errorf("注册过于频繁，请稍后再试")
		}
	}

	// 【注册安全规则3】检查手机号是否在黑名单中
	if repository.CheckPhoneBlacklist(req.PhoneNum) {
		return "", fmt.Errorf("当前手机号已被限制注册")
	}

	// 【注册安全规则5】检查设备是否在黑名单中
	if repository.CheckDeviceBlacklist(req.DeviceID) {
		return "", fmt.Errorf("当前设备已被限制注册")
	}

	// 【功能10】使用register业务标签验证短信验证码
	err := VerifySMSCode(req.PhoneNum, req.Code, "register")
	if err != nil {
		return "", err
	}

	// 【注册安全规则4】校验手机号是否已被占用（用户已存在）
	existingUser, _ := repository.GetUserByPhone(req.PhoneNum)
	if existingUser.ID != 0 {
		return "", fmt.Errorf("该手机号已注册")
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(""), bcrypt.DefaultCost)

	user := &model.User{
		Uid:           utils.GenerateUID(),
		PhoneNum:      req.PhoneNum,
		Password:      string(password),
		Gender:        -1,
		AccountStatus: 2,
	}

	err = repository.CreateUser(user)
	if err != nil {
		return "", fmt.Errorf("注册失败")
	}

	return middleware.GenerateToken(user.Uid)
}

// LoginByCode 验证码登录（带安全校验）
// 登录安全规则：
//   1. 【登录安全规则1】服务器检验通过IP是否位于黑名单中，是则拦截
//   2. 【登录安全规则2】服务器检验通过IP检查登录请求的频率，请求频率限制在1分钟10次
//   3. 【登录安全规则3】校验手机号是否位于黑名单中，是则拦截
//   4. 【登录安全规则4】检查该设备是否在黑名单中，是则拦截
func LoginByCode(req LoginRequest) (string, error) {
	if req.PhoneNum == "" || req.Code == "" {
		return "", fmt.Errorf("手机号和验证码不能为空")
	}

	// 【登录安全规则1】检查IP是否在黑名单中
	if req.IP != "" && repository.CheckIPBlacklist(req.IP) {
		return "", fmt.Errorf("当前IP已被限制登录")
	}

	// 【登录安全规则2】检查IP登录请求频率（1分钟10次）
	if req.IP != "" {
		rateLimitExceeded, err := repository.CheckLoginIPRateLimit(req.IP)
		if err != nil {
			return "", fmt.Errorf("验证失败")
		}
		if rateLimitExceeded {
			return "", fmt.Errorf("登录过于频繁，请稍后再试")
		}
	}

	// 【登录安全规则3】检查手机号是否在黑名单中
	if repository.CheckPhoneBlacklist(req.PhoneNum) {
		return "", fmt.Errorf("当前手机号已被限制登录")
	}

	// 【登录安全规则4】检查设备是否在黑名单中
	if repository.CheckDeviceBlacklist(req.DeviceID) {
		return "", fmt.Errorf("当前设备已被限制登录")
	}

	// 【功能10】使用login业务标签验证短信验证码
	err := VerifySMSCode(req.PhoneNum, req.Code, "login")
	if err != nil {
		return "", err
	}

	user, err := repository.GetUserByPhone(req.PhoneNum)
	if err != nil {
		return "", fmt.Errorf("用户不存在")
	}

	return middleware.GenerateToken(user.Uid)
}

// LoginByPassword 密码登录（带安全校验）
// 登录安全规则：
//   1. 【登录安全规则1】服务器检验通过IP是否位于黑名单中，是则拦截
//   2. 【登录安全规则2】服务器检验通过IP检查登录请求的频率，请求频率限制在1分钟10次
//   3. 【登录安全规则3】校验手机号是否位于黑名单中，是则拦截
//   4. 【登录安全规则4】检查该设备是否在黑名单中，是则拦截
func LoginByPassword(req LoginRequest) (string, error) {
	if req.PhoneNum == "" || req.Password == "" {
		return "", fmt.Errorf("手机号和密码不能为空")
	}

	// 【登录安全规则1】检查IP是否在黑名单中
	if req.IP != "" && repository.CheckIPBlacklist(req.IP) {
		return "", fmt.Errorf("当前IP已被限制登录")
	}

	// 【登录安全规则2】检查IP登录请求频率（1分钟10次）
	if req.IP != "" {
		rateLimitExceeded, err := repository.CheckLoginIPRateLimit(req.IP)
		if err != nil {
			return "", fmt.Errorf("验证失败")
		}
		if rateLimitExceeded {
			return "", fmt.Errorf("登录过于频繁，请稍后再试")
		}
	}

	// 【登录安全规则3】检查手机号是否在黑名单中
	if repository.CheckPhoneBlacklist(req.PhoneNum) {
		return "", fmt.Errorf("当前手机号已被限制登录")
	}

	// 【登录安全规则4】检查设备是否在黑名单中
	if repository.CheckDeviceBlacklist(req.DeviceID) {
		return "", fmt.Errorf("当前设备已被限制登录")
	}

	user, err := repository.GetUserByPhone(req.PhoneNum)
	if err != nil {
		return "", fmt.Errorf("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", fmt.Errorf("密码错误")
	}

	return middleware.GenerateToken(user.Uid)
}

// ResetPassword 重置密码
func ResetPassword(phoneNum, pwd1, pwd2 string) error {
	if pwd1 != pwd2 {
		return fmt.Errorf("两次密码不一致")
	}

	if !validatePassword(pwd1) {
		return fmt.Errorf("密码格式不正确")
	}

	user, err := repository.GetUserByPhone(phoneNum)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(pwd1), bcrypt.DefaultCost)
	user.Password = string(password)
	return repository.UpdateUser(user)
}

// validatePassword 验证密码格式
func validatePassword(password string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]{6,20}$`, password)
	return matched
}