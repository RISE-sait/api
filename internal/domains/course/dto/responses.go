package course

import (
	values "api/internal/domains/course/values"
	"github.com/google/uuid"
	"time"
)

type ResponseDto struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Capacity    int32     `json:"capacity"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewCourseResponse(course values.ReadDetails) ResponseDto {

	return ResponseDto{
		ID:          course.ID,
		Name:        course.Name,
		Description: course.Description,
		Capacity:    course.Capacity,
		CreatedAt:   course.CreatedAt,
		UpdatedAt:   course.UpdatedAt,
	}
}
