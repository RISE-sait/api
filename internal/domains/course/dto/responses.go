package course

import (
	entity "api/internal/domains/course/entity"
	"github.com/google/uuid"
)

type ResponseDto struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}

func NewCourseResponse(course entity.Course) *ResponseDto {
	return &ResponseDto{
		ID:          course.ID,
		Name:        course.Name,
		Description: course.Description,
	}
}
