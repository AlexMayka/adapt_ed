package sessions

import (
	"backend/internal/storage/interfaces"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// SessionRepository предоставляет кэширование данных сессий в Redis.
type SessionRepository struct {
	cache interfaces.CacheStorage
}

// NewSessionRepository создаёт репозиторий сессий.
func NewSessionRepository(cache interfaces.CacheStorage) *SessionRepository {
	return &SessionRepository{cache: cache}
}

func sessionVersionKey(userID uuid.UUID) string {
	return fmt.Sprintf("session_version:%s", userID.String())
}

func refreshTokenKey(userID uuid.UUID) string {
	return fmt.Sprintf("refresh_token:%s", userID.String())
}

// SetSessionVersion кэширует версию сессии пользователя.
func (r *SessionRepository) SetSessionVersion(ctx context.Context, userID uuid.UUID, version int, ttl time.Duration) error {
	return r.cache.Set(ctx, sessionVersionKey(userID), strconv.Itoa(version), ttl)
}

// GetSessionVersion возвращает кэшированную версию сессии. Возвращает -1 при cache miss.
func (r *SessionRepository) GetSessionVersion(ctx context.Context, userID uuid.UUID) (int, error) {
	val, err := r.cache.Get(ctx, sessionVersionKey(userID))
	if err != nil {
		return -1, err
	}
	if val == "" {
		return -1, nil
	}
	return strconv.Atoi(val)
}

// SetRefreshTokenHash кэширует хэш refresh token пользователя.
func (r *SessionRepository) SetRefreshTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string, ttl time.Duration) error {
	return r.cache.Set(ctx, refreshTokenKey(userID), tokenHash, ttl)
}

// GetRefreshTokenHash возвращает кэшированный хэш refresh token. Пустая строка при cache miss.
func (r *SessionRepository) GetRefreshTokenHash(ctx context.Context, userID uuid.UUID) (string, error) {
	return r.cache.Get(ctx, refreshTokenKey(userID))
}

// DelSession удаляет кэш сессии пользователя (version + token).
func (r *SessionRepository) DelSession(ctx context.Context, userID uuid.UUID) error {
	err1 := r.cache.Del(ctx, sessionVersionKey(userID))
	err2 := r.cache.Del(ctx, refreshTokenKey(userID))
	if err1 != nil {
		return err1
	}
	return err2
}
