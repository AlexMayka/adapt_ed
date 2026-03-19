package profiles

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// ProfileRepository предоставляет доступ к профилям учеников в PostgreSQL.
type ProfileRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// NewProfileRepository создаёт репозиторий профилей.
func NewProfileRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *ProfileRepository {
	return &ProfileRepository{pool: pool, queryTimeout: queryTimeout}
}

func scanProfile(row pgx.Row) (*dto.StudentProfile, error) {
	var p dto.StudentProfile
	err := row.Scan(&p.ID, &p.UserID, &p.DefaultLevel, &p.Interests, &p.IsActive, &p.Version, &p.CreatedAt)
	return &p, err
}

// GetActiveByUserID возвращает текущий активный профиль ученика.
func (r *ProfileRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*dto.StudentProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		SELECT id, user_id, default_level, interests, is_active, version, created_at
		FROM student_profiles
		WHERE user_id = $1 AND is_active = true
	`

	p, err := scanProfile(r.pool.QueryRow(ctx, query, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrProfileNotFound
	}
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Create создаёт новый профиль ученика.
func (r *ProfileRepository) Create(ctx context.Context, profile *dto.StudentProfile) (*dto.StudentProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		INSERT INTO student_profiles (id, user_id, default_level, interests, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, default_level, interests, is_active, version, created_at
	`

	p, err := scanProfile(r.pool.QueryRow(ctx, query,
		profile.ID, profile.UserID, profile.DefaultLevel, profile.Interests,
		profile.IsActive, profile.CreatedAt,
	))
	if err != nil {
		return nil, err
	}

	return p, nil
}

// DeactivateByUserID деактивирует текущий активный профиль (для версионирования).
func (r *ProfileRepository) DeactivateByUserID(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `UPDATE student_profiles SET is_active = false WHERE user_id = $1 AND is_active = true`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
