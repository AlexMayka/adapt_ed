package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashValue хэширует значение через bcrypt.
func HashValue(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckValuesHash сравнивает значение с bcrypt-хэшем.
func CheckValuesHash(value string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(value))
	return err == nil
}
