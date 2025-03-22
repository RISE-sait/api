package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

func GetClientsMembershipPlans() (dbSeed.InsertClientsMembershipPlansParams, error) {

	var params dbSeed.InsertClientsMembershipPlansParams

	var (
		clientArray      []string
		planArray        []string
		renewalDateArray []time.Time
		startDateArray   []time.Time
	)

	file, err := os.Open("cmd/seed/clients.csv")
	if err != nil {
		return params, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		return params, fmt.Errorf("error reading CSV: %w", err)
	}

	// Loop through the records (skipping the header row)
	for i, record := range records {
		if i == 0 {
			// Skip the header row
			continue
		}

		lastStartDate := time.Now()

		var renewalDate time.Time
		if record[21] != "" { // Check if the renewal date is not empty
			renewalDate, err = time.Parse("1/2/2006", record[21]) // Adjust the date format as needed
			if err != nil {
				return params, fmt.Errorf("error parsing renewal date: %w", err)
			}
		} else {
			// Use a default "empty" date (e.g., time.Time{} or a specific placeholder)
			renewalDate = time.Time{}
		}

		membershipPlan := record[20]
		email := record[3]

		// Append to the arrays
		clientArray = append(clientArray, email)
		planArray = append(planArray, membershipPlan)
		renewalDateArray = append(renewalDateArray, renewalDate)
		startDateArray = append(startDateArray, lastStartDate)
	}

	return dbSeed.InsertClientsMembershipPlansParams{
		CustomerEmailArray: clientArray,
		MembershipPlanName: planArray,
		StartDateArray:     startDateArray,
		RenewalDateArray:   renewalDateArray,
	}, nil
}
