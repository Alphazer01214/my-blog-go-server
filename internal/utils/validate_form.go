package utils

func IsPasswordValid(password string) bool {
	if len(password) < 6 || len(password) >= 256 {
		return false
	}
	return true
}
