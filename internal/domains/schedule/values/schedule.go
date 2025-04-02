package schedule

import (
	"api/internal/custom_types"
	"time"

	"github.com/google/uuid"
)

//goland:noinspection GoNameStartsWithPackageName
type ScheduleDetails struct {
	Day               string
	RecurrenceStartAt time.Time
	RecurrenceEndAt   *time.Time
	EventStartTime    custom_types.TimeWithTimeZone
	EventEndTime      custom_types.TimeWithTimeZone
	ProgramID         uuid.UUID
	LocationID        uuid.UUID
	TeamID            uuid.UUID
}

type CreateScheduleValues struct {
	ScheduleDetails
}

type UpdateScheduleValues struct {
	ID uuid.UUID
	ScheduleDetails
}

type (
	ReadScheduleLocationValues struct {
		ID      uuid.UUID
		Name    string
		Address string
	}

	ReadScheduleProgramValues struct {
		ID   uuid.UUID
		Name string
		Type string
	}

	ReadScheduleTeamValues struct {
		ID   uuid.UUID
		Name string
	}

	ReadScheduleValues struct {
		ID                uuid.UUID
		CreatedAt         time.Time
		UpdatedAt         time.Time
		Day               string
		RecurrenceStartAt time.Time
		RecurrenceEndAt   *time.Time
		EventStartTime    custom_types.TimeWithTimeZone
		EventEndTime      custom_types.TimeWithTimeZone
		ReadScheduleLocationValues
		*ReadScheduleProgramValues
		*ReadScheduleTeamValues
	}
)
