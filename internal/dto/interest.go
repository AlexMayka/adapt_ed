package dto

import (
	"github.com/google/uuid"
	"time"
)

// Interest содержит данные интереса для передачи между слоями.
type Interest struct {
	ID         uuid.UUID
	Name       string
	IconKey    *string
	IsVerified bool
	CreatedAt  *time.Time
}

// InterestFilter параметры фильтрации и пагинации для списка интересов.
type InterestFilter struct {
	Name         *string
	IsVerified   *bool
	Limit        int
	Offset       int
}
