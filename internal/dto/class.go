package dto

import (
	"github.com/google/uuid"
	"time"
)

// Class содержит данные класса для передачи между слоями.
type Class struct {
	ID                uuid.UUID
	SchoolID          uuid.UUID
	NumberOfClass     int
	SuffixesOfClass   string
	AcademicYearStart *time.Time
	AcademicYearEnd   *time.Time
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
	DeletedAt         *time.Time
}

// ClassFilter параметры фильтрации и пагинации для списка классов.
type ClassFilter struct {
	SchoolID      uuid.UUID
	NumberOfClass *int
	Limit         int
	Offset        int
}
