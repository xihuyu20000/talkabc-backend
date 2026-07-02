package middleware

import (
	"backend/internal/config"       // 配置模块，获取JWT密钥等配置
	"backend/internal/repository"   // 数据访问层，用于更新用户最后活跃时间
	"backend/pkg/response"         // 统一响应模块
	"github.com/gin-gonic/gin"     // Gin框架
	"github.com/golang-jwt/jwt/v5"  // JWT库，用于生成和验证令牌
	"strings"                      // 字符串处理
)

type Claims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

// JWT JWT认证中间件
// 功能说明：
//   1. 从请求头获取Authorization令牌
//   2. 验证令牌格式和有效性
//   3. 解析令牌获取用户ID
//   4. 将用户ID存入Gin上下文，供后续处理函数使用
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

func GenerateToken(uid string) (string, error) {
	claims := Claims{
		UID: uid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}
