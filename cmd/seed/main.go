package main

//
//import (
//	"context"
//	"log"
//	"time"
//
//	"api/config"
//	courseSQLC "api/internal/domains/course/persistence/sqlc/generated"
//	facilitySQLC "api/internal/domains/facility/persistence/sqlc/generated"
//	membershipSQLC "api/internal/domains/membership/persistence/sqlc/generated"
//	"database/sql"
//
//	"github.com/google/uuid"
//	_ "github.com/lib/pq" // Add this import
//)
//
//func seedCourses(ctx context.Context, q *courseSQLC.Queries) error {
//	courses := []courseSQLC.CreateCourseParams{
//		{
//			Name:        "Yoga Basics",
//			Description: sql.NullString{String: "Introduction to Yoga", Valid: true},
//			StartDate:   time.Now().AddDate(0, 0, 7),
//			EndDate:     time.Now().AddDate(0, 1, 7),
//		},
//		{
//			Name:        "Advanced HIIT",
//			Description: sql.NullString{String: "High Intensity Training", Valid: true},
//			StartDate:   time.Now().AddDate(0, 0, 14),
//			EndDate:     time.Now().AddDate(0, 2, 14),
//		},
//	}
//
//	for _, course := range courses {
//		if _, err := q.CreateCourse(ctx, course); err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func seedFacilities(ctx context.Context, q *facilitySQLC.Queries) error {
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
//		if _, err := q.CreateFacility(ctx, facility); err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func seedMemberships(ctx context.Context, q *membershipSQLC.Queries) error {
//	memberships := []membershipSQLC.CreateMembershipParams{
//		{
//			Name:        "Basic Plan",
//			Description: sql.NullString{String: "Access to basic facilities", Valid: true},
//			StartDate:   time.Now(),
//			EndDate:     time.Now().AddDate(0, 1, 0),
//		},
//		{
//			Name:        "Premium Plan",
//			Description: sql.NullString{String: "Full access to all facilities", Valid: true},
//			StartDate:   time.Now(),
//			EndDate:     time.Now().AddDate(1, 0, 0),
//		},
//	}
//
//	for _, membership := range memberships {
//		if _, err := q.CreateMembership(ctx, membership); err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func main() {
//	db := config.GetDBConnection()
//	defer db.Close()
//
//	ctx := context.Background()
//
//	// Initialize queries
//	courseQueries := courseSQLC.New(db)
//	// facilityQueries := facilitySQLC.New(db)
//	membershipQueries := membershipSQLC.New(db)
//
//	// Seed data
//	if err := seedCourses(ctx, courseQueries); err != nil {
//		log.Fatalf("Error seeding courses: %v", err)
//	}
//
//	// if err := seedFacilities(ctx, facilityQueries); err != nil {
//	// 	log.Fatalf("Error seeding facilities: %v", err)
//	// }
//
//	if err := seedMemberships(ctx, membershipQueries); err != nil {
//		log.Fatalf("Error seeding memberships: %v", err)
//	}
//
//	log.Println("Seeding completed successfully")
//}
