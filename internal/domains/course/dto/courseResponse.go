package dto

import (
	"time"

	"github.com/google/uuid"
)

type CourseResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Capacity    int32     `json:"capacity"`
}
