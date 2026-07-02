package response

import (
	"net/http" // Go标准库HTTP包

	"github.com/gin-gonic/gin" // Gin框架，用于Web服务
)

const (
	Code0   = 0   // 成功
	Code400 = 400 // 请求参数错误
	Code401 = 401 // 未授权，需要登录
	Code403 = 403 // 禁止访问
	Code404 = 404 // 资源不存在
	Code500 = 500 // 服务器内部错误
)

// Response 统一API响应结构
// 所有API接口都使用此结构返回数据
type Response struct {
	Code int         `json:"code"` // 状态码：0-成功，其他-失败
	Msg  string      `json:"msg"`  // 响应消息
	Data interface{} `json:"data"` // 响应数据
}

// Success 返回成功响应
// 参数说明：
//   - c: Gin上下文对象，包含请求和响应信息
//   - data: 要返回的数据，可以是任意类型
//
// 使用示例：
//   response.Success(c, gin.H{"id": 1, "name": "test"})
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: Code0,      // 成功
		Msg:  "success",        // 默认成功消息
		Data: data,             // 返回的数据
	})
}

// SuccessMsg 返回带自定义消息的成功响应
// 参数说明：
//   - c: Gin上下文对象
//   - msg: 自定义成功消息
//   - data: 要返回的数据
//
// 使用示例：
//   response.SuccessMsg(c, "登录成功", gin.H{"token": "xxx"})
func SuccessMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: Code0,       // 成功
		Msg:  msg,     // 自定义消息
		Data: data,    // 返回的数据
	})
}

// Error 返回错误响应
// 参数说明：
//   - c: Gin上下文对象
//   - code: 错误码（非0值）
//   - msg: 错误消息
//
// 使用示例：
//   response.Error(c, 1001, "用户不存在")
func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,              // 自定义错误码
		Msg:  msg,               // 错误消息
		Data: nil,               // 错误响应不返回数据
	})
}

// BadRequest 返回400错误（参数错误）
// 参数说明：
//   - c: Gin上下文对象
//   - msg: 错误消息
//
// 使用示例：
//   response.BadRequest(c, "参数不能为空")
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Response{
		Code: Code400,           // 请求参数错误
		Msg:  msg,
		Data: nil,
	})
}

// Unauthorized 返回401错误（未授权）
// 参数说明：
//   - c: Gin上下文对象
//   - msg: 错误消息
//
// 使用示例：
//   response.Unauthorized(c, "请先登录")
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code: Code401,           // 未授权，需要登录
		Msg:  msg,
		Data: nil,
	})
}

// Forbidden 返回403错误（禁止访问）
// 参数说明：
//   - c: Gin上下文对象
//   - msg: 错误消息
//
// 使用示例：
//   response.Forbidden(c, "您没有权限访问此资源")
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Response{
		Code: Code403,           // 禁止访问
		Msg:  msg,
		Data: nil,
	})
}

// NotFound 返回404错误（资源不存在）
// 参数说明：
//   - c: Gin上下文对象
//   - msg: 错误消息
//
// 使用示例：
//   response.NotFound(c, "用户不存在")
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Response{
		Code: Code404,           // 资源不存在
		Msg:  msg,
		Data: nil,
	})
}

// InternalError 返回500错误（服务器内部错误）
// 参数说明：
//   - c: Gin上下文对象
//   - msg: 错误消息
//
// 使用示例：
//   response.InternalError(c, "服务器繁忙，请稍后重试")
func InternalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code: Code500,           // 服务器内部错误
		Msg:  msg,
		Data: nil,
	})
}
