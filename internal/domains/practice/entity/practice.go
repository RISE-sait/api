package entity

import (
	"github.com/google/uuid"
)

type Practice struct {
	ID          uuid.UUID
	Name        string
	Description string
}
