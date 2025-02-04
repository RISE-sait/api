package dto

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleResponse struct {
	ID            uuid.UUID `json:"id"`
	BeginDatetime time.Time `json:"begin_datetime"`
	EndDatetime   time.Time `json:"end_datetime"`
	Course        string    `json:"course"`
	Facility      string    `json:"facility" `
	Day           string    `json:"day" `
}
