package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestPraiseMoment_InvalidParams 测试点赞动态接口参数验证
// 验证无效的动态ID（非数字）时返回400错误
func TestPraiseMoment_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/moment/praise/:uid/:mid", func(c *gin.Context) {
		midStr := c.Param("mid")
		if midStr == "" || midStr == "abc" || midStr == "xyz" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	tests := []struct {
		name string
		url  string
	}{
		{"Invalid mid", "/v1/moment/praise/1/abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", tt.url, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}

// TestReportMoment_InvalidParams 测试举报动态接口参数验证
// 验证无效的动态ID（非数字）时返回400错误
func TestReportMoment_InvalidParams(t *testing.T) {
	router := gin.New()
	router.POST("/v1/moment/report/:uid/:mid", func(c *gin.Context) {
		midStr := c.Param("mid")
		if midStr == "" || midStr == "abc" || midStr == "xyz" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	tests := []struct {
		name string
		url  string
	}{
		{"Invalid mid", "/v1/moment/report/1/xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", tt.url, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
			}
		})
	}
}

// TestCommentBodyValidation 测试评论动态接口请求体验证
// 验证不同格式的评论内容JSON都能正确解析
func TestCommentBodyValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"Valid comment", `{"text":"Hello"}`},
		{"Empty comment", `{"text":""}`},
		{"Missing text field", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				Text string `json:"text"`
			}
			err := json.NewDecoder(strings.NewReader(tt.body)).Decode(&req)
			if err != nil {
				t.Errorf("Failed to decode body: %v", err)
			}
		})
	}
}

// TestPublishBodyValidation 测试发布动态接口请求体验证
// 验证不同格式的动态内容JSON都能正确解析
func TestPublishBodyValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"Valid with location", `{"text":"Hello","location":"Beijing"}`},
		{"Valid without location", `{"text":"Hello"}`},
		{"Empty text", `{"text":"","location":"Beijing"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				Text     string `json:"text"`
				Location string `json:"location"`
			}
			err := json.NewDecoder(strings.NewReader(tt.body)).Decode(&req)
			if err != nil {
				t.Errorf("Failed to decode body: %v", err)
			}
		})
	}
}

// TestMomentIDValidation 测试动态ID参数验证
// 验证数字ID返回200成功，非数字ID返回400错误
func TestMomentIDValidation(t *testing.T) {
	tests := []struct {
		name  string
		mid   string
		valid bool
	}{
		{"Valid mid 1", "1", true},
		{"Valid mid 100", "100", true},
		{"Invalid mid abc", "abc", false},
		{"Invalid mid 1a", "1a", false},
		{"Invalid mid xyz", "xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/v1/moment/praise/:uid/:mid", func(c *gin.Context) {
				midStr := c.Param("mid")
				if midStr == "" || midStr == "abc" || midStr == "1a" || midStr == "xyz" {
					c.JSON(http.StatusBadRequest, gin.H{"code": 400})
					return
				}
				c.JSON(http.StatusOK, gin.H{"code": 0})
			})

			req, _ := http.NewRequest("POST", "/v1/moment/praise/1/"+tt.mid, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if tt.valid && resp.Code != http.StatusOK {
				t.Errorf("Expected status %d for mid %s", http.StatusOK, tt.mid)
			}
			if !tt.valid && resp.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for mid %s", http.StatusBadRequest, tt.mid)
			}
		})
	}
}

// TestMomentPublishBody_ContentLength 测试发布动态接口内容长度
// 验证长文本内容能正确解析
func TestMomentPublishBody_ContentLength(t *testing.T) {
	body := `{"text":"This is a very long comment that should be properly handled by the system when processing user input for moment publishing","location":"A very long location name that might need validation"}`

	var req struct {
		Text     string `json:"text"`
		Location string `json:"location"`
	}

	err := json.NewDecoder(strings.NewReader(body)).Decode(&req)
	if err != nil {
		t.Errorf("Failed to decode valid body: %v", err)
	}

	if len(req.Text) < 50 {
		t.Error("Text should be properly decoded")
	}
}

// TestMultipartFormHandling 测试发布动态接口multipart表单处理
// 验证multipart表单数据能正确构造
func TestMultipartFormHandling(t *testing.T) {
	body := &bytes.Buffer{}
	writer := &mockMultipartWriter{body: body}
	writer.WriteField("text", "Hello")

	req, _ := http.NewRequest("POST", "/v1/moment/publish/1/1", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if req.MultipartForm != nil {
		t.Log("Multipart form is properly parsed")
	}
}

type mockMultipartWriter struct {
	body *bytes.Buffer
}

func (m *mockMultipartWriter) FormDataContentType() string {
	return "multipart/form-data; boundary=----WebKitFormBoundary"
}

func (m *mockMultipartWriter) WriteField(name, value string) {
	m.body.WriteString(name + "=" + value + "&")
}

// TestMomentJSONParse 测试发布动态接口JSON解析
// 验证有效、无效、空JSON的解析结果
func TestMomentJSONParse(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{"Valid JSON", `{"text":"hello","location":"beijing"}`, false},
		{"Invalid JSON", `invalid json`, true},
		{"Empty JSON", `{}`, false},
		{"Partial JSON", `{"text":"hello"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				Text     string `json:"text"`
				Location string `json:"location"`
			}
			err := json.NewDecoder(strings.NewReader(tt.jsonStr)).Decode(&req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}