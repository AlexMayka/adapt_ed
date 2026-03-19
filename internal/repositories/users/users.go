package users

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"context"
	"errors"
	"fmt"
	"strings"

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

const userColumns = `id, role, class_id, school_id, email, password_hash,
	last_name, first_name, middle_name, avatar_key,
	session_version, is_active, created_at, updated_at, deleted_at`

func scanUser(row pgx.Row) (*dto.User, error) {
	var u dto.User
	err := row.Scan(
		&u.ID, &u.Role, &u.ClassID, &u.SchoolID, &u.Email, &u.PasswordHash,
		&u.LastName, &u.FirstName, &u.MiddleName, &u.AvatarKey,
		&u.SessionVersion, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	return &u, err
}

// List возвращает список пользователей по фильтрам и общее количество.
func (r *UserRepository) List(ctx context.Context, filter dto.UserFilter) ([]*dto.User, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.SchoolID != nil {
		conditions = append(conditions, fmt.Sprintf("school_id = $%d", argIdx))
		args = append(args, *filter.SchoolID)
		argIdx++
	}

	if filter.ClassID != nil {
		conditions = append(conditions, fmt.Sprintf("class_id = $%d", argIdx))
		args = append(args, *filter.ClassID)
		argIdx++
	}

	if filter.Role != nil {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, *filter.Role)
		argIdx++
	}

	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(last_name ILIKE $%d OR first_name ILIKE $%d OR middle_name ILIKE $%d)",
			argIdx, argIdx, argIdx,
		))
		args = append(args, "%"+*filter.Name+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := "SELECT COUNT(*) FROM users " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM users %s ORDER BY last_name, first_name LIMIT $%d OFFSET $%d`,
		userColumns, where, argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*dto.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, rows.Err()
}

// Update обновляет данные пользователя и возвращает обновлённую запись.
func (r *UserRepository) Update(ctx context.Context, user *dto.User) (*dto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := fmt.Sprintf(`
		UPDATE users
		SET email = $2, last_name = $3, first_name = $4, middle_name = $5,
		    avatar_key = $6, class_id = $7, school_id = $8, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING %s
	`, userColumns)

	u, err := scanUser(r.pool.QueryRow(ctx, query,
		user.ID, user.Email, user.LastName, user.FirstName, user.MiddleName,
		user.AvatarKey, user.ClassID, user.SchoolID,
	))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrUserNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uq_users_email" {
			return nil, appErr.ErrEmailAlreadyExists
		}
		return nil, err
	}

	return u, nil
}

// UpdatePassword обновляет хэш пароля пользователя.
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, query, userID, passwordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appErr.ErrUserNotFound
	}

	return nil
}

// SetActive устанавливает is_active для пользователя.
func (r *UserRepository) SetActive(ctx context.Context, userID uuid.UUID, active bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `UPDATE users SET is_active = $2, updated_at = now() WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, query, userID, active)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appErr.ErrUserNotFound
	}

	return nil
}

// SoftDelete выполняет мягкое удаление пользователя.
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `UPDATE users SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appErr.ErrUserNotFound
	}

	return nil
}

// Restore восстанавливает мягко удалённого пользователя.
func (r *UserRepository) Restore(ctx context.Context, id uuid.UUID) (*dto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := fmt.Sprintf(`
		UPDATE users SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NOT NULL
		RETURNING %s
	`, userColumns)

	u, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}
