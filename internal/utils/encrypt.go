package utils

import (
	"crypto/sha256"
	"encoding/hex"

	uuid2 "github.com/google/uuid"
)

func EncryptPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func IsPasswordCorrect(cleartext string, hash string) bool {
	return hash == EncryptPassword(cleartext)
}

func GenerateUUID() string {
	u, err := uuid2.NewV7()
	if err != nil {
		return ""
	}
	return u.String()
}
