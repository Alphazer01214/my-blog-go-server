package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func EncryptPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func IsPasswordCorrect(cleartext string, hash string) bool {
	return hash == EncryptPassword(cleartext)
}
