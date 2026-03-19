package auth

import (
	"backend/internal/dto"
	"context"
	"github.com/google/uuid"
)

type AuthService interface {
	// Login выполняет аутентификацию по email/password.
	Login(ctx context.Context, email, password, userAgent, ip string) (*dto.User, *dto.TokenPair, error)

	// Registration создаёт нового пользователя (самостоятельная регистрация).
	Registration(ctx context.Context, user *dto.User, password string, userAgent, ip string) (*dto.User, *dto.TokenPair, error)

	// RegistrationByAdmin создаёт пользователя через админку с генерацией пароля.
	RegistrationByAdmin(ctx context.Context, user *dto.User) (*dto.User, string, error)

	// GetMe возвращает данные текущего пользователя.
	GetMe(ctx context.Context, id uuid.UUID) (*dto.User, error)

	// Refresh обновляет пару токенов по refresh token.
	Refresh(ctx context.Context, userID uuid.UUID, refreshToken, userAgent, ip string) (*dto.TokenPair, error)

	// Logout инвалидирует текущий refresh token.
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error

	// LogoutAll инвалидирует все refresh token пользователя.
	LogoutAll(ctx context.Context, userID uuid.UUID) error
}
