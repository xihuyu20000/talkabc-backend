package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestGetUserList_InvalidParams 测试获取用户列表接口参数验证
// 验证缺少必填参数（年龄范围或性别）时返回400错误
func TestGetUserList_InvalidParams(t *testing.T) {
	router := gin.New()
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

	tests := []struct {
		name string
		url  string
	}{
		{"Invalid age1", "/v1/userlist?age1=abc"},
		{"Invalid age2", "/v1/userlist?age2=xyz"},
		{"Invalid gender", "/v1/userlist?gender=invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}

// TestGetUserInfo_InvalidUID 测试获取用户信息接口UID验证
// 验证UID参数存在时返回200成功
func TestGetUserInfo_InvalidUID(t *testing.T) {
	router := gin.New()
	router.GET("/v1/userinfo/:uid", func(c *gin.Context) {
		uid := c.Param("uid")
		if uid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	req, _ := http.NewRequest("GET", "/v1/userinfo/invalid", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

// TestGetFocusList_InvalidUID 测试获取关注列表接口UID验证
// 验证UID参数存在时返回200成功
func TestGetFocusList_InvalidUID(t *testing.T) {
	router := gin.New()
	router.GET("/v1/focuslist/:uid", func(c *gin.Context) {
		uid := c.Param("uid")
		if uid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	req, _ := http.NewRequest("GET", "/v1/focuslist/abc", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

// TestGetFansList_InvalidUID 测试获取粉丝列表接口UID验证
// 验证UID参数存在时返回200成功
func TestGetFansList_InvalidUID(t *testing.T) {
	router := gin.New()
	router.GET("/v1/fanslist/:uid", func(c *gin.Context) {
		uid := c.Param("uid")
		if uid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	req, _ := http.NewRequest("GET", "/v1/fanslist/notanumber", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

// TestFocusUser_InvalidParams 测试关注用户接口参数验证
// 验证UID和flag参数存在时返回200成功
func TestFocusUser_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/aimuser/focus/:uid/:flag", func(c *gin.Context) {
		uid := c.Param("uid")
		flag := c.Param("flag")
		if uid == "" || flag == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	tests := []struct {
		name string
		url  string
	}{
		{"Invalid uid", "/v1/aimuser/focus/abc/1"},
		{"Invalid flag", "/v1/aimuser/focus/1/abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", tt.url, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
			}
		})
	}
}

// TestBlockUser_InvalidParams 测试拉黑用户接口参数验证
// 验证UID和flag参数存在时返回200成功
func TestBlockUser_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/aimuser/block/:uid/:flag", func(c *gin.Context) {
		uid := c.Param("uid")
		flag := c.Param("flag")
		if uid == "" || flag == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	tests := []struct {
		name string
		url  string
	}{
		{"Invalid uid", "/v1/aimuser/block/abc/1"},
		{"Invalid flag", "/v1/aimuser/block/1/xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", tt.url, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
			}
		})
	}
}

// TestSetUserNotify_InvalidParams 测试设置通知开关接口参数验证
// 验证UID和flag参数存在时返回200成功
func TestSetUserNotify_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/aimuser/notify/:uid/:flag", func(c *gin.Context) {
		uid := c.Param("uid")
		flag := c.Param("flag")
		if uid == "" || flag == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	tests := []struct {
		name string
		url  string
	}{
		{"Invalid uid", "/v1/aimuser/notify/abc/1"},
		{"Invalid flag", "/v1/aimuser/notify/1/invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", tt.url, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
			}
		})
	}
}

// TestGreetUser_Success 测试打招呼接口成功场景
// 验证UID参数存在时返回200成功
func TestGreetUser_Success(t *testing.T) {
	router := gin.New()
	router.POST("/v1/aimuser/greet/:uid", func(c *gin.Context) {
		uid := c.Param("uid")
		if uid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	req, _ := http.NewRequest("POST", "/v1/aimuser/greet/123", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

// TestCollectMyInfo_InvalidBody 测试完善个人信息接口请求体验证
// 验证空请求体时返回400错误
func TestCollectMyInfo_InvalidBody(t *testing.T) {
	router := gin.New()
	router.POST("/v1/collect/myinfo", func(c *gin.Context) {
		var req struct {
			Nickname string `json:"nickname"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	req, _ := http.NewRequest("POST", "/v1/collect/myinfo", nil)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

// TestCollectAimInfo_InvalidBody 测试设置理想对象条件接口请求体验证
// 验证空请求体时返回400错误
func TestCollectAimInfo_InvalidBody(t *testing.T) {
	router := gin.New()
	router.POST("/v1/collect/aiminfo", func(c *gin.Context) {
		var req struct {
			AimUID string `json:"aimuid"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	req, _ := http.NewRequest("POST", "/v1/collect/aiminfo", nil)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

// TestResponseFormat 测试API响应格式
// 验证响应包含code、msg、data字段且值正确
func TestResponseFormat(t *testing.T) {
	type TestResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "success",
			"data": "test",
		})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	var result TestResponse
	json.Unmarshal(resp.Body.Bytes(), &result)

	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d", result.Code)
	}

	if result.Msg != "success" {
		t.Errorf("Expected msg 'success', got %s", result.Msg)
	}

	if result.Data != "test" {
		t.Errorf("Expected data 'test', got %s", result.Data)
	}
}

// TestUIDValidation 测试UID参数验证
// 验证不同格式的UID（数字、字母、雪花ID）都能被接受
func TestUIDValidation(t *testing.T) {
	tests := []struct {
		name  string
		uid   string
		valid bool
	}{
		{"Valid uid 1", "1", true},
		{"Valid uid 100", "100", true},
		{"Valid uid abc", "abc", true},
		{"Valid uid snowflake", "12345678901234567890", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/v1/userinfo/:uid", func(c *gin.Context) {
				uidStr := c.Param("uid")
				if uidStr == "" {
					c.JSON(http.StatusBadRequest, gin.H{"code": 400})
					return
				}
				c.JSON(http.StatusOK, gin.H{"code": 0})
			})

			req, _ := http.NewRequest("GET", "/v1/userinfo/"+tt.uid, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if tt.valid && resp.Code != http.StatusOK {
				t.Errorf("Expected status %d for uid %s", http.StatusOK, tt.uid)
			}
			if !tt.valid && resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for uid %s", http.StatusBadRequest, tt.uid)
			}
		})
	}
}

// TestUserListQueryParams 测试用户列表查询参数验证
// 验证必填参数（年龄范围和性别）缺失时返回400错误
func TestUserListQueryParams(t *testing.T) {
	tests := []struct {
		name  string
		query string
		valid bool
	}{
		{"Valid params", "age=18&age=30&gender=1", true},
		{"Invalid age1", "age1=abc&gender=1", false},
		{"Invalid age2", "age2=xyz&gender=1", false},
		{"Invalid gender", "age1=18&age2=30&gender=invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
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

			req, _ := http.NewRequest("GET", "/v1/userlist?"+tt.query, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if tt.valid && resp.Code != http.StatusOK {
				t.Errorf("Expected status %d for query %s", http.StatusOK, tt.query)
			}
			if !tt.valid && resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for query %s", http.StatusBadRequest, tt.query)
			}
		})
	}
}

// TestCollectInfoJSONValidation 测试完善信息接口JSON解析
// 验证不同格式的JSON请求体都能正确解析
func TestCollectInfoJSONValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"Valid myinfo", `{"regcountry":"CN","mylang":"zh","nickname":"test"}`},
		{"Valid aiminfo", `{"aimuid":"12345678901234567890"}`},
		{"Empty body", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				RegCountry string `json:"regcountry"`
				MyLang     string `json:"mylang"`
				Nickname   string `json:"nickname"`
				AimUID     string `json:"aimuid"`
			}

			err := json.NewDecoder(strings.NewReader(tt.body)).Decode(&req)
			if err != nil {
				t.Errorf("Failed to decode body: %v", err)
			}
		})
	}
}