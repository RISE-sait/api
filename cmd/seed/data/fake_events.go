package data

import (
	"math/rand"
	"time"

	dbSeed "api/cmd/seed/sqlc/generated"
)

// GenerateSeedEvents creates mock event data for seeding the database.
// It supports both recurring and one-time event generation.
func GenerateSeedEvents(programs, locations []string, isRecurring bool) dbSeed.InsertEventsParams {
	var (
		startAtArray        []time.Time // Start time of each event
		endAtArray          []time.Time // End time of each event
		locationNameArray   []string    // Location names for events
		createdByEmailArray []string    // Creator email (static for all entries)
		updatedByEmailArray []string    // Updater email (same as creator here)
		programNameArray    []string    // Program name the event belongs to
	)

	for _, program := range programs {
		startDate := time.Now().Truncate(24 * time.Hour) // Today's date, no time component
		endDate := startDate.AddDate(0, 3, 0)            // 3 months from today

		randomHour := rand.Intn(8) + 9                                  // Random hour between 9 AM and 5 PM
		eventStart := time.Date(0, 1, 1, randomHour, 0, 0, 0, time.UTC) // Placeholder time
		eventEnd := eventStart.Add(2 * time.Hour)                       // Each event lasts 2 hours

		randomDay := time.Weekday(rand.Intn(7)) // Random weekday (0–6)

		if isRecurring {
			// Create weekly recurring events on the same weekday
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
			// Create a single event on the startDate
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

	// Return a full InsertEventsParams struct for SQL seeding
	return dbSeed.InsertEventsParams{
		StartAtArray:        startAtArray,
		EndAtArray:          endAtArray,
		LocationNameArray:   locationNameArray,
		CreatedByEmailArray: createdByEmailArray,
		UpdatedByEmailArray: updatedByEmailArray,
		ProgramNameArray:    programNameArray,
	}
}
