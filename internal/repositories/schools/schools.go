package schools

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

// SchoolRepository предоставляет доступ к данным школ в PostgreSQL.
type SchoolRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// NewSchoolRepository создаёт репозиторий школ.
func NewSchoolRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *SchoolRepository {
	return &SchoolRepository{pool: pool, queryTimeout: queryTimeout}
}

// GetByID возвращает школу по ID (исключая удалённые).
func (r *SchoolRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.School, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		SELECT id, name, city, logo_key, created_at, updated_at, deleted_at
		FROM schools
		WHERE id = $1 AND deleted_at IS NULL
	`

	var s dto.School
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.City, &s.LogoKey,
		&s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrSchoolNotFound
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// List возвращает список школ по фильтрам и общее количество.
func (r *SchoolRepository) List(ctx context.Context, filter dto.SchoolFilter) ([]*dto.School, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Name+"%")
		argIdx++
	}

	if filter.City != nil && *filter.City != "" {
		conditions = append(conditions, fmt.Sprintf("city ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.City+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Получаем общее количество
	countQuery := "SELECT COUNT(*) FROM schools " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Получаем записи с пагинацией
	query := fmt.Sprintf(`
		SELECT id, name, city, logo_key, created_at, updated_at, deleted_at
		FROM schools
		%s
		ORDER BY name
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var schools []*dto.School
	for rows.Next() {
		var s dto.School
		if err := rows.Scan(
			&s.ID, &s.Name, &s.City, &s.LogoKey,
			&s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		schools = append(schools, &s)
	}

	return schools, total, rows.Err()
}

// Create создаёт новую школу и возвращает созданную запись.
func (r *SchoolRepository) Create(ctx context.Context, school *dto.School) (*dto.School, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		INSERT INTO schools (id, name, city, logo_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, city, logo_key, created_at, updated_at, deleted_at
	`

	var s dto.School
	err := r.pool.QueryRow(ctx, query,
		school.ID, school.Name, school.City, school.LogoKey,
		school.CreatedAt, school.UpdatedAt,
	).Scan(
		&s.ID, &s.Name, &s.City, &s.LogoKey,
		&s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// Update обновляет данные школы и возвращает обновлённую запись.
func (r *SchoolRepository) Update(ctx context.Context, school *dto.School) (*dto.School, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		UPDATE schools
		SET name = $2, city = $3, logo_key = $4, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, city, logo_key, created_at, updated_at, deleted_at
	`

	var s dto.School
	err := r.pool.QueryRow(ctx, query,
		school.ID, school.Name, school.City, school.LogoKey,
	).Scan(
		&s.ID, &s.Name, &s.City, &s.LogoKey,
		&s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrSchoolNotFound
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// SoftDelete выполняет мягкое удаление школы.
func (r *SchoolRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		UPDATE schools
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appErr.ErrSchoolNotFound
	}

	return nil
}
