package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/sms"
	"backend/pkg/utils"
	"context"
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
	PhoneNum string // 手机号
	Code     string // 验证码
	Password string // 密码
	IP       string // 客户端IP
	DeviceID string // 设备ID
	UA       string // 用户代理
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

	// 通过短信网关发送实际短信
	gateway := sms.GetGateway()
	if gateway != nil {
		err = gateway.SendVerificationCode(context.Background(), req.PhoneNum, code)
		if err != nil {
			return fmt.Errorf("发送失败: %v", err)
		}
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
	PhoneNum string // 手机号
	Code     string // 验证码
	Password string // 密码
	IP       string // 客户端IP
	DeviceID string // 设备ID
	UA       string // 用户代理
}

// Register 用户注册函数（带安全校验）
// 注册安全规则：
//   1. 【注册安全规则1】服务器检验通过IP是否位于黑名单中，是则拦截
//   2. 【注册安全规则2】服务器检验通过IP检查注册请求的频率，请求频率限制在1分钟10次
//   3. 【注册安全规则3】校验手机号是否位于黑名单中，是则拦截
//   4. 【注册安全规则4】校验手机号是否已被占用（用户已存在）
//   5. 【注册安全规则5】检查该设备是否在黑名单中，是则拦截
//   6. 【注册安全规则6】密码复杂度校验（最低安全策略）
//   7. 【注册安全规则7】注册成功后清理验证码，防止二次复用
//   8. 【注册安全规则8】记录注册操作日志（不可删除）
func Register(req RegisterRequest) (string, error) {
	if req.PhoneNum == "" || req.Code == "" || req.Password == "" {
		return "", fmt.Errorf("手机号、验证码和密码不能为空")
	}

	// 【注册安全规则1】检查IP是否在黑名单中
	if req.IP != "" && repository.CheckIPBlacklist(req.IP) {
		repository.LogOperation(0, req.IP, req.UA, "register", false, "IP在黑名单中")
		return "", fmt.Errorf("当前IP已被限制注册")
	}

	// 【注册安全规则2】检查IP注册请求频率（1分钟10次）
	if req.IP != "" {
		rateLimitExceeded, err := repository.CheckRegisterIPRateLimit(req.IP)
		if err != nil {
			return "", fmt.Errorf("验证失败")
		}
		if rateLimitExceeded {
			repository.LogOperation(0, req.IP, req.UA, "register", false, "注册频率超限")
			return "", fmt.Errorf("注册过于频繁，请稍后再试")
		}
	}

	// 【注册安全规则3】检查手机号是否在黑名单中
	if repository.CheckPhoneBlacklist(req.PhoneNum) {
		repository.LogOperation(0, req.IP, req.UA, "register", false, "手机号在黑名单中")
		return "", fmt.Errorf("当前手机号已被限制注册")
	}

	// 【注册安全规则5】检查设备是否在黑名单中
	if repository.CheckDeviceBlacklist(req.DeviceID) {
		repository.LogOperation(0, req.IP, req.UA, "register", false, "设备在黑名单中")
		return "", fmt.Errorf("当前设备已被限制注册")
	}

	// 【功能10】使用register业务标签验证短信验证码
	err := VerifySMSCode(req.PhoneNum, req.Code, "register")
	if err != nil {
		repository.LogOperation(0, req.IP, req.UA, "register", false, "验证码验证失败")
		return "", err
	}

	// 【注册安全规则4】校验手机号是否已被占用（用户已存在）
	existingUser, _ := repository.GetUserByPhone(req.PhoneNum)
	if existingUser.ID != 0 {
		repository.LogOperation(0, req.IP, req.UA, "register", false, "手机号已注册")
		return "", fmt.Errorf("该手机号已注册")
	}

	// 【注册安全规则6】密码复杂度校验（最低安全策略）
	emptyUser := &model.User{PhoneNum: req.PhoneNum}
	if err := validatePasswordComplexity(req.Password, emptyUser); err != nil {
		repository.LogOperation(0, req.IP, req.UA, "register", false, "密码复杂度校验失败")
		return "", err
	}

	// 【密码存储加密】使用bcrypt加密（cost=10，自动内置盐）
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("注册失败")
	}

	user := &model.User{
		Uid:           utils.GenerateUID(),
		PhoneNum:      req.PhoneNum,
		Password:      string(passwordHash),
		Gender:        -1,
		AccountStatus: 1,
	}

	err = repository.CreateUser(user)
	if err != nil {
		repository.LogOperation(0, req.IP, req.UA, "register", false, "用户创建失败")
		return "", fmt.Errorf("注册失败")
	}

	// 【注册安全规则7】注册成功后清理验证码，防止二次复用
	repository.DeleteSMSVerificationCode(req.PhoneNum, "register")

	// 【注册安全规则8】记录注册操作日志（不可删除）
	repository.LogOperation(user.ID, req.IP, req.UA, "register", true, "注册成功")

	return middleware.GenerateToken(user.Uid)
}

