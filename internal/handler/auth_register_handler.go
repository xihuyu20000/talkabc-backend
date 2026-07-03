package handler

import (
	"backend/internal/service"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

/**
注册时的安全性规则如下：
1. 服务器检验通过IP是否位于黑名单中，是则拦截。
2. 服务器检验通过IP检查注册请求的频率，请求频率限制在1分钟10次。
3. 校验手机号手机号是否位于黑名单中，是则拦截
4. 校验手机号是否已被占用（用户已存在）。
5. 检查该设备是否在黑名单中，是则拦截。
6. 密码复杂度校验（最低安全策略）。
7. 注册成功后清理验证码，防止二次复用。
8. 记录注册操作日志（不可删除）。
*/

// Register 用户注册
// @Summary 用户注册
// @Description 使用手机号、验证码和密码进行注册。安全规则：1. IP黑名单检查；2. IP注册频率限制（1分钟10次）；3. 手机号黑名单检查；4. 手机号唯一性检查；5. 设备黑名单检查；6. 密码复杂度校验（≥8位，至少包含两种字符类型：大写字母、小写字母、数字、特殊符号；禁止弱密码；禁止包含手机号/昵称/邮箱前缀）；7. 注册成功后清理验证码；8. 记录注册操作日志（不可删除）
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum formData string true "手机号"
// @Param code formData string true "验证码"
// @Param password formData string true "密码（≥8位，至少包含两种字符类型）"
// @Param device_id formData string false "设备ID"
// @Success 200 {object} map[string]interface{} "注册成功，返回token"
// @Failure 400 {object} map[string]interface{} "请求参数错误或安全校验失败"
// @Failure 500 {object} map[string]interface{} "注册失败"
// @Router /api/v1/auth/register [post]
func Register(c *gin.Context) {
	phoneNum := c.PostForm("phonenum")
	code := c.PostForm("code")
	password := c.PostForm("password")
	// 【注册安全规则5】获取设备ID（由客户端传递，用于设备黑名单校验）
	deviceID := c.PostForm("device_id")

	if phoneNum == "" || code == "" || password == "" {
		response.BadRequest(c, "手机号、验证码和密码不能为空")
		return
	}

	// 获取客户端真实IP和UA（【注册安全规则8】记录操作日志）
	clientIP := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	// 【注册安全规则】组装注册请求，包含IP、设备ID、UA用于安全校验和日志记录
	req := service.RegisterRequest{
		PhoneNum: phoneNum,
		Code:     code,
		Password: password,
		IP:       clientIP,
		DeviceID: deviceID,
		UA:       ua,
	}

	token, err := service.Register(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"token": token})
}

/**
重置凭证（最核心，找回 / 重置链路）
1. 重置 Token 设计（推荐，优于短信验证码长期使用）
- 单次有效：使用一次立即销毁，不可重复使用
- 短有效期：邮箱重置 15～30min；短信重置 5～10min
- 不可预测：Go 生成：crypto/rand 随机字节，禁止 math/rand
- 绑定唯一信息：Token 必须绑定 userID + 设备标识 + 创建时间，防止跨账号盗用
- 禁止明文存库：数据库只存 Token 哈希，不存原文
- 存储：sha256 (token) 或 bcrypt 加盐存储
- 校验逻辑：用户提交 token 后，先哈希再和库中对比


重置流程行为风控：
1. 记录用户常用 IP、设备 UA、地区
2. 同一账号 24h 最多允许 3 次密码重置，超限锁定重置通道 24h
3. 敏感操作日志落地（不可删除日志）：
用户 ID、操作时间、IP、UA、操作类型（发起重置 / 完成重置）、是否成功
4. 重置完成后推送通知：APP推送、短信、邮箱告知「密码已修改」，用户非本人操作可快速冻结账号


密码存储加密:
1. 禁止 MD5、SHA1、SHA256 明文哈希，必须使用慢哈希算法
推荐：bcrypt / argon2id（Go 官方标准库 / 第三方库）
bcrypt cost 设置 10～12，平衡安全与性能
2. 禁止全局固定盐，每个用户独立随机盐（bcrypt 自动内置盐）
3. 重置成功后：
- 清空该用户全部登录态（Redis token、JWT、设备登录记录全部销毁）
- 清空所有未使用重置 Token / 验证码，防止二次复用
- 绝不返回原始密码、加密密码到前端

最低安全策略（生产标准）
1. 长度：≥8 位，推荐 12 位以上
2. 必须包含三类中至少两种：
大写字母 A-Z
小写字母 a-z
数字 0-9
特殊符号 !@#$%^&*()_+-=
3. 禁止弱密码黑名单（内置常用弱密码库：123456、password、手机号后 6 位、生日）
4. 禁止和历史 5 次旧密码重复
5. 禁止包含用户名、手机号、邮箱前缀
6. 密码前后空格自动修剪，空白密码直接拦截
*/

