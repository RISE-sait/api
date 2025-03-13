package main

import (
	"api/cmd/seed/data"
	dbSeed "api/cmd/seed/sqlc/generated"
	"api/config"
	"github.com/google/uuid"

	//"github.com/google/uuid"
	"github.com/shopspring/decimal"
	//"time"

	//"github.com/google/uuid"
	//"github.com/shopspring/decimal"

	"fmt"
	"strings"

	"context"
	"database/sql"
	_ "github.com/lib/pq" // Add this import
	"log"
)

func clearTables(ctx context.Context, db *sql.DB) error {
	// Define the schemas you want to truncate tables from
	schemas := []string{"public", "location", "users", "course", "barber", "audit", "membership", "waiver"}

	// Build the TRUNCATE query
	var tables []string
	for _, schema := range schemas {
		// Query for tables in the specified schema
		rows, err := db.QueryContext(ctx, "SELECT tablename FROM pg_tables WHERE schemaname = $1", schema)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var table string
			if err := rows.Scan(&table); err != nil {
				return err
			}
			if schema == "public" && table == "goose_db_version" {
				continue
			}
			tables = append(tables, fmt.Sprintf("%s.%s", schema, table))
		}

		if err := rows.Err(); err != nil {
			return err
		}
	}

	// If there are no tables, return early
	if len(tables) == 0 {
		return nil
	}

	// Generate the TRUNCATE statement with CASCADE and RESTART IDENTITY
	truncateQuery := "TRUNCATE TABLE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE"

	// Execute the TRUNCATE query
	_, err := db.ExecContext(ctx, truncateQuery)
	return err
}