// LoginByCode 验证码登录（带安全校验）
// 登录安全规则：
//   1. 【登录安全规则1】服务器检验通过IP是否位于黑名单中，是则拦截
//   2. 【登录安全规则2】服务器检验通过IP检查登录请求的频率，请求频率限制在1分钟10次
//   3. 【登录安全规则3】校验手机号是否位于黑名单中，是则拦截
//   4. 【登录安全规则4】检查该设备是否在黑名单中，是则拦截
//   5. 【登录安全规则5】检查登录失败次数（5分钟内5次失败锁定15分钟）
//   6. 【登录安全规则6】检查用户账号状态（正常/封禁/注销）
//   7. 【登录安全规则7】登录成功后清理验证码，防止二次复用
//   8. 【登录安全规则8】记录登录操作日志（不可删除）
func LoginByCode(req LoginRequest) (string, error) {
	if req.PhoneNum == "" || req.Code == "" {
		return "", fmt.Errorf("手机号和验证码不能为空")
	}

	// 【登录安全规则1】检查IP是否在黑名单中
	if req.IP != "" && repository.CheckIPBlacklist(req.IP) {
		repository.LogOperation(0, req.IP, req.UA, "login_code", false, "IP在黑名单中")
		return "", fmt.Errorf("当前IP已被限制登录")
	}

	// 【登录安全规则2】检查IP登录请求频率（1分钟10次）
	if req.IP != "" {
		rateLimitExceeded, err := repository.CheckLoginIPRateLimit(req.IP)
		if err != nil {
			return "", fmt.Errorf("验证失败")
		}
		if rateLimitExceeded {
			repository.LogOperation(0, req.IP, req.UA, "login_code", false, "登录频率超限")
			return "", fmt.Errorf("登录过于频繁，请稍后再试")
		}
	}

	// 【登录安全规则3】检查手机号是否在黑名单中
	if repository.CheckPhoneBlacklist(req.PhoneNum) {
		repository.LogOperation(0, req.IP, req.UA, "login_code", false, "手机号在黑名单中")
		return "", fmt.Errorf("当前手机号已被限制登录")
	}

	// 【登录安全规则4】检查设备是否在黑名单中
	if repository.CheckDeviceBlacklist(req.DeviceID) {
		repository.LogOperation(0, req.IP, req.UA, "login_code", false, "设备在黑名单中")
		return "", fmt.Errorf("当前设备已被限制登录")
	}

	// 【登录安全规则5】检查登录失败次数（5分钟内5次失败锁定15分钟）
	failedLocked, err := repository.CheckLoginFailedAttempt(req.PhoneNum)
	if err != nil {
		return "", fmt.Errorf("验证失败")
	}
	if failedLocked {
		repository.LogOperation(0, req.IP, req.UA, "login_code", false, "登录失败次数超限")
		return "", fmt.Errorf("登录失败次数过多，请15分钟后再试")
	}

	// 【功能10】使用login业务标签验证短信验证码
	err = VerifySMSCode(req.PhoneNum, req.Code, "login")
	if err != nil {
		repository.LogOperation(0, req.IP, req.UA, "login_code", false, "验证码验证失败")
		return "", err
	}

	user, err := repository.GetUserByPhone(req.PhoneNum)
	if err != nil {
		repository.LogOperation(0, req.IP, req.UA, "login_code", false, "用户不存在")
		return "", fmt.Errorf("用户不存在")
	}

	// 【登录安全规则6】检查用户账号状态（正常/封禁/注销）
	if user.AccountStatus == 0 {
		repository.LogOperation(user.ID, req.IP, req.UA, "login_code", false, "账号被封禁")
		return "", fmt.Errorf("账号已被封禁")
	}
	if user.AccountStatus == 2 {
		repository.LogOperation(user.ID, req.IP, req.UA, "login_code", false, "账号已注销")
		return "", fmt.Errorf("账号已注销")
	}

	// 【登录安全规则7】登录成功后清理验证码，防止二次复用
	repository.DeleteSMSVerificationCode(req.PhoneNum, "login")

	// 【登录安全规则8】登录成功后重置失败次数
	repository.ResetLoginFailedAttempt(req.PhoneNum)

	// 【登录安全规则8】记录登录操作日志（不可删除）
	repository.LogOperation(user.ID, req.IP, req.UA, "login_code", true, "验证码登录成功")

	return middleware.GenerateToken(user.Uid)
}

