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
	PayGPrice   string    `json:"pay_as_u_go_price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewCourseResponse(course values.ReadDetails) ResponseDto {

	response := ResponseDto{
		ID:          course.ID,
		Name:        course.Name,
		Description: course.Description,
		CreatedAt:   course.CreatedAt,
		UpdatedAt:   course.UpdatedAt,
	}

	if course.PayGPrice != nil {
		response.PayGPrice = course.PayGPrice.String()
	}

	return response
}
