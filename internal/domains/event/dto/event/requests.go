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
	StartAt    string `json:"start_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	EndAt      string `json:"end_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	ProgramID  string `json:"program_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	LocationID string `json:"location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	TeamID     string `json:"team_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	Capacity   int32  `json:"capacity" example:"100"`
}

type CreateRequestDto struct {
	RequestDto
	Capacity int32 `json:"capacity" example:"100"`
}

type UpdateRequestDto struct {
	RequestDto
	Capacity *int32 `json:"capacity" example:"100"`
}

// validate validates the request DTO and returns parsed values.
// It performs the following validations:
//   - Validates the DTO structure using validators.ValidateDto
//   - Ensures the StartAt and EndAt strings are valid date-time formats
//   - Verifies ProgramID and LocationID and TeamID are valid UUIDs
//
// Returns:
//   - programID (uuid.UUID): The parsed program UUID
//   - locationID (uuid.UUID): The parsed location UUID
//   - teamID (uuid.UUID): The parsed team UUID
//   - startAt (time.Time): The parsed start date-time
//   - endAt (time.Time): The parsed end date-time
//   - error (*errLib.CommonError): Error information if validation fails, nil otherwise
func (dto RequestDto) validate() (uuid.UUID, uuid.UUID, uuid.UUID, time.Time, time.Time, *errLib.CommonError) {

	if err := validators.ValidateDto(&dto); err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, err
	}

	startAt, err := validators.ParseDateTime(dto.StartAt)

	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, err
	}

	endAt, err := validators.ParseDateTime(dto.EndAt)

	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, err
	}

	programID, err := validators.ParseUUID(dto.ProgramID)

	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, err
	}

	locationID, err := validators.ParseUUID(dto.LocationID)

	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, err
	}

	teamID, err := validators.ParseUUID(dto.TeamID)

	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, err
	}

	return programID, locationID, teamID, startAt, endAt, nil

}

func (dto CreateRequestDto) ToCreateEventValues(creator uuid.UUID) (values.CreateEventValues, *errLib.CommonError) {
	programID, locationID, teamID, startAt, endAt, err := dto.validate()
	if err != nil {
		return values.CreateEventValues{}, err
	}

	if dto.Capacity <= 0 {
		return values.CreateEventValues{}, errLib.New("Capacity must be greater than 0", http.StatusBadRequest)
	}

	return values.CreateEventValues{
		CreatedBy: creator,
		Capacity:  dto.Capacity,
		Details: values.Details{
			StartAt: startAt,
			EndAt:   endAt,
		},
		MutationValues: values.MutationValues{
			ProgramID:  programID,
			LocationID: locationID,
			TeamID:     teamID,
		},
	}, nil
}

func (dto UpdateRequestDto) ToUpdateEventValues(idStr string, updater uuid.UUID) (values.UpdateEventValues, *errLib.CommonError) {
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		return values.UpdateEventValues{}, err
	}

	programID, locationID, teamID, startAt, endAt, err := dto.validate()
	if err != nil {
		return values.UpdateEventValues{}, err
	}

	v := values.UpdateEventValues{
		ID:        id,
		UpdatedBy: updater,
		Details: values.Details{
			StartAt: startAt,
			EndAt:   endAt,
		},
		MutationValues: values.MutationValues{
			ProgramID:  programID,
			LocationID: locationID,
			TeamID:     teamID,
		},
	}

	if dto.Capacity != nil {
		if *dto.Capacity <= 0 {
			return values.UpdateEventValues{}, errLib.New("Capacity must be greater than 0", http.StatusBadRequest)
		}
	}

	v.Capacity = dto.Capacity

	return v, nil
}
