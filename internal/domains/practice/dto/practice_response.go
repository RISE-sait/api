package practice

import (
	"github.com/google/uuid"
	"time"
)

type Response struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Level       string    `json:"level"`
	Capacity    int32     `json:"capacity"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type LevelsResponse struct {
	Name []string `json:"levels"`
}
