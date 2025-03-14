package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"github.com/google/uuid"
	"math/rand"
	"time"
)

func GetEvents(practices, courses, games, locations []uuid.UUID) dbSeed.InsertEventsParams {
	var events dbSeed.InsertEventsParams

	const sessionTimeLayout = "15:04:05-07:00" // "HH:mm:ss+00:00"
	const sessionDuration = 120                // Duration for each session in minutes

	// Placeholder for generated data
	var (
		eventStartAtArray []time.Time
		eventEndAtArray   []time.Time
		practiceIDArray   []uuid.UUID
		courseIDArray     []uuid.UUID
		gameIDArray       []uuid.UUID
		locationIDArray   []uuid.UUID
	)

	numEvents := 10
	numSessions := 15 // Number of sessions per event

	months := 6

	for m := 1; m < months+1; m++ {

		for e := 0; e < numEvents; e++ {
			// Set event start time and end time for each event
			eventStart := time.Now().Add(time.Duration(rand.Intn(m*30)) * 24 * time.Hour) // Random start date within 30 days
			eventEnd := eventStart.Add(30 * 24 * time.Hour)                               // Random course duration between 90 to 180 days

			for i := 0; i < numSessions; i++ {
				// Randomly determine session date and time within the course duration
				sessionDate := eventStart.Add(time.Duration(rand.Int63n(eventEnd.Unix()-eventStart.Unix())) * time.Second)
				sessionStartHour := rand.Intn(12) + 8 // Random session start time between 8 AM and 8 PM

				sessionStartMinute := 1
				sessionStart := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(), sessionStartHour, sessionStartMinute, 0, 0, time.UTC)

				sessionEndHour := sessionStart.Hour()
				sessionEndMinute := 0
				sessionEnd := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(), sessionEndHour, sessionEndMinute, 0, 0, time.UTC)
				sessionEnd = sessionEnd.Add(time.Duration(sessionDuration) * time.Minute) // Add session duration

				eventStartAtArray = append(eventStartAtArray, eventStart)
				eventEndAtArray = append(eventEndAtArray, eventEnd)

				locationIDArray = append(locationIDArray, locations[rand.Intn(len(locations))]) // Random location assignment

				// Randomly assign foreign keys for practice, course, or game
				switch rand.Intn(3) {
				case 0:
					courseIDArray = append(courseIDArray, courses[rand.Intn(len(courses))]) // Random course
					practiceIDArray = append(practiceIDArray, uuid.Nil)
					gameIDArray = append(gameIDArray, uuid.Nil)
				case 1:
					practiceIDArray = append(practiceIDArray, practices[rand.Intn(len(practices))]) // Random practice
					courseIDArray = append(courseIDArray, uuid.Nil)
					gameIDArray = append(gameIDArray, uuid.Nil)
				default:
					gameIDArray = append(gameIDArray, games[rand.Intn(len(games))]) // Random game
					courseIDArray = append(courseIDArray, uuid.Nil)
					practiceIDArray = append(practiceIDArray, uuid.Nil)
				}
			}
		}

		// Return the generated events data
		events = dbSeed.InsertEventsParams{
			EventStartAtArray: eventStartAtArray,
			EventEndAtArray:   eventEndAtArray,
			PracticeIDArray:   practiceIDArray,
			CourseIDArray:     courseIDArray,
			GameIDArray:       gameIDArray,
			LocationIDArray:   locationIDArray,
		}

	}

	return events

}
