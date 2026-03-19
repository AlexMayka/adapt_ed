package auth

import (
	"backend/internal/auth"
	"backend/internal/dto"
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*dto.User, error)
	CreateUser(ctx context.Context, user *dto.User) (*dto.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*dto.User, error)
	IncrementSessionVersion(ctx context.Context, userID uuid.UUID) (int, error)
}

type TokenRepository interface {
	SetTokenByUser(ctx context.Context, userID uuid.UUID, tokenHash, deviceInfo string, expires time.Time) error
	RevokeToken(ctx context.Context, tokenHash string) error
	RevokeAllByUser(ctx context.Context, userID uuid.UUID) error
	GetActiveTokenHashesByUser(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type AuthManager interface {
	GenerateAccessToken(userID uuid.UUID, schoolID *uuid.UUID, sessionVersion int, role dto.UserRole) (string, error)
	ParseAccessToken(tokenString string) (*auth.AccessToken, error)
	GenerateRefreshToken() (string, time.Time)
	CheckRefreshToken(tokenString string, hashToken string) bool
	AccessTTL() time.Duration
	RefreshTTL() time.Duration
}

type SessionCache interface {
	SetSessionVersion(ctx context.Context, userID uuid.UUID, version int, ttl time.Duration) error
	SetRefreshTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string, ttl time.Duration) error
	DelSession(ctx context.Context, userID uuid.UUID) error
}