// LoginByPassword 密码登录（带安全校验）
// 登录安全规则：
//   1. 【登录安全规则1】服务器检验通过IP是否位于黑名单中，是则拦截
//   2. 【登录安全规则2】服务器检验通过IP检查登录请求的频率，请求频率限制在1分钟10次
//   3. 【登录安全规则3】校验手机号是否位于黑名单中，是则拦截
//   4. 【登录安全规则4】检查该设备是否在黑名单中，是则拦截
//   5. 【登录安全规则5】检查登录失败次数（5分钟内5次失败锁定15分钟）
//   6. 【登录安全规则6】检查用户账号状态（正常/封禁/注销）
//   7. 【登录安全规则7】记录登录操作日志（不可删除）
func LoginByPassword(req LoginRequest) (string, error) {
	if req.PhoneNum == "" || req.Password == "" {
		return "", fmt.Errorf("手机号和密码不能为空")
	}

	// 【登录安全规则1】检查IP是否在黑名单中
	if req.IP != "" && repository.CheckIPBlacklist(req.IP) {
		repository.LogOperation(0, req.IP, req.UA, "login_password", false, "IP在黑名单中")
		return "", fmt.Errorf("当前IP已被限制登录")
	}

	// 【登录安全规则2】检查IP登录请求频率（1分钟10次）
	if req.IP != "" {
		rateLimitExceeded, err := repository.CheckLoginIPRateLimit(req.IP)
		if err != nil {
			return "", fmt.Errorf("验证失败")
		}
		if rateLimitExceeded {
			repository.LogOperation(0, req.IP, req.UA, "login_password", false, "登录频率超限")
			return "", fmt.Errorf("登录过于频繁，请稍后再试")
		}
	}

	// 【登录安全规则3】检查手机号是否在黑名单中
	if repository.CheckPhoneBlacklist(req.PhoneNum) {
		repository.LogOperation(0, req.IP, req.UA, "login_password", false, "手机号在黑名单中")
		return "", fmt.Errorf("当前手机号已被限制登录")
	}

	// 【登录安全规则4】检查设备是否在黑名单中
	if repository.CheckDeviceBlacklist(req.DeviceID) {
		repository.LogOperation(0, req.IP, req.UA, "login_password", false, "设备在黑名单中")
		return "", fmt.Errorf("当前设备已被限制登录")
	}

	// 【登录安全规则5】检查登录失败次数（5分钟内5次失败锁定15分钟）
	failedLocked, err := repository.CheckLoginFailedAttempt(req.PhoneNum)
	if err != nil {
		return "", fmt.Errorf("验证失败")
	}
	if failedLocked {
		repository.LogOperation(0, req.IP, req.UA, "login_password", false, "登录失败次数超限")
		return "", fmt.Errorf("登录失败次数过多，请15分钟后再试")
	}

	user, err := repository.GetUserByPhone(req.PhoneNum)
	if err != nil {
		repository.LogOperation(0, req.IP, req.UA, "login_password", false, "用户不存在")
		return "", fmt.Errorf("用户不存在")
	}

	// 【登录安全规则6】检查用户账号状态（正常/封禁/注销）
	if user.AccountStatus == 0 {
		repository.LogOperation(user.ID, req.IP, req.UA, "login_password", false, "账号被封禁")
		return "", fmt.Errorf("账号已被封禁")
	}
	if user.AccountStatus == 2 {
		repository.LogOperation(user.ID, req.IP, req.UA, "login_password", false, "账号已注销")
		return "", fmt.Errorf("账号已注销")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		repository.LogOperation(user.ID, req.IP, req.UA, "login_password", false, "密码错误")
		return "", fmt.Errorf("密码错误")
	}

	// 【登录安全规则7】登录成功后重置失败次数
	repository.ResetLoginFailedAttempt(req.PhoneNum)

	// 【登录安全规则7】记录登录操作日志（不可删除）
	repository.LogOperation(user.ID, req.IP, req.UA, "login_password", true, "密码登录成功")

	return middleware.GenerateToken(user.Uid)
}

// ResetPasswordRequest 发起密码重置请求参数
type ResetPasswordRequest struct {
	PhoneNum string // 手机号
	DeviceID string // 设备标识，绑定唯一信息
	IP       string // 操作IP
	UA       string // 设备UA
}

