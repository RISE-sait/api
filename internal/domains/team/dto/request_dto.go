package team

import (
	values "api/internal/domains/team/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/google/uuid"
)

type RequestDto struct {
	Name     string    `json:"name" validate:"required,notwhitespace"`
	Capacity int32     `json:"capacity" validate:"required,gt=0"`
	CoachID  uuid.UUID `json:"coach_id" example:"faae4b3a-ad9f-463c-ae4b-3aad9fb63c9b"`
	LogoURL  *string   `json:"logo_url,omitempty"`
}

func (dto RequestDto) ToCreateValueObjects() (values.CreateTeamValues, *errLib.CommonError) {

	if err := validators.ValidateDto(&dto); err != nil {
		return values.CreateTeamValues{}, err
	}

	return values.CreateTeamValues{
		Details: values.Details{
			Name:     dto.Name,
			Capacity: dto.Capacity,
			CoachID:  dto.CoachID,
			LogoURL:  dto.LogoURL,
		},
	}, nil
}

func (dto RequestDto) ToUpdateValueObjects(idStr string) (values.UpdateTeamValues, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return values.UpdateTeamValues{}, err
	}

	if err = validators.ValidateDto(&dto); err != nil {
		return values.UpdateTeamValues{}, err
	}

	return values.UpdateTeamValues{
		ID: id,
		TeamDetails: values.Details{
			Name:     dto.Name,
			Capacity: dto.Capacity,
			CoachID:  dto.CoachID,
			LogoURL:  dto.LogoURL,
		},
	}, nil
}
