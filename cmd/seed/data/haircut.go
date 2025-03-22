package data

import (
	db "api/cmd/seed/sqlc/generated"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"math/rand"
	"time"
)

var barberEmails = []string{
	"barber@test.com", "barber.anthony@test.com",
}

var services = []string{
	"Regular Haircut", "Beard Trim", "Razor", "Buzz",
}

// GetHaircutEvents generates 50 random schedule events for barbers and clients
func GetHaircutEvents(clientIDs []uuid.UUID) db.InsertHaircutEventsParams {
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
	var haircutNamesArray []string

	for i := 0; i < 50; i++ {
		// Randomly select a client ID
		clientID := clientIDs[rand.Intn(len(clientIDs))]

		// Randomly select a barber email
		barberEmail := barberEmails[rand.Intn(len(barberEmails))]

		haircut := services[rand.Intn(len(services))]

		// Generate a random start time within the range
		randomStart := randomTime(r, startDate, endDate)
		randomEnd := randomStart.Add(30 * time.Minute) // Assume each event is 30 minutes long

		// Append to arrays
		beginDateTimeArray = append(beginDateTimeArray, randomStart)
		endDateTimeArray = append(endDateTimeArray, randomEnd)
		customerIDArray = append(customerIDArray, clientID)
		barberEmailArray = append(barberEmailArray, barberEmail)
		haircutNamesArray = append(haircutNamesArray, haircut)
	}

	// Return the params for the SQL query
	return db.InsertHaircutEventsParams{
		BeginDateTimeArray: beginDateTimeArray,
		EndDateTimeArray:   endDateTimeArray,
		CustomerIDArray:    customerIDArray,
		BarberEmailArray:   barberEmailArray,
		HaircutNameArray:   haircutNamesArray,
	}
}

func GetHaircutServices() db.InsertHaircutServicesParams {

	var prices = []float64{
		25.00, // Regular Haircut
		15.00, // Beard Trim
		20.00, // Razor
		18.00, // Buzz Cut
	}

	var durations = []int32{
		30, // Regular Haircut (30 minutes)
		15, // Beard Trim (15 minutes)
		20, // Razor (20 minutes)
		25, // Buzz Cut (25 minutes)
	}

	var descriptions = []string{
		"Standard men's haircut with styling",
		"Trimming and shaping of the beard",
		"Straight razor shave for a clean look",
		"Short, even-length buzz cut",
	}

	var (
		pricesArray []decimal.Decimal
	)

	for _, price := range prices {
		pricesArray = append(pricesArray, decimal.NewFromFloat(price))
	}

	return db.InsertHaircutServicesParams{
		NameArray:          services,
		DescriptionArray:   descriptions,
		PriceArray:         pricesArray,
		DurationInMinArray: durations,
	}
}

func GetBarberServices() db.InsertBarberServicesParams {

	var barberEmailArray []string
	var serviceNameArray []string

	// Randomly assign a barber to each service
	for _, service := range services {
		for _, barber := range barberEmails {
			barberEmailArray = append(barberEmailArray, barber)
			serviceNameArray = append(serviceNameArray, service)
		}
	}

	return db.InsertBarberServicesParams{
		BarberEmailArray: barberEmailArray,
		ServiceNameArray: serviceNameArray,
	}
}

// randomTime generates a random time within a given range
func randomTime(r *rand.Rand, start, end time.Time) time.Time {
	delta := end.Unix() - start.Unix()
	randomSeconds := r.Int63n(delta)
	return start.Add(time.Duration(randomSeconds) * time.Second)
}
