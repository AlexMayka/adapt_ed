package dto

import (
	"github.com/google/uuid"
	"time"
)

// Program содержит данные учебной программы (курса).
type Program struct {
	ID          uuid.UUID
	SubjectID   uuid.UUID
	GradeNumber int
	Slug        string
	Title       string
	Author      *string
	Textbook    *string
	Description *string
	IsActive    bool
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// ProgramFilter параметры фильтрации для списка программ.
type ProgramFilter struct {
	SubjectID *uuid.UUID
	Grade     *int
	Title     *string
	Limit     int
	Offset    int
}
