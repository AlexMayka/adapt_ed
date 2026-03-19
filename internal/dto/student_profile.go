package dto

import (
	"github.com/google/uuid"
	"time"
)

// DifficultyLevel определяет уровень сложности.
type DifficultyLevel string

const (
	LevelSimple   DifficultyLevel = "simple"
	LevelMedium   DifficultyLevel = "medium"
	LevelAdvanced DifficultyLevel = "advanced"
)

// StudentProfile содержит данные профиля ученика.
type StudentProfile struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	DefaultLevel DifficultyLevel
	Interests    []uuid.UUID
	IsActive     bool
	Version      int
	CreatedAt    *time.Time
}
