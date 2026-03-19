package interest

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"backend/internal/logger/interfaces"
	"backend/internal/utils"
	"context"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// InterestService реализует бизнес-логику работы с интересами.
type InterestService struct {
	log         interfaces.Logger
	interestRep InterestRepository
}

// NewInterestService создаёт сервис интересов.
func NewInterestService(log interfaces.Logger, interestRep InterestRepository) *InterestService {
	return &InterestService{log: log, interestRep: interestRep}
}

// GetInterest возвращает интерес по ID.
func (s *InterestService) GetInterest(ctx context.Context, id uuid.UUID) (*dto.Interest, error) {
	interest, err := s.interestRep.GetByID(ctx, id)

	if errors.Is(err, appErr.ErrInterestNotFound) {
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "интерес не найден")
	}
	if err != nil {
		s.log.Error("ошибка получения интереса", "err", err, "interest_id", id)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	return interest, nil
}

// ListInterests возвращает список интересов.
func (s *InterestService) ListInterests(ctx context.Context, filter dto.InterestFilter) ([]*dto.Interest, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	list, total, err := s.interestRep.List(ctx, filter)
	if err != nil {
		s.log.Error("ошибка получения списка интересов", "err", err)
		return nil, 0, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	return list, total, nil
}

// CreateInterest создаёт новый интерес.
func (s *InterestService) CreateInterest(ctx context.Context, interest *dto.Interest) (*dto.Interest, error) {
	now := time.Now()
	interest.ID = utils.GetUniqUUID()
	interest.IsVerified = true
	interest.CreatedAt = &now

	created, err := s.interestRep.Create(ctx, interest)

	if errors.Is(err, appErr.ErrInterestAlreadyExists) {
		return nil, appErr.NewAppError(http.StatusConflict, appErr.ErrInterestAlreadyExists, "интерес с таким названием уже существует")
	}
	if err != nil {
		s.log.Error("ошибка создания интереса", "err", err)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка создания интереса")
	}

	return created, nil
}

// UpdateInterest обновляет интерес.
func (s *InterestService) UpdateInterest(ctx context.Context, interest *dto.Interest) (*dto.Interest, error) {
	existing, err := s.interestRep.GetByID(ctx, interest.ID)

	if errors.Is(err, appErr.ErrInterestNotFound) {
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "интерес не найден")
	}
	if err != nil {
		s.log.Error("ошибка получения интереса", "err", err, "interest_id", interest.ID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	if interest.Name != "" {
		existing.Name = interest.Name
	}
	if interest.IconKey != nil {
		existing.IconKey = interest.IconKey
	}

	updated, err := s.interestRep.Update(ctx, existing)

	if errors.Is(err, appErr.ErrInterestAlreadyExists) {
		return nil, appErr.NewAppError(http.StatusConflict, appErr.ErrInterestAlreadyExists, "интерес с таким названием уже существует")
	}
	if errors.Is(err, appErr.ErrInterestNotFound) {
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "интерес не найден")
	}
	if err != nil {
		s.log.Error("ошибка обновления интереса", "err", err, "interest_id", interest.ID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка обновления интереса")
	}

	return updated, nil
}

// DeleteInterest удаляет интерес.
func (s *InterestService) DeleteInterest(ctx context.Context, id uuid.UUID) error {
	err := s.interestRep.Delete(ctx, id)

	if errors.Is(err, appErr.ErrInterestNotFound) {
		return appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "интерес не найден")
	}
	if err != nil {
		s.log.Error("ошибка удаления интереса", "err", err, "interest_id", id)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка удаления интереса")
	}

	s.log.Info("интерес удалён", "interest_id", id)
	return nil
}

// VerifyInterests верифицирует интересы по списку ID.
func (s *InterestService) VerifyInterests(ctx context.Context, ids []uuid.UUID) (int, error) {
	count, err := s.interestRep.VerifyBatch(ctx, ids)
	if err != nil {
		s.log.Error("ошибка верификации интересов", "err", err)
		return 0, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка верификации интересов")
	}

	s.log.Info("интересы верифицированы", "count", count)
	return count, nil
}