// CompleteResetPasswordRequest 完成密码重置请求参数
type CompleteResetPasswordRequest struct {
	Token    string // 重置Token
	Pwd1     string // 新密码
	Pwd2     string // 确认密码
	DeviceID string // 设备标识，验证绑定信息
	IP       string // 操作IP
	UA       string // 设备UA
}

// InitiateResetPassword 发起密码重置（生成重置Token）
// 【重置凭证】生成随机Token，绑定userID+设备标识，短有效期（5分钟）
func InitiateResetPassword(req ResetPasswordRequest) error {
	if req.PhoneNum == "" {
		return fmt.Errorf("手机号不能为空")
	}

	// 查询用户
	user, err := repository.GetUserByPhone(req.PhoneNum)
	if err != nil {
		// 【重置流程行为风控】记录敏感操作日志（用户不存在）
		repository.LogOperation(0, req.IP, req.UA, "initiate_reset", false, "用户不存在: "+req.PhoneNum)
		return fmt.Errorf("用户不存在")
	}

	// 【重置流程行为风控】检查同一账号24h内重置次数（最多3次）
	rateLimitExceeded, err := repository.CheckResetPasswordRateLimit(user.ID)
	if err != nil {
		return fmt.Errorf("验证失败")
	}
	if rateLimitExceeded {
		// 【重置流程行为风控】记录敏感操作日志（超限）
		repository.LogOperation(user.ID, req.IP, req.UA, "initiate_reset", false, "24小时内重置次数超限")
		return fmt.Errorf("重置次数过多，请24小时后再试")
	}

	// 【重置凭证】使用crypto/rand生成不可预测的随机Token
	token, err := generateResetToken()
	if err != nil {
		return fmt.Errorf("生成重置链接失败")
	}

	// 【重置凭证】创建重置Token（存储哈希值，绑定userID+设备标识，5分钟有效期）
	err = repository.CreateResetToken(token, req.DeviceID, user.ID, 5)
	if err != nil {
		return fmt.Errorf("生成重置链接失败")
	}

	// 【重置流程行为风控】记录敏感操作日志（发起重置成功）
	repository.LogOperation(user.ID, req.IP, req.UA, "initiate_reset", true, "发起密码重置")

	// TODO: 发送重置链接到用户手机/邮箱（实际项目中调用短信/邮件服务）
	// 这里仅返回token供测试使用（生产环境不应返回原始token）
	return nil
}

// generateResetToken 生成重置密码Token
// 【重置凭证】使用crypto/rand生成随机字节，不可预测
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 32)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		result[i] = letters[num.Int64()]
	}
	
	return string(result), nil
}

// ValidateResetToken 验证重置Token是否有效
// 【重置凭证】检查Token是否存在、未使用、未过期
func ValidateResetToken(token string) (*model.ResetToken, error) {
	if token == "" {
		return nil, fmt.Errorf("重置链接无效")
	}

	resetToken, err := repository.VerifyResetToken(token)
	if err != nil {
		return nil, fmt.Errorf("重置链接无效或已过期")
	}

	return resetToken, nil
}

// CompleteResetPassword 完成密码重置
// 【密码存储加密】使用bcrypt加密，重置成功后清空登录态和未使用Token
// 【最低安全策略】密码复杂度校验、禁止重复密码
func CompleteResetPassword(req CompleteResetPasswordRequest) error {
	if req.Token == "" || req.Pwd1 == "" || req.Pwd2 == "" {
		return fmt.Errorf("参数不能为空")
	}

	if req.Pwd1 != req.Pwd2 {
		return fmt.Errorf("两次密码不一致")
	}

	// 【重置凭证】验证重置Token（验证后自动标记为已使用）
	resetToken, err := repository.VerifyResetToken(req.Token)
	if err != nil {
		// 【重置流程行为风控】记录敏感操作日志（Token验证失败）
		repository.LogOperation(0, req.IP, req.UA, "complete_reset", false, "重置Token无效")
		return fmt.Errorf("重置链接无效或已过期")
	}

	// 【重置凭证】验证设备标识（防止跨账号盗用）
	if req.DeviceID != "" && resetToken.DeviceID != "" && req.DeviceID != resetToken.DeviceID {
		repository.LogOperation(resetToken.UserID, req.IP, req.UA, "complete_reset", false, "设备标识不匹配")
		return fmt.Errorf("设备验证失败")
	}

	// 查询用户
	user, err := repository.GetUserByID(resetToken.UserID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 【最低安全策略】密码复杂度校验
	if err := validatePasswordComplexity(req.Pwd1, user); err != nil {
		repository.LogOperation(user.ID, req.IP, req.UA, "complete_reset", false, "密码复杂度校验失败")
		return err
	}

	// 【最低安全策略】检查新密码是否与历史密码重复（最近5次）
	historyRepeat, err := repository.CheckPasswordHistory(user.ID, req.Pwd1)
	if err != nil {
		return fmt.Errorf("验证失败")
	}
	if historyRepeat {
		repository.LogOperation(user.ID, req.IP, req.UA, "complete_reset", false, "新密码与历史密码重复")
		return fmt.Errorf("新密码不能与最近5次使用过的密码相同")
	}

	// 【密码存储加密】使用bcrypt加密（cost=10，自动内置盐）
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Pwd1), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("重置失败")
	}

	// 【密码存储加密】保存旧密码到历史记录
	if user.Password != "" {
		repository.SavePasswordHistory(user.ID, user.Password)
	}

	// 更新密码
	user.Password = string(passwordHash)
	if err := repository.UpdateUser(user); err != nil {
		return fmt.Errorf("重置失败")
	}

	// 【密码存储加密】重置成功后清空该用户全部登录态（Redis token等）
	repository.ClearUserLoginState(user.ID)

	// 【密码存储加密】清空所有未使用重置Token，防止二次复用
	repository.DeleteAllResetTokensByUser(user.ID)

	// 【重置流程行为风控】记录敏感操作日志（完成重置成功）
	repository.LogOperation(user.ID, req.IP, req.UA, "complete_reset", true, "密码重置成功")

	return nil
}

