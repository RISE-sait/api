package practice

import (
	values "api/internal/domains/practice/values"
	"github.com/google/uuid"
	"time"
)

type ResponseDto struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Level       string    `json:"level"`
	PayGPrice   string    `json:"pay_as_u_go_price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewPracticeResponse(practice values.GetPracticeValues) ResponseDto {

	response := ResponseDto{
		ID:          practice.ID,
		Name:        practice.Name,
		Description: practice.Description,
		CreatedAt:   practice.CreatedAt,
		UpdatedAt:   practice.UpdatedAt,
	}

	if practice.PayGPrice != nil {
		response.PayGPrice = practice.PayGPrice.String()
	}

	return response
}

type LevelsResponse struct {
	Name []string `json:"levels"`
}
