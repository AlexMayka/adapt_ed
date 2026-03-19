package profile

import (
	"backend/internal/dto"
	"github.com/google/uuid"
	"time"
)

type ProfileResponse struct {
	// UUID профиля
	ID uuid.UUID `json:"id" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	// UUID пользователя
	UserID uuid.UUID `json:"user_id" example:"10cb44c1-18f7-4a3e-b0bd-5d2609619d65"`
	// Уровень сложности по умолчанию
	DefaultLevel dto.DifficultyLevel `json:"default_level" enums:"simple,medium,advanced" example:"simple"`
	// Список UUID интересов
	Interests []uuid.UUID `json:"interests"`
	// Версия профиля
	Version int `json:"version" example:"1"`
	// Дата создания версии
	CreatedAt *time.Time `json:"created_at" example:"2026-03-16T10:30:00Z"`
}

// UpdateProfileRequest обновление профиля ученика.
type UpdateProfileRequest struct {
	// Уровень сложности
	DefaultLevel *dto.DifficultyLevel `json:"default_level,omitempty" enums:"simple,medium,advanced" example:"medium"`
	// Список UUID интересов
	Interests []uuid.UUID `json:"interests,omitempty"`
}
