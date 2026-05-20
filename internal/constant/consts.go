package constant

// Todo：非法字符定义

// 错误码
const (
	SUCCESS = 0
	ERROR   = 7
)

// LoginType 登录方式
type LoginType int

const (
	Unknown LoginType = iota
	Password
	SMS
)

type TargetType int

const (
	TargetPost TargetType = iota + 1
	TargetComment
	TargetVideo
	TargetUser
	TargetTag
)

// ActionType 指定几个交互行为
type ActionType int

const (
	ActionLike ActionType = iota + 1
	ActionDislike
	ActionFavorite
	ActionShare
)
