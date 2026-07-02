package router

import (
	"backend/internal/config"     // 配置模块
	"backend/internal/handler"    // 处理器函数
	"backend/internal/middleware" // 中间件（JWT认证等）

	"github.com/gin-contrib/cors" // CORS跨域资源共享
	"github.com/gin-gonic/gin"    // Gin Web框架
)

// InitRouter 初始化路由配置
// 返回值：
//   - *gin.Engine: 配置好的Gin引擎实例
//
// 路由分组说明：
//   - 所有API都在/api/v1路径下
//   - 分为公开接口（无需认证）和私有接口（需要JWT认证）
//   - 使用路由分组统一添加中间件，避免每个接口重复配置
func initCORSConfig() cors.Config {
	cfg := cors.Config{
		AllowMethods:     config.AppConfig.CORS.Methods,
		AllowHeaders:     config.AppConfig.CORS.Headers,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: config.AppConfig.CORS.Credentials,
	}

	for _, origin := range config.AppConfig.CORS.Origins {
		if origin == "*" {
			cfg.AllowOriginFunc = func(origin string) bool {
				return true
			}
			return cfg
		}
	}

	cfg.AllowOrigins = config.AppConfig.CORS.Origins
	return cfg
}

func InitRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(initCORSConfig()))
	

	apiV1 := r.Group("/api/v1")
	{
		// ==================== 公开接口（无需认证） ====================
		public := apiV1.Group("/")
		{
			// 系统模块 - 公开接口
			public.GET("/sys/code-sms", handler.SendSMSCode)
			public.GET("/sys/code-alnum", handler.GenerateAlnumCode)
			public.POST("/sys/register", handler.Register)
			public.POST("/sys/login-code", handler.LoginByCode)
			public.POST("/sys/login-pwd", handler.LoginByPassword)
			public.POST("/sys/reset-pwd", handler.ResetPassword)
		}

		// ==================== 私有接口（需要JWT认证） ====================
		// 在组级别统一添加JWT中间件，组内所有接口自动继承
		private := apiV1.Group("/", middleware.JWT())
		{
			// 系统模块 - 认证接口
			private.POST("/sys/logout", handler.Logout)

			// 用户模块（关注/拉黑改用 WebSocket）
			private.GET("/user/users", handler.GetUserList)
			private.GET("/user/info/:uid", handler.GetUserInfo)
			private.GET("/user/focuslist/:uid", handler.GetFocusList)
			private.GET("/user/fanslist/:uid", handler.GetFansList)
			private.POST("/user/notify/:uid/:flag", handler.SetUserNotify)
			private.POST("/user/greet/:uid", handler.GreetUser)
			private.POST("/user/upload-avatar", handler.UploadAvatar)
			private.POST("/user/collect-myinfo", handler.CollectMyInfo)
			private.POST("/user/collect-aiminfo", handler.CollectAimInfo)
			private.GET("/user/adbanner", handler.GetLatestAdBanner)
			private.POST("/user/gift/:uid/:giftid", handler.SendGift)
			private.GET("/user/praise-me", handler.GetPraiseMeList)
			private.GET("/user/comment-me", handler.GetCommentMeList)
			private.GET("/user/add-me", handler.GetAddMeList)
			private.GET("/user/visit-me", handler.GetVisitMeList)
			private.GET("/user/like-me", handler.GetLikeMeList)
			private.POST("/user/agree-friend/:uid/:flag", handler.AgreeFriendRequest)
			private.POST("/user/add/:uid/:flag", handler.AddFriend)

			// 聊天消息模块（消息收发改用 WebSocket）
			private.GET("/msg/sysmsgs", handler.GetSystemMsgList)
			private.GET("/msg/latest", handler.GetLatestUserMsg)
			private.GET("/msg/:uid", handler.GetUserMsgHistory)
			private.POST("/msg/pintop/:uid/:flag", handler.SetMessageTop)
			private.POST("/msg/clear/:uid", handler.ClearChatHistory)

			// 文件上传模块（上传后通过 WebSocket 发送消息）
			private.POST("/upload/image", handler.UploadImage)
			private.POST("/upload/audio", handler.UploadAudio)
			private.POST("/upload/video", handler.UploadVideo)
			private.POST("/upload/file", handler.UploadFile)

			// 动态模块（点赞/评论/举报改用 WebSocket）
			private.GET("/moment/latest", handler.GetLatestMoment)
			private.GET("/moment/user/:uid/latest", handler.GetUserMoment)
			private.GET("/moment/:mid/comments", handler.GetMomentComments)

			// 钻石模块
			private.POST("/diamond/buy/:did", handler.BuyDiamond)
			private.GET("/diamond/stock", handler.GetDiamondStock)
			private.GET("/diamond/history", handler.GetDiamondHistory)

			// 会员模块
			private.POST("/member/buy/:vid", handler.BuyMember)
			private.GET("/member/history", handler.GetMemberHistory)
		}
	}

	r.Static("/uploads", "./uploads")

	// ==================== WebSocket模块 ====================
	r.GET("/ws", handler.WebSocketHandler)

	return r
}
