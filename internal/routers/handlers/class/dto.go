package class

import (
	"github.com/google/uuid"
	"time"
)

// ── Компоненты DTO ──────────────────────────────────────────────────────────

type ClassID struct {
	// UUID класса
	ID uuid.UUID `json:"id" example:"1db34ac1-23f1-4b47-82e3-9e8ad16ee1fb"`
}

type ClassSchoolID struct {
	// UUID школы
	SchoolID uuid.UUID `json:"school_id" example:"3e1b8139-1f9d-458f-ac70-41f815c8b128"`
}

type ClassNumber struct {
	// Номер класса (5, 6, 7...)
	NumberOfClass int `json:"number_of_class" binding:"required" example:"7"`
}

type ClassSuffix struct {
	// Суффикс класса (А, Б, В...)
	SuffixesOfClass string `json:"suffixes_of_class" binding:"required" example:"А"`
}

type ClassAcademicYear struct {
	// Дата начала учебного года
	AcademicYearStart *time.Time `json:"academic_year_start,omitempty" example:"2025-09-01T00:00:00Z"`
	// Дата окончания учебного года
	AcademicYearEnd *time.Time `json:"academic_year_end,omitempty" example:"2026-08-31T00:00:00Z"`
}

type ClassMeta struct {
	// Дата создания записи
	CreatedAt *time.Time `json:"created_at" example:"2026-03-16T10:30:00Z"`
	// Дата последнего обновления записи
	UpdatedAt *time.Time `json:"updated_at" example:"2026-03-16T10:30:00Z"`
}

// ── Составные DTO ───────────────────────────────────────────────────────────

type ClassResponse struct {
	ClassID
	ClassSchoolID
	ClassNumber
	ClassSuffix
	ClassAcademicYear
	ClassMeta
}

// ── Создание ────────────────────────────────────────────────────────────────

// CreateRequest входные данные для создания класса.
type CreateRequest struct {
	ClassNumber
	ClassSuffix
	ClassAcademicYear
}

// ── Обновление ──────────────────────────────────────────────────────────────

// UpdateRequest входные данные для обновления класса. ID берётся из URL.
type UpdateRequest struct {
	// Номер класса
	NumberOfClass *int `json:"number_of_class,omitempty" example:"7"`
	// Суффикс класса
	SuffixesOfClass *string `json:"suffixes_of_class,omitempty" example:"Б"`
	ClassAcademicYear
}

// ── Список ──────────────────────────────────────────────────────────────────

// ListRequest параметры фильтрации и пагинации.
type ListRequest struct {
	// Фильтр по номеру класса
	NumberOfClass *int `form:"number_of_class" example:"7"`
	// Количество записей
	Limit int `form:"limit" example:"20"`
	// Смещение
	Offset int `form:"offset" example:"0"`
}

// ListResponse список классов с пагинацией.
type ListResponse struct {
	// Список классов
	Classes []ClassResponse `json:"classes"`
	// Общее количество записей
	Total int `json:"total" example:"12"`
}
