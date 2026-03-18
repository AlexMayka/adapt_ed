package utils

import "github.com/google/uuid"

// GetUniqUUID генерирует новый UUID v4.
func GetUniqUUID() uuid.UUID {
	return uuid.New()
}
