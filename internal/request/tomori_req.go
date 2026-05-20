package request

type TomoriBanRequest struct {
	UserId uint `json:"user_id" binding:"required"`
	Banned bool `json:"banned"`
}

type TomoriRoleRequest struct {
	UserId uint            `json:"user_id" binding:"required"`
	Role   int             `json:"role" binding:"required"`
}

type TomoriResetPasswordRequest struct {
	UserId      uint   `json:"user_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
