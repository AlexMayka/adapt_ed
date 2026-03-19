package utils

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const passwordChars = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%&*"

// GeneratePassword генерирует случайный пароль заданной длины.
func GeneratePassword(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordChars))))
		if err != nil {
			return "", err
		}
		result[i] = passwordChars[idx.Int64()]
	}
	return string(result), nil
}

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
