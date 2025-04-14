package service

import (
	values "api/internal/domains/event/values"
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateEventsFromRecurrence(t *testing.T) {
	t.Run("Valid input generates events", func(t *testing.T) {

		day := time.Monday

		recurrence := values.CreateEventsRecurrenceValues{
			CreatedBy:         uuid.New(),
			Day:               &day,
			RecurrenceStartAt: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC),  // Monday
			RecurrenceEndAt:   time.Date(2023, 10, 30, 0, 0, 0, 0, time.UTC), // 4 weeks later
			EventStartTime:    "10:00",
			EventEndTime:      "12:00",
			ProgramID:         uuid.New(),
			LocationID:        uuid.New(),
			TeamID:            uuid.New(),
			Capacity:          50,
		}

		events, err := generateEventsFromRecurrence(recurrence)

		assert.Nil(t, err)
		assert.Len(t, events, 5) // 5 Mondays in the range
		for _, event := range events {
			assert.Equal(t, recurrence.CreatedBy, event.CreatedBy)
			assert.Equal(t, recurrence.ProgramID, event.ProgramID)
			assert.Equal(t, recurrence.LocationID, event.LocationID)
			assert.Equal(t, recurrence.TeamID, event.TeamID)
			assert.Equal(t, recurrence.Capacity, event.Capacity)
			assert.Equal(t, time.Monday, event.StartAt.Weekday())
			assert.Equal(t, "10:00", event.StartAt.Format("15:04"))
			assert.Equal(t, "12:00", event.EndAt.Format("15:04"))
		}
	})

	t.Run("Invalid start time format", func(t *testing.T) {

		day := time.Monday

		recurrence := values.CreateEventsRecurrenceValues{
			CreatedBy:         uuid.New(),
			Day:               &day,
			RecurrenceStartAt: time.Now(),
			RecurrenceEndAt:   time.Now().AddDate(0, 0, 7),
			EventStartTime:    "invalid-time",
			EventEndTime:      "12:00",
			ProgramID:         uuid.New(),
			LocationID:        uuid.New(),
			TeamID:            uuid.New(),
			Capacity:          50,
		}

		events, err := generateEventsFromRecurrence(recurrence)

		assert.Nil(t, events)
		assert.NotNil(t, err)
		assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
		assert.Contains(t, err.Message, "Invalid start time format")
	})

	t.Run("Invalid end time format", func(t *testing.T) {

		day := time.Monday

		recurrence := values.CreateEventsRecurrenceValues{
			CreatedBy:         uuid.New(),
			Day:               &day,
			RecurrenceStartAt: time.Now(),
			RecurrenceEndAt:   time.Now().AddDate(0, 0, 7),
			EventStartTime:    "10:00",
			EventEndTime:      "invalid-time",
			ProgramID:         uuid.New(),
			LocationID:        uuid.New(),
			TeamID:            uuid.New(),
			Capacity:          50,
		}

		events, err := generateEventsFromRecurrence(recurrence)

		assert.Nil(t, events)
		assert.NotNil(t, err)
		assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
		assert.Contains(t, err.Message, "Invalid end time format")
	})

	t.Run("End time before start time (crosses midnight)", func(t *testing.T) {

		day := time.Monday

		recurrence := values.CreateEventsRecurrenceValues{
			CreatedBy:         uuid.New(),
			Day:               &day,
			RecurrenceStartAt: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC),  // Monday
			RecurrenceEndAt:   time.Date(2023, 10, 30, 0, 0, 0, 0, time.UTC), // 4 weeks later
			EventStartTime:    "23:00",
			EventEndTime:      "01:00",
			ProgramID:         uuid.New(),
			LocationID:        uuid.New(),
			TeamID:            uuid.New(),
			Capacity:          50,
		}

		events, err := generateEventsFromRecurrence(recurrence)

		assert.Nil(t, err)
		assert.Len(t, events, 5) // 5 Mondays in the range
		for _, event := range events {
			assert.Equal(t, "23:00", event.StartAt.Format("15:04"))
			assert.Equal(t, "01:00", event.EndAt.Format("15:04"))
			assert.True(t, event.EndAt.After(event.StartAt))
		}
	})

	t.Run("No events generated if recurrence period is invalid", func(t *testing.T) {

		day := time.Monday

		recurrence := values.CreateEventsRecurrenceValues{
			CreatedBy:         uuid.New(),
			Day:               &day,
			RecurrenceStartAt: time.Date(2023, 10, 30, 0, 0, 0, 0, time.UTC), // Monday
			RecurrenceEndAt:   time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC),  // Earlier date
			EventStartTime:    "10:00",
			EventEndTime:      "12:00",
			ProgramID:         uuid.New(),
			LocationID:        uuid.New(),
			TeamID:            uuid.New(),
			Capacity:          50,
		}

		events, err := generateEventsFromRecurrence(recurrence)

		assert.NotNil(t, err)
		assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
		assert.Contains(t, err.Message, "Recurrence start date must be before the end date")
		assert.Empty(t, events)
	})

	t.Run("Single-day recurrence period", func(t *testing.T) {
		recurrence := values.CreateEventsRecurrenceValues{
			CreatedBy:         uuid.New(),
			Day:               nil,
			RecurrenceStartAt: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), // Monday
			RecurrenceEndAt:   time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), // Same day
			EventStartTime:    "10:00",
			EventEndTime:      "12:00",
			ProgramID:         uuid.New(),
			LocationID:        uuid.New(),
			TeamID:            uuid.New(),
			Capacity:          50,
		}

		events, err := generateEventsFromRecurrence(recurrence)

		assert.Nil(t, err)
		assert.Len(t, events, 1)
		assert.Equal(t, time.Monday, events[0].StartAt.Weekday())
	})

	t.Run("No matching weekdays in recurrence period", func(t *testing.T) {

		day := time.Sunday

		recurrence := values.CreateEventsRecurrenceValues{
			CreatedBy:         uuid.New(),
			Day:               &day,                                         // No Sundays in the range
			RecurrenceStartAt: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), // Monday
			RecurrenceEndAt:   time.Date(2023, 10, 6, 0, 0, 0, 0, time.UTC), // Friday
			EventStartTime:    "10:00",
			EventEndTime:      "12:00",
			ProgramID:         uuid.New(),
			LocationID:        uuid.New(),
			TeamID:            uuid.New(),
			Capacity:          50,
		}

		events, err := generateEventsFromRecurrence(recurrence)

		assert.Nil(t, err)
		assert.Empty(t, events)
	})
}
