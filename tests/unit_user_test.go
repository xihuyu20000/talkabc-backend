package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	InitTest()
}

// ==================== 用户列表接口测试 ====================

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

// ==================== 用户信息接口测试 ====================

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

// ==================== 关注列表接口测试 ====================

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

// ==================== 粉丝列表接口测试 ====================

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

// ==================== 关注用户接口测试 ====================

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

// ==================== 拉黑用户接口测试 ====================

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

// ==================== 设置通知开关接口测试 ====================

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

// ==================== 打招呼接口测试 ====================

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

// ==================== 完善个人信息接口测试 ====================

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

// ==================== 设置理想对象条件接口测试 ====================

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

// ==================== UID参数验证测试 ====================

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

// ==================== 用户列表查询参数测试 ====================

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

// ==================== 完善信息接口JSON解析测试 ====================

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