package game

import "github.com/google/uuid"

type BaseValue struct {
	Name      string
	VideoLink *string
}
type CreateGameValue struct {
	BaseValue
}

type UpdateGameValue struct {
	ID uuid.UUID
	BaseValue
}

type ReadValue struct {
	ID uuid.UUID
	BaseValue
}
