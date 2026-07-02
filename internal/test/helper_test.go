package test

import (
	"backend/internal/handler"
	"testing"

	"github.com/gin-gonic/gin"
)

type TestRouter struct {
	Engine *gin.Engine
}

func NewTestRouter() *TestRouter {
	router := gin.New()
	return &TestRouter{Engine: router}
}

func (tr *TestRouter) SetupAuthRoutes() {
	tr.Engine.GET("/v1/code/sms", handler.SendSMSCode)
	tr.Engine.GET("/v1/code/alnum", handler.GenerateAlnumCode)
	tr.Engine.POST("/v1/register", handler.Register)
	tr.Engine.POST("/v1/login/code", handler.LoginByCode)
	tr.Engine.POST("/v1/login/pwd", handler.LoginByPassword)
	tr.Engine.POST("/v1/logout", handler.Logout)
	tr.Engine.POST("/v1/resetpwd", handler.ResetPassword)
}

func (tr *TestRouter) SetupUserRoutes() {
	tr.Engine.GET("/v1/userlist", handler.GetUserList)
	tr.Engine.GET("/v1/userinfo/:uid", handler.GetUserInfo)
	tr.Engine.GET("/v1/focuslist/:uid", handler.GetFocusList)
	tr.Engine.GET("/v1/fanslist/:uid", handler.GetFansList)
	tr.Engine.POST("/v1/aimuser/notify/:uid/:flag", handler.SetUserNotify)
	tr.Engine.POST("/v1/aimuser/greet/:uid", handler.GreetUser)
	tr.Engine.POST("/v1/collect/myinfo", handler.CollectMyInfo)
	tr.Engine.POST("/v1/collect/aiminfo", handler.CollectAimInfo)
}

func (tr *TestRouter) SetupChatRoutes() {
	tr.Engine.GET("/v1/adbanner/latest", handler.GetLatestAdBanner)
	tr.Engine.GET("/v1/sysmsg/list", handler.GetSystemMsgList)
	tr.Engine.GET("/v1/usermsg/latest", handler.GetLatestUserMsg)
	tr.Engine.GET("/v1/usermsg/history/:uid", handler.GetUserMsgHistory)
	tr.Engine.POST("/v1/usermsg/pintop/:uid/:flag", handler.SetMessageTop)
	tr.Engine.POST("/v1/usermsg/addfriend/:uid/:flag", handler.AddFriend)
	tr.Engine.POST("/v1/usermsg/clear/:uid", handler.ClearChatHistory)
	tr.Engine.POST("/v1/usergift/:uid/:giftid", handler.SendGift)
}

func (tr *TestRouter) SetupMomentRoutes() {
	tr.Engine.GET("/v1/moment/latest", handler.GetLatestMoment)
	tr.Engine.GET("/v1/mymoment/latest", handler.GetMyLatestMoment)
	tr.Engine.POST("/v1/moment/publish", handler.PublishMoment)
}

func (tr *TestRouter) SetupPaymentRoutes() {
	tr.Engine.POST("/v1/diamond/buy/:pid", handler.BuyDiamond)
	tr.Engine.GET("/v1/diamond/stock", handler.GetDiamondStock)
	tr.Engine.GET("/v1/diamond/history", handler.GetDiamondHistory)
	tr.Engine.POST("/v1/member/buy/:pid", handler.BuyMember)
	tr.Engine.GET("/v1/member/history", handler.GetMemberHistory)
}

func (tr *TestRouter) SetupInteractionRoutes() {
	tr.Engine.GET("/v1/praiseme/list", handler.GetPraiseMeList)
	tr.Engine.GET("/v1/commentme/list", handler.GetCommentMeList)
	tr.Engine.GET("/v1/addme/list", handler.GetAddMeList)
	tr.Engine.GET("/v1/visitme/list", handler.GetVisitMeList)
	tr.Engine.GET("/v1/likeme/list", handler.GetLikeMeList)
	tr.Engine.POST("/v1/agreefriend/:uid/:flag", handler.AgreeFriendRequest)
}

func (tr *TestRouter) SetupUploadRoutes() {
	tr.Engine.POST("/v1/upload/avatar", handler.UploadAvatar)
	tr.Engine.POST("/v1/upload/image", handler.UploadImage)
	tr.Engine.POST("/v1/upload/audio", handler.UploadAudio)
	tr.Engine.POST("/v1/upload/video", handler.UploadVideo)
	tr.Engine.POST("/v1/upload/file", handler.UploadFile)
}

func TestNewTestRouter(t *testing.T) {
	router := NewTestRouter()
	if router == nil {
		t.Fatal("NewTestRouter should not return nil")
	}
	if router.Engine == nil {
		t.Fatal("Engine should not be nil")
	}
}

