package haircut_event

import (
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// SetAvailabilityDto represents a request to set barber availability for a specific day
type SetAvailabilityDto struct {
	DayOfWeek int    `json:"day_of_week" validate:"required,min=0,max=6" example:"1"`     // 0=Sunday, 6=Saturday
	StartTime string `json:"start_time" validate:"required" example:"09:00"`             // HH:MM format
	EndTime   string `json:"end_time" validate:"required" example:"17:00"`               // HH:MM format
	IsActive  *bool  `json:"is_active,omitempty" example:"true"`                         // Optional, defaults to true
}

// UpdateAvailabilityDto represents a request to update existing availability
type UpdateAvailabilityDto struct {
	StartTime string `json:"start_time" validate:"required" example:"09:00"`
	EndTime   string `json:"end_time" validate:"required" example:"17:00"`
	IsActive  *bool  `json:"is_active,omitempty" example:"true"`
}

// BulkSetAvailabilityDto allows setting multiple days at once
type BulkSetAvailabilityDto struct {
	Availability []SetAvailabilityDto `json:"availability" validate:"required,dive"`
}

func (dto *SetAvailabilityDto) Validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}

	// Parse and validate time formats
	startTime, err := time.Parse("15:04", dto.StartTime)
	if err != nil {
		return errLib.New("invalid start_time format, expected HH:MM", http.StatusBadRequest)
	}

	endTime, err := time.Parse("15:04", dto.EndTime)
	if err != nil {
		return errLib.New("invalid end_time format, expected HH:MM", http.StatusBadRequest)
	}

	// Validate time range
	if !endTime.After(startTime) {
		return errLib.New("end_time must be after start_time", http.StatusBadRequest)
	}

	return nil
}

func (dto *UpdateAvailabilityDto) Validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}

	// Parse and validate time formats
	startTime, err := time.Parse("15:04", dto.StartTime)
	if err != nil {
		return errLib.New("invalid start_time format, expected HH:MM", http.StatusBadRequest)
	}

	endTime, err := time.Parse("15:04", dto.EndTime)
	if err != nil {
		return errLib.New("invalid end_time format, expected HH:MM", http.StatusBadRequest)
	}

	// Validate time range
	if !endTime.After(startTime) {
		return errLib.New("end_time must be after start_time", http.StatusBadRequest)
	}

	return nil
}

func (dto *BulkSetAvailabilityDto) Validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}

	// Validate each availability entry
	for i, avail := range dto.Availability {
		if err := avail.Validate(); err != nil {
			return errLib.New(
				fmt.Sprintf("availability[%d]: %s", i, err.Error()),
				http.StatusBadRequest,
			)
		}
	}

	return nil
}

// ToCreateParams converts DTO to database parameters
func (dto *SetAvailabilityDto) ToCreateParams(barberID uuid.UUID) CreateAvailabilityParams {
	startTime, _ := time.Parse("15:04", dto.StartTime)
	endTime, _ := time.Parse("15:04", dto.EndTime)
	
	isActive := true
	if dto.IsActive != nil {
		isActive = *dto.IsActive
	}

	return CreateAvailabilityParams{
		BarberID:  barberID,
		DayOfWeek: int32(dto.DayOfWeek),
		StartTime: startTime,
		EndTime:   endTime,
		IsActive:  isActive,
	}
}

// CreateAvailabilityParams represents parameters for creating availability
type CreateAvailabilityParams struct {
	BarberID  uuid.UUID
	DayOfWeek int32
	StartTime time.Time
	EndTime   time.Time
	IsActive  bool
}

// UpdateAvailabilityParams represents parameters for updating availability
type UpdateAvailabilityParams struct {
	ID        uuid.UUID
	StartTime time.Time
	EndTime   time.Time
	IsActive  bool
}