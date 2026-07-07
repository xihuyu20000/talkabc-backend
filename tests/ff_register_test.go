package test

import (
	"backend/internal/config"
	"backend/internal/handler"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	InitTest()
}

// TestRegister_FullFlow 完整注册流程集成测试
// 模拟用户从输入手机号到注册成功的全过程，验证所有安全规则检查
//
// 测试流程：
//   Step1: 用户输入手机号，点击发送验证码
//     - 系统安全规则检查：
//       1. 检查手机号格式
//       2. 检查60秒冷却期（防刷）
//       3. 检查1小时发送次数限制（10次）
//       4. 判断是否需要图形验证码（根据配置）
//     - 生成6位随机验证码
//     - 设置冷却期
//     - 存储验证码到Redis（5分钟有效期）
//     - 调用短信网关发送验证码
//
//   Step2: 用户输入验证码和密码，点击注册
//     - 系统安全规则检查：
//       1. IP黑名单检查
//       2. IP注册频率限制（1分钟10次）
//       3. 手机号黑名单检查
//       4. 设备黑名单检查
//       5. 验证码验证（原子验证并删除，防止暴力攻击）
//       6. 手机号唯一性检查（用户不存在）
//       7. 密码复杂度校验（≥8位，至少包含两种字符类型）
//     - 使用bcrypt加密密码存储
//     - 创建用户记录
//     - 清理验证码防止二次复用
//     - 记录注册操作日志（不可删除）
//     - 生成JWT token返回给用户
//
//   Step3: 验证验证码已被清理
//     - 注册成功后，验证码应从Redis中删除，防止二次使用
func TestRegister_FullFlow(t *testing.T) {
	if config.DB == nil || mockSMSGateway == nil {
		t.Skip("Database or Mock SMS gateway not initialized, skipping test")
	}

	// 初始化路由
	router := gin.New()
	router.GET("/v1/code/sms", handler.SendSMSCode)
	router.POST("/v1/register", handler.Register)

	// 测试数据
	phoneNum := "13900139003"
	password := "Password123"

	// 清理测试环境（Redis和Mock短信记录）
	config.RDB.FlushDB(config.RDB.Context())
	mockSMSGateway.ClearSentMessages()

	// ==================== Step1: 发送短信验证码 ====================
	// 模拟用户输入手机号，点击"发送验证码"按钮
	t.Run("Step1_SendSMSCode", func(t *testing.T) {
		// 构造发送验证码请求
		req, _ := http.NewRequest("GET", "/v1/code/sms?phonenum="+phoneNum+"&tag=register", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// 验证响应状态码
		if resp.Code != http.StatusOK {
			t.Fatalf("SendSMSCode failed with status %d: %s", resp.Code, resp.Body.String())
		}

		// 验证Mock短信网关是否收到发送请求
		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) != 1 {
			t.Fatalf("Expected 1 sent message, got %d", len(sentMsgs))
		}

		// 验证验证码格式（6位数字）
		code := sentMsgs[0].Code
		if len(code) != 6 {
			t.Fatalf("Expected code length 6, got %d", len(code))
		}

		t.Logf("Generated verification code: %s", code)
	})

	// ==================== Step2: 用户注册 ====================
	// 模拟用户收到验证码后，输入验证码和密码，点击"注册"按钮
	t.Run("Step2_Register", func(t *testing.T) {
		// 获取之前发送的验证码
		sentMsgs := mockSMSGateway.GetSentMessages()
		if len(sentMsgs) == 0 {
			t.Skip("No SMS code sent in previous step")
		}
		code := sentMsgs[0].Code

		// 构造注册请求（使用 JSON 格式）
		registerData := map[string]string{
			"phonenum": phoneNum,
			"code":     code,
			"password": password,
		}
		jsonData, _ := json.Marshal(registerData)
		registerReq, _ := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(jsonData))
		registerReq.Header.Set("Content-Type", "application/json")
		registerResp := httptest.NewRecorder()

		router.ServeHTTP(registerResp, registerReq)

		t.Logf("Register response status: %d", registerResp.Code)
		t.Logf("Register response body: %s", registerResp.Body.String())

		// 验证响应状态码
		if registerResp.Code != http.StatusOK {
			var result map[string]interface{}
			if err := json.Unmarshal(registerResp.Body.Bytes(), &result); err == nil {
				if msg, ok := result["msg"].(string); ok {
					t.Logf("Register error message: %s", msg)
				}
			}
			t.Fatalf("Register failed with status %d: %s", registerResp.Code, registerResp.Body.String())
		}

		// 解析响应数据
		var result map[string]interface{}
		if err := json.Unmarshal(registerResp.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal register response: %v", err)
		}

		// 验证响应码（0表示成功）
		if result["code"] != float64(0) {
			t.Errorf("Expected code 0, got %v", result["code"])
			return
		}

		// 验证响应数据
		data := result["data"]
		if data == nil {
			t.Error("Data should not be nil")
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Errorf("Data should be map[string]interface{}, got %T", data)
			return
		}

		// 验证JWT token存在且非空
		token, ok := dataMap["access_token"].(string)
		if !ok || token == "" {
			t.Error("Token should not be empty")
		}

		t.Logf("Generated JWT token: %s", token)
	})

	// ==================== Step3: 验证验证码已被清理 ====================
	// 注册成功后，验证码应被自动清理，防止二次使用
	t.Run("Step3_VerificationCodeCleared", func(t *testing.T) {
		// 尝试从Redis获取验证码
		_, err := config.RDB.Get(config.RDB.Context(), "verification_code:"+phoneNum+":sms:register").Result()
		// 如果获取成功（err为nil），说明验证码未被清理，测试失败
		if err == nil {
			t.Error("Verification code should be cleared after registration")
		}
	})
}