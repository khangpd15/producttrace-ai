package cache

import "fmt"

func RefreshTokenKey(userID int) string {
	return fmt.Sprintf("refresh:token:%d", userID)
}

func OTPKey(email string) string {
	return fmt.Sprintf("otp:%s", email)
}
