package interests

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

// InterestRepository предоставляет доступ к данным интересов в PostgreSQL.
type InterestRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// NewInterestRepository создаёт репозиторий интересов.
func NewInterestRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *InterestRepository {
	return &InterestRepository{pool: pool, queryTimeout: queryTimeout}
}

const interestColumns = `id, name, icon_key, is_verified, created_at`

func scanInterest(row pgx.Row) (*dto.Interest, error) {
	var i dto.Interest
	err := row.Scan(&i.ID, &i.Name, &i.IconKey, &i.IsVerified, &i.CreatedAt)
	return &i, err
}

// GetByID возвращает интерес по ID.
func (r *InterestRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.Interest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := fmt.Sprintf(`SELECT %s FROM interests WHERE id = $1`, interestColumns)

	i, err := scanInterest(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrInterestNotFound
	}
	if err != nil {
		return nil, err
	}

	return i, nil
}

// List возвращает список интересов с фильтрацией и пагинацией.
func (r *InterestRepository) List(ctx context.Context, filter dto.InterestFilter) ([]*dto.Interest, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Name+"%")
		argIdx++
	}

	if filter.IsVerified != nil {
		conditions = append(conditions, fmt.Sprintf("is_verified = $%d", argIdx))
		args = append(args, *filter.IsVerified)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM interests " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM interests %s ORDER BY name LIMIT $%d OFFSET $%d`,
		interestColumns, where, argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*dto.Interest
	for rows.Next() {
		i, err := scanInterest(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, i)
	}

	return list, total, rows.Err()
}

// Create создаёт новый интерес.
func (r *InterestRepository) Create(ctx context.Context, interest *dto.Interest) (*dto.Interest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := fmt.Sprintf(`
		INSERT INTO interests (id, name, icon_key, is_verified, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING %s
	`, interestColumns)

	i, err := scanInterest(r.pool.QueryRow(ctx, query,
		interest.ID, interest.Name, interest.IconKey, interest.IsVerified, interest.CreatedAt,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, appErr.ErrInterestAlreadyExists
		}
		return nil, err
	}

	return i, nil
}

// Update обновляет интерес.
func (r *InterestRepository) Update(ctx context.Context, interest *dto.Interest) (*dto.Interest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := fmt.Sprintf(`
		UPDATE interests SET name = $2, icon_key = $3, is_verified = $4
		WHERE id = $1
		RETURNING %s
	`, interestColumns)

	i, err := scanInterest(r.pool.QueryRow(ctx, query,
		interest.ID, interest.Name, interest.IconKey, interest.IsVerified,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErr.ErrInterestNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, appErr.ErrInterestAlreadyExists
		}
		return nil, err
	}

	return i, nil
}

// Delete удаляет интерес (физически — справочник, не soft delete).
func (r *InterestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	tag, err := r.pool.Exec(ctx, `DELETE FROM interests WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appErr.ErrInterestNotFound
	}

	return nil
}

// VerifyBatch верифицирует интересы по списку ID.
func (r *InterestRepository) VerifyBatch(ctx context.Context, ids []uuid.UUID) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query := `UPDATE interests SET is_verified = true WHERE id = ANY($1) AND is_verified = false`

	tag, err := r.pool.Exec(ctx, query, ids)
	if err != nil {
		return 0, err
	}

	return int(tag.RowsAffected()), nil
}
