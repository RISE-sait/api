package event

import (
	"api/internal/custom_types"
	entity "api/internal/domains/event/entity"
	"api/internal/domains/event/persistence/sqlc/generated"
	"api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type RequestDto struct {
	BeginTime  string    `json:"begin_time" validate:"required" example:"07:00:00+00:00"`
	EndTime    string    `json:"end_time" validate:"required" example:"08:00:00+00:00"`
	PracticeID uuid.UUID `json:"practice_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	CourseID   uuid.UUID `json:"course_id" example:"00000000-0000-0000-0000-000000000000"`
	LocationID uuid.UUID `json:"location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	Day        string    `json:"day" validate:"required" example:"MONDAY"`
}

func (dto *RequestDto) validate() (string, string, *errLib.CommonError) {
	if err := validators.ValidateDto(dto); err != nil {
		return "", "", err
	}

	beginTime, err := validators.ParseTime(dto.BeginTime)

	if err != nil {
		return "", "", err
	}

	endTime, err := validators.ParseTime(dto.EndTime)

	if err != nil {
		return "", "", err
	}

	if !db.DayEnum(dto.Day).Valid() {
		validDaysEnum := db.AllDayEnumValues()

		validDays := make([]string, 0, len(validDaysEnum))

		for _, day := range validDaysEnum {
			validDays = append(validDays, string(day))
		}

		return "", "", errLib.New(
			fmt.Sprintf("Invalid day. Valid days are: %v", validDays),
			http.StatusBadRequest,
		)
	}

	return beginTime, endTime, nil
}

func (dto *RequestDto) ToDetails() (*values.EventDetails, *errLib.CommonError) {

	beginTime, endTime, err := dto.validate()

	if err != nil {

		return nil, err
	}

	return &values.EventDetails{

		BeginTime:  custom_types.TimeWithTimeZone{Time: beginTime},
		EndTime:    custom_types.TimeWithTimeZone{Time: endTime},
		PracticeID: dto.PracticeID,
		LocationID: dto.LocationID,
		Day:        dto.Day,
	}, nil
}

func (dto *RequestDto) ToEntity(idStr string) (*entity.Event, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	beginTime, endTime, err := dto.validate()

	if err != nil {

		log.Println("Error: ", err)
		return nil, err
	}

	return &entity.Event{
		ID:         id,
		BeginTime:  custom_types.TimeWithTimeZone{Time: beginTime},
		EndTime:    custom_types.TimeWithTimeZone{Time: endTime},
		LocationID: dto.LocationID,
		Day:        db.DayEnum(dto.Day),
	}, nil
}
