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

// TODO(CHAO): 这里需要调用外部短信网关接口发送验证码短信
// GenerateSMSCode 生成6位数字短信验证码
// 参数说明：
//   - phoneNum: 用户手机号
//
// 返回值：
//   - error: 错误信息，如果生成成功则返回nil
//
// 验证码生成逻辑：
//   1. 生成6位随机数字
//   2. 存入数据库，设置过期时间
//   3. 实际项目中会调用短信网关发送验证码
func GenerateSMSCode(phoneNum string) error {
	code := generateRandomCode(6) // 生成6位数字验证码
	return repository.CreateVerificationCode(phoneNum, code, repository.VerificationCodeTypeSMS)
}

// GenerateAlnumCode 生成4位字母数字混合验证码
// 参数说明：
//   - phoneNum: 用户手机号
//
// 返回值：
//   - error: 错误信息
//
// 用途：用于图形验证码或其他验证场景
func GenerateAlnumCode(phoneNum string) error {
	code := generateRandomAlnum(4) // 生成4位混合验证码
	return repository.CreateVerificationCode(phoneNum, code, repository.VerificationCodeTypeAlnum)
}

// generateRandomCode 生成指定长度的数字验证码
// 参数说明：
//   - length: 验证码长度
//
// 返回值：
//   - string: 生成的数字字符串
//
// 算法说明：
//   使用crypto/rand安全随机数生成器
//   每次生成一个0-9之间的随机数字
func generateRandomCode(length int) string {
	var result string
	for i := 0; i < length; i++ {
		// rand.Int生成0-9之间的随机整数
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		// fmt.Sprintf将数字转换为字符串并拼接
		result += fmt.Sprintf("%d", num)
	}
	return result
}

// generateRandomAlnum 生成指定长度的字母数字混合验证码
// 参数说明：
//   - length: 验证码长度
//
// 返回值：
//   - string: 生成的混合字符串
//
// 字符集说明：
//   使用0-9和A-Z的组合，共36个字符
func generateRandomAlnum(length int) string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var result string
	for i := 0; i < length; i++ {
		// 在字符集中随机选择一个字符
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result += string(chars[num.Int64()])
	}
	return result
}

// Register 用户注册函数
// 参数说明：
//   - phoneNum: 用户手机号
//   - code: 用户输入的验证码
//
// 返回值：
//   - string: JWT令牌，注册成功时返回
//   - error: 错误信息
//
// 注册流程：
//   1. 验证验证码是否有效且未过期
//   2. 检查手机号是否已注册
//   3. 生成密码盐值
//   4. 创建新用户记录
//   5. 删除已使用的验证码
//   6. 生成JWT令牌返回
func Register(phoneNum, code string) (string, error) {
	// 验证验证码有效性
	err := repository.GetValidVerificationCode(phoneNum, code, repository.VerificationCodeTypeSMS)
	if err != nil {
		return "", fmt.Errorf("验证码无效或已过期")
	}

	// 检查手机号是否已注册
	existingUser, _ := repository.GetUserByPhone(phoneNum)
	if existingUser.ID != 0 {
		return "", fmt.Errorf("该手机号已注册")
	}

	// 使用bcrypt加密空密码（新注册用户首次不设置密码）
	password, _ := bcrypt.GenerateFromPassword([]byte(""), bcrypt.DefaultCost)

	// 创建用户对象
	user := &model.User{
		Uid:           utils.GenerateUID(),
		PhoneNum:      phoneNum,
		Password:      string(password),
		Gender:        -1,
		AccountStatus: 2,
	}

	// 保存用户到数据库
	err = repository.CreateUser(user)
	if err != nil {
		return "", fmt.Errorf("注册失败")
	}

	repository.DeleteVerificationCode(phoneNum, repository.VerificationCodeTypeSMS)
	return middleware.GenerateToken(user.Uid)
}

// LoginByCode 验证码登录
// 参数说明：
//   - phoneNum: 用户手机号
//   - code: 用户输入的验证码
//
// 返回值：
//   - string: JWT令牌
//   - error: 错误信息
//
// 登录流程：
//   1. 验证验证码有效性
//   2. 查找用户是否存在
//   3. 删除已使用的验证码
//   4. 生成JWT令牌
func LoginByCode(phoneNum, code string) (string, error) {
	// 验证验证码
	err := repository.GetValidVerificationCode(phoneNum, code, repository.VerificationCodeTypeSMS)
	if err != nil {
		return "", fmt.Errorf("验证码无效或已过期")
	}

	// 查找用户
	user, err := repository.GetUserByPhone(phoneNum)
	if err != nil {
		return "", fmt.Errorf("用户不存在")
	}

	repository.DeleteVerificationCode(phoneNum, repository.VerificationCodeTypeSMS)
	return middleware.GenerateToken(user.Uid)
}

// LoginByPassword 密码登录
// 参数说明：
//   - phoneNum: 用户手机号
//   - password: 用户输入的密码（明文）
//
// 返回值：
//   - string: JWT令牌
//   - error: 错误信息
//
// 登录流程：
//   1. 根据手机号查找用户
//   2. 使用相同的盐值和算法加密输入密码
//   3. 与数据库中存储的密码哈希比对
//   4. 比对成功生成JWT令牌
func LoginByPassword(phoneNum, password string) (string, error) {
	// 查找用户
	user, err := repository.GetUserByPhone(phoneNum)
	if err != nil {
		return "", fmt.Errorf("用户不存在")
	}

	// 使用bcrypt比对密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("密码错误")
	}

	return middleware.GenerateToken(user.Uid)
}

// ResetPassword 重置密码
// 参数说明：
//   - phoneNum: 用户手机号
//   - pwd1: 新密码
//   - pwd2: 确认密码
//
// 返回值：
//   - error: 错误信息
//
// 重置流程：
//   1. 验证两次密码输入一致
//   2. 验证密码格式（6-20位字母数字）
//   3. 查找用户
//   4. 更新密码（使用新的盐值加密）
func ResetPassword(phoneNum, pwd1, pwd2 string) error {
	// 验证两次密码一致
	if pwd1 != pwd2 {
		return fmt.Errorf("两次密码不一致")
	}

	// 验证密码格式：6-20位字母或数字
	if !validatePassword(pwd1) {
		return fmt.Errorf("密码格式不正确")
	}

	// 查找用户
	user, err := repository.GetUserByPhone(phoneNum)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 使用bcrypt加密新密码
	password, _ := bcrypt.GenerateFromPassword([]byte(pwd1), bcrypt.DefaultCost)
	user.Password = string(password)
	return repository.UpdateUser(user)
}

// validatePassword 验证密码格式
// 参数说明：
//   - password: 要验证的密码
//
// 返回值：
//   - bool: 格式是否有效
//
// 验证规则：
//   - 长度：6-20个字符
//   - 字符：只能包含字母和数字
func validatePassword(password string) bool {
	// 正则表达式：匹配6-20位字母数字组合
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]{6,20}$`, password)
	return matched
}
