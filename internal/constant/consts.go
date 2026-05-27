package constant

import "time"

// Todo：非法字符定义

// 错误码
const (
	SUCCESS = 0
	ERROR   = 7
)

// UploadChunkSize 数值
const (
	UploadChunkSize                       = 5 * 1024 * 1024
	UploadSessionExpireTime time.Duration = 1145 * time.Second
	UploadVideoBaseDir                    = "./upload/video"
	UploadChunkBaseDir                    = "./upload/chunk"
)

// LoginType 登录方式
type LoginType string

const (
	Unknown  LoginType = "unknown"
	Password LoginType = "password"
	SMS      LoginType = "sms"
)

type TargetType string

const (
	TargetPost    TargetType = "post"
	TargetComment TargetType = "comment"
	TargetVideo   TargetType = "video"
	TargetUser    TargetType = "user"
	TargetTag     TargetType = "tag"
)

// ActionType 指定几个交互行为
type ActionType string

const (
	ActionLike     ActionType = "like"
	ActionDislike  ActionType = "dislike"
	ActionFavorite ActionType = "favorite"
	ActionShare    ActionType = "share"
)

type UploadStatus string

const (
	UploadPending    UploadStatus = "pending"
	UploadProcessing UploadStatus = "processing"
	UploadCompleted  UploadStatus = "completed"
	UploadFailed     UploadStatus = "failed"
	UploadCanceled   UploadStatus = "canceled"
)
