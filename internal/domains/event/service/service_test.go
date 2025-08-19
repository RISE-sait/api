package service

import (
	"net/http"
	"testing"
	"time"

	values "api/internal/domains/event/values"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateEventsFromRecurrence(t *testing.T) {
	t.Run("Valid input generates events", func(t *testing.T) {
		recurrence := values.CreateRecurrenceValues{
			CreatedBy: uuid.New(),
			BaseRecurrenceValues: values.BaseRecurrenceValues{
				DayOfWeek:       time.Monday,
				FirstOccurrence: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC),  // Monday
				LastOccurrence:  time.Date(2023, 10, 30, 0, 0, 0, 0, time.UTC), // 4 weeks later
				StartTime:       "10:00:00+00:00",
				EndTime:         "12:00:00+00:00",
			},

			ProgramID:  uuid.New(),
			LocationID: uuid.New(),
			TeamID:     uuid.New(),
		}

		events, err := generateEventsFromRecurrence(
			recurrence.FirstOccurrence,
			recurrence.LastOccurrence,
			recurrence.StartTime,
			recurrence.EndTime,
			recurrence.CreatedBy,
			recurrence.ProgramID,
			recurrence.LocationID,
			recurrence.CourtID,
			recurrence.TeamID,
			recurrence.RequiredMembershipPlanID,
			recurrence.PriceID,
			recurrence.DayOfWeek,
		)

		assert.Nil(t, err)
		assert.Len(t, events, 5) // 5 Mondays in the range
		for _, event := range events {
			assert.Equal(t, recurrence.CreatedBy, event.CreatedBy)
			assert.Equal(t, recurrence.ProgramID, event.ProgramID)
			assert.Equal(t, recurrence.LocationID, event.LocationID)
			assert.Equal(t, recurrence.TeamID, event.TeamID)
			assert.Equal(t, time.Monday, event.StartAt.Weekday())
			assert.Equal(t, "10:00", event.StartAt.Format("15:04"))
			assert.Equal(t, "12:00", event.EndAt.Format("15:04"))
		}
	})

	t.Run("Invalid start time format", func(t *testing.T) {
		day := time.Monday

		recurrence := values.CreateRecurrenceValues{
			CreatedBy: uuid.New(),
			BaseRecurrenceValues: values.BaseRecurrenceValues{
				DayOfWeek:       day,
				FirstOccurrence: time.Now(),
				LastOccurrence:  time.Now().AddDate(0, 0, 7),
				StartTime:       "invalid-time",
				EndTime:         "12:00",
			},
			ProgramID:  uuid.New(),
			LocationID: uuid.New(),
			TeamID:     uuid.New(),
		}

		events, err := generateEventsFromRecurrence(
			recurrence.FirstOccurrence,
			recurrence.LastOccurrence,
			recurrence.StartTime,
			recurrence.EndTime,
			recurrence.CreatedBy,
			recurrence.ProgramID,
			recurrence.LocationID,
			recurrence.CourtID,
			recurrence.TeamID,
			recurrence.RequiredMembershipPlanID,
			recurrence.PriceID,
			recurrence.DayOfWeek,
		)

		assert.Nil(t, events)
		assert.NotNil(t, err)
		assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
		assert.Contains(t, err.Message, "Invalid start time format")
	})

	t.Run("Invalid end time format", func(t *testing.T) {
		day := time.Monday

		recurrence := values.CreateRecurrenceValues{
			CreatedBy: uuid.New(),
			BaseRecurrenceValues: values.BaseRecurrenceValues{
				DayOfWeek:       day,
				FirstOccurrence: time.Now(),
				LastOccurrence:  time.Now().AddDate(0, 0, 7),
				StartTime:       "10:00",
				EndTime:         "invalid-time",
			},
			ProgramID:  uuid.New(),
			LocationID: uuid.New(),
			TeamID:     uuid.New(),
		}

		events, err := generateEventsFromRecurrence(
			recurrence.FirstOccurrence,
			recurrence.LastOccurrence,
			recurrence.StartTime,
			recurrence.EndTime,
			recurrence.CreatedBy,
			recurrence.ProgramID,
			recurrence.LocationID,
			recurrence.CourtID,
			recurrence.TeamID,
			recurrence.RequiredMembershipPlanID,
			recurrence.PriceID,
			recurrence.DayOfWeek,
		)

		assert.Nil(t, events)
		assert.NotNil(t, err)
		assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
		assert.Contains(t, err.Message, "Invalid start time format - must be HH:MM:SS±HH:MM (e.g. 09:00:00+00:00)")
	})

	t.Run("End time before start time (crosses midnight)", func(t *testing.T) {
		day := time.Monday

		recurrence := values.CreateRecurrenceValues{
			CreatedBy: uuid.New(),
			BaseRecurrenceValues: values.BaseRecurrenceValues{
				DayOfWeek:       day,
				FirstOccurrence: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC),  // Monday
				LastOccurrence:  time.Date(2023, 10, 30, 0, 0, 0, 0, time.UTC), // 4 weeks later
				StartTime:       "23:00:00+00:00",                              // Changed from "23:00"
				EndTime:         "01:00:00+00:00",                              // Changed from "01:00"
			},
			ProgramID:  uuid.New(),
			LocationID: uuid.New(),
			TeamID:     uuid.New(),
		}

		events, err := generateEventsFromRecurrence(
			recurrence.FirstOccurrence,
			recurrence.LastOccurrence,
			recurrence.StartTime,
			recurrence.EndTime,
			recurrence.CreatedBy,
			recurrence.ProgramID,
			recurrence.LocationID,
			recurrence.CourtID,
			recurrence.TeamID,
			recurrence.RequiredMembershipPlanID,
			recurrence.PriceID,
			recurrence.DayOfWeek,
		)

		assert.Nil(t, err)
		assert.Len(t, events, 5) // 5 Mondays in the range
		for _, event := range events {
			assert.Equal(t, "23:00", event.StartAt.Format("15:04"))
			assert.Equal(t, "01:00", event.EndAt.Format("15:04"))
			assert.True(t, event.EndAt.After(event.StartAt))
		}
	})

	t.Run("No events generated if recurrence period is invalid", func(t *testing.T) {
		recurrence := values.CreateRecurrenceValues{
			CreatedBy: uuid.New(),
			BaseRecurrenceValues: values.BaseRecurrenceValues{
				DayOfWeek:       time.Monday,
				FirstOccurrence: time.Date(2023, 10, 30, 0, 0, 0, 0, time.UTC), // Monday
				LastOccurrence:  time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC),  // Earlier date
				StartTime:       "10:00:00+00:00",                              // Changed from "10:00"
				EndTime:         "12:00:00+00:00",
			},
			ProgramID:  uuid.New(),
			LocationID: uuid.New(),
			TeamID:     uuid.New(),
		}

		events, err := generateEventsFromRecurrence(
			recurrence.FirstOccurrence,
			recurrence.LastOccurrence,
			recurrence.StartTime,
			recurrence.EndTime,
			recurrence.CreatedBy,
			recurrence.ProgramID,
			recurrence.LocationID,
			recurrence.CourtID,
			recurrence.TeamID,
			recurrence.RequiredMembershipPlanID,
			recurrence.PriceID,
			recurrence.DayOfWeek,
		)

		assert.NotNil(t, err)
		assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
		assert.Contains(t, err.Message, "Recurrence start date must be before the end date")
		assert.Empty(t, events)
	})

	t.Run("No matching weekdays in recurrence period", func(t *testing.T) {
		recurrence := values.CreateRecurrenceValues{
			CreatedBy: uuid.New(),
			BaseRecurrenceValues: values.BaseRecurrenceValues{
				DayOfWeek:       time.Sunday,                                  // No Sundays in the range
				FirstOccurrence: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), // Monday
				LastOccurrence:  time.Date(2023, 10, 6, 0, 0, 0, 0, time.UTC), // Friday
				StartTime:       "10:00:00+00:00",                             // Changed from "10:00"
				EndTime:         "12:00:00+00:00",
			},
			ProgramID:  uuid.New(),
			LocationID: uuid.New(),
			TeamID:     uuid.New(),
		}

		events, err := generateEventsFromRecurrence(
			recurrence.FirstOccurrence,
			recurrence.LastOccurrence,
			recurrence.StartTime,
			recurrence.EndTime,
			recurrence.CreatedBy,
			recurrence.ProgramID,
			recurrence.LocationID,
			recurrence.CourtID,
			recurrence.TeamID,
			recurrence.RequiredMembershipPlanID,
			recurrence.PriceID,
			recurrence.DayOfWeek,
		)

		assert.Nil(t, err)
		assert.Empty(t, events)
	})
}
