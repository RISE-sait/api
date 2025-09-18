package dto

import (
	values "api/internal/domains/notification/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

type RegisterPushTokenRequestDto struct {
	ExpoPushToken string `json:"expo_push_token" validate:"required,notwhitespace"`
	DeviceType    string `json:"device_type" validate:"required,oneof=ios android"`
}

func (dto RegisterPushTokenRequestDto) Validate() *errLib.CommonError {
	return validators.ValidateDto(&dto)
}

type SendNotificationRequestDto struct {
	Type   string                 `json:"type" validate:"required,notwhitespace"`
	Title  string                 `json:"title" validate:"required,notwhitespace"`
	Body   string                 `json:"body" validate:"required,notwhitespace"`
	TeamID string                 `json:"team_id" validate:"required,uuid"`
	Data   map[string]interface{} `json:"data"`
}

func (dto SendNotificationRequestDto) Validate() *errLib.CommonError {
	return validators.ValidateDto(&dto)
}

func (dto SendNotificationRequestDto) ToTeamNotification() (values.TeamNotification, *errLib.CommonError) {
	var vo values.TeamNotification
	
	err := validators.ValidateDto(&dto)
	if err != nil {
		return vo, err
	}

	teamID, parseErr := uuid.Parse(dto.TeamID)
	if parseErr != nil {
		return vo, errLib.New("invalid team_id format", 400)
	}

	return values.TeamNotification{
		Type:   dto.Type,
		Title:  dto.Title,
		Body:   dto.Body,
		TeamID: teamID,
		Data:   dto.Data,
	}, nil
}