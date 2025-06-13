package playground


import (
	values "api/internal/domains/playground/values"
	"time"
)

type ResponseDto struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func NewResponse(v values.System) ResponseDto {
	return ResponseDto{
		ID:        v.ID.String(),
		Name:      v.Name,
		CreatedAt: v.CreatedAt.Format(time.RFC3339),
		UpdatedAt: v.UpdatedAt.Format(time.RFC3339),
	}
}