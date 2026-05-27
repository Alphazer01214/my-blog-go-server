package request

import "blog.alphazer01214.top/internal/entity"

type TomoriBanRequest struct {
	UserId uint `json:"user_id" binding:"required"`
	Banned bool `json:"banned"`
}

type TomoriRoleRequest struct {
	UserId uint            `json:"user_id" binding:"required"`
	Role   entity.RoleType `json:"role" binding:"required"`
}

type TomoriResetPasswordRequest struct {
	UserId      uint   `json:"user_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
