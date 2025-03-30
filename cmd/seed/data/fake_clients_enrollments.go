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

	// Create a shuffled copy of client IDs to randomize selection
	shuffledClients := make([]uuid.UUID, len(clientIds))
	copy(shuffledClients, clientIds)
	rand.Shuffle(len(shuffledClients), func(i, j int) {
		shuffledClients[i], shuffledClients[j] = shuffledClients[j], shuffledClients[i]
	})

	for _, eventID := range eventIds {
		// Determine how many customers to assign to this event (up to 50)
		maxCustomers := min(50, len(shuffledClients))
		if maxCustomers == 0 {
			continue
		}

		// Random number of customers for this event (between 1 and maxCustomers)
		numCustomers := 1 + randomGenerator.Intn(maxCustomers)

		// Track which customers we've assigned to this event
		assignedCustomers := make(map[uuid.UUID]bool)

		for i := 0; i < numCustomers; i++ {
			// Find next available client that hasn't been assigned to this event
			var clientID uuid.UUID
			for _, c := range shuffledClients {
				if !assignedCustomers[c] {
					clientID = c
					break
				}
			}
			if clientID == uuid.Nil {
				break // No more unique customers available
			}

			// Mark customer as assigned to this event
			assignedCustomers[clientID] = true

			// Randomize enrollment details
			isCancelled := randomGenerator.Intn(4) == 0 // 25% probability of cancellation

			var checkedInAt time.Time
			if !isCancelled && randomGenerator.Intn(2) == 0 { // 50% probability of checked in if not cancelled
				daysAgo := randomGenerator.Intn(365) // Random day within the last year
				checkedInAt = time.Now().AddDate(0, 0, -daysAgo)
			}

			// Append to the arrays
			clientArray = append(clientArray, clientID)
			eventArray = append(eventArray, eventID)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
