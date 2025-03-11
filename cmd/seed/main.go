package main

import (
	"api/config"
	courseSQLC "api/internal/domains/course/persistence/sqlc/generated"
	identitySQLC "api/internal/domains/identity/persistence/sqlc/generated"
	membershipSQLC "api/internal/domains/membership/persistence/sqlc/generated"
	"api/internal/services/hubspot"
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Add this import
	"log"
	"os"
)

func seedUser(ctx context.Context, db *sql.DB, clients []Client) {

	identityQueries := identitySQLC.New(db)

	apiKey := os.Getenv("HUBSPOT_API_KEY")

	if apiKey == "" {
		log.Fatalln("HubSpot API key not set")
	}

	hubspotService := hubspot.GetHubSpotService(&apiKey)

	limit := 5

	for i, client := range clients {

		if i == limit {
			break
		}

		// Get the HubSpot IDs for the current batch
		hubspotResponse, err := hubspotService.GetUserByEmail(client.Email)

		if err != nil {
			log.Printf("Error getting user by email %v. Err: %v", client.Email, err.Message)
			continue
		}

		user, qErr := identityQueries.CreateUser(ctx, hubspotResponse.HubSpotId)

		if qErr != nil {
			log.Printf("Error creating user %v", qErr)
			continue
		}

		log.Printf("User created: %v ", user.ID)
	}
}

func seedCourses(ctx context.Context, q *courseSQLC.Queries) error {
	courses := []courseSQLC.CreateCourseParams{
		{
			Name:        "All Hs Girls (Gr. 10-12s) Spring Club Tryouts // March 10th, 2025",
			Description: sql.NullString{String: "March 10th, 2025 6PM - 8PM Non-refundable", Valid: true},
			Capacity:    300,
		},
		{
			Name:        "APRIL Spring Break Camp",
			Description: sql.NullString{String: "Join us for skills, drills and fun on the Court! DATES: April 22, 23, 24, 25 TIMES: 10AM-3:30PM Please bring indoor shoes, a ball, water bottles, Lunch and Snacks", Valid: true},
			Capacity:    300,
		},
		{
			Name:        "BOYS U12/U13 Spring Club Tryouts",
			Description: sql.NullString{String: "JANUARY 10, 2025 8:00 - 9:45PM Court 1, 2 and 3 AND JANUARY 12, 2025 11:00 - 12:45PM Court 1, 2 and 3 Address: Rise Facility Non-refundable", Valid: true},
			Capacity:    500,
		},
		{
			Name:        "BOYS U11 Spring Club Tryouts",
			Description: sql.NullString{String: "January 09, 2025 6:00PM - 8:15PM Court 1 and 2 January 10, 2025 6:00 - 7:30 PM Court 2 Address: Rise Facility Non-refundable", Valid: true},
			Capacity:    500,
		},
		{
			Name:        "BOYS U14/U15 Spring Club Tryouts",
			Description: sql.NullString{String: "January 12, 2025 7:15 - 9:30PM Court 1, 2 and 3 January 13, 2025 8:00 - 9:45PM Court 1, 2 and 3 Address: Rise Facility Non-refundable", Valid: true},
			Capacity:    483,
		},
		{
			Name:        "Drop In",
			Description: sql.NullString{String: "Drop in access to Rise 3 courts", Valid: true},
			Capacity:    100,
		},
		{
			Name:        "GIRLS U11 Spring Club Tryouts",
			Description: sql.NullString{String: "January 10, 2025 6:00PM - 7:30PM Address: Rise Facility Court 1 Non-refundable", Valid: true},
			Capacity:    100,
		},
		{
			Name:        "GIRLS U12/U13 Spring Club Tryouts",
			Description: sql.NullString{String: "January 10, 2025 5:30PM - 7:30PM Court 3 & Surge Court January 12, 2025 1:00 - 2:30PM Court 1&2 Non-refundable Address: Rise Facility", Valid: true},
			Capacity:    500,
		},
		{
			Name:        "GIRLS U14/U15 Spring Club Tryouts",
			Description: sql.NullString{String: "January 12, 2025 5:00 -7:00PM Court 1, 2 and 3 January 13, 2025 5:30 - 7:30PM Court 1, 2 and 3 Address: Rise Facility", Valid: true},
			Capacity:    500,
		},
		{
			Name:        "Holiday Hoops Academy (Winter Camp 2024) Ages 9-17",
			Description: sql.NullString{String: "Join us for 4 days filled with Skills, Games, Fundamentals and Competitive drills all with a holiday twist! DEC 21 1PM-7PM DEC 22 130PM-730PM DEC 23 9AM-4PM", Valid: true},
			Capacity:    300,
		},
		{
			Name:        "MARCH Spring Break Camp",
			Description: sql.NullString{String: "Join us for skills, drills and fun on the court. DATES: March 24, 25, 26, 27, 28 TIME: 10AM-3:30PM Please bring indoor shoes, a ball, water bottles, Lunch and Snacks", Valid: true},
			Capacity:    300,
		},
		{
			Name:        "Pro Rise Club",
			Description: sql.NullString{String: "WARNING ** SERIOUS ATHLETES ONLY** 4x per week Strength & Conditioning to get you prepared and stronger. Designed to build endurance and resilience for the season ahead. $650+gst for non-members", Valid: true},
			Capacity:    400,
		},
		{
			Name:        "Rise & Honor Memorial Cup",
			Description: sql.NullString{String: "Available Age Groups: U11 Boys & Girls U13 Boys & Girls U15 Boys & Girls U17 Boys & Girls U18 Boys & Girls", Valid: true},
			Capacity:    100,
		},
		{
			Name:        "Rising Stars Camp (Winter Camp 2024) Ages 10-13",
			Description: sql.NullString{String: "Join us for 3 days filled with Skills, Games, Fundamentals and Competitive drills all with a holiday twist! DEC 20 6PM-8PM, DEC 21 9AM-12PM, DEC 22 10AM-1PM", Valid: true},
			Capacity:    300,
		},
		{
			Name:        "U11 CO-ED Winter League Assessments (Tier 3)",
			Description: sql.NullString{String: "January 17, 2025 6:00-7:45PM Court 1 and 2 Address: Rise Facility Non-Refundable", Valid: true},
			Capacity:    500,
		},
		{
			Name:        "U13 BOYS Winter Rise League Assessments (Tier 3)",
			Description: sql.NullString{String: "January 17, 2025 8:00 -9:30 PM Court 1 and 2 Address: Rise Facility Non-Refundable", Valid: true},
			Capacity:    500,
		},
		{
			Name:        "U13/U15 GIRLS Winter Rise League Assessments (Tier 3)",
			Description: sql.NullString{String: "January 19, 2025 12:30 - 2:30PM Court 1, 2 and 3 Address: Rise Facility Non-Refundable", Valid: true},
			Capacity:    500,
		},
		{
			Name:        "U15 BOYS Winter League Assessments (Tier 3)",
			Description: sql.NullString{String: "January 19, 2025 10:00 - 12:00PM Court 1, 2 and 3 Address: Rise Facility Non-Refundable", Valid: true},
			Capacity:    500,
		},
		{
			Name:        "U16 Boys (Gr. 10) Spring Club Tryouts // March 10th, 2025",
			Description: sql.NullString{String: "March 10th, 2025 8PM - 10PM", Valid: true},
			Capacity:    300,
		},
		{
			Name:        "U17/U18 Boys (Gr. 11 & 12) Spring Club Tryouts // March 11th, 2025",
			Description: sql.NullString{String: "March 11th, 2025 6PM - 8PM Non-refundable", Valid: true},
			Capacity:    300,
		},
		{
			Name:        "U17/U18 Girls (Gr. 11 & 12) Spring Club Tryouts // March 10th, 2025",
			Description: sql.NullString{String: "March 10th, 2025 6PM - 8PM Non-refundable", Valid: true},
			Capacity:    300,
		},
	}

	for _, course := range courses {
		if _, err := q.CreateCourse(ctx, course); err != nil {
			return err
		}
	}
	return nil
}

