package event

import (
	"api/internal/custom_types"
	"api/internal/domains/event/types"
	values "api/internal/domains/event/values"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

type ResponseDtoEventWithoutPeople struct {
	DetailsResponseDto
	DaysResponseDto  *[]DayResponseDto  `json:"schedule_days,omitempty"`
	DatesResponseDto *[]DateResponseDto `json:"schedule_dates,omitempty"`
}

type ResponseDtoEventWithPeople struct {
	DetailsResponseDto
	*DayResponseDto  `json:",omitempty"`
	DatesResponseDto *[]DateResponseDto    `json:"dates,omitempty"`
	Customers        []CustomerResponseDto `json:"customers"`
	Staff            []StaffResponseDto    `json:"staff"`
}

type DetailsResponseDto struct {
	ID              uuid.UUID  `json:"id"`
	ProgramID       *uuid.UUID `json:"program_id,omitempty"`
	ProgramName     *string    `json:"program_name,omitempty"`
	ProgramType     *string    `json:"program_type,omitempty"`
	LocationID      uuid.UUID  `json:"location_id,omitempty"`
	LocationName    string     `json:"location_name"`
	LocationAddress string     `json:"location_address"`
	TeamID          *uuid.UUID `json:"team_id,omitempty"`
	TeamName        *string    `json:"team_name,omitempty"`
	Capacity        *int32     `json:"capacity,omitempty"`
}

type DayResponseDto struct {
	ProgramStartAt string `json:"program_start_at"`
	ProgramEndAt   string `json:"program_end_at"`
	SessionStart   string `json:"session_start_at"`
	SessionEnd     string `json:"session_end_at"`
	Day            string `json:"day"`
}

type DateResponseDto struct {
	StartAt string `json:"start_at"`
	EndAt   string `json:"end_at"`
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

func NewEventDetailsResponseDto(event values.ReadEventValues) DetailsResponseDto {
	response := DetailsResponseDto{
		ID:              event.ID,
		LocationID:      event.LocationID,
		LocationName:    event.LocationName,
		LocationAddress: event.LocationAddress,
		Capacity:        event.Capacity,
	}

	if event.ProgramID != uuid.Nil && event.ProgramName != "" && event.ProgramType != "" {
		response.ProgramID = &event.ProgramID
		response.ProgramName = &event.ProgramName
		response.ProgramType = &event.ProgramType
	}

	if event.TeamID != uuid.Nil && event.TeamName != "" {
		response.TeamID = &event.TeamID
		response.TeamName = &event.TeamName
	}

	return response
}

func NewEventDayResponseDto(event values.ReadEventValues) DayResponseDto {
	return DayResponseDto{
		ProgramStartAt: event.ProgramStartAt.String(),
		ProgramEndAt:   event.ProgramEndAt.String(),
		SessionStart:   event.EventStartTime.Time,
		SessionEnd:     event.EventEndTime.Time,
		Day:            event.Day,
	}
}

func NewEventDatesResponseDto(event values.ReadEventValues) []DateResponseDto {
	var results []DateResponseDto

	// Parse the day of week (e.g., "Monday", "Tuesday")
	day, err := parseWeekday(event.Details.Day)
	if err != nil {
		// Handle error (invalid day string)
		return results
	}

	// Calculate all dates for the recurring event within program dates
	eventDates := calculateEventDates(
		event.Details.ProgramStartAt,
		event.Details.ProgramEndAt,
		day,
		event.Details.EventStartTime,
		event.Details.EventEndTime,
	)

	// Convert each date to a DateResponseDto
	for _, ed := range eventDates {
		dto := DateResponseDto{
			StartAt: ed.Start.Format("Monday, Jan 2, 2006 at 15:04"),
			EndAt:   ed.End.Format("Monday, Jan 2, 2006 at 15:04"),
		}

		results = append(results, dto)
	}

	return results
}

func CreateCustomersResponseDto(customers []values.Customer) []CustomerResponseDto {

	var response []CustomerResponseDto

	for _, customer := range customers {
		response = append(response, CustomerResponseDto{
			ID:                     customer.ID,
			FirstName:              customer.FirstName,
			LastName:               customer.LastName,
			Email:                  customer.Email,
			Phone:                  customer.Phone,
			Gender:                 customer.Gender,
			HasCancelledEnrollment: customer.IsEnrollmentCancelled,
		})
	}

	return response
}

func CreateStaffsResponseDto(staffs []values.Staff) []StaffResponseDto {

	var response []StaffResponseDto

	for _, staff := range staffs {
		response = append(response, StaffResponseDto{
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

// Helper to parse weekday from string
func parseWeekday(day string) (time.Weekday, error) {
	switch strings.ToLower(day) {
	case "monday":
		return time.Monday, nil
	case "tuesday":
		return time.Tuesday, nil
	case "wednesday":
		return time.Wednesday, nil
	case "thursday":
		return time.Thursday, nil
	case "friday":
		return time.Friday, nil
	case "saturday":
		return time.Saturday, nil
	case "sunday":
		return time.Sunday, nil
	default:
		return time.Sunday, fmt.Errorf("invalid weekday: %s", day)
	}
}

// todo: write test

// Helper to calculate all event dates in range
func calculateEventDates(programStart, programEnd time.Time, day time.Weekday,
	startTime, endTime custom_types.TimeWithTimeZone) []struct{ Start, End time.Time } {

	var dates []struct{ Start, End time.Time }

	startParsed, err := time.Parse("15:04:05-07:00", startTime.Time)
	if err != nil {
		// Handle error (maybe log and return empty)
		return dates
	}

	endParsed, err := time.Parse("15:04:05-07:00", endTime.Time)
	if err != nil {
		// Handle error
		return dates
	}

	startHour, startMin, startSec := startParsed.Hour(), startParsed.Minute(), startParsed.Second()
	endHour, endMin, endSec := endParsed.Hour(), endParsed.Minute(), endParsed.Second()
	loc := startParsed.Location()

	// Find first occurrence of the target weekday on or after program start
	firstDate := programStart
	for firstDate.Weekday() != day {
		firstDate = firstDate.AddDate(0, 0, 1)
		if firstDate.After(programEnd) {
			return dates
		}
	}

	// Iterate through each week until program end
	for date := firstDate; !date.After(programEnd); date = date.AddDate(0, 0, 7) {
		start := time.Date(
			date.Year(), date.Month(), date.Day(),
			startHour, startMin, startSec, 0,
			loc,
		)

		end := time.Date(
			date.Year(), date.Month(), date.Day(),
			endHour, endMin, endSec, 0,
			loc,
		)

		// Handle events that cross midnight
		if end.Before(start) {
			end = end.AddDate(0, 0, 1)
		}

		dates = append(dates, struct {
			Start time.Time
			End   time.Time
		}{start, end})
	}

	return dates
}

// todo: write test

func TransformEventToDtoWithoutPeople(event values.ReadEventValues, view types.ViewOption) ResponseDtoEventWithoutPeople {
	eventDto := ResponseDtoEventWithoutPeople{
		DetailsResponseDto: NewEventDetailsResponseDto(event),
	}

	if view == types.ViewOptionDate {

		var datesDto []DateResponseDto

		dateDto := NewEventDatesResponseDto(event)

		datesDto = append(datesDto, dateDto...)

		eventDto.DatesResponseDto = &datesDto
	} else {
		var daysDto []DayResponseDto

		dayDto := NewEventDayResponseDto(event)
		daysDto = append(daysDto, dayDto)
		eventDto.DaysResponseDto = &daysDto
	}

	return eventDto
}

// todo: write test

func TransformEventToDtoWithPeople(event values.ReadEventValues, view types.ViewOption) ResponseDtoEventWithPeople {
	var responseDto ResponseDtoEventWithPeople

	if view == types.ViewOptionDay {
		dayResponseDto := NewEventDayResponseDto(event)
		responseDto.DayResponseDto = &dayResponseDto
	} else {
		dateResponseDto := NewEventDatesResponseDto(event)
		responseDto.DatesResponseDto = &dateResponseDto
	}

	responseDto.DetailsResponseDto = NewEventDetailsResponseDto(event)
	responseDto.Customers = CreateCustomersResponseDto(event.Customers)
	responseDto.Staff = CreateStaffsResponseDto(event.Staffs)

	if len(responseDto.Customers) == 0 {
		responseDto.Customers = make([]CustomerResponseDto, 0)
	}

	if len(responseDto.Staff) == 0 {
		responseDto.Staff = make([]StaffResponseDto, 0)
	}

	return responseDto
}
