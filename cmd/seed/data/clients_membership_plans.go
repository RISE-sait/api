package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	dbSeed "api/cmd/seed/sqlc/generated"
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
		if record[20] != "" && record[20] != "0" {
			renewalDate, err = time.Parse("1/2/2006", record[20])
			if err != nil {
				fmt.Printf("Skipping invalid renewal date (%s): %v\n", record[20], err)
				renewalDate = time.Time{} // or continue if you'd rather skip the row
			}
		} else {
			renewalDate = time.Time{}
		}
		

		membershipPlan := record[19]

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
