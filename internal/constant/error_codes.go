package constant

import "net/http"

// 业务错误码定义
const (
	// 通用
	CodeSuccess       = 0
	CodeError         = 7
	CodeInvalidParams = 1001
	CodeUnauthorized  = 1002
	CodeForbidden     = 1003
	CodeNotFound      = 1004
	CodeRateLimit     = 1005
	CodeInternal      = 1999

	// 用户相关 2xxx
	CodeUserNotFound    = 2001
	CodeUserExists      = 2002
	CodePasswordWrong   = 2003
	CodeUserBanned      = 2004
	CodePasswordInvalid = 2005
	CodeFollowSelf      = 2006

	// 帖子相关 3xxx
	CodePostNotFound   = 3001
	CodePostPrivate    = 3002
	CodePostForbidOp   = 3003
	CodePostNotOwner   = 3004

	// 评论相关 4xxx
	CodeCommentNotFound  = 4001
	CodeCommentForbidden = 4002
	CodeCommentNotOwner  = 4003

	// 视频相关 5xxx
	CodeVideoNotFound = 5001

	// AI 相关 6xxx
	CodeAgentNotFound = 6001
	CodeChatNotFound  = 6002

	// 文件相关 7xxx
	CodeFileNotFound = 7001
	CodeUploadFailed = 7002
)

// AppError 业务错误类型
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError 创建业务错误
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusOK, // 默认 200，业务错误通过 code 区分
	}
}

// NewAppErrorWithStatus 创建带 HTTP 状态码的业务错误
func NewAppErrorWithStatus(code int, httpStatus int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// 预定义常用错误
var (
	ErrInvalidParams  = NewAppError(CodeInvalidParams, "invalid parameters")
	ErrUnauthorized   = NewAppErrorWithStatus(CodeUnauthorized, http.StatusUnauthorized, "unauthorized")
	ErrForbidden      = NewAppErrorWithStatus(CodeForbidden, http.StatusForbidden, "forbidden")
	ErrNotFound       = NewAppError(CodeNotFound, "resource not found")
	ErrInternal       = NewAppError(CodeInternal, "internal server error")

	ErrUserNotFound   = NewAppError(CodeUserNotFound, "user not found")
	ErrUserExists     = NewAppError(CodeUserExists, "username already exists")
	ErrPasswordWrong  = NewAppError(CodePasswordWrong, "wrong password")
	ErrUserBanned     = NewAppError(CodeUserBanned, "user is banned")
	ErrPasswordInvalid = NewAppError(CodePasswordInvalid, "password invalid")
	ErrFollowSelf     = NewAppError(CodeFollowSelf, "cannot follow yourself")

	ErrPostNotFound  = NewAppError(CodePostNotFound, "post not found")
	ErrPostNotOwner  = NewAppError(CodePostNotOwner, "not the post owner")
	ErrPostForbidOp  = NewAppError(CodePostForbidOp, "this operation is forbidden on this post")

	ErrCommentNotFound  = NewAppError(CodeCommentNotFound, "comment not found")
	ErrCommentNotOwner  = NewAppError(CodeCommentNotOwner, "not the comment owner")
	ErrCommentForbidden = NewAppError(CodeCommentForbidden, "commenting is forbidden")

	ErrVideoNotFound = NewAppError(CodeVideoNotFound, "video not found")

	ErrAgentNotFound = NewAppError(CodeAgentNotFound, "agent not found")
	ErrChatNotFound  = NewAppError(CodeChatNotFound, "chat session not found")

	ErrFileNotFound  = NewAppError(CodeFileNotFound, "file not found")
	ErrUploadFailed  = NewAppError(CodeUploadFailed, "upload failed")
)
