package utils

import "blog.alphazer01214.top/internal/entity"

func IsPasswordValid(password string) bool {
	if len(password) < 6 || len(password) >= 256 {
		return false
	}
	return true
}

func IsSameUser(extUser *entity.User, dbUser *entity.User) bool {
	if extUser.ID == dbUser.ID {
		return true
	}
	return false
}
