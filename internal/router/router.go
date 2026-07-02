package router

import (
	"backend/internal/config"     // 配置模块
	"backend/internal/handler"    // 处理器函数
	"backend/internal/middleware" // 中间件（JWT认证等）

	"github.com/gin-contrib/cors" // CORS跨域资源共享
	"github.com/gin-gonic/gin"    // Gin Web框架
)

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
		public := apiV1.Group("/")
		{
			// ==================== auth & sms ====================
			public.GET("/auth/code-sms", handler.SendSMSCode)
			public.GET("/auth/code-alnum", handler.GenerateAlnumCode)
			public.POST("/auth/register", handler.Register)
			public.POST("/auth/login/code", handler.LoginByCode)
			public.POST("/auth/login/password", handler.LoginByPassword)
			public.POST("/auth/reset-password", handler.ResetPassword)
		}

		private := apiV1.Group("/", middleware.JWT())
		{
			// ==================== auth ====================
			private.POST("/auth/logout", handler.Logout)

			// ==================== users ====================
			private.GET("/users", handler.GetUserList)
			private.GET("/users/:uid", handler.GetUserInfo)
			private.GET("/users/:uid/following", handler.GetFocusList)
			private.GET("/users/:uid/fans", handler.GetFansList)
			private.POST("/users/:uid/greet", handler.GreetUser)
			private.POST("/users/:uid/notification/:flag", handler.SetUserNotify)

			// ==================== profile ====================
			private.POST("/profile/me", handler.CollectMyInfo)
			private.POST("/profile/preferences", handler.CollectAimInfo)

			// ==================== uploads ====================
			private.POST("/users/avatar", handler.UploadAvatar)
			private.POST("/uploads/image", handler.UploadImage)
			private.POST("/uploads/audio", handler.UploadAudio)
			private.POST("/uploads/video", handler.UploadVideo)
			private.POST("/uploads/file", handler.UploadFile)

			// ==================== notifications ====================
			private.GET("/notifications/praise", handler.GetPraiseMeList)
			private.GET("/notifications/comment", handler.GetCommentMeList)
			private.GET("/notifications/friend", handler.GetAddMeList)
			private.GET("/notifications/visit", handler.GetVisitMeList)
			private.GET("/notifications/like", handler.GetLikeMeList)

			// ==================== friendships ====================
			private.POST("/friendships/:uid/:flag", handler.AddFriend)
			private.POST("/friendships/agree/:uid/:flag", handler.AgreeFriendRequest)

			// ==================== ads ====================
			private.GET("/ads/latest", handler.GetLatestAdBanner)

			// ==================== gifts ====================
			private.POST("/gifts/send/:uid/:giftid", handler.SendGift)

			// ==================== messages ====================
			private.GET("/messages/system", handler.GetSystemMsgList)
			private.GET("/messages/latest", handler.GetLatestUserMsg)
			private.GET("/messages/:uid", handler.GetUserMsgHistory)
			private.POST("/messages/top/:uid/:flag", handler.SetMessageTop)
			private.DELETE("/messages/:uid", handler.ClearChatHistory)

			// ==================== moments ====================
			private.GET("/moments/latest", handler.GetLatestMoment)
			private.GET("/users/:uid/moments", handler.GetUserMoment)
			private.GET("/moments/:mid/comments", handler.GetMomentComments)

			// ==================== diamonds ====================
			private.POST("/diamonds/buy/:did", handler.BuyDiamond)
			private.GET("/diamonds/stock", handler.GetDiamondStock)
			private.GET("/diamonds/history", handler.GetDiamondHistory)

			// ==================== memberships ====================
			private.POST("/memberships/buy/:vid", handler.BuyMember)
			private.GET("/memberships/history", handler.GetMemberHistory)
		}
	}

	r.Static("/uploads", "./uploads")

	// ==================== websocket ====================
	r.GET("/ws", handler.WebSocketHandler)

	return r
}