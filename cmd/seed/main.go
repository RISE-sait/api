package main

import (
	"api/cmd/seed/data"
	dbSeed "api/cmd/seed/sqlc/generated"
	"api/config"
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
	schemas := []string{"audit", "events", "haircut",
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

func seedPractices(ctx context.Context, db *sql.DB) []uuid.UUID {

	seedQueries := dbSeed.New(db)

	practices := data.Practices

	var (
		nameArray        []string
		descriptionArray []string
		levelArray       []dbSeed.ProgramProgramLevel
	)
	for i := 0; i < len(practices); i++ {

		nameArray = append(nameArray, practices[i].Name)
		descriptionArray = append(descriptionArray, practices[i].Description)
		levelArray = append(levelArray, dbSeed.ProgramProgramLevelAll)
	}

	ids, err := seedQueries.InsertPractices(ctx, dbSeed.InsertPracticesParams{
		NameArray:        nameArray,
		DescriptionArray: descriptionArray,
		LevelArray:       levelArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert practices: %v", err)
	}

	return ids
}

func seedStaffRoles(ctx context.Context, db *sql.DB) {

	seedQueries := dbSeed.New(db)

	err := seedQueries.InsertStaffRoles(ctx)

	if err != nil {
		log.Fatalf("Failed to insert roles: %v", err)
		return
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

func seedFakeCourses(ctx context.Context, db *sql.DB) {

	seedQueries := dbSeed.New(db)

	if err := seedQueries.InsertCourses(ctx, data.GetCourses()); err != nil {
		log.Fatalf("Failed to insert courses: %v", err)
		return
	}

	return
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

func seedFakeGames(ctx context.Context, db *sql.DB, teamIds []uuid.UUID) {
	seedQueries := dbSeed.New(db)

	gamesData := data.GetGames(10, teamIds) // Generate 20 games

	if err := seedQueries.InsertGames(ctx, gamesData); err != nil {
		log.Fatalf("Failed to insert games: %v", err)
	}
}

func seedLocations(ctx context.Context, db *sql.DB) {

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
}

func seedMembershipPlans(ctx context.Context, db *sql.DB) {
	seedQueries := dbSeed.New(db)

	if err := seedQueries.InsertMembershipPlans(ctx, data.GetMembershipPlans()); err != nil {
		log.Fatalf("Failed to insert membership plans: %v", err)
		return
	}
}

func seedMemberships(ctx context.Context, db *sql.DB) []uuid.UUID {

	seedQueries := dbSeed.New(db)

	var (
		nameArray        []string
		descriptionArray []string
	)
	for i := 0; i < len(data.Memberships); i++ {

		nameArray = append(nameArray, data.Memberships[i].Name)
		descriptionArray = append(descriptionArray, data.Memberships[i].Description)
	}

	ids, err := seedQueries.InsertMemberships(ctx, dbSeed.InsertMembershipsParams{
		NameArray:        nameArray,
		DescriptionArray: descriptionArray,
	})

	if err != nil {
		log.Fatalf("Failed to insert memberships: %v", err)
	}

	return ids
}

func seedFakeEnrollmentFees(ctx context.Context, db *sql.DB, programIds, membershipIds []uuid.UUID) {
	seedQueries := dbSeed.New(db)

	arg := data.GetEnrollmentFees(programIds, membershipIds)

	// Insert events and sessions into the database
	if err := seedQueries.InsertEnrollmentFees(ctx, arg); err != nil {
		log.Fatalf("Failed to insert events: %v", err)
		return
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

	clientIds := seedUsers(ctx, db)

	staffIds := seedStaff(ctx, db)

	seedFakeCoachStats(ctx, db)

	teamIds := seedFakeTeams(ctx, db)

	practiceIds := seedPractices(ctx, db)

	seedFakeCourses(ctx, db)

	seedFakeGames(ctx, db, teamIds)

	seedLocations(ctx, db)

	eventIds := seedEvents(ctx, db)

	membershipIds := seedMemberships(ctx, db)

	seedMembershipPlans(ctx, db)

	updateFakeParents(ctx, db)

	seedFakeAthletes(ctx, db, clientIds)

	seedClientsMembershipPlans(ctx, db)

	seedFakeClientsEnrollments(ctx, db, clientIds, eventIds)

	seedFakeEventStaff(ctx, db, eventIds, staffIds)

	seedFakeEnrollmentFees(ctx, db, practiceIds, membershipIds)

	seedHaircutServices(ctx, db)

	seedFakeBarberServices(ctx, db)

	seedFakeHaircutEvents(ctx, db, clientIds)
}
