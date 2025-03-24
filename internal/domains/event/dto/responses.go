package event

import (
	values "api/internal/domains/event/values"
	"log"

	"github.com/google/uuid"
)

type ResponseDto struct {
	ID              uuid.UUID             `json:"id"`
	ProgramStartAt  string                `json:"program_start_at"`
	ProgramEndAt    string                `json:"program_end_at"`
	SessionStart    string                `json:"session_start_at"`
	SessionEnd      string                `json:"session_end_at"`
	Day             string                `json:"day"`
	ProgramID       *uuid.UUID            `json:"program_id,omitempty"`
	ProgramName     *string               `json:"program_name,omitempty"`
	ProgramType     *string               `json:"program_type,omitempty"`
	LocationID      *uuid.UUID            `json:"location_id,omitempty"`
	LocationName    *string               `json:"location_name,omitempty"`
	LocationAddress *string               `json:"location_address,omitempty"`
	Capacity        *int32                `json:"capacity,omitempty"`
	Customers       []CustomerResponseDto `json:"customers"`
	Staff           []StaffResponseDto    `json:"staff"`
}

type CustomerResponseDto struct {
	ID                     uuid.UUID `json:"id"`
	FirstName              string    `json:"first_name"`
	LastName               string    `json:"last_name"`
	Email                  *string   `json:"email,omitempty"`
	Phone                  *string   `json:"phone,omitempty"`
	Gender                 *string   `json:"gender,omitempty"`
	HasCancelledEnrollment bool      `json:"has_cancelled_enrollment"`
}

type StaffResponseDto struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Gender    *string   `json:"gender,omitempty"`
	RoleName  string    `json:"role_name"`
}

func NewEventResponse(event values.ReadEventValues) ResponseDto {
	response := ResponseDto{
		ID:             event.ID,
		ProgramStartAt: event.ProgramStartAt.String(),
		ProgramEndAt:   event.ProgramStartAt.String(),
		SessionStart:   event.EventStartTime.Time,
		SessionEnd:     event.EventEndTime.Time,
		Day:            event.Day,
		Capacity:       event.Capacity,
	}

	if event.ProgramID != uuid.Nil && event.ProgramName != "" && event.ProgramType != "" {
		response.ProgramID = &event.ProgramID
		response.ProgramName = &event.ProgramName
		response.ProgramType = &event.ProgramType
	}

	if event.LocationID != uuid.Nil && event.LocationName != "" && event.LocationAddress != "" {
		response.LocationID = &event.LocationID
		response.LocationName = &event.LocationName
		response.LocationAddress = &event.LocationAddress
	}

	if event.Capacity != nil {
		response.Capacity = event.Capacity
	}

	for _, customer := range event.Customers {
		response.Customers = append(response.Customers, CustomerResponseDto{
			ID:                     customer.ID,
			FirstName:              customer.FirstName,
			LastName:               customer.LastName,
			Email:                  customer.Email,
			Phone:                  customer.Phone,
			Gender:                 customer.Gender,
			HasCancelledEnrollment: customer.IsEnrollmentCancelled,
		})
	}

	log.Println(event.Staffs)

	for _, staff := range event.Staffs {
		response.Staff = append(response.Staff, StaffResponseDto{
			ID:        staff.ID,
			FirstName: staff.FirstName,
			LastName:  staff.LastName,
			Email:     staff.Email,
			Phone:     staff.Phone,
			Gender:    staff.Gender,
			RoleName:  staff.RoleName,
		})
	}

	return response
}
