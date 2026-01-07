package event

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

type RecurrenceRequestDto struct {
	Day                       string      `json:"day" example:"THURSDAY"`
	RecurrenceStartAt         string      `json:"recurrence_start_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	RecurrenceEndAt           string      `json:"recurrence_end_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	EventStartTime            string      `json:"event_start_at" validate:"required" example:"23:00:00+00:00"`
	EventEndTime              string      `json:"event_end_at" validate:"required" example:"23:00:00+00:00"`
	ProgramID                 uuid.UUID   `json:"program_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	LocationID                uuid.UUID   `json:"location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	CourtId                   uuid.UUID   `json:"court_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	TeamID                    uuid.UUID   `json:"team_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	RequiredMembershipPlanIDs []uuid.UUID `json:"required_membership_plan_ids" example:"[\"f0e21457-75d4-4de6-b765-5ee13221fd72\"]"`
	PriceID                   string      `json:"price_id" example:"price_123"`
	CreditCost                *int32      `json:"credit_cost" validate:"omitempty,gte=0" example:"5"`
	RegistrationRequired      *bool       `json:"registration_required" example:"true"` // Defaults to true if not provided
	// Fields for Stripe auto-creation (when PriceID is not provided)
	UnitAmount *int64 `json:"unit_amount" example:"2500"` // Price in cents (e.g., 2500 = $25.00)
	Currency   string `json:"currency" example:"cad"`     // "cad" or "usd", defaults to "cad"
}

//goland:noinspection GoNameStartsWithPackageName
type EventRequestDto struct {
	StartAt                   string      `json:"start_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	EndAt                     string      `json:"end_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	ProgramID                 uuid.UUID   `json:"program_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	LocationID                uuid.UUID   `json:"location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	CourtID                   uuid.UUID   `json:"court_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a111"`
	TeamID                    uuid.UUID   `json:"team_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	RequiredMembershipPlanIDs []uuid.UUID `json:"required_membership_plan_ids" example:"[\"f0e21457-75d4-4de6-b765-5ee13221fd72\"]"`
	PriceID                   string      `json:"price_id" example:"price_123"`
	CreditCost                *int32      `json:"credit_cost" validate:"omitempty,gte=0" example:"5"`
	RegistrationRequired      *bool       `json:"registration_required" example:"true"` // Defaults to true if not provided
	// Fields for Stripe auto-creation (when PriceID is not provided)
	UnitAmount       *int64 `json:"unit_amount" example:"2500"`        // Price in cents (e.g., 2500 = $25.00)
	Currency         string `json:"currency" example:"cad"`           // "cad" or "usd", defaults to "cad"
	SkipNotification *bool  `json:"skip_notification" example:"false"` // Skip auto-notification on update (for minor changes)
}

type DeleteRequestDto struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1"`
}

// validate validates the request DTO and returns parsed values.
// It performs the following validations:
//   - Validates the DTO structure using validators.ValidateDto
//   - Ensures the StartAt and EndAt strings are valid date-time formats
//   - Verifies ProgramID and LocationID and TeamID are valid UUIDs
//
// Returns:
//   - startAt (time.Time): The parsed start date-time
//   - endAt (time.Time): The parsed end date-time
//   - error (*errLib.CommonError): Error information if validation fails, nil otherwise
func (dto EventRequestDto) validate() (time.Time, time.Time, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return time.Time{}, time.Time{}, err
	}

	startAt, err := validators.ParseDateTime(dto.StartAt)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endAt, err := validators.ParseDateTime(dto.EndAt)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startAt, endAt, nil
}

func (dto RecurrenceRequestDto) ToCreateRecurrenceValues(creator uuid.UUID) (values.CreateRecurrenceValues, *errLib.CommonError) {
	recurrence, err := dto.ToBaseRecurrenceValues()
	if err != nil {
		return values.CreateRecurrenceValues{}, err
	}

	// Default registration_required to true if not provided
	registrationRequired := true
	if dto.RegistrationRequired != nil {
		registrationRequired = *dto.RegistrationRequired
	}

	createRecurrenceValues := values.CreateRecurrenceValues{
		CreatedBy:                 creator,
		BaseRecurrenceValues:      recurrence,
		TeamID:                    dto.TeamID,
		LocationID:                dto.LocationID,
		CourtID:                   dto.CourtId,
		ProgramID:                 dto.ProgramID,
		RequiredMembershipPlanIDs: dto.RequiredMembershipPlanIDs,
		PriceID:                   dto.PriceID,
		CreditCost:                dto.CreditCost,
		RegistrationRequired:      registrationRequired,
		UnitAmount:                dto.UnitAmount,
		Currency:                  dto.Currency,
	}

	return createRecurrenceValues, nil
}

