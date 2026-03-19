package school

import (
	"github.com/google/uuid"
	"time"
)

// ── Компоненты DTO ──────────────────────────────────────────────────────────

type SchoolName struct {
	// Наименование школы
	Name string `json:"name" binding:"required" example:"МКОУ 'Никольская СОШ'"`
}

type SchoolCity struct {
	// Город школы
	City string `json:"city" binding:"required" example:"Воронеж"`
}

type SchoolID struct {
	// UUID школы
	ID uuid.UUID `json:"id" example:"10cb44c1-18f7-4a3e-b0bd-5d2609619d65"`
}

type SchoolMeta struct {
	// Дата создания школы
	CreatedAt *time.Time `json:"created_at" example:"2026-03-16T10:30:00Z"`
	// Дата последнего обновления школы
	UpdatedAt *time.Time `json:"updated_at" example:"2026-03-16T10:30:00Z"`
}

type Logo struct {
	// Ключ логотипа в S3-хранилище
	LogoKey *string `json:"logo_key,omitempty" example:"logos/10cb44c1/photo.png"`
}

// ── Составные DTO ───────────────────────────────────────────────────────────

type SchoolBase struct {
	SchoolName
	SchoolCity
}

type SchoolResponse struct {
	SchoolID
	SchoolBase
	Logo
	SchoolMeta
}

// ── Создание ────────────────────────────────────────────────────────────────

// CreateRequest входные данные для создания школы.
type CreateRequest struct {
	SchoolBase
}

// ── Обновление ──────────────────────────────────────────────────────────────

// UpdateRequest входные данные для обновления школы. ID берётся из URL.
type UpdateRequest struct {
	// Наименование школы
	Name *string `json:"name,omitempty" example:"МКОУ 'Никольская СОШ'"`
	// Город школы
	City *string `json:"city,omitempty" example:"Воронеж"`
	// Ключ логотипа в S3-хранилище
	LogoKey *string `json:"logo_key,omitempty" example:"logos/10cb44c1/photo.png"`
}

// ── Список ──────────────────────────────────────────────────────────────────

// ListRequest параметры фильтрации и пагинации для списка школ.
type ListRequest struct {
	// Поиск по названию (подстрока)
	Name *string `form:"name" example:"Никольская"`
	// Фильтр по городу
	City *string `form:"city" example:"Воронеж"`
	// Количество записей
	Limit int `form:"limit" example:"20"`
	// Смещение
	Offset int `form:"offset" example:"0"`
}

// ListResponse список школ с пагинацией.
type ListResponse struct {
	// Список школ
	Schools []SchoolResponse `json:"schools"`
	// Общее количество записей
	Total int `json:"total" example:"42"`
}
