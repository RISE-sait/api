package dto

import (
	"github.com/google/uuid"
)

type EventResponse struct {
	ID         uuid.UUID `json:"id"`
	BeginTime  string    `json:"begin_time"`
	EndTime    string    `json:"end_time"`
	Course     string    `json:"course"`
	CourseID   uuid.UUID `json:"course_id"`
	Facility   string    `json:"facility" `
	FacilityID uuid.UUID `json:"facility_id" `
	Day        string    `json:"day" `
}