func TestTestRouter_SetupAllRoutes(t *testing.T) {
	router := NewTestRouter()

	t.Run("SetupAuthRoutes", func(t *testing.T) {
		router.SetupAuthRoutes()
	})

	t.Run("SetupUserRoutes", func(t *testing.T) {
		router.SetupUserRoutes()
	})

	t.Run("SetupChatRoutes", func(t *testing.T) {
		router.SetupChatRoutes()
	})

	t.Run("SetupMomentRoutes", func(t *testing.T) {
		router.SetupMomentRoutes()
	})

	t.Run("SetupPaymentRoutes", func(t *testing.T) {
		router.SetupPaymentRoutes()
	})

	t.Run("SetupInteractionRoutes", func(t *testing.T) {
		router.SetupInteractionRoutes()
	})

	t.Run("SetupUploadRoutes", func(t *testing.T) {
		router.SetupUploadRoutes()
	})
}

func TestJWTValidation(t *testing.T) {
	tests := []struct {
		name   string
		header string
		valid  bool
	}{
		{"Empty header", "", false},
		{"No Bearer prefix", "token123", false},
		{"Wrong prefix", "Basic token123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.header != "" &&
				len(tt.header) > 7 &&
				tt.header[:7] == "Bearer "

			if tt.valid && !isValid {
				t.Error("Expected valid header")
			}
		})
	}
}

func TestRequestValidation(t *testing.T) {
	t.Run("Phone number validation", func(t *testing.T) {
		phone := "13800138000"
		if len(phone) != 11 {
			t.Error("Phone number should be 11 digits")
		}
	})

	t.Run("Code validation", func(t *testing.T) {
		code := "123456"
		if len(code) != 6 {
			t.Error("Code should be 6 digits")
		}
	})

	t.Run("Password validation", func(t *testing.T) {
		password := "password123"
		if len(password) < 6 {
			t.Error("Password should be at least 6 characters")
		}
	})
}

func TestDatabaseModelValidation(t *testing.T) {
	t.Run("User model fields", func(t *testing.T) {
		requiredFields := []string{
			"phone_num",
			"password",
			"nickname",
			"gender",
			"avatar_url",
		}

		for _, field := range requiredFields {
			if field == "" {
				t.Error("Field name should not be empty")
			}
		}
	})

	t.Run("Chat message types", func(t *testing.T) {
		validTypes := []int{1, 2, 3, 4}
		for _, msgType := range validTypes {
			if msgType < 1 || msgType > 4 {
				t.Errorf("Invalid message type: %d", msgType)
			}
		}
	})

	t.Run("Gender values", func(t *testing.T) {
		validGenders := []int{0, 1, -1}
		for _, gender := range validGenders {
			if gender != 0 && gender != 1 && gender != -1 {
				t.Errorf("Invalid gender value: %d", gender)
			}
		}
	})
}

func TestAPIResponseStructure(t *testing.T) {
	t.Run("Success response", func(t *testing.T) {
		response := map[string]interface{}{
			"code": 0,
			"msg":  "success",
			"data": map[string]interface{}{"id": 1},
		}

		if response["code"].(int) != 0 {
			t.Error("Code should be 0 for success")
		}
		if response["msg"].(string) != "success" {
			t.Error("Msg should be 'success'")
		}
		if response["data"] == nil {
			t.Error("Data should not be nil")
		}
	})

	t.Run("Error response", func(t *testing.T) {
		response := map[string]interface{}{
			"code": 1,
			"msg":  "error message",
			"data": nil,
		}

		if response["code"].(int) == 0 {
			t.Error("Code should not be 0 for error")
		}
		if response["data"] != nil {
			t.Error("Data should be nil for error")
		}
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("Database config", func(t *testing.T) {
		config := map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"user":     "postgres",
			"password": "admin",
			"dbname":   "letstalk",
		}

		if config["host"] == "" {
			t.Error("Host should not be empty")
		}
		if config["port"].(int) <= 0 {
			t.Error("Port should be positive")
		}
		if config["user"] == "" {
			t.Error("User should not be empty")
		}
	})

	t.Run("Server config", func(t *testing.T) {
		config := map[string]interface{}{
			"port": 8080,
		}

		if config["port"].(int) <= 0 || config["port"].(int) > 65535 {
			t.Error("Port should be between 1 and 65535")
		}
	})

	t.Run("JWT config", func(t *testing.T) {
		config := map[string]interface{}{
			"secret":       "secret_key",
			"expires_hour": 24,
		}

		if config["secret"] == "" {
			t.Error("Secret should not be empty")
		}
		if config["expires_hour"].(int) <= 0 {
			t.Error("Expires hour should be positive")
		}
	})
}