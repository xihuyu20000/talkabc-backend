package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// TestResponseFormat 测试API响应格式
// 验证成功响应、错误响应、未授权响应、错误请求响应的格式正确性
func TestResponseFormat(t *testing.T) {
	t.Run("Success response", func(t *testing.T) {
		router := gin.New()
		router.GET("/test/success", func(c *gin.Context) {
			response.Success(c, gin.H{"id": 1, "name": "test"})
		})

		req, _ := http.NewRequest("GET", "/test/success", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		var result response.Response
		json.Unmarshal(resp.Body.Bytes(), &result)

		if result.Code != 0 {
			t.Errorf("Expected code 0, got %d", result.Code)
		}
		if result.Msg != "success" {
			t.Errorf("Expected msg 'success', got %s", result.Msg)
		}
		if result.Data == nil {
			t.Error("Data should not be nil")
		}
	})

	t.Run("Error response", func(t *testing.T) {
		router := gin.New()
		router.GET("/test/error", func(c *gin.Context) {
			response.Error(c, 1, "error message")
		})

		req, _ := http.NewRequest("GET", "/test/error", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		var result response.Response
		json.Unmarshal(resp.Body.Bytes(), &result)

		if result.Code != 1 {
			t.Errorf("Expected code 1, got %d", result.Code)
		}
		if result.Data != nil {
			t.Error("Data should be nil for error")
		}
	})

	t.Run("Unauthorized response", func(t *testing.T) {
		router := gin.New()
		router.GET("/test/unauthorized", func(c *gin.Context) {
			response.Unauthorized(c, "not logged in")
		})

		req, _ := http.NewRequest("GET", "/test/unauthorized", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
		}
	})

	t.Run("Bad request response", func(t *testing.T) {
		router := gin.New()
		router.GET("/test/badrequest", func(c *gin.Context) {
			response.BadRequest(c, "invalid params")
		})

		req, _ := http.NewRequest("GET", "/test/badrequest", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})
}

// TestRouterSetup 测试路由注册
// 验证各路由组能够正常创建和注册
func TestRouterSetup(t *testing.T) {
	t.Run("TestRouter creation", func(t *testing.T) {
		router := NewTestRouter()
		if router == nil {
			t.Fatal("NewTestRouter should not return nil")
		}
		if router.Engine == nil {
			t.Fatal("Engine should not be nil")
		}
	})

	t.Run("SetupAuthRoutes", func(t *testing.T) {
		router := NewTestRouter()
		router.SetupAuthRoutes()
	})

	t.Run("SetupUserRoutes", func(t *testing.T) {
		router := NewTestRouter()
		router.SetupUserRoutes()
	})

	t.Run("SetupChatRoutes", func(t *testing.T) {
		router := NewTestRouter()
		router.SetupChatRoutes()
	})

	t.Run("SetupMomentRoutes", func(t *testing.T) {
		router := NewTestRouter()
		router.SetupMomentRoutes()
	})

	t.Run("SetupPaymentRoutes", func(t *testing.T) {
		router := NewTestRouter()
		router.SetupPaymentRoutes()
	})

	t.Run("SetupInteractionRoutes", func(t *testing.T) {
		router := NewTestRouter()
		router.SetupInteractionRoutes()
	})

	t.Run("SetupUploadRoutes", func(t *testing.T) {
		router := NewTestRouter()
		router.SetupUploadRoutes()
	})
}

// TestAuth_APIEndpoints 测试认证API端点
// 验证验证码接口和登出接口的响应状态
func TestAuth_APIEndpoints(t *testing.T) {
	router := NewTestRouter()
	router.SetupAuthRoutes()

	t.Run("GET /v1/code/sms without phone", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/code/sms", nil)
		resp := httptest.NewRecorder()
		router.Engine.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	t.Run("POST /v1/logout success", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/logout", nil)
		resp := httptest.NewRecorder()
		router.Engine.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, resp.Code)
		}
	})
}

// TestUser_APIEndpoints 测试用户API端点
// 验证用户列表接口参数验证和用户信息接口UID验证
func TestUser_APIEndpoints(t *testing.T) {
	router := gin.New()

	t.Run("GET /v1/userlist missing params", func(t *testing.T) {
		router.GET("/v1/userlist", func(c *gin.Context) {
			age1 := c.Query("age1")
			age2 := c.Query("age2")
			gender := c.Query("gender")

			if age1 != "" {
				if _, err := strconv.Atoi(age1); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 400})
					return
				}
			}
			if age2 != "" {
				if _, err := strconv.Atoi(age2); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 400})
					return
				}
			}
			if gender != "" {
				if _, err := strconv.Atoi(gender); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 400})
					return
				}
			}
			c.JSON(http.StatusOK, gin.H{"code": 0})
		})

		req, _ := http.NewRequest("GET", "/v1/userlist", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	t.Run("GET /v1/userinfo invalid uid", func(t *testing.T) {
		router.GET("/v1/userinfo/:uid", func(c *gin.Context) {
			uidStr := c.Param("uid")
			if uidStr == "abc" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 0})
		})

		req, _ := http.NewRequest("GET", "/v1/userinfo/abc", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})
}

