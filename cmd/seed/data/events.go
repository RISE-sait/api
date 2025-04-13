package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"log"
	"math/rand"
	"strings"
	"time"
)

func GetEvents() dbSeed.InsertEventsParams {

	var (
		startAtArray        []time.Time
		endAtArray          []time.Time
		locationNameArray   []string
		createdByEmailArray []string
		updatedByEmailArray []string
		programNameArray    []string
		capacityArray       []int32
	)

	for _, practice := range Practices {

		capacity := practice.Capacity

		for _, schedule := range practice.Schedules {

			var programStartAt, programEndAt time.Time

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

			if schedule.ProgramEndDate != "" {
				end, err := time.Parse("2006-01-02 15:04", schedule.ProgramEndDate+" 17:00")
				if err != nil {
					log.Printf("Invalid program end date '%s': %v", schedule.ProgramEndDate, err)
					continue
				}

				programEndAt = end
			}

			endDate := programStartAt.AddDate(0, 5, 0) // Default: 5 months from start
			if !programEndAt.IsZero() {
				endDate = programEndAt
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
				programNameArray = append(programNameArray, practice.Name)
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
		ProgramNameArray:    programNameArray,
	}
}

func GetFakeEvents(programs, locations []string) dbSeed.InsertEventsParams {

	var (
		startAtArray        []time.Time
		endAtArray          []time.Time
		locationNameArray   []string
		createdByEmailArray []string
		updatedByEmailArray []string
		programNameArray    []string
		capacityArray       []int32
	)

	for _, game := range programs {

		programStartAt := time.Now().Truncate(24 * time.Hour) // Start at midnight today

		programEndAt := programStartAt.AddDate(0, 5, 0) // Default: 5 months from start

		randomHour := rand.Intn(13) + 8 // Random hour between 8 and 20
		randomMinute := rand.Intn(60)   // Random minute between 0 and 59
		eventStart := time.Date(0, 1, 1, randomHour, randomMinute, 0, 0, time.UTC)

		eventEnd := eventStart.Add(2 * time.Hour) // Default event duration: 2 hours

		randomDay := time.Weekday(rand.Intn(7)) // Random value between 0 (Sunday) and 6 (Saturday)

		// Generate events for each matching day in range
		for d := programStartAt; !d.After(programEndAt); d = d.AddDate(0, 0, 1) {
			if d.Weekday() != randomDay {
				continue
			}

			randomLocation := locations[rand.Intn(len(locations))]

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
			locationNameArray = append(locationNameArray, randomLocation)
			createdByEmailArray = append(createdByEmailArray, "klintlee1@gmail.com")
			updatedByEmailArray = append(updatedByEmailArray, "klintlee1@gmail.com")
			capacityArray = append(capacityArray, int32(40))
			programNameArray = append(programNameArray, game)
		}
	}

	return dbSeed.InsertEventsParams{
		StartAtArray:        startAtArray,
		EndAtArray:          endAtArray,
		LocationNameArray:   locationNameArray,
		CreatedByEmailArray: createdByEmailArray,
		UpdatedByEmailArray: updatedByEmailArray,
		CapacityArray:       capacityArray,
		ProgramNameArray:    programNameArray,
	}
}
