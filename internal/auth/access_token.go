package auth

import (
	"backend/internal/dto"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// AccessToken содержит JWT claims для access-токена.
type AccessToken struct {
	UserID         uuid.UUID
	SchoolID       uuid.UUID
	SessionVersion int
	Role           dto.UserRole

	jwt.RegisteredClaims
}
