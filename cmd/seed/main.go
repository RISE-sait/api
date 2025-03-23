package main

import (
	"api/cmd/seed/data"
	dbSeed "api/cmd/seed/sqlc/generated"
	"api/config"
	"api/internal/custom_types"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"time"

	"github.com/shopspring/decimal"

	"fmt"
	"strings"

	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func clearTables(ctx context.Context, db *sql.DB) error {
	// Define the schemas you want to truncate tables from
	schemas := []string{"public", "location", "users", "course", "haircut", "audit", "membership", "waiver"}

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

func seedUsers(ctx context.Context, db *sql.DB) ([]uuid.UUID, error) {
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

	ids, err := seedQueries.InsertUsers(ctx, dbSeed.InsertUsersParams{
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

	_, err = seedQueries.InsertUsers(ctx, staffs)

	if err != nil {
		log.Fatalf("Failed to insert staff as clients: %v", err)
		return nil, err
	}

	return ids, nil
}

func seedHaircutServices(ctx context.Context, db *sql.DB) error {

	seedQueries := dbSeed.New(db)

	services := data.GetHaircutServices()

	if err := seedQueries.InsertHaircutServices(ctx, services); err != nil {
		log.Fatalf("Failed to insert haircut services: %v", err)
		return err
	}

	return nil
}

func seedBarberServices(ctx context.Context, db *sql.DB) error {

	seedQueries := dbSeed.New(db)

	services := data.GetBarberServices()

	if err := seedQueries.InsertBarberServices(ctx, services); err != nil {
		log.Fatalf("Failed to insert barber services: %v", err)
		return err
	}

	return nil
}

func seedHaircutEvents(ctx context.Context, db *sql.DB, clientIds []uuid.UUID) error {

	seedQueries := dbSeed.New(db)

	events := data.GetHaircutEvents(clientIds)

	if err := seedQueries.InsertHaircutEvents(ctx, events); err != nil {
		log.Fatalf("Failed to insert haircut events: %v", err)
		return err
	}

	return nil
}

func seedAthletes(ctx context.Context, db *sql.DB, ids []uuid.UUID) ([]uuid.UUID, error) {

	seedQueries := dbSeed.New(db)

	ids, err := seedQueries.InsertAthletes(ctx, ids)

	if err != nil {
		log.Fatalf("Failed to insert athletes: %v", err)
		return nil, err
	}

	return ids, nil
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

	if err := seedQueries.InsertPractices(ctx, dbSeed.InsertPracticesParams{
		NameArray:        nameArray,
		DescriptionArray: descriptionArray,
		LevelArray:       levelArray,
	}); err != nil {
		log.Fatalf("Failed to insert practices: %v", err)
		return err
	}

	return nil
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

func seedCourses(ctx context.Context, db *sql.DB) error {

	seedQueries := dbSeed.New(db)

	if err := seedQueries.InsertCourses(ctx, data.GetCourses()); err != nil {
		log.Fatalf("Failed to insert courses: %v", err)
		return err
	}

	return nil
}

func getGames(numGames int) []string {
	names := make([]string, numGames)
	for i := 0; i < numGames; i++ {
		names[i] = data.GenerateGameName(i)
	}
	return names
}

func seedGames(ctx context.Context, db *sql.DB) error {
	seedQueries := dbSeed.New(db)

	gamesData := getGames(10) // Generate 20 games

	if err := seedQueries.InsertGames(ctx, gamesData); err != nil {
		log.Fatalf("Failed to insert games: %v", err)
		return err
	}

	return nil
}

//func seedMembershipCoursesEligibility(ctx context.Context, db *sql.DB, membershipsIds, courseIds []uuid.UUID) error {
//	seedQueries := dbSeed.New(db)
//
//	membershipsData := data.Memberships
//
//	err := seedQueries.InsertCourseMembershipsEligibility(ctx, eligibilityData)
//
//	if err != nil {
//		log.Fatalf("Failed to insert membership courses eligibility: %v", err)
//		return err
//	}
//
//	return nil
//}

func seedMembershipPracticeEligibility(ctx context.Context, db *sql.DB) error {
	seedQueries := dbSeed.New(db)

	practicesData := data.Practices

	var practiceNamesArray, membershipNamesArray []string
	var isEligibleArray []bool
	var pricePerBookingArray []decimal.Decimal

	for _, d := range practicesData {

		for _, eligibility := range d.MembershipsEligibility {
			practiceNamesArray = append(practiceNamesArray, d.Name)

			membershipNamesArray = append(membershipNamesArray, eligibility.Name)

			if eligibility.EligibleMembership != nil {
				isEligibleArray = append(isEligibleArray, true)
				pricePerBookingArray = append(pricePerBookingArray, decimal.NewFromInt(int64(eligibility.Price)))
			} else {
				isEligibleArray = append(isEligibleArray, false)
				pricePerBookingArray = append(pricePerBookingArray, decimal.NewFromInt(int64(0)))
			}

		}
	}

	args := dbSeed.InsertPracticeMembershipsEligibilityParams{
		PracticeNamesArray:   practiceNamesArray,
		MembershipNamesArray: membershipNamesArray,
		IsEligibleArray:      isEligibleArray,
		PricePerBookingArray: pricePerBookingArray,
	}

	err := seedQueries.InsertPracticeMembershipsEligibility(ctx, args)

	if err != nil {
		log.Fatalf("Failed to insert membership practices eligibility: %v", err)
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
	if err := seedQueries.InsertLocations(ctx, dbSeed.InsertLocationsParams{
		NameArray:    nameArray,
		AddressArray: addressArray,
	}); err != nil {
		log.Fatalf("Failed to insert locations batch: %v", err)
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
		if err := seedQueries.InsertMembershipPlans(ctx, dbSeed.InsertMembershipPlansParams{
			NameArray:             nameArray,
			PriceArray:            priceArray,
			JoiningFeeArray:       joiningFeeArray,
			AutoRenewArray:        autoRenewArray,
			MembershipNameArray:   membershipNameArray,
			PaymentFrequencyArray: paymentFrequencyArray,
			AmtPeriodsArray:       amtPeriodsArray,
		}); err != nil {
			log.Fatalf("Failed to insert membership plans: %v", err)
			return nil
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

	if err := seedQueries.InsertMemberships(ctx, dbSeed.InsertMembershipsParams{
		NameArray:        nameArray,
		DescriptionArray: descriptionArray,
	}); err != nil {
		log.Fatalf("Failed to insert memberships: %v", err)
		return err
	}

	return nil
}

func seedEvents(ctx context.Context, db *sql.DB) ([]uuid.UUID, error) {
	seedQueries := dbSeed.New(db)

	practices := data.Practices

	var (
		programStartAtArray []time.Time
		programEndAtArray   []time.Time
		eventStartTimeArray []custom_types.TimeWithTimeZone
		eventEndTimeArray   []custom_types.TimeWithTimeZone
		dayArray            []dbSeed.DayEnum
		practiceNameArray   []string
		//courseNameArray       []string
		//gameNameArray         []string
		//locationNameArray     []string
	)

	for _, practice := range practices {

		for _, schedule := range practice.Schedules {

			programStartAtArray = append(programStartAtArray, time.Now())
			programEndAtArray = append(programEndAtArray, time.Now().Add(time.Hour*2440))

			eventStartTime, err := validators.ParseTime(schedule.EventStartTime + ":00+00:00")

			if err != nil {
				log.Fatalf("Failed to parse session start time: %v", err)
				return nil, err
			}

			eventEndTime, err := validators.ParseTime(schedule.EventEndTime + ":00+00:00")

			if err != nil {
				log.Fatalf("Failed to parse session end time: %v", err)
				return nil, err
			}

			eventStartTimeArray = append(eventStartTimeArray, eventStartTime)
			eventEndTimeArray = append(eventEndTimeArray, eventEndTime)

			day := dbSeed.DayEnum(strings.ToUpper(schedule.Day))

			if !day.Valid() {
				log.Fatalf("Invalid day: %v", schedule.Day)
				return nil, err
			}

			dayArray = append(dayArray, day)

			practiceNameArray = append(practiceNameArray, practice.Name)

		}
	}

	arg := dbSeed.InsertEventsParams{
		ProgramStartAtArray: programStartAtArray,
		ProgramEndAtArray:   programEndAtArray,
		EventStartTimeArray: eventStartTimeArray,
		EventEndTimeArray:   eventEndTimeArray,
		DayArray:            dayArray,
		PracticeNameArray:   practiceNameArray,
		CourseNameArray:     nil,
		GameNameArray:       nil,
		LocationNameArray:   nil,
	}

	// Insert events and sessions into the database
	ids, err := seedQueries.InsertEvents(ctx, arg)

	if err != nil {
		log.Fatalf("Failed to insert events: %v", err)
		return nil, err
	}

	return ids, nil
}

func seedClientsMembershipPlans(ctx context.Context, db *sql.DB) error {
	seedQueries := dbSeed.New(db)

	plans, err := data.GetClientsMembershipPlans()

	if err != nil {
		log.Fatalf("Failed to insert client membership plans: %v", err)
		return err
	}

	if err := seedQueries.InsertClientsMembershipPlans(ctx, plans); err != nil {
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

func updateParents(ctx context.Context, db *sql.DB) error {
	seedQueries := dbSeed.New(db)

	rows, err := seedQueries.UpdateParents(ctx)

	if err != nil {
		log.Fatalf("Failed to insert client enrollments: %v", err)
		return err
	}

	if rows == 0 {
		log.Fatalf("Failed to update parents. Rows affected: %d", rows)
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

	if err := seedCourses(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err := seedGames(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err := seedLocations(ctx, db); err != nil {
		log.Println(err)
		return
	}

	eventIds, err := seedEvents(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}

	if err = seedMemberships(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err = seedMembershipPlans(ctx, db); err != nil {
		log.Println(err)
		return
	}

	clientIds, err := seedUsers(ctx, db)

	if err != nil {
		log.Println(err)
		return
	}

	if err = updateParents(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if _, err = seedAthletes(ctx, db, clientIds); err != nil {
		log.Println(err)
		return
	}

	if err := seedClientsMembershipPlans(ctx, db); err != nil {
		log.Println(err)
		return
	}

	err = seedClientsEnrollments(ctx, db, clientIds, eventIds)

	if err != nil {
		log.Println(err)
		return
	}

	//err = seedMembershipCoursesEligibility(ctx, db, membershipIds, courseIds)
	//
	//if err != nil {
	//	log.Println(err)
	//	return
	//}

	if err = seedMembershipPracticeEligibility(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err = seedStaffRoles(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err = seedStaff(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err = seedHaircutServices(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err = seedBarberServices(ctx, db); err != nil {
		log.Println(err)
		return
	}

	if err = seedHaircutEvents(ctx, db, clientIds); err != nil {
		log.Println(err)
		return
	}
}