// validatePasswordComplexity 验证密码复杂度（最低安全策略）
// 【最低安全策略】
// 1. 长度：≥8位，推荐12位以上
// 2. 必须包含三类中至少两种：大写字母、小写字母、数字、特殊符号
// 3. 禁止弱密码黑名单
// 4. 禁止包含用户名、手机号、邮箱前缀
// 5. 密码前后空格自动修剪，空白密码直接拦截
func validatePasswordComplexity(password string, user *model.User) error {
	// 【最低安全策略6】密码前后空格自动修剪，空白密码直接拦截
	password = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(password, "")
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}

	// 【最低安全策略1】长度：≥8位
	if len(password) < 8 {
		return fmt.Errorf("密码长度不能少于8位")
	}

	// 【最低安全策略3】禁止弱密码黑名单
	if isWeakPassword(password) {
		return fmt.Errorf("密码过于简单，请使用更复杂的密码")
	}

	// 【最低安全策略5】禁止包含用户名、手机号、邮箱前缀
	if user.PhoneNum != "" && containsSubstring(password, user.PhoneNum) {
		return fmt.Errorf("密码不能包含手机号")
	}
	if user.Nickname != "" && containsSubstring(password, user.Nickname) {
		return fmt.Errorf("密码不能包含昵称")
	}
	if user.Email != "" {
		emailPrefix := regexp.MustCompile(`^([^@]+)`).FindString(user.Email)
		if emailPrefix != "" && containsSubstring(password, emailPrefix) {
			return fmt.Errorf("密码不能包含邮箱前缀")
		}
	}

	// 【最低安全策略2】必须包含三类中至少两种
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case regexp.MustCompile(`[!@#$%^&*()_+\-=]`).MatchString(string(c)):
			hasSpecial = true
		}
	}

	categoryCount := 0
	if hasUpper {
		categoryCount++
	}
	if hasLower {
		categoryCount++
	}
	if hasDigit {
		categoryCount++
	}
	if hasSpecial {
		categoryCount++
	}

	if categoryCount < 2 {
		return fmt.Errorf("密码必须包含大写字母、小写字母、数字、特殊符号中的至少两种")
	}

	return nil
}

// isWeakPassword 检查是否为弱密码
// 【最低安全策略3】禁止弱密码黑名单（内置常用弱密码库）
func isWeakPassword(password string) bool {
	weakPasswords := []string{
		"123456", "password", "12345678", "qwerty", "abc123", "monkey",
		"1234567", "letmein", "trustno1", "dragon", "baseball", "iloveyou",
		"master", "sunshine", "ashley", "bailey", "shadow", "123123",
		"654321", "superman", "qazwsx", "michael", "football", "password1",
	}
	
	for _, weak := range weakPasswords {
		if password == weak {
			return true
		}
	}
	
	return false
}

// containsSubstring 检查字符串是否包含子串
func containsSubstring(str, substr string) bool {
	if len(substr) < 3 {
		return false
	}
	return len(str) >= len(substr) && (regexp.MustCompile(regexp.QuoteMeta(substr)).MatchString(str))
}