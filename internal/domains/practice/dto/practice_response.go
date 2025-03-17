package dto

import (
	"github.com/google/uuid"
	"time"
)

type PracticeResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Level       string    `json:"level"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PracticeLevelsResponse struct {
	Name []string `json:"levels"`
}
