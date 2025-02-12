package entity

import (
	"github.com/google/uuid"
)

type Course struct {
	ID          uuid.UUID
	Name        string
	Description string
}
