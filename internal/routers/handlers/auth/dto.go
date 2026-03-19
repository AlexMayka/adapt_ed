package auth

import (
	"backend/internal/dto"
	"github.com/google/uuid"
	"time"
)

// ── Компоненты DTO ──────────────────────────────────────────────────────────

type AuthEmail struct {
	Email string `json:"email" binding:"required,email" example:"example@example.com" description:"Email пользователя"`
}

type AuthPassword struct {
	Password string `json:"password" binding:"required,min=8" minLength:"8" example:"NFMpC9!fm;ARqoh" description:"Пароль пользователя, минимум 8 символов"`
}

type GeneratedPassword struct {
	Password string `json:"password" example:"NFMpC9!fm;ARqoh" description:"Сгенерированный временный пароль пользователя"`
}

type Role struct {
	Role dto.UserRole `json:"role" binding:"required,oneof=student teacher school_admin super_admin" enums:"student,teacher,school_admin,super_admin" example:"student" description:"Роль пользователя в системе"`
}

type AuthParamResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." description:"JWT access token"`
	RefreshToken string `json:"refresh_token" example:"b7d0e8d1-47d8-4d3e-8c2b-7b1d0e8a1234" description:"Refresh token для обновления access token"`
}

type UserMeta struct {
	IsActive  bool       `json:"is_active" example:"true" description:"Признак активности пользователя"`
	CreatedAt *time.Time `json:"created_at" example:"2026-03-16T10:30:00Z" description:"Дата создания пользователя"`
	UpdatedAt *time.Time `json:"updated_at" example:"2026-03-16T10:30:00Z" description:"Дата последнего обновления пользователя"`
}

type FIO struct {
	LastName   string  `json:"last_name" binding:"required" example:"Иванов" description:"Фамилия пользователя"`
	FirstName  string  `json:"first_name" binding:"required" example:"Иван" description:"Имя пользователя"`
	MiddleName *string `json:"middle_name,omitempty" example:"Иванович" description:"Отчество пользователя, если имеется"`
}

type Education struct {
	ClassID  *uuid.UUID `json:"class_id,omitempty" example:"1db34ac1-23f1-4b47-82e3-9e8ad16ee1fb" description:"UUID класса пользователя"`
	SchoolID *uuid.UUID `json:"school_id,omitempty" example:"3e1b8139-1f9d-458f-ac70-41f815c8b128" description:"UUID школы пользователя"`
}

type Avatar struct {
	AvatarKey *string `json:"avatar_key,omitempty" example:"avatars/10cb44c1/photo.png" description:"Ключ аватара в S3-хранилище"`
}

type UserID struct {
	ID uuid.UUID `json:"id" example:"10cb44c1-18f7-4a3e-b0bd-5d2609619d65" description:"UUID пользователя"`
}

// ── Составные DTO ───────────────────────────────────────────────────────────

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
	UserID       uuid.UUID `json:"user_id" binding:"required" example:"10cb44c1-18f7-4a3e-b0bd-5d2609619d65" description:"UUID пользователя"`
	RefreshToken string    `json:"refresh_token" binding:"required" example:"b7d0e8d1-47d8-4d3e-8c2b-7b1d0e8a1234" description:"Текущий refresh token"`
}

// RefreshResponse выходные данные с новой парой токенов.
type RefreshResponse struct {
	AuthParamResponse
}

// ── Logout ──────────────────────────────────────────────────────────────────

// LogoutRequest входные данные для выхода из текущей сессии.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"b7d0e8d1-47d8-4d3e-8c2b-7b1d0e8a1234" description:"Refresh token текущей сессии"`
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
