package dto

import (
	"github.com/google/uuid"
	"time"
)

// Subtopic содержит данные подтемы — атомарной единицы обучения (версионируется).
type Subtopic struct {
	ID        uuid.UUID
	TopicID   uuid.UUID
	Title     string
	SortOrder int
	IconKey   *string
	IsActive  bool
	Version   int
	CreatedAt *time.Time
}

// SubtopicFilter параметры фильтрации для списка подтем.
type SubtopicFilter struct {
	TopicID uuid.UUID
	Limit   int
	Offset  int
}
