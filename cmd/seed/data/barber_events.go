package data

import (
	db "api/cmd/seed/sqlc/generated"
	"github.com/google/uuid"
	"math/rand"
	"time"
)

var barberEmails = []string{
	"barber@test.com", "barber.anthony@test.com",
}

// GetBarberEvents generates 50 random schedule events for barbers and clients
func GetBarberEvents(clientIDs []uuid.UUID) db.InsertBarberEventsParams {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Define the time range for scheduling (e.g., next 7 days)
	now := time.Now()
	startDate := now.Add(24 * time.Hour) // Start from tomorrow
	endDate := startDate.Add(7 * 24 * time.Hour)

	// Generate 50 random events
	var beginDateTimeArray []time.Time
	var endDateTimeArray []time.Time
	var customerIDArray []uuid.UUID
	var barberEmailArray []string

	for i := 0; i < 50; i++ {
		// Randomly select a client ID
		clientID := clientIDs[rand.Intn(len(clientIDs))]

		// Randomly select a barber email
		barberEmail := barberEmails[rand.Intn(len(barberEmails))]

		// Generate a random start time within the range
		randomStart := randomTime(r, startDate, endDate)
		randomEnd := randomStart.Add(30 * time.Minute) // Assume each event is 30 minutes long

		// Append to arrays
		beginDateTimeArray = append(beginDateTimeArray, randomStart)
		endDateTimeArray = append(endDateTimeArray, randomEnd)
		customerIDArray = append(customerIDArray, clientID)
		barberEmailArray = append(barberEmailArray, barberEmail)
	}

	// Return the params for the SQL query
	return db.InsertBarberEventsParams{
		BeginDateTimeArray: beginDateTimeArray,
		EndDateTimeArray:   endDateTimeArray,
		CustomerIDArray:    customerIDArray,
		BarberEmailArray:   barberEmailArray,
	}
}

// randomTime generates a random time within a given range
func randomTime(r *rand.Rand, start, end time.Time) time.Time {
	delta := end.Unix() - start.Unix()
	randomSeconds := r.Int63n(delta)
	return start.Add(time.Duration(randomSeconds) * time.Second)
}
