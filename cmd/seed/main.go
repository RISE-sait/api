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

func seedClients(ctx context.Context, db *sql.DB) ([]uuid.UUID, error) {
	clients, err := data.GetClients()

	if err != nil {
		return nil, err
	}

	staffs := data.GetStaffsAsClients()

	if err != nil {
		return nil, err
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
		genderArray                 []string
		hasMarketingEmailConsentArr []bool
		hasSMSConsentArray          []bool
	)

	for _, client := range clients {
		countryAlpha2CodeArray = append(countryAlpha2CodeArray, client.CountryAlpha2)
		firstNameArray = append(firstNameArray, client.FirstName)
		lastNameArray = append(lastNameArray, client.LastName)
		genderArray = append(genderArray, client.Gender)
		ageArray = append(ageArray, int32(client.Age))
		parentIDArray = append(parentIDArray, uuid.Nil)
		phoneArray = append(phoneArray, client.Phone)
		emailArray = append(emailArray, client.Email)
		hasMarketingEmailConsentArr = append(hasMarketingEmailConsentArr, client.EmailConsent)
		hasSMSConsentArray = append(hasSMSConsentArray, client.SMSConsent)
	}

	ids, err := seedQueries.InsertClients(ctx, dbSeed.InsertClientsParams{
		CountryAlpha2CodeArray:        countryAlpha2CodeArray,
		FirstNameArray:                firstNameArray,
		LastNameArray:                 lastNameArray,
		AgeArray:                      ageArray,
		GenderArray:                   genderArray,
		ParentIDArray:                 parentIDArray,
		PhoneArray:                    phoneArray,
		EmailArray:                    emailArray,
		HasMarketingEmailConsentArray: hasMarketingEmailConsentArr,
		HasSmsConsentArray:            hasSMSConsentArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert clients: %v", err)
		return nil, err
	}
	seedQueries.InsertClientsMembershipPlans(ctx, dbSeed.InsertClientsMembershipPlansParams{
		CustomerID:       nil,
		PlansArray:       nil,
		StartDateArray:   nil,
		RenewalDateArray: nil,
	})

	seedQueries.InsertClients(ctx, staffs)

	return ids, nil
}

func seedPractices(ctx context.Context, db *sql.DB) ([]uuid.UUID, error) {

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

	createdPractices, err := seedQueries.InsertPractices(ctx, dbSeed.InsertPracticesParams{
		NameArray:        nameArray,
		DescriptionArray: descriptionArray,
		LevelArray:       levelArray,
		CapacityArray:    capacityArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert practices: %v", err)
		return nil, err
	}

	return createdPractices, nil
}

func seedStaffRoles(ctx context.Context, db *sql.DB) error {

	seedQueries := dbSeed.New(db)

	err := seedQueries.InsertStaffRoles(ctx)

	if err != nil {
		log.Fatalf("Failed to insert roles: %v", err)
		return err
	}

	return nil
}

func seedStaff(ctx context.Context, db *sql.DB) error {

	seedQueries := dbSeed.New(db)

	staffs := data.GetStaffs()

	err := seedQueries.InsertStaff(ctx, staffs)

	if err != nil {
		log.Fatalf("Failed to insert roles: %v", err)
		return err
	}

	return nil
}

func seedCourses(ctx context.Context, db *sql.DB) ([]uuid.UUID, error) {

	seedQueries := dbSeed.New(db)

	createdCourses, err := seedQueries.InsertCourses(ctx, data.GetCourses())

	if err != nil {
		log.Fatalf("Failed to insert courses: %v", err)
		return nil, err
	}

	return createdCourses, nil
}

func getGames(numGames int) []string {
	names := make([]string, numGames)
	for i := 0; i < numGames; i++ {
		names[i] = data.GenerateGameName(i)
	}
	return names
}

// Seed games into database
func seedGames(ctx context.Context, db *sql.DB) ([]uuid.UUID, error) {
	seedQueries := dbSeed.New(db)

	gamesData := getGames(10) // Generate 20 games

	createdGames, err := seedQueries.InsertGames(ctx, gamesData)

	if err != nil {
		log.Fatalf("Failed to insert games: %v", err)
		return nil, err
	}

	return createdGames, nil
}

func seedMembershipCoursesEligibility(ctx context.Context, db *sql.DB, membershipsIds, courseIds []uuid.UUID) error {
	seedQueries := dbSeed.New(db)

	eligibilityData := data.GetMembershipCoursesEligibility(membershipsIds, courseIds)

	err := seedQueries.InsertCourseMembershipsEligibility(ctx, eligibilityData)

	if err != nil {
		log.Fatalf("Failed to insert membership courses eligibility: %v", err)
		return err
	}

	return nil
}

func seedMembershipPracticeEligibility(ctx context.Context, db *sql.DB, membershipsIds, practiceIds []uuid.UUID) error {
	seedQueries := dbSeed.New(db)

	eligibilityData := data.GetMembershipPracticesEligibility(membershipsIds, practiceIds)

	err := seedQueries.InsertPracticeMembershipsEligibility(ctx, eligibilityData)

	if err != nil {
		log.Fatalf("Failed to insert membership practices eligibility: %v", err)
		return err
	}

	return nil
}

func seedLocations(ctx context.Context, db *sql.DB) ([]uuid.UUID, error) {

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
	createdLocations, err := seedQueries.InsertLocations(ctx, dbSeed.InsertLocationsParams{
		NameArray:    nameArray,
		AddressArray: addressArray,
	})
	if err != nil {
		log.Fatalf("Failed to insert locations batch: %v", err)
		return nil, err
	}

	return createdLocations, nil
}

func seedMembershipPlans(ctx context.Context, db *sql.DB) ([]uuid.UUID, error) {
	seedQueries := dbSeed.New(db)

	var response []uuid.UUID

	for i := 0; i < len(data.Memberships); i++ {

		membershipName := data.Memberships[i].Name
		plans := data.Memberships[i].MembershipPlans

		var (
			nameArray             []string
			priceArray            []decimal.Decimal
			joiningFeeArray       []decimal.Decimal
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

			price := decimal.NewFromFloat(plan.PaymentFrequency.Price)

			nameArray = append(nameArray, plan.PlanName)
			priceArray = append(priceArray, price)
			joiningFeeArray = append(joiningFeeArray, decimal.NewFromFloat(plan.PaymentFrequency.JoiningFee))
			autoRenewArray = append(autoRenewArray, plan.PaymentFrequency.HasEndDate.WillPlanAutoRenew)
			membershipNameArray = append(membershipNameArray, membershipName)
			paymentFrequencyArray = append(paymentFrequencyArray, dbSeed.PaymentFrequency(plan.PaymentFrequency.RecurringPeriod))
			amtPeriodsArray = append(amtPeriodsArray, periods)
		}

		// Perform the batch insert
		createdPlans, err := seedQueries.InsertMembershipPlans(ctx, dbSeed.InsertMembershipPlansParams{
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
			return nil, err
		}

		response = append(response, createdPlans...)
	}

	return response, nil
}

func seedMemberships(ctx context.Context, db *sql.DB) ([]uuid.UUID, error) {

	seedQueries := dbSeed.New(db)

	var (
		nameArray        []string
		descriptionArray []string
	)
	for i := 0; i < len(data.Memberships); i++ {

		nameArray = append(nameArray, data.Memberships[i].Name)
		descriptionArray = append(descriptionArray, data.Memberships[i].Description)
	}

	memberships, err := seedQueries.InsertMemberships(ctx, dbSeed.InsertMembershipsParams{
		NameArray:        nameArray,
		DescriptionArray: descriptionArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert membership plans: %v", err)
		return nil, err
	}

	return memberships, nil
}

func seedEvents(ctx context.Context, db *sql.DB, practices, courses, games, locations []uuid.UUID) ([]uuid.UUID, error) {
	seedQueries := dbSeed.New(db)

	if len(practices) == 0 || len(courses) == 0 || len(games) == 0 || len(locations) == 0 {
		return nil, fmt.Errorf("missing required foreign key data")
	}

	// Insert events and sessions into the database
	ids, err := seedQueries.InsertEvents(ctx, data.GetEvents(practices, courses, games, locations))

	if err != nil {
		log.Fatalf("Failed to insert events: %v", err)
		return nil, err
	}

	return ids, nil
}

func seedClientsMembershipPlans(ctx context.Context, db *sql.DB, clients, plans []uuid.UUID) error {
	seedQueries := dbSeed.New(db)

	_, err := seedQueries.InsertClientsMembershipPlans(ctx, data.GetClientsMembershipPlans(clients, plans))

	if err != nil {
		log.Fatalf("Failed to insert client membership plans: %v", err)
		return err
	}

	return nil
}

func seedClientsEnrollments(ctx context.Context, db *sql.DB, clients, events []uuid.UUID) error {
	seedQueries := dbSeed.New(db)

	_, err := seedQueries.InsertCustomersEnrollments(ctx, data.GetClientsEnrollments(clients, events))

	if err != nil {
		log.Fatalf("Failed to insert client enrollments: %v", err)
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

	practiceIds, err := seedPractices(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}

	courseIds, err := seedCourses(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}

	gameIds, err := seedGames(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}

	locationIds, err := seedLocations(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}

	eventIds, err := seedEvents(ctx, db, practiceIds, courseIds, gameIds, locationIds)

	if err != nil {
		log.Println(err)
		return
	}

	membershipIds, err := seedMemberships(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}

	plansIds, err := seedMembershipPlans(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}

	clientIds, err := seedClients(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}

	err = seedClientsMembershipPlans(ctx, db, clientIds, plansIds)

	if err != nil {
		log.Println(err)
		return
	}

	err = seedClientsEnrollments(ctx, db, clientIds, eventIds)

	if err != nil {
		log.Println(err)
		return
	}

	err = seedMembershipCoursesEligibility(ctx, db, membershipIds, courseIds)

	if err != nil {
		log.Println(err)
		return
	}

	err = seedMembershipPracticeEligibility(ctx, db, membershipIds, practiceIds)

	if err != nil {
		log.Println(err)
		return
	}

	err = seedStaffRoles(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}

	err = seedStaff(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}
}
