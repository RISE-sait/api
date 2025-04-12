package event

import (
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type RequestDto struct {
	StartAt    string    `json:"start_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	EndAt      string    `json:"end_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	ProgramID  uuid.UUID `json:"program_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	LocationID uuid.UUID `json:"location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	TeamID     uuid.UUID `json:"team_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	Capacity   int32     `json:"capacity" example:"100"`
}

type CreateRequestDto struct {
	Events []RequestDto `json:"events" validate:"required"`
}

type UpdateRequestDto struct {
	ID uuid.UUID `json:"id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	RequestDto
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
func (dto RequestDto) validate() (time.Time, time.Time, *errLib.CommonError) {

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

func (dto CreateRequestDto) ToCreateEventsValues(creator uuid.UUID) (values.CreateEventsValues, *errLib.CommonError) {

	var events []values.Details

	for _, event := range dto.Events {
		start, end, err := event.validate()

		if err != nil {
			return values.CreateEventsValues{}, err
		}

		if event.Capacity <= 0 {
			return values.CreateEventsValues{}, errLib.New("Capacity must be provided, and greater than 0", http.StatusBadRequest)
		}

		events = append(events, values.Details{
			Capacity:   event.Capacity,
			StartAt:    start,
			EndAt:      end,
			ProgramID:  event.ProgramID,
			LocationID: event.LocationID,
			TeamID:     event.TeamID,
		})
	}

	return values.CreateEventsValues{
		CreatedBy: creator,
		Events:    events,
	}, nil
}

func (dto UpdateRequestDto) ToUpdateEventValues(idStr string, updater uuid.UUID) (values.UpdateEventValues, *errLib.CommonError) {
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		return values.UpdateEventValues{}, err
	}

	startAt, endAt, err := dto.validate()
	if err != nil {
		return values.UpdateEventValues{}, err
	}

	v := values.UpdateEventValues{
		ID:        id,
		UpdatedBy: updater,
		Details: values.Details{
			Capacity:   dto.Capacity,
			StartAt:    startAt,
			EndAt:      endAt,
			ProgramID:  dto.ProgramID,
			LocationID: dto.LocationID,
			TeamID:     dto.TeamID,
		},
	}

	return v, nil
}
