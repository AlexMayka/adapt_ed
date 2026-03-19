package user

import (
	"backend/internal/dto"
	"github.com/google/uuid"
	"time"
)

// ── Компоненты DTO ──────────────────────────────────────────────────────────

type UserID struct {
	// UUID пользователя
	ID uuid.UUID `json:"id" example:"10cb44c1-18f7-4a3e-b0bd-5d2609619d65"`
}

type UserFIO struct {
	// Фамилия
	LastName string `json:"last_name" example:"Иванов"`
	// Имя
	FirstName string `json:"first_name" example:"Иван"`
	// Отчество
	MiddleName *string `json:"middle_name,omitempty" example:"Иванович"`
}

type UserEmail struct {
	// Email пользователя
	Email string `json:"email" example:"example@example.com"`
}

type UserRole struct {
	// Роль пользователя
	Role dto.UserRole `json:"role" enums:"student,teacher,school_admin,super_admin" example:"student"`
}

type UserEducation struct {
	// UUID класса
	ClassID *uuid.UUID `json:"class_id,omitempty" example:"1db34ac1-23f1-4b47-82e3-9e8ad16ee1fb"`
	// UUID школы
	SchoolID *uuid.UUID `json:"school_id,omitempty" example:"3e1b8139-1f9d-458f-ac70-41f815c8b128"`
}

type UserAvatar struct {
	// Ключ аватара в S3-хранилище
	AvatarKey *string `json:"avatar_key,omitempty" example:"avatars/10cb44c1/photo.png"`
}

type UserMeta struct {
	// Признак активности
	IsActive bool `json:"is_active" example:"true"`
	// Дата создания
	CreatedAt *time.Time `json:"created_at" example:"2026-03-16T10:30:00Z"`
	// Дата обновления
	UpdatedAt *time.Time `json:"updated_at" example:"2026-03-16T10:30:00Z"`
}

// ── Составные DTO ───────────────────────────────────────────────────────────

type UserResponse struct {
	UserID
	UserEmail
	UserFIO
	UserRole
	UserEducation
	UserAvatar
	UserMeta
}

// ── Обновление профиля ──────────────────────────────────────────────────────

// UpdateProfileRequest обновление своего профиля.
type UpdateProfileRequest struct {
	// Email
	Email *string `json:"email,omitempty" binding:"omitempty,email" example:"new@example.com"`
	// Фамилия
	LastName *string `json:"last_name,omitempty" example:"Петров"`
	// Имя
	FirstName *string `json:"first_name,omitempty" example:"Пётр"`
	// Отчество
	MiddleName *string `json:"middle_name,omitempty" example:"Петрович"`
	// Ключ аватара
	AvatarKey *string `json:"avatar_key,omitempty" example:"avatars/new/photo.png"`
}

// UpdateUserRequest обновление пользователя админом.
type UpdateUserRequest struct {
	UpdateProfileRequest
	// UUID класса
	ClassID *uuid.UUID `json:"class_id,omitempty" example:"1db34ac1-23f1-4b47-82e3-9e8ad16ee1fb"`
	// UUID школы
	SchoolID *uuid.UUID `json:"school_id,omitempty" example:"3e1b8139-1f9d-458f-ac70-41f815c8b128"`
}

// ── Смена пароля ────────────────────────────────────────────────────────────

// ChangePasswordRequest смена пароля.
type ChangePasswordRequest struct {
	// Текущий пароль
	OldPassword string `json:"old_password" binding:"required" example:"OldPass1!"`
	// Новый пароль
	NewPassword string `json:"new_password" binding:"required,min=8" example:"NewPass1!"`
}

// ── Активация/деактивация ───────────────────────────────────────────────────

// SetActiveRequest установка активности пользователя.
type SetActiveRequest struct {
	// Активен ли пользователь
	IsActive bool `json:"is_active" example:"false"`
}

// ── Список ──────────────────────────────────────────────────────────────────

// ListRequest параметры фильтрации.
type ListRequest struct {
	// Фильтр по школе
	SchoolID *uuid.UUID `form:"school_id"`
	// Фильтр по классу
	ClassID *uuid.UUID `form:"class_id"`
	// Фильтр по роли
	Role *dto.UserRole `form:"role"`
	// Поиск по ФИО
	Name *string `form:"name" example:"Иван"`
	// Количество записей
	Limit int `form:"limit" example:"20"`
	// Смещение
	Offset int `form:"offset" example:"0"`
}

// ListResponse список пользователей.
type ListResponse struct {
	// Список пользователей
	Users []UserResponse `json:"users"`
	// Общее количество
	Total int `json:"total" example:"42"`
}