// InitiateResetPassword 发起密码重置（生成重置Token）
// @Summary 发起密码重置
// @Description 根据手机号发起密码重置，生成重置Token并发送。安全规则：1. 同一账号24h内最多允许3次密码重置，超限锁定重置通道24h；2. 记录敏感操作日志（用户ID、IP、UA、操作类型、是否成功，不可删除）；3. Token设计：单次有效（使用后立即销毁）、短有效期（5分钟）、不可预测（crypto/rand生成）、绑定userID+设备标识、禁止明文存库（仅存储sha256哈希）
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum formData string true "手机号"
// @Param device_id formData string false "设备ID"
// @Success 200 {object} map[string]interface{} "发起成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误或频率限制"
// @Failure 500 {object} map[string]interface{} "发起失败"
// @Router /api/v1/auth/reset-password/initiate [post]
func InitiateResetPassword(c *gin.Context) {
	phoneNum := c.PostForm("phonenum")
	// 【重置凭证】获取设备ID，用于绑定唯一信息，防止跨账号盗用
	deviceID := c.PostForm("device_id")

	if phoneNum == "" {
		response.BadRequest(c, "手机号不能为空")
		return
	}

	// 获取客户端真实IP和UA（【重置流程行为风控】记录用户常用IP、设备UA）
	clientIP := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	// 【重置凭证】组装发起重置请求，包含IP、设备ID、UA用于安全校验和日志记录
	req := service.ResetPasswordRequest{
		PhoneNum: phoneNum,
		DeviceID: deviceID,
		IP:       clientIP,
		UA:       ua,
	}

	err := service.InitiateResetPassword(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "重置链接已发送，有效期5分钟"})
}

// ValidateResetToken 验证重置Token是否有效
// @Summary 验证重置Token
// @Description 验证重置Token是否存在、未使用、未过期。Token设计：单次有效、短有效期（5分钟）、绑定userID+设备标识、禁止明文存库（仅存储sha256哈希）
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param token query string true "重置Token"
// @Success 200 {object} map[string]interface{} "Token有效"
// @Failure 400 {object} map[string]interface{} "Token无效或已过期"
// @Router /api/v1/auth/reset-password/validate [get]
func ValidateResetToken(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		response.BadRequest(c, "重置链接无效")
		return
	}

	// 【重置凭证】验证Token是否有效（不标记为已使用，允许前端预先校验）
	_, err := service.ValidateResetToken(token)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "重置链接有效"})
}

// CompleteResetPassword 完成密码重置
// @Summary 完成密码重置
// @Description 使用重置Token设置新密码。安全规则：1. 密码复杂度校验（≥8位，至少包含两种字符类型：大写字母、小写字母、数字、特殊符号；禁止弱密码；禁止包含手机号/昵称/邮箱前缀；禁止与历史5次密码重复）；2. 密码存储：使用bcrypt加密（cost=10，自动内置盐）；3. 重置成功后：清空该用户全部登录态（Redis token等）、清空所有未使用重置Token、绝不返回原始密码或加密密码；4. 记录敏感操作日志（不可删除）；5. 设备验证：验证Token绑定的设备标识，防止跨账号盗用
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param token formData string true "重置Token"
// @Param pwd1 formData string true "新密码"
// @Param pwd2 formData string true "确认密码"
// @Param device_id formData string false "设备ID"
// @Success 200 {object} map[string]interface{} "重置成功"
// @Failure 400 {object} map[string]interface{} "参数错误或密码校验失败"
// @Failure 500 {object} map[string]interface{} "重置失败"
// @Router /api/v1/auth/reset-password/complete [post]
func CompleteResetPassword(c *gin.Context) {
	token := c.PostForm("token")
	pwd1 := c.PostForm("pwd1")
	pwd2 := c.PostForm("pwd2")
	// 【重置凭证】获取设备ID，验证绑定信息，防止跨账号盗用
	deviceID := c.PostForm("device_id")

	if token == "" || pwd1 == "" || pwd2 == "" {
		response.BadRequest(c, "参数不能为空")
		return
	}

	// 获取客户端真实IP和UA（【重置流程行为风控】记录操作日志）
	clientIP := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	// 【密码存储加密】组装完成重置请求，包含IP、设备ID、UA用于安全校验和日志记录
	req := service.CompleteResetPasswordRequest{
		Token:    token,
		Pwd1:     pwd1,
		Pwd2:     pwd2,
		DeviceID: deviceID,
		IP:       clientIP,
		UA:       ua,
	}

	err := service.CompleteResetPassword(req)
	if err != nil {
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功，请重新登录"})
}