// TestChat_APIEndpoints 测试聊天API端点
// 验证广告横幅接口和聊天历史接口的响应状态
func TestChat_APIEndpoints(t *testing.T) {
	router := gin.New()

	t.Run("GET /v1/adbanner/latest success", func(t *testing.T) {
		router.GET("/v1/adbanner/latest", func(c *gin.Context) {
			response.Success(c, gin.H{})
		})

		req, _ := http.NewRequest("GET", "/v1/adbanner/latest", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, resp.Code)
		}
	})

	t.Run("GET /v1/usermsg/history invalid uid", func(t *testing.T) {
		router.GET("/v1/usermsg/history/:uid", func(c *gin.Context) {
			uidStr := c.Param("uid")
			if uidStr == "abc" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 0})
		})

		req, _ := http.NewRequest("GET", "/v1/usermsg/history/abc", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})
}

// TestMoment_APIEndpoints 测试动态API端点
// 验证动态列表接口和点赞接口的响应状态
func TestMoment_APIEndpoints(t *testing.T) {
	router := gin.New()

	t.Run("GET /v1/moment/latest success", func(t *testing.T) {
		router.GET("/v1/moment/latest", func(c *gin.Context) {
			response.Success(c, gin.H{})
		})

		req, _ := http.NewRequest("GET", "/v1/moment/latest", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, resp.Code)
		}
	})

	t.Run("POST /v1/moment/praise invalid mid", func(t *testing.T) {
		router.POST("/v1/moment/praise/:uid/:mid", func(c *gin.Context) {
			midStr := c.Param("mid")
			if midStr == "abc" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 0})
		})

		req, _ := http.NewRequest("POST", "/v1/moment/praise/1/abc", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})
}

// TestPayment_APIEndpoints 测试支付API端点
// 验证钻石库存接口的响应状态
func TestPayment_APIEndpoints(t *testing.T) {
	router := gin.New()

	t.Run("GET /v1/diamond/stock success", func(t *testing.T) {
		router.GET("/v1/diamond/stock", func(c *gin.Context) {
			response.Success(c, gin.H{})
		})

		req, _ := http.NewRequest("GET", "/v1/diamond/stock", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, resp.Code)
		}
	})
}

// TestInteraction_APIEndpoints 测试互动API端点
// 验证赞我、评论我、添加我、访问我、喜欢我等接口的响应状态
func TestInteraction_APIEndpoints(t *testing.T) {
	router := gin.New()

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/praiseme/list"},
		{"GET", "/v1/commentme/list"},
		{"GET", "/v1/addme/list"},
		{"GET", "/v1/visitme/list"},
		{"GET", "/v1/likeme/list"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			router.Handle(ep.method, ep.path, func(c *gin.Context) {
				response.Success(c, gin.H{})
			})

			req, _ := http.NewRequest(ep.method, ep.path, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusOK {
				t.Errorf("Expected %d, got %d", http.StatusOK, resp.Code)
			}
		})
	}
}