func seedClients(ctx context.Context, db *sql.DB) error {
	clients, err := data.GetClients()
	if err != nil {
		return err
	}

	seedQueries := dbSeed.New(db)

	var (
		countryAlpha2CodeArray      []string
		firstNameArray              []string
		lastNameArray               []string
		ageArray                    []int32
		parentIDArray               []uuid.UUID
		phoneArray                  []string
		emailArray                  []string
		hasMarketingEmailConsentArr []bool
		hasSMSConsentArray          []bool
	)

	for _, client := range clients {
		countryAlpha2CodeArray = append(countryAlpha2CodeArray, client.CountryAlpha2)
		firstNameArray = append(firstNameArray, client.FirstName)
		lastNameArray = append(lastNameArray, client.LastName)
		ageArray = append(ageArray, int32(client.Age))
		parentIDArray = append(parentIDArray, uuid.Nil)
		emailArray = append(emailArray, client.Email)
		hasMarketingEmailConsentArr = append(hasMarketingEmailConsentArr, client.EmailConsent)
		hasSMSConsentArray = append(hasSMSConsentArray, client.SMSConsent)
	}

	err = seedQueries.InsertClients(ctx, dbSeed.InsertClientsParams{
		CountryAlpha2CodeArray:        countryAlpha2CodeArray,
		FirstNameArray:                firstNameArray,
		LastNameArray:                 lastNameArray,
		AgeArray:                      ageArray,
		ParentIDArray:                 parentIDArray,
		PhoneArray:                    phoneArray,
		EmailArray:                    emailArray,
		HasMarketingEmailConsentArray: hasMarketingEmailConsentArr,
		HasSmsConsentArray:            hasSMSConsentArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert clients: %v", err)
		return err
	}

	return nil
}

func seedPractices(ctx context.Context, db *sql.DB) error {

	seedQueries := dbSeed.New(db)

	practices := data.Practices

	var (
		nameArray        []string
		descriptionArray []string
		levelArray       []dbSeed.PracticeLevel
		capacityArray    []int32
	)
	for i := 0; i < len(practices); i++ {

		nameArray = append(nameArray, practices[i].Name)
		descriptionArray = append(descriptionArray, practices[i].Description)
		levelArray = append(levelArray, dbSeed.PracticeLevelAll)
		capacityArray = append(capacityArray, int32(practices[i].Capacity))
	}

	err := seedQueries.InsertPractices(ctx, dbSeed.InsertPracticesParams{
		NameArray:        nameArray,
		DescriptionArray: descriptionArray,
		LevelArray:       levelArray,
		CapacityArray:    capacityArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert membership plans: %v", err)
		return err
	}

	return nil
}

func seedLocations(ctx context.Context, db *sql.DB) error {

	seedQueries := dbSeed.New(db)

	var (
		nameArray    []string
		addressArray []string
	)
	for i := 0; i < len(data.Locations); i++ {

		nameArray = append(nameArray, data.Locations[i].Name)
		addressArray = append(addressArray, data.Locations[i].Address)
	}
	// Batch insert
	err := seedQueries.InsertLocations(ctx, dbSeed.InsertLocationsParams{
		NameArray:    nameArray,
		AddressArray: addressArray,
	})
	if err != nil {
		log.Fatalf("Failed to insert batch: %v", err)
		return err
	}

	return nil
}

func seedMembershipPlans(ctx context.Context, db *sql.DB) error {
	seedQueries := dbSeed.New(db)

	for i := 0; i < len(data.Memberships); i++ {

		membershipName := data.Memberships[i].Name
		plans := data.Memberships[i].MembershipPlans

		var (
			nameArray             []string
			priceArray            []string
			joiningFeeArray       []string
			autoRenewArray        []bool
			membershipNameArray   []string
			paymentFrequencyArray []dbSeed.PaymentFrequency
			amtPeriodsArray       []int32
		)

		for _, plan := range plans {
			hasEndDate := plan.PaymentFrequency.HasEndDate
			periods := int32(0)

			if hasEndDate.Value {
				periods = int32(hasEndDate.NoOfPeriods)
			}

			price := decimal.NewFromFloat(plan.PaymentFrequency.Price).String()

			nameArray = append(nameArray, plan.PlanName)
			priceArray = append(priceArray, price)
			joiningFeeArray = append(joiningFeeArray, fmt.Sprintf("%f", plan.PaymentFrequency.JoiningFee))
			autoRenewArray = append(autoRenewArray, plan.PaymentFrequency.HasEndDate.WillPlanAutoRenew)
			membershipNameArray = append(membershipNameArray, membershipName)
			paymentFrequencyArray = append(paymentFrequencyArray, dbSeed.PaymentFrequency(plan.PaymentFrequency.RecurringPeriod))
			amtPeriodsArray = append(amtPeriodsArray, periods)
		}

		// Perform the batch insert
		err := seedQueries.InsertMembershipPlans(ctx, dbSeed.InsertMembershipPlansParams{
			NameArray:             nameArray,
			PriceArray:            priceArray,
			JoiningFeeArray:       joiningFeeArray,
			AutoRenewArray:        autoRenewArray,
			MembershipNameArray:   membershipNameArray,
			PaymentFrequencyArray: paymentFrequencyArray,
			AmtPeriodsArray:       amtPeriodsArray,
		})

		if err != nil {
			log.Fatalf("Failed to insert membership plans: %v", err)
			return err
		}
	}

	return nil
}

func seedMemberships(ctx context.Context, db *sql.DB) error {

	seedQueries := dbSeed.New(db)

	var (
		nameArray        []string
		descriptionArray []string
	)
	for i := 0; i < len(data.Memberships); i++ {

		nameArray = append(nameArray, data.Memberships[i].Name)
		descriptionArray = append(descriptionArray, data.Memberships[i].Description)
	}

	err := seedQueries.InsertMemberships(ctx, dbSeed.InsertMembershipsParams{
		NameArray:        nameArray,
		DescriptionArray: descriptionArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert membership plans: %v", err)
		return err
	}

	return nil
}

func main() {

	ctx := context.Background()

	db := config.GetDBConnection()

	defer db.Close()

	if err := clearTables(ctx, db); err != nil {
		log.Println("Failed to clear tables:", err)
		return
	}

	if err := seedPractices(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err := seedLocations(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err := seedMemberships(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err := seedMembershipPlans(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err := seedClients(ctx, db); err != nil {
		log.Println(err)
		return
	}

}