func (dto RecurrenceRequestDto) ToUpdateRecurrenceValues(updater, recurrenceID uuid.UUID) (values.UpdateRecurrenceValues, *errLib.CommonError) {
	recurrence, err := dto.ToBaseRecurrenceValues()
	if err != nil {
		return values.UpdateRecurrenceValues{}, err
	}

	// Default registration_required to true if not provided
	registrationRequired := true
	if dto.RegistrationRequired != nil {
		registrationRequired = *dto.RegistrationRequired
	}

	updateRecurrenceValues := values.UpdateRecurrenceValues{
		UpdatedBy:                 updater,
		BaseRecurrenceValues:      recurrence,
		ID:                        recurrenceID,
		TeamID:                    dto.TeamID,
		LocationID:                dto.LocationID,
		CourtID:                   dto.CourtId,
		ProgramID:                 dto.ProgramID,
		RequiredMembershipPlanIDs: dto.RequiredMembershipPlanIDs,
		PriceID:                   dto.PriceID,
		CreditCost:                dto.CreditCost,
		RegistrationRequired:      registrationRequired,
		UnitAmount:                dto.UnitAmount,
		Currency:                  dto.Currency,
	}

	return updateRecurrenceValues, nil
}

func (dto RecurrenceRequestDto) ToBaseRecurrenceValues() (values.BaseRecurrenceValues, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return values.BaseRecurrenceValues{}, err
	}

	recurrenceStartAt, err := validators.ParseDateTime(dto.RecurrenceStartAt)
	if err != nil {
		return values.BaseRecurrenceValues{}, err
	}

	recurrenceEndAt, err := validators.ParseDateTime(dto.RecurrenceEndAt)
	if err != nil {
		return values.BaseRecurrenceValues{}, err
	}

	eventStartTime, err := validators.ParseTime(dto.EventStartTime)
	if err != nil {
		return values.BaseRecurrenceValues{}, err
	}

	eventEndTime, err := validators.ParseTime(dto.EventEndTime)
	if err != nil {
		return values.BaseRecurrenceValues{}, err
	}

	day, err := validateWeekday(dto.Day)
	if err != nil {
		return values.BaseRecurrenceValues{}, err
	}

	return values.BaseRecurrenceValues{
		DayOfWeek:       day,
		FirstOccurrence: recurrenceStartAt,
		LastOccurrence:  recurrenceEndAt,
		StartTime:       eventStartTime,
		EndTime:         eventEndTime,
	}, nil
}

func validateWeekday(day string) (time.Weekday, *errLib.CommonError) {
	// Map of valid weekdays
	weekdays := map[string]time.Weekday{
		"SUNDAY":    time.Sunday,
		"MONDAY":    time.Monday,
		"TUESDAY":   time.Tuesday,
		"WEDNESDAY": time.Wednesday,
		"THURSDAY":  time.Thursday,
		"FRIDAY":    time.Friday,
		"SATURDAY":  time.Saturday,
	}

	// Convert input to uppercase for case-insensitive comparison
	day = strings.ToUpper(day)

	// Check if the input matches a valid weekday
	if weekday, exists := weekdays[day]; exists {
		return weekday, nil
	}

	// Return an error if the input is invalid
	errMsg := fmt.Sprintf("Invalid weekday: %s. Expected one of: SUNDAY, MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY", day)
	return time.Weekday(0), errLib.New(errMsg, http.StatusBadRequest)
}

func (dto EventRequestDto) ToCreateEventValues(creator uuid.UUID) (values.CreateEventValues, *errLib.CommonError) {
	startAt, endAt, err := dto.validate()
	if err != nil {
		return values.CreateEventValues{}, err
	}

	// Default registration_required to true if not provided
	registrationRequired := true
	if dto.RegistrationRequired != nil {
		registrationRequired = *dto.RegistrationRequired
	}

	v := values.CreateEventValues{
		CreatedBy: creator,
		EventDetails: values.EventDetails{
			StartAt:                   startAt,
			EndAt:                     endAt,
			ProgramID:                 dto.ProgramID,
			LocationID:                dto.LocationID,
			CourtID:                   dto.CourtID,
			TeamID:                    dto.TeamID,
			RequiredMembershipPlanIDs: dto.RequiredMembershipPlanIDs,
			PriceID:                   dto.PriceID,
			CreditCost:                dto.CreditCost,
			RegistrationRequired:      registrationRequired,
			UnitAmount:                dto.UnitAmount,
			Currency:                  dto.Currency,
		},
	}

	return v, nil
}

func (dto EventRequestDto) ToUpdateEventValues(idStr string, updater uuid.UUID) (values.UpdateEventValues, *errLib.CommonError) {
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		return values.UpdateEventValues{}, err
	}

	startAt, endAt, err := dto.validate()
	if err != nil {
		return values.UpdateEventValues{}, err
	}

	// Default registration_required to true if not provided
	registrationRequired := true
	if dto.RegistrationRequired != nil {
		registrationRequired = *dto.RegistrationRequired
	}

	// Default skip_notification to false if not provided
	skipNotification := false
	if dto.SkipNotification != nil {
		skipNotification = *dto.SkipNotification
	}

	v := values.UpdateEventValues{
		ID:               id,
		UpdatedBy:        updater,
		SkipNotification: skipNotification,
		EventDetails: values.EventDetails{
			StartAt:                   startAt,
			EndAt:                     endAt,
			ProgramID:                 dto.ProgramID,
			LocationID:                dto.LocationID,
			CourtID:                   dto.CourtID,
			TeamID:                    dto.TeamID,
			RequiredMembershipPlanIDs: dto.RequiredMembershipPlanIDs,
			PriceID:                   dto.PriceID,
			CreditCost:                dto.CreditCost,
			RegistrationRequired:      registrationRequired,
			UnitAmount:                dto.UnitAmount,
			Currency:                  dto.Currency,
		},
	}

	return v, nil
}
