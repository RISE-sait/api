package data

import (
	"math/rand"
	"time"

	dbSeed "api/cmd/seed/sqlc/generated"
)

func GenerateSeedEvents(programs, locations []string, isRecurring bool) dbSeed.InsertEventsParams {
	var (
		startAtArray        []time.Time
		endAtArray          []time.Time
		locationNameArray   []string
		createdByEmailArray []string
		updatedByEmailArray []string
		programNameArray    []string
	)

	for _, program := range programs {
		startDate := time.Now().Truncate(24 * time.Hour)
		endDate := startDate.AddDate(0, 3, 0)

		randomHour := rand.Intn(8) + 9 // between 9AM and 5PM
		eventStart := time.Date(0, 1, 1, randomHour, 0, 0, 0, time.UTC)
		eventEnd := eventStart.Add(2 * time.Hour)

		randomDay := time.Weekday(rand.Intn(7))

		if isRecurring {
			for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
				if d.Weekday() != randomDay {
					continue
				}
				loc := locations[rand.Intn(len(locations))]
				start := time.Date(d.Year(), d.Month(), d.Day(), eventStart.Hour(), 0, 0, 0, time.UTC)
				end := time.Date(d.Year(), d.Month(), d.Day(), eventEnd.Hour(), 0, 0, 0, time.UTC)

				startAtArray = append(startAtArray, start)
				endAtArray = append(endAtArray, end)
				locationNameArray = append(locationNameArray, loc)
				createdByEmailArray = append(createdByEmailArray, "c.davison18@gmail.com")
				updatedByEmailArray = append(updatedByEmailArray, "c.davison18@gmail.com")
				programNameArray = append(programNameArray, program)
			}
		} else {
			loc := locations[rand.Intn(len(locations))]
			start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), eventStart.Hour(), 0, 0, 0, time.UTC)
			end := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), eventEnd.Hour(), 0, 0, 0, time.UTC)

			startAtArray = append(startAtArray, start)
			endAtArray = append(endAtArray, end)
			locationNameArray = append(locationNameArray, loc)
			createdByEmailArray = append(createdByEmailArray, "c.davison18@gmail.com")
			updatedByEmailArray = append(updatedByEmailArray, "c.davison18@gmail.com")
			programNameArray = append(programNameArray, program)
		}
	}

	return dbSeed.InsertEventsParams{
		StartAtArray:        startAtArray,
		EndAtArray:          endAtArray,
		LocationNameArray:   locationNameArray,
		CreatedByEmailArray: createdByEmailArray,
		UpdatedByEmailArray: updatedByEmailArray,
		ProgramNameArray:    programNameArray,
	}
}