//func seedFacilities(ctx context.Context, q *facilitySQLC.IdentityQueries) error {
//	facilities := []facilitySQLC.CreateFacilityParams{
//		{
//			Name:           "Main Gym",
//			Location:       "First Floor",
//			FacilityTypeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), // Add proper UUID
//		},
//		{
//			Name:           "Yoga Studio",
//			Location:       "Second Floor",
//			FacilityTypeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), // Add proper UUID
//		},
//	}
//
//	for _, facility := range facilities {
//		if _, err := q.CreateLocation(ctx, facility); err != nil {
//			return err
//		}
//	}
//	return nil
//}

func seedMemberships(ctx context.Context, q *membershipSQLC.Queries) error {
	memberships := []membershipSQLC.CreateMembershipParams{
		{
			Name:        "Basic Plan",
			Description: sql.NullString{String: "Access to basic facilities", Valid: true},
		},
		{
			Name:        "Premium Plan",
			Description: sql.NullString{String: "Full access to all facilities", Valid: true},
		},
	}

	for _, membership := range memberships {
		if _, err := q.CreateMembership(ctx, membership); err != nil {
			return err
		}
	}
	return nil
}

func loadEnv() {

	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting working directory:", err)
		return
	}
	fmt.Println("Current Working Directory:", wd)

	if err := godotenv.Load("config/.env"); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	if err := godotenv.Load("config/.env.local"); err != nil {
		log.Printf("Error loading .env.local file: %v", err)
	}
}

func main() {

	loadEnv()

	ctx := context.Background()

	db := config.GetDBConnection()

	defer db.Close()

	clients, err := extract()

	if err != nil {
		log.Println(err)
		return
	}

	seedUser(ctx, db, clients)

	//
	//// Initialize queries
	//courseQueries := courseSQLC.New(db)
	//// facilityQueries := facilitySQLC.New(db)
	//membershipQueries := membershipSQLC.New(db)
	//
	//// Seed data
	//if err := seedCourses(ctx, courseQueries); err != nil {
	//	log.Fatalf("Error seeding courses: %v", err)
	//}
	//
	//// if err := seedFacilities(ctx, facilityQueries); err != nil {
	//// 	log.Fatalf("Error seeding facilities: %v", err)
	//// }
	//
	//if err := seedMemberships(ctx, membershipQueries); err != nil {
	//	log.Fatalf("Error seeding memberships: %v", err)
	//}
	//
	//log.Println("Seeding completed successfully")
}
