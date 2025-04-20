package main

import (
	"api/cmd/seed/data"
	dbSeed "api/cmd/seed/sqlc/generated"
	"api/config"
	"time"

	"github.com/google/uuid"

	"fmt"
	"strings"

	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func clearTables(ctx context.Context, db *sql.DB) {
	// Define the schemas you want to truncate tables from
	schemas := []string{"athletic", "audit", "events", "haircut",
		"location", "membership", "program", "public", "staff", "users", "waiver"}

	// Build the TRUNCATE query
	var tables []string
	for _, schema := range schemas {
		// Query for tables in the specified schema
		rows, err := db.QueryContext(ctx, "SELECT tablename FROM pg_tables WHERE schemaname = $1", schema)
		if err != nil {
			log.Fatalf("Failed to query tables: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var table string
			if err := rows.Scan(&table); err != nil {
				log.Fatalf("Failed to scan tables: %v", err)
				return
			}
			if schema == "public" && table == "goose_db_version" {
				continue
			}
			tables = append(tables, fmt.Sprintf("%s.%s", schema, table))
		}

		if err := rows.Err(); err != nil {
			log.Fatalf("Failed to scan tables: %v", err)
			return
		}
	}

	// If there are no tables, return early
	if len(tables) == 0 {
		return
	}

	// Generate the TRUNCATE statement with CASCADE and RESTART IDENTITY
	truncateQuery := "TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE"

	// Execute the TRUNCATE query
	if _, err := db.ExecContext(ctx, truncateQuery); err != nil {
		log.Fatalf("Failed to truncate tables: %v", err)
	}

}

func seedUsers(ctx context.Context, db *sql.DB) []uuid.UUID {
	clients, err := data.GetClients()

	if err != nil {
		log.Fatalf("Failed to get clients: %v", err)
		return nil
	}

	staffs := data.GetStaffsAsClients()

	if err != nil {
		log.Fatalf("Failed to get staffs: %v", err)
		return nil
	}

	seedQueries := dbSeed.New(db)

	var (
		countryAlpha2CodeArray      []string
		firstNameArray              []string
		lastNameArray               []string
		dobArray                    []time.Time
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
		dobArray = append(dobArray, client.DOB)
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
		DobArray:                      dobArray,
		GenderArray:                   genderArray,
		ParentIDArray:                 parentIDArray,
		PhoneArray:                    phoneArray,
		EmailArray:                    emailArray,
		HasMarketingEmailConsentArray: hasMarketingEmailConsentArr,
		HasSmsConsentArray:            hasSMSConsentArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert clients: %v", err)
		return nil
	}

	_, err = seedQueries.InsertUsers(ctx, staffs)

	if err != nil {
		log.Fatalf("Failed to insert staff as clients: %v", err)
		return nil
	}

	return ids
}

func seedHaircutServices(ctx context.Context, db *sql.DB) {

	seedQueries := dbSeed.New(db)

	services := data.GetHaircutServices()

	if err := seedQueries.InsertHaircutServices(ctx, services); err != nil {
		log.Fatalf("Failed to insert haircut services: %v", err)
		return
	}
}

func seedFakeWaivers(ctx context.Context, db *sql.DB) {

	seedQueries := dbSeed.New(db)

	if err := seedQueries.InsertWaivers(ctx); err != nil {
		log.Fatalf("Failed to insert waivers: %v", err)
	}
}

func seedFakeBarberServices(ctx context.Context, db *sql.DB) {

	seedQueries := dbSeed.New(db)

	services := data.GetBarberServices()

	if err := seedQueries.InsertBarberServices(ctx, services); err != nil {
		log.Fatalf("Failed to insert barber services: %v", err)
		return
	}
}

func seedFakeHaircutEvents(ctx context.Context, db *sql.DB, clientIds []uuid.UUID) {

	seedQueries := dbSeed.New(db)

	events := data.GetHaircutEvents(clientIds)

	if err := seedQueries.InsertHaircutEvents(ctx, events); err != nil {
		log.Fatalf("Failed to insert haircut events: %v", err)
		return
	}
}

func seedFakeAthletes(ctx context.Context, db *sql.DB, ids []uuid.UUID) {

	seedQueries := dbSeed.New(db)

	if _, err := seedQueries.InsertAthletes(ctx, ids); err != nil {
		log.Fatalf("Failed to insert athletes: %v", err)
		return
	}
}

func seedPractices(ctx context.Context, db *sql.DB) {

	seedQueries := dbSeed.New(db)

	practices := data.Practices

	var (
		nameArray        []string
		descriptionArray []string
		levelArray       []dbSeed.ProgramProgramLevel
		isPayPerEvent    []bool
	)
	for _, practice := range practices {

		nameArray = append(nameArray, practice.Name)
		descriptionArray = append(descriptionArray, practice.Description)
		levelArray = append(levelArray, dbSeed.ProgramProgramLevelAll)
		isPayPerEvent = append(isPayPerEvent, practice.IsPayPerEvent)
	}

	_, err := seedQueries.InsertPractices(ctx, dbSeed.InsertPracticesParams{
		NameArray:          nameArray,
		DescriptionArray:   descriptionArray,
		LevelArray:         levelArray,
		IsPayPerEventArray: isPayPerEvent,
	})

	if err != nil {
		log.Fatalf("Failed to insert practices: %v", err)
	}
}

func seedProgramsFees(ctx context.Context, db *sql.DB) {

	seedQueries := dbSeed.New(db)

	practices := data.Practices

	var (
		programNameArray    []string
		membershipNameArray []string
		stripePriceIDArray  []string
	)

	for _, practice := range practices {
		for _, eligibility := range practice.MembershipsEligibility {
			programNameArray = append(programNameArray, practice.Name)
			membershipNameArray = append(membershipNameArray, eligibility.Name)

			if eligibility.StripePriceID == nil {
				stripePriceIDArray = append(stripePriceIDArray, "")
			} else {
				stripePriceIDArray = append(stripePriceIDArray, *eligibility.StripePriceID)
			}
		}
	}

	err := seedQueries.InsertProgramFees(ctx, dbSeed.InsertProgramFeesParams{
		ProgramNameArray:          programNameArray,
		MembershipNameArray:       membershipNameArray,
		StripeProgramPriceIDArray: stripePriceIDArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert program fees: %v", err)
	}
}

func seedStaffRoles(ctx context.Context, db *sql.DB) {

	seedQueries := dbSeed.New(db)

	err := seedQueries.InsertStaffRoles(ctx)

	if err != nil {
		log.Fatalf("Failed to insert roles: %v", err)
		return
	}
}

func seedStaff(ctx context.Context, db *sql.DB) []uuid.UUID {

	seedQueries := dbSeed.New(db)

	staffs := data.GetStaffs()

	ids, err := seedQueries.InsertStaff(ctx, staffs)

	if err != nil {
		log.Fatalf("Failed to insert staff: %v", err)
	}

	return ids
}

func seedFakeCoachStats(ctx context.Context, db *sql.DB) {

	seedQueries := dbSeed.New(db)

	if err := seedQueries.InsertCoachStats(ctx); err != nil {
		log.Fatalf("Failed to insert coach stats: %v", err)
		return
	}
}

func seedFakeCourses(ctx context.Context, db *sql.DB) []string {

	seedQueries := dbSeed.New(db)

	courses := data.GetCourses()

	if err := seedQueries.InsertCourses(ctx, courses); err != nil {
		log.Fatalf("Failed to insert courses: %v", err)
		return nil
	} else {
		return courses.NameArray
	}
}

func seedFakeTeams(ctx context.Context, db *sql.DB) []uuid.UUID {

	seedQueries := dbSeed.New(db)

	teams := dbSeed.InsertTeamsParams{
		CoachEmailArray: []string{
			"viktor.djurasic+1@abcfitness.com",
			"coach@test.com",
		},
		NameArray:     []string{"Team 1", "Team 2"},
		CapacityArray: []int32{10, 10},
	}

	teamIds, err := seedQueries.InsertTeams(ctx, teams)

	if err != nil {
		log.Fatalf("Failed to insert teams: %v", err)
		return nil
	}

	return teamIds
}

func seedFakeGames(ctx context.Context, db *sql.DB, teamIds []uuid.UUID) []string {
	seedQueries := dbSeed.New(db)

	gamesData := data.GetGames(10, teamIds) // Generate 20 games

	if err := seedQueries.InsertGames(ctx, gamesData); err != nil {
		log.Fatalf("Failed to insert games: %v", err)
		return nil
	} else {
		return gamesData.NameArray
	}
}

func seedLocations(ctx context.Context, db *sql.DB) []string {

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
	}

	return nameArray
}

func seedMembershipPlans(ctx context.Context, db *sql.DB) {
	seedQueries := dbSeed.New(db)

	if err := seedQueries.InsertMembershipPlans(ctx, data.GetMembershipPlans()); err != nil {
		log.Fatalf("Failed to insert membership plans: %v", err)
		return
	}
}

func seedMemberships(ctx context.Context, db *sql.DB) {

	seedQueries := dbSeed.New(db)

	benefits := []string{
		"24/7 unlimited access to all of our facilities, including premium locations in major cities worldwide, with no blackout periods or restricted hours.",
		"Complimentary high-speed Wi-Fi throughout all facilities with enterprise-grade security and bandwidth capable of supporting video conferencing and streaming.",
		"A personal locker with biometric fingerprint access that can be used at any location, cleaned and sanitized daily by our staff.",
		"25% discount on all merchandise including premium athletic apparel, fitness equipment, and recovery tools, applicable both in-store and online.",
		"Ten complimentary guest passes per month that allow your friends or family members to experience our facilities, with advance booking required.",
		"A comprehensive fitness assessment every six months conducted by our certified trainers, including body composition analysis and personalized workout recommendations.",
		"Unlimited participation in all group fitness classes across all modalities, from high-intensity interval training to restorative yoga, with guaranteed spot reservations.",
		"Exclusive 48-hour priority booking window for all premium classes and personal training sessions before they open to general members.",
		"Daily towel service featuring premium Egyptian cotton towels, plus access to luxury toiletries from top wellness brands in all locker rooms.",
		"Invitations to members-only events including celebrity trainer workshops, nutrition seminars, and exclusive product launches throughout the year.",
	}

	var (
		nameArray        []string
		descriptionArray []string
		benefitsArray    []string
	)

	for i := 0; i < len(data.Memberships); i++ {

		nameArray = append(nameArray, data.Memberships[i].Name)
		descriptionArray = append(descriptionArray, data.Memberships[i].Description)

		// append a random benefit from benefits list
		randomIndex := i % len(benefits)

		benefitsArray = append(benefitsArray, benefits[randomIndex])

	}

	_, err := seedQueries.InsertMemberships(ctx, dbSeed.InsertMembershipsParams{
		NameArray:        nameArray,
		DescriptionArray: descriptionArray,
		BenefitsArray:    benefitsArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert memberships: %v", err)
	}
}

func seedEvents(ctx context.Context, db *sql.DB) []uuid.UUID {
	seedQueries := dbSeed.New(db)

	arg := data.GetEvents()

	// Insert events and sessions into the database
	ids, err := seedQueries.InsertEvents(ctx, arg)

	if err != nil {
		log.Fatalf("Failed to insert events: %v", err)
		return nil
	}

	return ids
}

func seedFakeEvents(ctx context.Context, db *sql.DB, programs, locations []string, isRecurring bool) []uuid.UUID {
	seedQueries := dbSeed.New(db)

	arg := data.GetFakeEvents(programs, locations, isRecurring)

	// Insert events and sessions into the database
	ids, err := seedQueries.InsertEvents(ctx, arg)

	if err != nil {
		log.Fatalf("Failed to insert events: %v", err)
		return nil
	}

	return ids
}

func seedClientsMembershipPlans(ctx context.Context, db *sql.DB) {
	seedQueries := dbSeed.New(db)

	plans, err := data.GetClientsMembershipPlans()

	if err != nil {
		log.Fatalf("Failed to insert client membership plans: %v", err)
		return
	}

	if err = seedQueries.InsertClientsMembershipPlans(ctx, plans); err != nil {
		log.Fatalf("Failed to insert client membership plans: %v", err)
		return
	}
}

func seedFakeClientsEnrollments(ctx context.Context, db *sql.DB, clients, events []uuid.UUID) {
	seedQueries := dbSeed.New(db)

	_, err := seedQueries.InsertCustomersEnrollments(ctx, data.GetClientsEnrollments(clients, events))

	if err != nil {
		log.Fatalf("Failed to insert client enrollments: %v", err)
		return
	}
}

func seedFakeEventStaff(ctx context.Context, db *sql.DB, eventIds, staffIds []uuid.UUID) {
	seedQueries := dbSeed.New(db)

	err := seedQueries.InsertEventsStaff(ctx, data.GetEventStaff(eventIds, staffIds))

	if err != nil {
		log.Fatalf("Failed to insert client enrollments: %v", err)
	}
}

func updateFakeParents(ctx context.Context, db *sql.DB) {
	seedQueries := dbSeed.New(db)

	rows, err := seedQueries.UpdateParents(ctx)

	if err != nil {
		log.Fatalf("Failed to insert client enrollments: %v", err)
		return
	}

	if rows == 0 {
		log.Fatalf("Failed to update parents. Rows affected: %d", rows)
	}
}

func main() {

	ctx := context.Background()

	db := config.GetDBConnection()

	defer db.Close()

	clearTables(ctx, db)

	seedStaffRoles(ctx, db)

	seedFakeWaivers(ctx, db)

	clientIds := seedUsers(ctx, db)

	staffIds := seedStaff(ctx, db)

	seedFakeCoachStats(ctx, db)

	teamIds := seedFakeTeams(ctx, db)

	seedPractices(ctx, db)

	courses := seedFakeCourses(ctx, db)

	games := seedFakeGames(ctx, db, teamIds)

	locations := seedLocations(ctx, db)

	eventIds := seedEvents(ctx, db)

	_ = seedFakeEvents(ctx, db, games, locations, false)
	_ = seedFakeEvents(ctx, db, courses, locations, true)

	seedMemberships(ctx, db)

	seedMembershipPlans(ctx, db)

	seedProgramsFees(ctx, db)

	updateFakeParents(ctx, db)

	seedFakeAthletes(ctx, db, clientIds)

	seedClientsMembershipPlans(ctx, db)

	seedFakeClientsEnrollments(ctx, db, clientIds, eventIds)

	seedFakeEventStaff(ctx, db, eventIds, staffIds)

	seedHaircutServices(ctx, db)

	seedFakeBarberServices(ctx, db)

	seedFakeHaircutEvents(ctx, db, clientIds)
}
