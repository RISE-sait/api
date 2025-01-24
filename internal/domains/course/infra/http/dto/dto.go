package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateCourseRequestBody struct {
	Name        string    `json:"name" validate:"notwhitespace"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required"`
}

type UpdateCourseRequest struct {
	Name        string    `json:"name" validate:"notwhitespace"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required"`
	ID          uuid.UUID `json:"id" validate:"required"`
}

type CourseResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}
