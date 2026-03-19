package classes

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

// ClassRepository предоставляет доступ к данным классов в PostgreSQL.
type ClassRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// NewClassRepository создаёт репозиторий классов.
func NewClassRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *ClassRepository {
	return &ClassRepository{pool: pool, queryTimeout: queryTimeout}
}

// GetByID возвращает класс по ID (исключая удалённые).
func (r *ClassRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.Class, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		SELECT id, school_id, number_of_class, suffixes_of_class,
		       academic_year_start, academic_year_finish,
		       created_at, updated_at, deleted_at
		FROM classes
		WHERE id = $1 AND deleted_at IS NULL
	`

	var c dto.Class
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.SchoolID, &c.NumberOfClass, &c.SuffixesOfClass,
		&c.AcademicYearStart, &c.AcademicYearEnd,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrClassNotFound
	}
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// List возвращает список классов школы с фильтрацией и пагинацией.
func (r *ClassRepository) List(ctx context.Context, filter dto.ClassFilter) ([]*dto.Class, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("school_id = $%d", argIdx))
	args = append(args, filter.SchoolID)
	argIdx++

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.NumberOfClass != nil {
		conditions = append(conditions, fmt.Sprintf("number_of_class = $%d", argIdx))
		args = append(args, *filter.NumberOfClass)
		argIdx++
	}

	where := "WHERE " + joinAnd(conditions)

	countQuery := "SELECT COUNT(*) FROM classes " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, school_id, number_of_class, suffixes_of_class,
		       academic_year_start, academic_year_finish,
		       created_at, updated_at, deleted_at
		FROM classes
		%s
		ORDER BY number_of_class, suffixes_of_class
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var classList []*dto.Class
	for rows.Next() {
		var c dto.Class
		if err := rows.Scan(
			&c.ID, &c.SchoolID, &c.NumberOfClass, &c.SuffixesOfClass,
			&c.AcademicYearStart, &c.AcademicYearEnd,
			&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		classList = append(classList, &c)
	}

	return classList, total, rows.Err()
}

// Create создаёт новый класс.
func (r *ClassRepository) Create(ctx context.Context, class *dto.Class) (*dto.Class, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		INSERT INTO classes (id, school_id, number_of_class, suffixes_of_class,
		                     academic_year_start, academic_year_finish, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, school_id, number_of_class, suffixes_of_class,
		          academic_year_start, academic_year_finish,
		          created_at, updated_at, deleted_at
	`

	var c dto.Class
	err := r.pool.QueryRow(ctx, query,
		class.ID, class.SchoolID, class.NumberOfClass, class.SuffixesOfClass,
		class.AcademicYearStart, class.AcademicYearEnd,
		class.CreatedAt, class.UpdatedAt,
	).Scan(
		&c.ID, &c.SchoolID, &c.NumberOfClass, &c.SuffixesOfClass,
		&c.AcademicYearStart, &c.AcademicYearEnd,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, appErr.ErrClassAlreadyExists
		}
		return nil, err
	}

	return &c, nil
}

// Update обновляет данные класса.
func (r *ClassRepository) Update(ctx context.Context, class *dto.Class) (*dto.Class, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		UPDATE classes
		SET number_of_class = $2, suffixes_of_class = $3,
		    academic_year_start = $4, academic_year_finish = $5, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, school_id, number_of_class, suffixes_of_class,
		          academic_year_start, academic_year_finish,
		          created_at, updated_at, deleted_at
	`

	var c dto.Class
	err := r.pool.QueryRow(ctx, query,
		class.ID, class.NumberOfClass, class.SuffixesOfClass,
		class.AcademicYearStart, class.AcademicYearEnd,
	).Scan(
		&c.ID, &c.SchoolID, &c.NumberOfClass, &c.SuffixesOfClass,
		&c.AcademicYearStart, &c.AcademicYearEnd,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrClassNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, appErr.ErrClassAlreadyExists
		}
		return nil, err
	}

	return &c, nil
}

// SoftDelete выполняет мягкое удаление класса.
func (r *ClassRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `UPDATE classes SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appErr.ErrClassNotFound
	}

	return nil
}

// Restore восстанавливает мягко удалённый класс.
func (r *ClassRepository) Restore(ctx context.Context, id uuid.UUID) (*dto.Class, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `
		UPDATE classes
		SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NOT NULL
		RETURNING id, school_id, number_of_class, suffixes_of_class,
		          academic_year_start, academic_year_finish,
		          created_at, updated_at, deleted_at
	`

	var c dto.Class
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.SchoolID, &c.NumberOfClass, &c.SuffixesOfClass,
		&c.AcademicYearStart, &c.AcademicYearEnd,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrClassNotFound
	}
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func joinAnd(parts []string) string {
	result := parts[0]
	for _, p := range parts[1:] {
		result += " AND " + p
	}
	return result
}
