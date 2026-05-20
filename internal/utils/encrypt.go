package utils

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func EncryptPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

func IsPasswordCorrect(cleartext string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(cleartext)) == nil
}

func GenerateUUID() string {
	u, err := uuid.NewV7()
	if err != nil {
		return ""
	}
	return u.String()
}
