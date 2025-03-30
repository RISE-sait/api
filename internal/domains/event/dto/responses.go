package event

import (
	"api/internal/domains/event/types"
	values "api/internal/domains/event/values"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

//goland:noinspection GoNameStartsWithPackageName
type (
	ScheduleResponseDto struct {
		*DayResponseDto  `json:",omitempty"`
		DatesResponseDto *[]DateResponseDto `json:"schedule,omitempty"`
	}

	DayResponseDto struct {
		RecurrenceStartAt string  `json:"recurrence_start_at"`
		RecurrenceEndAt   *string `json:"recurrence_end_at,omitempty"`
		SessionStart      string  `json:"session_start_at"`
		SessionEnd        string  `json:"session_end_at"`
		Day               string  `json:"day"`
	}

	DateResponseDto struct {
		StartAt string `json:"start_at"`
		EndAt   string `json:"end_at"`
	}

	CustomerResponseDto struct {
		ID                     uuid.UUID `json:"id"`
		FirstName              string    `json:"first_name"`
		LastName               string    `json:"last_name"`
		Email                  *string   `json:"email,omitempty"`
		Phone                  *string   `json:"phone,omitempty"`
		Gender                 *string   `json:"gender,omitempty"`
		HasCancelledEnrollment bool      `json:"has_cancelled_enrollment"`
	}

	StaffResponseDto struct {
		ID        uuid.UUID `json:"id"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		Email     string    `json:"email,omitempty"`
		Phone     string    `json:"phone,omitempty"`
		Gender    *string   `json:"gender,omitempty"`
		RoleName  string    `json:"role_name"`
	}

	Participants struct {
		Customers []CustomerResponseDto `json:"customers"`
		Staff     []StaffResponseDto    `json:"staff"`
	}

	ProgramInfo struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
		Type string    `json:"type"`
	}

	LocationInfo struct {
		ID      uuid.UUID `json:"id"`
		Name    string    `json:"name"`
		Address string    `json:"address"`
	}

	TeamInfo struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	EventResponseDto struct {
		ID        uuid.UUID    `json:"id"`
		Program   *ProgramInfo `json:"program,omitempty"`
		Location  LocationInfo `json:"location"`
		Capacity  *int32       `json:"capacity,omitempty"`
		CreatedBy uuid.UUID    `json:"created_by"`
		UpdatedBy uuid.UUID    `json:"updated_by"`
		Team      *TeamInfo    `json:"team,omitempty"`
		ScheduleResponseDto
		*Participants
	}
)

func NewEventInfoResponseDto(event values.ReadEventValues, includePeople bool, view types.ViewOption) EventResponseDto {
	response := EventResponseDto{
		ID:        event.ID,
		Capacity:  event.Capacity,
		CreatedBy: event.CreatedBy,
		UpdatedBy: event.UpdatedBy,
	}

	if event.ProgramID != uuid.Nil {
		response.Program = &ProgramInfo{
			ID:   event.ProgramID,
			Name: event.ProgramName,
			Type: event.ProgramType,
		}
	}

	if event.LocationID != uuid.Nil {
		response.Location = LocationInfo{
			ID:      event.LocationID,
			Name:    event.LocationName,
			Address: event.LocationAddress,
		}
	}

	if event.TeamID != uuid.Nil && event.TeamName != "" {
		response.Team = &TeamInfo{
			ID:   event.TeamID,
			Name: event.TeamName,
		}
	}

	if includePeople {
		response.Participants = &Participants{
			Customers: mapCustomers(event.Customers),
			Staff:     mapStaffs(event.Staffs),
		}
	}

	response.ScheduleResponseDto = newScheduleView(event, view)

	return response
}

func newScheduleView(event values.ReadEventValues, view types.ViewOption) ScheduleResponseDto {
	switch view {
	case types.ViewOptionDate:

		dates := calculateEventDates(event)

		return ScheduleResponseDto{
			DatesResponseDto: &dates,
		}
	default:
		response := ScheduleResponseDto{
			DayResponseDto: &DayResponseDto{
				RecurrenceStartAt: event.RecurrenceStartAt.Format(time.RFC3339),
				SessionStart:      event.EventStartTime.Time,
				SessionEnd:        event.EventEndTime.Time,
				Day:               event.Day,
			},
		}

		if event.RecurrenceEndAt != nil {
			end := event.RecurrenceEndAt.Format(time.RFC3339)
			response.DayResponseDto.RecurrenceEndAt = &end
		}

		return response
	}
}

func mapCustomers(customers []values.Customer) []CustomerResponseDto {
	result := make([]CustomerResponseDto, 0, len(customers))
	for _, c := range customers {
		result = append(result, CustomerResponseDto{
			ID:                     c.ID,
			FirstName:              c.FirstName,
			LastName:               c.LastName,
			Email:                  c.Email,
			Phone:                  c.Phone,
			Gender:                 c.Gender,
			HasCancelledEnrollment: c.IsEnrollmentCancelled,
		})
	}
	return result
}

func mapStaffs(staffs []values.Staff) []StaffResponseDto {
	result := make([]StaffResponseDto, 0, len(staffs))
	for _, staff := range staffs {
		result = append(result, StaffResponseDto{
			ID:        staff.ID,
			FirstName: staff.FirstName,
			LastName:  staff.LastName,
			Email:     staff.Email,
			Phone:     staff.Phone,
			Gender:    staff.Gender,
			RoleName:  staff.RoleName,
		})
	}
	return result
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

// Helper to calculate all event dates in range
func calculateEventDates(event values.ReadEventValues) []DateResponseDto {

	const maxDatesForNoEndDate = 20
	const timeFormat = "Monday, Jan 2, 2006 at 15:04"

	results := []DateResponseDto{}

	day, err := parseWeekday(event.Day)

	if err != nil {
		return results
	}

	startTime, err := time.Parse("15:04:05-07:00", event.EventStartTime.Time)
	if err != nil {
		return results
	}

	endTime, err := time.Parse("15:04:05-07:00", event.EventEndTime.Time)
	if err != nil {
		return results
	}

	loc := startTime.Location()
	startH, startM, startS := startTime.Clock()
	endH, endM, endS := endTime.Clock()

	// Find first occurrence of the target weekday
	currentDate := event.RecurrenceStartAt
	for currentDate.Weekday() != day {
		currentDate = currentDate.AddDate(0, 0, 1)
		if event.RecurrenceEndAt != nil && currentDate.After(*event.RecurrenceEndAt) {
			return results
		}
	}

	for dateCount := 0; ; dateCount++ {
		// Break conditions
		if event.RecurrenceEndAt != nil && currentDate.After(*event.RecurrenceEndAt) {
			break
		}
		if event.RecurrenceEndAt == nil && dateCount >= maxDatesForNoEndDate {
			break
		}

		// Create date times
		start := time.Date(
			currentDate.Year(), currentDate.Month(), currentDate.Day(),
			startH, startM, startS, 0, loc,
		)
		end := time.Date(
			currentDate.Year(), currentDate.Month(), currentDate.Day(),
			endH, endM, endS, 0, loc,
		)

		// Handle overnight events
		if end.Before(start) {
			end = end.AddDate(0, 0, 1)
		}

		results = append(results, DateResponseDto{
			StartAt: start.Format(timeFormat),
			EndAt:   end.Format(timeFormat),
		})

		// Move to next week
		currentDate = currentDate.AddDate(0, 0, 7)
	}

	return results
}
