package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PageData 分页数据结构
type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// PageResponse 分页响应
func PageResponse(c *gin.Context, message string, list interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
		Data: PageData{
			List:  list,
			Total: total,
			Page:  page,
			Size:  size,
		},
	})
}

// SuccessWithPagination 成功分页响应（别名）
func SuccessWithPagination(c *gin.Context, message string, list interface{}, total int64, page, size int) {
	PageResponse(c, message, list, total, page, size)
}

// 新增：不需要gin.Context的响应函数（用于返回值）

// NewSuccessResponse 创建成功响应（不需要gin.Context）
func NewSuccessResponse(message string, data interface{}) Response {
	return Response{
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应（不需要gin.Context）
func NewErrorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// NewPageResponse 创建分页响应（不需要gin.Context）
func NewPageResponse(message string, list interface{}, total int64, page, size int) Response {
	return Response{
		Code:    200,
		Message: message,
		Data: PageData{
			List:  list,
			Total: total,
			Page:  page,
			Size:  size,
		},
	}
}
