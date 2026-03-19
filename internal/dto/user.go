package dto

import (
	"github.com/google/uuid"
	"time"
)

// UserRole определяет роль пользователя в системе.
type UserRole string

const (
	RoleStudent     UserRole = "student"
	RoleTeacher     UserRole = "teacher"
	RoleSchoolAdmin UserRole = "school_admin"
	RoleSuperAdmin  UserRole = "super_admin"
)

// TokenPair содержит пару токенов авторизации.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// User содержит данные пользователя для передачи между слоями.
type User struct {
	ID             uuid.UUID
	Role           UserRole
	ClassID        *uuid.UUID
	SchoolID       *uuid.UUID
	Email          string
	PasswordHash   *string
	LastName       string
	FirstName      string
	MiddleName     *string
	AvatarKey      *string
	SessionVersion int
	IsActive       bool
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
}

// UserFilter параметры фильтрации и пагинации для списка пользователей.
type UserFilter struct {
	SchoolID *uuid.UUID
	ClassID  *uuid.UUID
	Role     *UserRole
	Name     *string
	Limit    int
	Offset   int
}
