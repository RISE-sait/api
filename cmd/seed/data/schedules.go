package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"api/internal/custom_types"
	"context"
	"database/sql"
	"github.com/google/uuid"
	"log"
	"strings"
	"time"
)

func InsertSchedulesReturnEvents(q *dbSeed.Queries) dbSeed.InsertEventsParams {

	var (
		startAtArray        []time.Time
		endAtArray          []time.Time
		locationNameArray   []string
		createdByEmailArray []string
		updatedByEmailArray []string
		capacityArray       []int32
		scheduleIDArray     []uuid.UUID
	)

	for _, practice := range Practices {

		capacity := practice.Capacity

		for _, schedule := range practice.Schedules {

			var programStartAt time.Time
			if schedule.ProgramStartDate == "" {
				programStartAt = time.Now().Truncate(24 * time.Hour) // Start at midnight today
				log.Printf("Using current date for practice '%s' with no start date", practice.Name)
			} else {
				var err error
				programStartAt, err = time.Parse("2006-01-02 15:04", schedule.ProgramStartDate+" 00:01")
				if err != nil {
					log.Printf("Invalid program start date '%s' for practice '%s': %v", schedule.ProgramStartDate, practice.Name, err)
					continue
				}
			}

			eventStartTimeStr := schedule.EventStartTime + ":00" // Add seconds
			eventStartTime := custom_types.TimeWithTimeZone{Time: eventStartTimeStr}

			eventEndTimeStr := schedule.EventEndTime + ":00" // Add seconds
			eventEndTime := custom_types.TimeWithTimeZone{Time: eventEndTimeStr}

			day := dbSeed.DayEnum(strings.ToUpper(schedule.Day))

			scheduleParam := dbSeed.InsertScheduleParams{
				RecurrenceStartAt: programStartAt, // Already at 00:01
				EventStartTime:    eventStartTime,
				EventEndTime:      eventEndTime,
				Day:               day,
				ProgramName:       sql.NullString{String: practice.Name, Valid: true},
				LocationName:      schedule.Location,
			}

			if schedule.ProgramEndDate != "" {
				programEndAt, err := time.Parse("2006-01-02 15:04", schedule.ProgramEndDate+" 17:00")
				if err != nil {
					log.Printf("Invalid program end date '%s': %v", schedule.ProgramEndDate, err)
					continue
				}

				scheduleParam.RecurrenceEndAt = sql.NullTime{
					Time:  programEndAt, // Set to 17:00
					Valid: true,
				}
			}

			scheduleID, err := q.InsertSchedule(context.Background(), scheduleParam) // Assume batch insert
			if err != nil {
				log.Fatalf("Failed to insert schedules: %v", err)
			}

			// for events

			endDate := programStartAt.AddDate(0, 5, 0) // Default: 5 months from start
			if scheduleParam.RecurrenceEndAt.Valid {
				endDate = scheduleParam.RecurrenceEndAt.Time
			}

			eventStart, err := time.Parse("15:04", schedule.EventStartTime)
			if err != nil {
				log.Printf("Failed to parse event start time: %v", err)
				continue
			}

			eventEnd, err := time.Parse("15:04", schedule.EventEndTime)
			if err != nil {
				log.Printf("Failed to parse event end time: %v", err)
				continue
			}

			// Generate events for each matching day in range
			for d := programStartAt; !d.After(endDate); d = d.AddDate(0, 0, 1) {
				if !strings.EqualFold(d.Weekday().String(), schedule.Day) {
					continue
				}

				// Combine date with time
				startAt := time.Date(
					d.Year(), d.Month(), d.Day(),
					eventStart.Hour(), eventStart.Minute(), 0, 0,
					time.UTC,
				)

				endAt := time.Date(
					d.Year(), d.Month(), d.Day(),
					eventEnd.Hour(), eventEnd.Minute(), 0, 0,
					time.UTC,
				)

				// Append event data
				startAtArray = append(startAtArray, startAt)
				endAtArray = append(endAtArray, endAt)
				locationNameArray = append(locationNameArray, schedule.Location)
				createdByEmailArray = append(createdByEmailArray, "klintlee1@gmail.com")
				updatedByEmailArray = append(updatedByEmailArray, "klintlee1@gmail.com")
				capacityArray = append(capacityArray, int32(capacity))
				scheduleIDArray = append(scheduleIDArray, scheduleID)
			}
		}
	}

	return dbSeed.InsertEventsParams{
		StartAtArray:        startAtArray,
		EndAtArray:          endAtArray,
		LocationNameArray:   locationNameArray,
		CreatedByEmailArray: createdByEmailArray,
		UpdatedByEmailArray: updatedByEmailArray,
		CapacityArray:       capacityArray,
		ScheduleIDArray:     scheduleIDArray,
	}
}
