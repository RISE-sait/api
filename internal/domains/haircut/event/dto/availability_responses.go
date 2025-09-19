package haircut_event

import (
	"time"

	"github.com/google/uuid"
)

// AvailabilityResponseDto represents barber availability data
type AvailabilityResponseDto struct {
	ID        uuid.UUID `json:"id"`
	DayOfWeek int32     `json:"day_of_week"` // 0=Sunday, 6=Saturday
	StartTime string    `json:"start_time"`  // HH:MM format
	EndTime   string    `json:"end_time"`    // HH:MM format
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WeeklyAvailabilityResponseDto represents a full week of availability
type WeeklyAvailabilityResponseDto struct {
	BarberID     uuid.UUID                  `json:"barber_id"`
	BarberName   string                     `json:"barber_name,omitempty"`
	Availability []AvailabilityResponseDto  `json:"availability"`
}

// DayNames maps day numbers to names for easier frontend usage
var DayNames = map[int32]string{
	0: "Sunday",
	1: "Monday", 
	2: "Tuesday",
	3: "Wednesday",
	4: "Thursday",
	5: "Friday",
	6: "Saturday",
}

// NewAvailabilityResponse creates a response DTO from database model
func NewAvailabilityResponse(id uuid.UUID, dayOfWeek int32, startTime, endTime time.Time, isActive bool, createdAt, updatedAt time.Time) AvailabilityResponseDto {
	return AvailabilityResponseDto{
		ID:        id,
		DayOfWeek: dayOfWeek,
		StartTime: startTime.Format("15:04"),
		EndTime:   endTime.Format("15:04"),
		IsActive:  isActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}