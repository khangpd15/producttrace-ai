package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a plain password using bcrypt with a cost of 12.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// ComparePassword compares a hashed password with a plain password.
func ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
