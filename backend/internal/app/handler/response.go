package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一 JSON 响应结构。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondJSON 成功响应。
func RespondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Code:    status,
		Message: "ok",
		Data:    data,
	})
}

// RespondError 错误响应。
func RespondError(c *gin.Context, status int, msg string) {
	c.JSON(status, Response{
		Code:    status,
		Message: msg,
	})
}

// BadRequest 400。
func BadRequest(c *gin.Context, msg string) {
	RespondError(c, http.StatusBadRequest, msg)
}

// Unauthorized 401。
func Unauthorized(c *gin.Context, msg string) {
	RespondError(c, http.StatusUnauthorized, msg)
}

// NotFound 404。
func NotFound(c *gin.Context, msg string) {
	RespondError(c, http.StatusNotFound, msg)
}

// Conflict 409。
func Conflict(c *gin.Context, msg string) {
	RespondError(c, http.StatusConflict, msg)
}

// InternalError 500。
func InternalError(c *gin.Context, msg string) {
	RespondError(c, http.StatusInternalServerError, msg)
}
