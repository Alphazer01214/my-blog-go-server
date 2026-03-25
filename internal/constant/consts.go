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

// RoleType 角色
type RoleType int

const (
	Evil RoleType = iota
	Guest
	NormalUser
	VIP
	Moderator
	TakamatsuTomori
)
