package service

import (
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
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
			EventStartTime:    "10:00:00+00:00",
			EventEndTime:      "12:00:00+00:00",
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
		assert.Contains(t, err.Message, "Invalid start time format - must be HH:MM:SS±HH:MM (e.g. 09:00:00+00:00)")
	})

	t.Run("End time before start time (crosses midnight)", func(t *testing.T) {

		day := time.Monday

		recurrence := values.CreateEventsRecurrenceValues{
			CreatedBy:         uuid.New(),
			Day:               &day,
			RecurrenceStartAt: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC),  // Monday
			RecurrenceEndAt:   time.Date(2023, 10, 30, 0, 0, 0, 0, time.UTC), // 4 weeks later
			EventStartTime:    "23:00:00+00:00",                              // Changed from "23:00"
			EventEndTime:      "01:00:00+00:00",                              // Changed from "01:00"
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
			EventStartTime:    "10:00:00+00:00",                              // Changed from "10:00"
			EventEndTime:      "12:00:00+00:00",                              // Changed from "12:00"
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
			EventStartTime:    "10:00:00+00:00",                             // Changed from "10:00"
			EventEndTime:      "12:00:00+00:00",                             // Changed from "12:00"
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
			EventStartTime:    "10:00:00+00:00",                             // Changed from "10:00"
			EventEndTime:      "12:00:00+00:00",                             // Changed from "12:00"
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

var testUUID = uuid.New()
var secondTestUUID = uuid.New() // Add this line
var updatedBy = uuid.New()

func TestConvertEventsForUpdate(t *testing.T) {

	location := time.UTC

	baseEvent := values.ReadEventValues{
		ID:        testUUID,
		StartAt:   time.Date(2023, 6, 15, 10, 0, 0, 0, location), // June 15, 2023 at 10:00
		EndAt:     time.Date(2023, 6, 15, 12, 0, 0, 0, location), // June 15, 2023 at 12:00
		CreatedBy: values.ReadPersonValues{ID: uuid.New()},
	}

	tests := []struct {
		name           string
		timeUpdate     EventTimeUpdate
		idUpdate       EventIDUpdate
		capacity       int32
		existingEvents []values.ReadEventValues
		want           []values.UpdateEventValues
		wantErr        *errLib.CommonError
	}{
		{
			name: "successful time update",
			timeUpdate: EventTimeUpdate{
				NewStartTime: "14:00:00+00:00",
				NewEndTime:   "16:00:00+00:00",
			},
			idUpdate: EventIDUpdate{
				NewProgramID:  testUUID,
				NewLocationID: testUUID,
				NewTeamID:     testUUID,
			},
			capacity:       20,
			existingEvents: []values.ReadEventValues{baseEvent},
			want: []values.UpdateEventValues{
				{
					ID:        testUUID,
					UpdatedBy: updatedBy,
					Details: values.Details{
						StartAt:    time.Date(2023, 6, 15, 14, 0, 0, 0, location),
						EndAt:      time.Date(2023, 6, 15, 16, 0, 0, 0, location),
						ProgramID:  testUUID,
						LocationID: testUUID,
						TeamID:     testUUID,
						Capacity:   20,
					},
				},
			},
		},
		{
			name: "midnight crossing",
			timeUpdate: EventTimeUpdate{
				NewStartTime: "22:00:00+00:00",
				NewEndTime:   "02:00:00+00:00",
			},
			idUpdate: EventIDUpdate{
				NewProgramID:  testUUID,
				NewLocationID: testUUID,
				NewTeamID:     testUUID,
			},
			capacity:       15,
			existingEvents: []values.ReadEventValues{baseEvent},
			want: []values.UpdateEventValues{
				{
					ID:        testUUID,
					UpdatedBy: updatedBy,
					Details: values.Details{
						StartAt:    time.Date(2023, 6, 15, 22, 0, 0, 0, location),
						EndAt:      time.Date(2023, 6, 16, 2, 0, 0, 0, location), // Next day
						ProgramID:  testUUID,
						LocationID: testUUID,
						TeamID:     testUUID,
						Capacity:   15,
					},
				},
			},
		},
		{
			name: "multiple events",
			timeUpdate: EventTimeUpdate{
				NewStartTime: "09:00:00+00:00",
				NewEndTime:   "11:00:00+00:00",
			},
			idUpdate: EventIDUpdate{
				NewProgramID:  testUUID,
				NewLocationID: testUUID,
				NewTeamID:     testUUID,
			},
			capacity: 10,
			existingEvents: []values.ReadEventValues{
				{
					ID:        testUUID, // First event with testUUID
					StartAt:   time.Date(2023, 6, 15, 10, 0, 0, 0, location),
					EndAt:     time.Date(2023, 6, 15, 12, 0, 0, 0, location),
					CreatedBy: values.ReadPersonValues{ID: uuid.New()},
				},
				{
					ID:        secondTestUUID, // Second event with different ID
					StartAt:   time.Date(2023, 6, 16, 10, 0, 0, 0, location),
					EndAt:     time.Date(2023, 6, 16, 12, 0, 0, 0, location),
					CreatedBy: values.ReadPersonValues{ID: uuid.New()},
				},
			},
			want: []values.UpdateEventValues{
				{
					ID:        testUUID, // Should match first event's ID
					UpdatedBy: updatedBy,
					Details: values.Details{
						StartAt:    time.Date(2023, 6, 15, 9, 0, 0, 0, location),
						EndAt:      time.Date(2023, 6, 15, 11, 0, 0, 0, location),
						ProgramID:  testUUID,
						LocationID: testUUID,
						TeamID:     testUUID,
						Capacity:   10,
					},
				},
				{
					ID:        secondTestUUID, // Should match second event's ID
					UpdatedBy: updatedBy,
					Details: values.Details{
						StartAt:    time.Date(2023, 6, 16, 9, 0, 0, 0, location),
						EndAt:      time.Date(2023, 6, 16, 11, 0, 0, 0, location),
						ProgramID:  testUUID,
						LocationID: testUUID,
						TeamID:     testUUID,
						Capacity:   10,
					},
				},
			},
		},
		{
			name: "invalid start time format",
			timeUpdate: EventTimeUpdate{
				NewStartTime: "25:00:00+00:00", // Invalid hour
				NewEndTime:   "16:00:00+00:00",
			},
			wantErr: errLib.New("Invalid start time format - must be HH:MM:SS±HH:MM (e.g. 09:00:00+00:00)", http.StatusBadRequest),
		},
		{
			name: "invalid end time format",
			timeUpdate: EventTimeUpdate{
				NewStartTime: "14:00:00+00:00",
				NewEndTime:   "16:60:00+00:00", // Invalid minute
			},
			wantErr: errLib.New("Invalid end time format - must be HH:MM:SS±HH:MM (e.g. 09:00:00+00:00)", http.StatusBadRequest),
		},
		{
			name: "invalid time format (missing timezone)",
			timeUpdate: EventTimeUpdate{
				NewStartTime: "14:00:00", // Missing timezone
				NewEndTime:   "16:00:00+00:00",
			},
			wantErr: errLib.New("Invalid start time format - must be HH:MM:SS±HH:MM (e.g. 09:00:00+00:00)", http.StatusBadRequest),
		},
		{
			name: "invalid timezone format",
			timeUpdate: EventTimeUpdate{
				NewStartTime: "14:00:00+0000", // Invalid timezone format
				NewEndTime:   "16:00:00+00:00",
			},
			wantErr: errLib.New("Invalid start time format - must be HH:MM:SS±HH:MM (e.g. 09:00:00+00:00)", http.StatusBadRequest),
		},
		{
			name: "zero capacity",
			timeUpdate: EventTimeUpdate{
				NewStartTime: "14:00:00+00:00",
				NewEndTime:   "16:00:00+00:00",
			},
			capacity: 0,
			wantErr:  errLib.New("capacity must be positive", http.StatusBadRequest),
		},
		{
			name: "negative capacity",
			timeUpdate: EventTimeUpdate{
				NewStartTime: "14:00:00+00:00",
				NewEndTime:   "16:00:00+00:00",
			},
			capacity: -5,
			wantErr:  errLib.New("capacity must be positive", http.StatusBadRequest),
		},
		{
			name:           "empty events list",
			timeUpdate:     EventTimeUpdate{NewStartTime: "09:00:00+00:00", NewEndTime: "17:00:00+00:00"},
			existingEvents: []values.ReadEventValues{},
			capacity:       5,
			wantErr:        errLib.New("no events to update", http.StatusBadRequest),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertEventsForUpdate(
				tt.timeUpdate,
				tt.idUpdate,
				tt.capacity,
				updatedBy,
				tt.existingEvents,
			)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateTimeKeepingDate(t *testing.T) {
	loc := time.UTC
	original := time.Date(2023, 6, 15, 10, 0, 0, 0, loc)
	newTime, _ := time.Parse("15:04", "14:30")

	got := updateTimeKeepingDate(original, newTime)
	want := time.Date(2023, 6, 15, 14, 30, 0, 0, loc)

	assert.Equal(t, want, got)
}
