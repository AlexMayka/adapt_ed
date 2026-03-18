package tokens

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// TokenRepository предоставляет доступ к refresh-токенам в PostgreSQL.
type TokenRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// NewTokenRepository создаёт репозиторий токенов.
func NewTokenRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *TokenRepository {
	return &TokenRepository{pool: pool, queryTimeout: queryTimeout}
}

// SetTokenByUser сохраняет хеш refresh token в БД.
func (r *TokenRepository) SetTokenByUser(ctx context.Context, userID uuid.UUID, tokenHash, deviceInfo string, expires time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, device_info, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, userID, tokenHash, deviceInfo, expires)
	return err
}

// RevokeToken отзывает конкретный refresh token по хешу.
func (r *TokenRepository) RevokeToken(ctx context.Context, tokenHash string) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, tokenHash)
	return err
}

// RevokeAllByUser отзывает все активные refresh token пользователя.
func (r *TokenRepository) RevokeAllByUser(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// GetActiveTokenHash проверяет наличие активного (не отозванного, не истёкшего) токена.
func (r *TokenRepository) GetActiveTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		SELECT EXISTS(
			SELECT 1 FROM refresh_tokens
			WHERE user_id = $1
			  AND token_hash = $2
			  AND revoked_at IS NULL
			  AND expires_at > now()
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, tokenHash).Scan(&exists)
	return exists, err
}
