package auth

import (
	"backend/internal/dto"
	"github.com/google/uuid"
	"time"
)

// ── Компоненты DTO ──────────────────────────────────────────────────────────

type AuthEmail struct {
	// Email пользователя
	Email string `json:"email" binding:"required,email" example:"example@example.com"`
}

type AuthPassword struct {
	// Пароль пользователя, минимум 8 символов
	Password string `json:"password" binding:"required,min=8" minLength:"8" example:"NFMpC9!fm;ARqoh"`
}

type GeneratedPassword struct {
	// Сгенерированный временный пароль пользователя
	Password string `json:"password" example:"NFMpC9!fm;ARqoh"`
}

type Role struct {
	// Роль пользователя в системе
	Role dto.UserRole `json:"role" binding:"required,oneof=student teacher school_admin super_admin" enums:"student,teacher,school_admin,super_admin" example:"student"`
}

type AuthParamResponse struct {
	// JWT access token
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// Refresh token для обновления access token
	RefreshToken string `json:"refresh_token" example:"b7d0e8d1-47d8-4d3e-8c2b-7b1d0e8a1234"`
}

type UserMeta struct {
	// Признак активности пользователя
	IsActive bool `json:"is_active" example:"true"`
	// Дата создания пользователя
	CreatedAt *time.Time `json:"created_at" example:"2026-03-16T10:30:00Z"`
	// Дата последнего обновления пользователя
	UpdatedAt *time.Time `json:"updated_at" example:"2026-03-16T10:30:00Z"`
}

type FIO struct {
	// Фамилия пользователя
	LastName string `json:"last_name" binding:"required" example:"Иванов"`
	// Имя пользователя
	FirstName string `json:"first_name" binding:"required" example:"Иван"`
	// Отчество пользователя, если имеется
	MiddleName *string `json:"middle_name,omitempty" example:"Иванович"`
}

type Education struct {
	// UUID класса пользователя
	ClassID *uuid.UUID `json:"class_id,omitempty" example:"1db34ac1-23f1-4b47-82e3-9e8ad16ee1fb"`
	// UUID школы пользователя
	SchoolID *uuid.UUID `json:"school_id,omitempty" example:"3e1b8139-1f9d-458f-ac70-41f815c8b128"`
}

type Avatar struct {
	// Ключ аватара в S3-хранилище
	AvatarKey *string `json:"avatar_key,omitempty" example:"avatars/10cb44c1/photo.png"`
}

type UserID struct {
	// UUID пользователя
	ID uuid.UUID `json:"id" example:"10cb44c1-18f7-4a3e-b0bd-5d2609619d65"`
}

// UserBase содержит общие поля пользователя, используемые в DTO.
type UserBase struct {
	AuthEmail
	Education
	FIO
}

// ── Регистрация (самостоятельная) ───────────────────────────────────────────

// RegistrationRequest входные данные для самостоятельной регистрации.
type RegistrationRequest struct {
	AuthEmail
	FIO
	AuthPassword
}

// RegistrationResponse выходные данные после самостоятельной регистрации.
type RegistrationResponse struct {
	UserID
	UserBase
	Role
	UserMeta
	AuthParamResponse
}

// ── Регистрация (через админку) ─────────────────────────────────────────────

// RegistrationRequestByAdmin входные данные для создания пользователя через админку.
type RegistrationRequestByAdmin struct {
	UserBase
	Role
}

// RegistrationResponseByAdmin выходные данные после создания пользователя админом.
type RegistrationResponseByAdmin struct {
	UserID
	UserBase
	Role
	GeneratedPassword
	UserMeta
}

// ── Логин ───────────────────────────────────────────────────────────────────

// LoginRequest входные данные для аутентификации.
type LoginRequest struct {
	AuthEmail
	AuthPassword
}

// LoginResponse выходные данные после успешного логина.
type LoginResponse struct {
	UserID
	UserBase
	UserMeta
	Role
	AuthParamResponse
}

// ── Refresh ─────────────────────────────────────────────────────────────────

// RefreshRequest входные данные для обновления токенов.
type RefreshRequest struct {
	// UUID пользователя
	UserID uuid.UUID `json:"user_id" binding:"required" example:"10cb44c1-18f7-4a3e-b0bd-5d2609619d65"`
	// Текущий refresh token
	RefreshToken string `json:"refresh_token" binding:"required" example:"b7d0e8d1-47d8-4d3e-8c2b-7b1d0e8a1234"`
}

// RefreshResponse выходные данные с новой парой токенов.
type RefreshResponse struct {
	AuthParamResponse
}

// ── Logout ──────────────────────────────────────────────────────────────────

// LogoutRequest входные данные для выхода из текущей сессии.
type LogoutRequest struct {
	// Refresh token текущей сессии
	RefreshToken string `json:"refresh_token" binding:"required" example:"b7d0e8d1-47d8-4d3e-8c2b-7b1d0e8a1234"`
}

// LogoutResponse выходные данные после выхода.
type LogoutResponse struct {
	UserID
}

// ── GetMe ───────────────────────────────────────────────────────────────────

// GetMeResponse данные текущего авторизованного пользователя.
type GetMeResponse struct {
	UserID
	UserBase
	UserMeta
	Role
	Avatar
}
