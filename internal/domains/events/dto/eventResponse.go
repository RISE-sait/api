package dto

import (
	"github.com/google/uuid"
)

type EventResponse struct {
	ID        uuid.UUID `json:"id"`
	BeginTime string    `json:"begin_time"`
	EndTime   string    `json:"end_time"`
	Course    string    `json:"course"`
	Facility  string    `json:"facility" `
	Day       string    `json:"day" `
}
