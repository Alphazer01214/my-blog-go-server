package response

import (
	"errors"
	"net/http"

	"blog.alphazer01214.top/internal/constant"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func Result(c *gin.Context, code int, data interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

// ResultWithStatus 返回指定 HTTP 状态码的响应
func ResultWithStatus(c *gin.Context, httpStatus int, code int, data interface{}, msg string) {
	c.JSON(httpStatus, Response{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

func Success(c *gin.Context) {
	Result(c, constant.SUCCESS, map[string]interface{}{}, "Success")
}

func SuccessWithMsg(c *gin.Context, msg string) {
	Result(c, constant.SUCCESS, map[string]interface{}{}, msg)
}

func SuccessWithDetail(c *gin.Context, det interface{}, msg string) {
	Result(c, constant.SUCCESS, det, msg)
}

func Error(c *gin.Context) {
	Result(c, constant.ERROR, map[string]interface{}{}, "Error")
}

func ErrorWithMsg(c *gin.Context, msg string) {
	Result(c, constant.ERROR, map[string]interface{}{}, msg)
}

func ErrorWithDetail(c *gin.Context, det interface{}, msg string) {
	Result(c, constant.ERROR, det, msg)
}

func ErrorAuth(c *gin.Context, msg string) {
	Result(c, constant.ERROR, gin.H{
		"reload": true,
	}, msg)
}

// ErrorWithAppError 返回结构化业务错误
func ErrorWithAppError(c *gin.Context, err error) {
	var appErr *constant.AppError
	if errors.As(err, &appErr) {
		if appErr.HTTPStatus > 0 && appErr.HTTPStatus != http.StatusOK {
			ResultWithStatus(c, appErr.HTTPStatus, appErr.Code, map[string]interface{}{}, appErr.Message)
		} else {
			Result(c, appErr.Code, map[string]interface{}{}, appErr.Message)
		}
		return
	}
	// 非 AppError，返回通用错误
	Result(c, constant.ERROR, map[string]interface{}{}, err.Error())
}
