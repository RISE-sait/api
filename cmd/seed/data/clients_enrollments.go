package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"github.com/google/uuid"
	"math/rand"
	"time"
)

func GetClientsEnrollments(clientIds, eventIds []uuid.UUID) dbSeed.InsertCustomersEnrollmentsParams {

	var (
		clientArray      []uuid.UUID
		eventArray       []uuid.UUID
		checkedInAtArray []time.Time
		isCancelledArray []bool
	)

	randomSource := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(randomSource)

	for _, clientID := range clientIds {

		// Decide randomly whether the client will have 1 or 2 events
		numEvents := 1 + randomGenerator.Intn(2) // Randomly 1 or 2 events

		for planIdx := 0; planIdx < numEvents; planIdx++ {

			// Select a random event
			randomEventID := eventIds[randomGenerator.Intn(len(eventIds))]

			isCancelled := randomGenerator.Intn(5) == 0

			var checkedInAt time.Time

			if !isCancelled && randomGenerator.Intn(2) == 0 { // 50% probability if not cancelled
				daysAgo := randomGenerator.Intn(365) // Random day within the last year
				checkedInAt = time.Now().AddDate(0, 0, -daysAgo)
			}

			// Append to the arrays
			clientArray = append(clientArray, clientID)
			eventArray = append(eventArray, randomEventID)
			checkedInAtArray = append(checkedInAtArray, checkedInAt)
			isCancelledArray = append(isCancelledArray, isCancelled)
		}
	}

	return dbSeed.InsertCustomersEnrollmentsParams{
		CustomerIDArray:  clientArray,
		EventIDArray:     eventArray,
		CheckedInAtArray: checkedInAtArray,
		IsCancelledArray: isCancelledArray,
	}
}
