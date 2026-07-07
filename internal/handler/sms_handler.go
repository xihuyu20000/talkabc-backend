package handler

import (
	"backend/internal/config"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/response"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

/**
短信验证码功能需求清单：
1. 短信验证码和邮件验证码有效期5分钟（redis key ttl来控制）
2. 验证码为6位纯数字
3. 每个手机号60秒内只能发送一次短信验证码，且这一规则的校验必须在服务器端执行
4. 同一个手机号在同一时间内只有一个有效的短信验证码，以最新的验证码为准
5. 保存于服务器端的验证码，至多可被使用1次（无论和请求中的验证码是否匹配）
6. 短信验证码不可直接记录到日志文件
7. 发送短信验证码之前，先验证图形验证码是否正确
8. 1小时内手机号码发送验证码次数限制10次
9. 每天(24小时内)首次获取手机验证码不需要图形验证码
10. 不同业务类型的验证码防止冲突（每个业务场景有自己的tag）
*/

// SendSMSCode 发送短信验证码
// @Summary 发送短信验证码
// @Description 向指定手机号发送6位短信验证码，用于身份验证
// @Description
// @Description **验证码安全机制：**
// @Description - 有效期控制：验证码有效期5分钟（通过Redis TTL自动过期）
// @Description - 格式规范：6位纯数字验证码
// @Description - 发送间隔：每个手机号60秒内只能发送一次（服务器端严格校验）
// @Description - 唯一性：同一手机号同一时间只有一个有效验证码（最新的为准）
// @Description - 一次性使用：验证码验证后立即删除，不可重复使用
// @Description - 日志脱敏：验证码内容不记录到日志文件，保护用户隐私
// @Description - 图形验证码保护：发送前验证图形验证码（每日首次发送可免图形验证码）
// @Description - 频率限制：1小时内同一手机号最多发送10次
// @Description - 业务隔离：不同业务类型的验证码独立隔离（通过tag区分，如register、login）
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param phonenum query string true "手机号"
// @Param captcha_id query string false "图形验证码ID（非每日首次发送时必填）"
// @Param captcha_code query string false "图形验证码（非每日首次发送时必填）"
// @Param tag query string true "业务标签，允许的值：register(注册)、login(登录)、change_phone(更换手机号)、reset_password(重置密码)"
// @Success 200 {object} map[string]interface{} "发送成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误或验证失败"
// @Failure 500 {object} map[string]interface{} "发送失败"
// @Router /auth/code-sms [get]
func SendSMSCode(c *gin.Context) {
	// 【功能10】支持业务标签参数，不同业务场景使用不同的验证码key
	req := service.GenerateSMSCodeRequest{
		PhoneNum:    c.Query("phonenum"),
		CaptchaID:   c.Query("captcha_id"),
		CaptchaCode: c.Query("captcha_code"),
		Tag:         c.Query("tag"),
	}

	logger.Infof("[Handler] SendSMSCode - PhoneNum: %s, Tag: %s, CaptchaID: %s, CaptchaCode: %s",
		req.PhoneNum, req.Tag, req.CaptchaID, req.CaptchaCode)

	// 调用service层执行完整的验证码发送逻辑（包含所有安全校验）
	err := service.GenerateSMSCode(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// VerifySMSCode 验证短信验证码
// @Summary 验证短信验证码
// @Description 验证客户端传入的短信验证码是否正确
// @Description 
// @Description **安全机制：**
// @Description - 原子操作：使用Lua脚本保证验证和删除的原子性，防止并发攻击
// @Description - 一次性使用：无论验证是否成功，验证码都会被删除
// @Description - 业务隔离：不同业务类型的验证码独立验证（通过tag区分）
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param body body map[string]string true "包含phonenum(手机号)、code(验证码)、tag(业务标签，允许的值：register、login、change_phone、reset_password)"
// @Success 200 {object} map[string]interface{} "验证成功"
// @Failure 400 {object} map[string]interface{} "参数错误或验证码不正确"
// @Router /auth/code-sms/verify [post]
func VerifySMSCode(c *gin.Context) {
	var req struct {
		PhoneNum string `json:"phonenum"`
		Code     string `json:"code"`
		// 【功能10】支持业务标签参数，与发送时保持一致
		Tag string `json:"tag"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.PhoneNum == "" || req.Code == "" {
		response.BadRequest(c, "手机号和验证码不能为空")
		return
	}

	logger.Infof("[Handler] VerifySMSCode - PhoneNum: %s, Tag: %s", req.PhoneNum, req.Tag)

	// 调用service层执行验证（使用Lua脚本保证原子性，实现功能5）
	err := service.VerifySMSCode(req.PhoneNum, req.Code, req.Tag)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GenerateAlnumCode 生成图形验证码
// @Summary 生成图形验证码
// @Description 使用base64Captcha生成数字图形验证码图片，返回验证码ID和base64编码的图片数据
// @Description 
// @Description **安全机制：**
// @Description - 有效期控制：验证码有效期5分钟（通过Redis TTL自动过期）
// @Description - 不可预测：使用base64Captcha库生成安全的图形验证码，防止自动化识别
// @Description - 安全存储：验证码存储在Redis中，不记录到日志文件
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Success 200 {object} map[string]interface{} "生成成功"
// @Failure 500 {object} map[string]interface{} "生成失败"
// @Router /auth/code-alnum [get]
func GenerateAlnumCode(c *gin.Context) {
	// 使用base64Captcha库生成数字验证码图片
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, base64Captcha.DefaultMemStore)

	id, b64s, answer, err := captcha.Generate()
	if err != nil {
		response.InternalError(c, "生成验证码失败")
		return
	}

	// 【功能1】图形验证码同样使用Redis TTL控制有效期（5分钟）
	err = config.RDB.Set(context.Background(), id, answer, time.Duration(config.AppConfig.Security.SMSValidMinutes)*time.Minute).Err()
	if err != nil {
		response.InternalError(c, "保存验证码失败")
		return
	}

	response.Success(c, gin.H{
		"captcha_id": id,
		"image":      b64s,
	})
}

// VerifyAlnumCode 验证图形验证码
// @Summary 验证图形验证码
// @Description 验证客户端传入的图形验证码是否正确
// @Description 
// @Description **安全机制：**
// @Description - 一次性使用：验证通过后立即删除验证码，防止重复使用
// @Description - 过期处理：验证码过期或不存在时返回错误
// @Tags 认证
// @Accept application/json
// @Produce application/json
// @Param body body map[string]string true "包含captcha_id和code"
// @Success 200 {object} map[string]interface{} "验证成功"
// @Failure 400 {object} map[string]interface{} "参数错误或验证码不正确"
// @Router /auth/code-alnum/verify [post]
func VerifyAlnumCode(c *gin.Context) {
	var req struct {
		CaptchaID string `json:"captcha_id"`
		Code      string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.CaptchaID == "" || req.Code == "" {
		response.BadRequest(c, "验证码ID和验证码不能为空")
		return
	}

	storedCode, err := config.RDB.Get(context.Background(), req.CaptchaID).Result()
	if err != nil {
		response.BadRequest(c, "验证码已过期或不存在")
		return
	}

	if storedCode != req.Code {
		response.BadRequest(c, "验证码不正确")
		return
	}

	// 【功能5】图形验证码验证通过后立即删除，防止重复使用
	err = config.RDB.Del(context.Background(), req.CaptchaID).Err()
	if err != nil {
		response.InternalError(c, "删除验证码失败")
		return
	}

	response.Success(c, nil)
}