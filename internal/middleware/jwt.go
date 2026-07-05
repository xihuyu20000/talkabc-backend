package middleware

import (
	"backend/internal/config"     // 配置模块，获取JWT密钥等配置
	"backend/internal/repository" // 数据访问层，用于更新用户最后活跃时间
	"backend/pkg/response"        // 统一响应模块
	"fmt"                         // 格式化字符串
	"math/rand"
	"time"

	// 随机数生成
	"strings" // 字符串处理

	// 时间处理
	"github.com/gin-gonic/gin"     // Gin框架
	"github.com/golang-jwt/jwt/v5" // JWT库，用于生成和验证令牌
)

type Claims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims 刷新令牌声明
// 【安全规则】刷新令牌需要比访问令牌更长的有效期，且包含用户ID用于验证
type RefreshTokenClaims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

// JWT JWT认证中间件
// 功能说明：
//   1. 从请求头获取Authorization令牌
//   2. 验证令牌格式和有效性
//   3. 解析令牌获取用户ID
//   4. 检查令牌是否在Redis中有效（支持主动失效）
//   5. 将用户ID存入Gin上下文，供后续处理函数使用
//
// 使用方式：
//   在路由配置中使用：v1.POST("/logout", middleware.JWT(), handler.Logout)
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization字段
		authHeader := c.GetHeader("Authorization")

		// 检查是否缺少令牌
		if authHeader == "" {
			response.Unauthorized(c, "未登录")
			c.Abort() // 阻止后续处理函数执行
			return
		}

		// 分割令牌，格式应为 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		// 验证格式：必须有两部分，且第一部分是"Bearer"
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "token格式错误")
			c.Abort()
			return
		}

		// 提取令牌主体
		token := parts[1]
		// 创建声明对象用于接收解析结果
		claims := &Claims{}

		// 解析并验证令牌
		// 第三个参数是回调函数，用于提供签名密钥
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			// 使用配置文件中的密钥对令牌进行签名验证
			return []byte(config.AppConfig.JWT.Secret), nil
		})

		// 检查解析是否成功
		if err != nil {
			response.Unauthorized(c, "token无效")
			c.Abort()
			return
		}

		// 【安全规则】检查令牌是否在Redis中有效（支持主动失效，如更换手机号后）
		tokenKey := fmt.Sprintf("user_token:%s", claims.UID)
		validToken, err := config.RDB.Get(c.Request.Context(), tokenKey).Result()
		if err != nil || validToken == "" || validToken != token {
			response.Unauthorized(c, "登录已失效，请重新登录")
			c.Abort()
			return
		}

		// 将用户ID存入Gin上下文
		// 后续处理函数可以通过GetUID函数获取当前登录用户ID
		c.Set("uid", claims.UID)

		// 异步更新用户最后活跃时间，不阻塞请求处理
		// 使用goroutine并发执行，避免影响主请求响应速度
		go func(uid string) {
			repository.UpdateLastSeenAt(uid)
		}(claims.UID)

		// 调用下一个中间件或处理函数
		c.Next()
	}
}

func GetUID(c *gin.Context) string {
	if uid, exists := c.Get("uid"); exists {
		return uid.(string)
	}
	return ""
}

// GenerateToken 生成访问令牌
// 【安全规则】访问令牌有效期为2小时，过期后需要使用刷新令牌获取新令牌
// 生成规则：包含用户ID、签发时间、过期时间和唯一标识(jti)，确保每次生成的token不同
func GenerateToken(uid string) (string, error) {
	claims := Claims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:         generateRandomString(16),
			IssuedAt:   jwt.NewNumericDate(time.Now()),
			ExpiresAt:  jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

// GenerateRefreshToken 生成刷新令牌
// 【安全规则】刷新令牌有效期为7天，比访问令牌长，用于获取新的访问令牌
// 生成规则：使用随机字符串+JWT签名，确保安全性
func GenerateRefreshToken(uid string) (string, error) {
	randomPart := generateRandomString(32)
	
	claims := RefreshTokenClaims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.AppConfig.JWT.Secret))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", randomPart, signedToken), nil
}

// generateRandomString 生成指定长度的随机字符串
// 【安全规则】用于生成刷新令牌的随机部分，增加令牌复杂度
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// ParseRefreshToken 解析刷新令牌
// 【安全规则】验证刷新令牌格式和有效性，提取用户ID
func ParseRefreshToken(refreshToken string) (string, error) {
	parts := strings.SplitN(refreshToken, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("刷新令牌格式错误")
	}

	signedToken := parts[1]
	claims := &RefreshTokenClaims{}

	_, err := jwt.ParseWithClaims(signedToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil {
		return "", fmt.Errorf("刷新令牌无效")
	}

	return claims.UID, nil
}
