package router

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"
	_ "backend/swagger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	r.Use(middleware.RequestLogger())
	r.Use(cors.New(initCORSConfig()))

	apiV1 := r.Group("/api/v1")
	{
		public := apiV1.Group("/")
		{
			// ===== 认证模块（公开）=====
			public.GET("/auth/code-sms", handler.SendSMSCode)
			public.POST("/auth/code-sms/verify", handler.VerifySMSCode)
			public.GET("/auth/code-alnum", handler.GenerateAlnumCode)
			public.POST("/auth/code-alnum/verify", handler.VerifyAlnumCode)
			public.POST("/auth/register", handler.Register)
			public.POST("/auth/login/code", handler.LoginByCode)
			public.POST("/auth/login/password", handler.LoginByPassword)
			public.POST("/auth/login/oauth", handler.OAuthLogin)
			public.POST("/auth/refresh-token", handler.RefreshToken)
			public.POST("/auth/reset-password/initiate", handler.InitiateResetPassword)
			public.GET("/auth/reset-password/validate", handler.ValidateResetToken)
			public.POST("/auth/reset-password/complete", handler.CompleteResetPassword)

			// ===== 广告模块（公开）=====
			public.GET("/ad/list", handler.GetAdList)
		}

		private := apiV1.Group("/", middleware.JWT())
		{
			// ===== 系统模块 =====
			private.GET("/system/log-level", handler.GetLogLevel)
			private.POST("/system/log-level", handler.SetLogLevel)

			// ===== 认证模块（私有）=====
			private.POST("/auth/logout", handler.Logout)
			private.POST("/auth/change-phone", handler.ChangePhone)
			private.POST("/auth/oauth/bind", handler.OAuthBind)
			private.POST("/auth/oauth/unbind", handler.OAuthUnbind)
			private.GET("/auth/oauth/list", handler.GetOAuthBindings)

			// ===== 用户模块 =====
			private.GET("/users", handler.GetUserList)
			private.GET("/users/me", handler.GetUserMe)
			private.PUT("/users/me", handler.UpdateUserMe)
			private.GET("/users/:uid", handler.GetUserInfo)
			private.POST("/users/:uid/follow", handler.FollowUser)
			private.DELETE("/users/:uid/follow", handler.UnfollowUser)
			private.GET("/users/:uid/follow/status", handler.CheckFollowStatus)
			private.GET("/users/:uid/following", handler.GetFocusList)
			private.GET("/users/:uid/fans", handler.GetFansList)
			private.POST("/users/:uid/greet", handler.GreetUser)
			private.POST("/users/:uid/notification/:flag", handler.SetUserNotify)
			private.GET("/users/:uid/online", handler.GetUserOnlineStatus)
			private.GET("/users/search", handler.SearchUsers)

			// ===== 用户资料模块 =====
			private.POST("/profile/me", handler.CollectMyInfo)
			private.POST("/profile/preferences", handler.CollectAimInfo)
			private.GET("/profile/status", handler.CheckProfileStatus)
			private.POST("/profile/sign", handler.SetSignText)
			private.POST("/profile/complete", handler.CompleteProfile)

			// ===== 文件上传模块 =====
			private.POST("/users/avatar", handler.UploadAvatar)
			private.POST("/uploads/image", handler.UploadImage)
			private.POST("/uploads/audio", handler.UploadAudio)
			private.POST("/uploads/video", handler.UploadVideo)
			private.POST("/uploads/file", handler.UploadFile)

			// ===== 通知模块 =====
			private.GET("/notifications/praise-me", handler.GetPraiseMeList)
			private.GET("/notifications/comment-me", handler.GetCommentMeList)
			private.GET("/notifications/add-me", handler.GetAddMeList)
			private.GET("/notifications/visit-me", handler.GetVisitMeList)
			private.GET("/notifications/like-me", handler.GetLikeMeList)
			private.POST("/notifications/friend-request/:uid/:flag", handler.AgreeFriendRequest)
			private.GET("/notifications", handler.GetNotificationList)
			private.POST("/notifications/read", handler.MarkNotificationsRead)
			private.GET("/notifications/count", handler.GetNotificationCount)

			// ===== 好友模块 =====
			private.POST("/friends/:uid/focus", handler.AddFriend)
			private.POST("/friends/:uid/notify/:flag", handler.SetUserNotify)
			private.GET("/friends/requests", handler.GetFriendRequests)
			private.POST("/friends/requests/:uid/:flag", handler.AgreeFriendRequest)
			private.POST("/friends/greet/:uid", handler.GreetUser)
			private.GET("/friends", handler.GetFriendList)
			private.DELETE("/friends/:uid", handler.DeleteFriend)

			// ===== 礼物模块 =====
			private.GET("/gifts", handler.GetGiftList)
			private.POST("/gifts/send/:uid/:giftid", handler.SendGift)

			// ===== 消息模块 =====
			private.GET("/messages/system", handler.GetSystemMsgList)
			private.GET("/messages/latest", handler.GetLatestUserMsg)
			private.GET("/messages/users/:uid", handler.GetUserMsgHistory)
			private.POST("/messages/users/:uid", handler.SendUserMessage)
			private.POST("/messages/users/:uid/read", handler.MarkMessagesRead)
			private.POST("/messages/users/:uid/clear", handler.ClearChatHistory)
			private.POST("/messages/users/:uid/top/:flag", handler.SetMessageTop)
			private.DELETE("/messages/:msgid", handler.DeleteMessage)
			private.POST("/messages/:msgid/recall", handler.RecallMessage)

			// ===== 朋友圈模块 =====
			private.GET("/moments", handler.GetLatestMoment)
			private.GET("/moments/following", handler.GetFollowingMoment)
			private.GET("/moments/me", handler.GetMyLatestMoment)
			private.GET("/moments/users/:uid", handler.GetUserMoment)
			private.GET("/moments/:mid", handler.GetMomentDetail)
			private.GET("/moments/:mid/comments", handler.GetMomentComments)
			private.POST("/moments/:mid/comments", handler.CommentMoment)
			private.POST("/moments/:mid/praise", handler.PraiseMoment)
			private.DELETE("/moments/:mid/praise", handler.CancelPraiseMoment)
			private.POST("/moments", handler.PublishMoment)
			private.DELETE("/moments/:mid", handler.DeleteMoment)
			private.DELETE("/moments/comments/:cid", handler.DeleteComment)

			// ===== VIP/充值模块 =====
			private.GET("/vip/info", handler.GetVipInfo)
			private.GET("/vip/products", handler.GetVipProducts)
			private.POST("/vip/buy", handler.BuyVip)
			private.GET("/vip/diamond", handler.GetDiamondStock)
			private.GET("/vip/recharge/products", handler.GetRechargeProducts)
			private.POST("/vip/recharge", handler.CreateRechargeOrder)
			private.POST("/vip/recharge/notify", handler.RechargeNotify)
			private.GET("/vip/recharge/history", handler.GetRechargeHistory)

			// ===== 收益模块 =====
			private.GET("/income/list", handler.GetIncomeList)
			private.GET("/income/total", handler.GetIncomeTotal)
			private.POST("/income/withdraw", handler.CreateWithdraw)
			private.GET("/income/withdraw/history", handler.GetWithdrawHistory)

			// ===== 实名认证模块 =====
			private.GET("/verification/real/status", handler.GetRealVerifyStatus)
			private.POST("/verification/real/apply", handler.ApplyRealVerify)
			private.GET("/verification/official/status", handler.GetOfficialVerifyStatus)
			private.POST("/verification/official/apply", handler.ApplyOfficialVerify)

			// ===== 设置模块 =====
			private.GET("/settings/blacklist", handler.GetBlacklist)
			private.POST("/settings/blacklist/:uid", handler.ToggleBlacklist)
			private.GET("/settings/language", handler.GetLanguage)
			private.PUT("/settings/language", handler.UpdateLanguage)
			private.GET("/settings/privacy", handler.GetPrivacySettings)
			private.PUT("/settings/privacy", handler.UpdatePrivacySettings)

			// ===== 分享模块 =====
			private.POST("/share/generate", handler.GenerateShareLink)
			private.GET("/share/invite/reward", handler.GetInviteReward)

			// ===== 广告模块（私有）=====
			private.POST("/ad/click", handler.TrackAdClick)
			private.POST("/ad/impression", handler.TrackAdImpression)

			// ===== WebSocket模块 =====
			private.GET("/websocket/connect", handler.WebSocketConnect)
			private.POST("/websocket/ping", handler.WebSocketPing)
		}
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	r.Static("/uploads", "./uploads")

	r.GET("/ws", handler.WebSocketHandler)

	r.Any("/api/v1/api/v1/*path", func(c *gin.Context) {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "请求路径错误：URL 前缀重复。请检查客户端配置，移除 baseURL 中的 /api/v1 前缀或请求路径中的重复前缀",
		})
	})

	return r
}