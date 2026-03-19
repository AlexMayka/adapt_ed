package auth

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"backend/internal/logger/interfaces"
	"backend/internal/utils"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"time"
)

type SessionsRepository interface {
	GetSessionVersion(ctx context.Context, userID uuid.UUID) (int, error)
	SetSessionVersion(ctx context.Context, userID uuid.UUID, version int, ttl time.Duration) error
}

type UserRepository interface {
	GetVersionToken(ctx context.Context, user uuid.UUID) (int, error)
}

// Manager управляет генерацией и валидацией JWT access-токенов и refresh-токенов.
type Manager struct {
	accessSecret string
	accessTTL    time.Duration
	refreshTTL   time.Duration
	log          interfaces.Logger

	sessions SessionsRepository
	user     UserRepository
}

// NewAuthManager создаёт менеджер авторизации с секретом и TTL токенов.
func NewAuthManager(log interfaces.Logger, accessSecret string, accessTTL, refreshTTL time.Duration, session SessionsRepository, user UserRepository) *Manager {
	return &Manager{
		log:          log,
		accessSecret: accessSecret,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
		sessions:     session,
		user:         user,
	}
}

// GenerateAccessToken генерирует подписанный JWT access-токен.
func (m *Manager) GenerateAccessToken(userID uuid.UUID, schoolID *uuid.UUID, sessionVersion int, role dto.UserRole) (string, error) {
	now := time.Now()

	var school uuid.UUID
	if schoolID != nil {
		school = *schoolID
	}

	claims := AccessToken{
		UserID:         userID,
		SchoolID:       school,
		SessionVersion: sessionVersion,
		Role:           role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ID:        utils.GetUniqUUID().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.accessSecret))
	if err != nil {
		m.log.Error("ошибка генерации access-токена", "err", err.Error())
		return "", fmt.Errorf("%w: %s", appErr.ErrJWTInvalid, err.Error())
	}

	return tokenString, nil
}

// ParseAccessToken парсит и валидирует JWT access-токен.
func (m *Manager) ParseAccessToken(tokenString string) (*AccessToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessToken{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", appErr.ErrJWTUnexpected, token.Header["alg"])
		}
		return []byte(m.accessSecret), nil
	})
	if err != nil {
		m.log.Error("ошибка парсинга access-токена", "err", err.Error())
		return nil, err
	}

	claims, ok := token.Claims.(*AccessToken)
	if !ok || !token.Valid {
		m.log.Info("невалидный access-токен", "status", ok)
		return nil, fmt.Errorf("%w: %v", appErr.ErrJWTInvalid, err)
	}

	return claims, nil
}

func (m *Manager) CheckToken(tokenString string) (*uuid.UUID, *uuid.UUID, int, *dto.UserRole, error) {
	token, err := m.ParseAccessToken(tokenString)
	if err != nil {
		return nil, nil, 0, nil, fmt.Errorf("%w: %s", appErr.ErrJWTInvalid, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	version, err := m.sessions.GetSessionVersion(ctx, token.UserID)
	if err == nil && token.SessionVersion >= version {
		m.log.Info("access-токен валиден", "user", token.UserID, "version", version, "version_token", token.SessionVersion)
		return &token.UserID, &token.SchoolID, token.SessionVersion, &token.Role, nil
	}

	if err != nil {
		m.log.Error("ошибка получения версии сессии", "err", err.Error(), "user", token.UserID)
	}

	version, err = m.user.GetVersionToken(ctx, token.UserID)
	if err == nil && token.SessionVersion >= version {
		m.log.Info("access-токен валиден", "user", token.UserID, "version", token.SessionVersion)

		if cacheErr := m.sessions.SetSessionVersion(ctx, token.UserID, version, m.accessTTL); cacheErr != nil {
			m.log.Warn("ошибка прогрева кэша сессии", "err", cacheErr, "user_id", token.UserID)
		}

		return &token.UserID, &token.SchoolID, token.SessionVersion, &token.Role, nil
	}

	if err != nil {
		m.log.Error("ошибка получения версии сессии", "err", err.Error(), "user", token.UserID, "version", version, "version_token", token.SessionVersion)
		return nil, nil, 0, nil, nil
	}

	m.log.Info("версия сессии устарела", "user", token.UserID, "version", version, "version_token", token.SessionVersion)
	return nil, nil, 0, nil, nil
}

// GenerateRefreshToken генерирует UUID refresh-токен и время его истечения.
func (m *Manager) GenerateRefreshToken() (string, time.Time) {
	return utils.GetUniqUUID().String(), time.Now().Add(m.refreshTTL)
}

// CheckRefreshToken сравнивает refresh-токен с его bcrypt-хэшем.
func (m *Manager) CheckRefreshToken(tokenString string, hashToken string) bool {
	return utils.CheckValuesHash(tokenString, hashToken)
}

// AccessTTL возвращает время жизни access-токена.
func (m *Manager) AccessTTL() time.Duration {
	return m.accessTTL
}

// RefreshTTL возвращает время жизни refresh-токена.
func (m *Manager) RefreshTTL() time.Duration {
	return m.refreshTTL
}
