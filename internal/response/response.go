package response

import (
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
