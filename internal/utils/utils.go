package utils

import "github.com/google/uuid"

func GetUniqUUID() uuid.UUID {
	return uuid.New()
}
