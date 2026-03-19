package users

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// UserRepository предоставляет доступ к данным пользователей в PostgreSQL.
type UserRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// NewUserRepository создаёт репозиторий пользователей.
func NewUserRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *UserRepository {
	return &UserRepository{pool: pool, queryTimeout: queryTimeout}
}

// GetUserByEmail находит пользователя по email (исключая удалённых).
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*dto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		SELECT id, role, class_id, school_id, email, password_hash,
		       last_name, first_name, middle_name, avatar_key,
		       session_version, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1
	`

	var user dto.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Role,
		&user.ClassID,
		&user.SchoolID,
		&user.Email,
		&user.PasswordHash,
		&user.LastName,
		&user.FirstName,
		&user.MiddleName,
		&user.AvatarKey,
		&user.SessionVersion,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *dto.User) (*dto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		INSERT INTO users (
			id, role, class_id, school_id, email, password_hash,
			last_name, first_name, middle_name, avatar_key,
			session_version, is_active, created_at, updated_at, deleted_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14, $15
		)
		RETURNING
			id, role, class_id, school_id, email, password_hash,
			last_name, first_name, middle_name, avatar_key,
			session_version, is_active, created_at, updated_at, deleted_at
	`

	userDB := &dto.User{}

	err := r.pool.QueryRow(
		ctx,
		query,
		user.ID,
		user.Role,
		user.ClassID,
		user.SchoolID,
		user.Email,
		user.PasswordHash,
		user.LastName,
		user.FirstName,
		user.MiddleName,
		user.AvatarKey,
		user.SessionVersion,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
		user.DeletedAt,
	).Scan(
		&userDB.ID,
		&userDB.Role,
		&userDB.ClassID,
		&userDB.SchoolID,
		&userDB.Email,
		&userDB.PasswordHash,
		&userDB.LastName,
		&userDB.FirstName,
		&userDB.MiddleName,
		&userDB.AvatarKey,
		&userDB.SessionVersion,
		&userDB.IsActive,
		&userDB.CreatedAt,
		&userDB.UpdatedAt,
		&userDB.DeletedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "uq_users_email" {
					return nil, appErr.ErrEmailAlreadyExists
				}
				return nil, fmt.Errorf("unique violation: %w", err)
			}
		}

		return nil, err
	}

	return userDB, nil
}

func (r *UserRepository) GetVersionToken(ctx context.Context, user uuid.UUID) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		SELECT session_version
		FROM users
		WHERE id = $1
	`

	var sessionVersion int
	err := r.pool.QueryRow(ctx, query, user).Scan(&sessionVersion)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, appErr.ErrUserNotFound
	}

	if err != nil {
		return 0, err
	}

	return sessionVersion, nil
}

// IncrementSessionVersion атомарно увеличивает версию сессии и возвращает новое значение.
func (r *UserRepository) IncrementSessionVersion(ctx context.Context, userID uuid.UUID) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		UPDATE users
		SET session_version = session_version + 1
		WHERE id = $1
		RETURNING session_version
	`

	var version int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, appErr.ErrUserNotFound
	}
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*dto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		SELECT id, role, class_id, school_id, email, password_hash,
		       last_name, first_name, middle_name, avatar_key,
		       session_version, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1
	`

	var user dto.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Role,
		&user.ClassID,
		&user.SchoolID,
		&user.Email,
		&user.PasswordHash,
		&user.LastName,
		&user.FirstName,
		&user.MiddleName,
		&user.AvatarKey,
		&user.SessionVersion,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
