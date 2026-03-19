package interest

import (
	"github.com/google/uuid"
	"time"
)

// ── Компоненты DTO ──────────────────────────────────────────────────────────

type InterestID struct {
	// UUID интереса
	ID uuid.UUID `json:"id" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
}

type InterestResponse struct {
	InterestID
	// Название интереса
	Name string `json:"name" example:"Футбол"`
	// Ключ иконки в S3
	IconKey *string `json:"icon_key,omitempty" example:"interests/football.png"`
	// Верифицирован ли интерес
	IsVerified bool `json:"is_verified" example:"true"`
	// Дата создания
	CreatedAt *time.Time `json:"created_at" example:"2026-03-16T10:30:00Z"`
}

// ── Создание ────────────────────────────────────────────────────────────────

// CreateRequest входные данные для создания интереса.
type CreateRequest struct {
	// Название интереса
	Name string `json:"name" binding:"required" example:"Футбол"`
}

// ── Обновление ──────────────────────────────────────────────────────────────

// UpdateRequest входные данные для обновления интереса.
type UpdateRequest struct {
	// Название интереса
	Name *string `json:"name,omitempty" example:"Баскетбол"`
	// Ключ иконки
	IconKey *string `json:"icon_key,omitempty" example:"interests/basketball.png"`
}

// ── Верификация ─────────────────────────────────────────────────────────────

// VerifyRequest массовая верификация интересов.
type VerifyRequest struct {
	// Список UUID интересов для верификации
	IDs []uuid.UUID `json:"ids" binding:"required" example:"[\"a1b2c3d4-e5f6-7890-abcd-ef1234567890\"]"`
}

// VerifyResponse результат верификации.
type VerifyResponse struct {
	// Количество верифицированных интересов
	Verified int `json:"verified" example:"3"`
}

// ── Список ──────────────────────────────────────────────────────────────────

// ListRequest параметры фильтрации.
type ListRequest struct {
	// Поиск по названию
	Name *string `form:"name" example:"футбол"`
	// Фильтр по верификации
	IsVerified *bool `form:"is_verified"`
	// Количество записей
	Limit int `form:"limit" example:"20"`
	// Смещение
	Offset int `form:"offset" example:"0"`
}

// ListResponse список интересов.
type ListResponse struct {
	// Список интересов
	Interests []InterestResponse `json:"interests"`
	// Общее количество
	Total int `json:"total" example:"15"`
}
