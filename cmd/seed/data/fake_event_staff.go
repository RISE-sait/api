package data

import (
	"math/rand"
	"time"

	dbSeed "api/cmd/seed/sqlc/generated"
	"github.com/google/uuid"
)

func GetEventStaff(eventIds, staffIds []uuid.UUID) dbSeed.InsertEventsStaffParams {
	var (
		eventIDArray []uuid.UUID
		staffIDArray []uuid.UUID
	)

	// Initialize random seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Loop through each event ID and assign a random staff member
	for _, eventID := range eventIds {
		if len(staffIds) == 0 {
			continue // Skip if no staff available
		}

		// Select a random staff ID
		randomIndex := r.Intn(len(staffIds))
		selectedStaffID := staffIds[randomIndex]

		// Append to arrays
		eventIDArray = append(eventIDArray, eventID)
		staffIDArray = append(staffIDArray, selectedStaffID)
	}

	return dbSeed.InsertEventsStaffParams{
		EventIDArray: eventIDArray,
		StaffIDArray: staffIDArray,
	}
}
