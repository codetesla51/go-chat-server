package utils

import "fmt"

const (
	MaxUsernameLength = 20
	MinUsernameLength = 2
)

func IsValidUsername(username string) (bool, string) {
	if len(username) < MinUsernameLength {
		return false, fmt.Sprintf("Username too short (min %d characters)", MinUsernameLength)
	}
	if len(username) > MaxUsernameLength {
		return false, fmt.Sprintf("Username too long (max %d characters)", MaxUsernameLength)
	}
	for _, ch := range username {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || 
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '-') {
			return false, "Username can only contain letters, numbers, - and _"
		}
	}
	return true, ""
}
