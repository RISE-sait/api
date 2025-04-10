package enrollment_test

import (
	enrollment "api/internal/domains/enrollment/persistence/repository"
	errLib "api/internal/libs/errors"
	dbTestUtils "api/utils/test_utils"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"net/http"
	"testing"
	"time"

	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	dbidentity "api/internal/domains/identity/persistence/sqlc/generated"
	dbMembership "api/internal/domains/membership/persistence/sqlc/generated"
	dbProgram "api/internal/domains/program/persistence/sqlc/generated"
)

func TestEnrollCustomerInProgramEvents_ACID_Serializable(t *testing.T) {
	for run := 1; run <= 20; run++ {
		t.Run(fmt.Sprintf("Run %d", run), func(t *testing.T) {
			dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

			enrollmentQ := dbEnrollment.New(dbConn)
			identityQ := dbidentity.New(dbConn)
			programQ := dbProgram.New(dbConn)
			membershipQ := dbMembership.New(dbConn)

			defer cleanup()

			enrollmentRepo := enrollment.NewEnrollmentRepository(dbConn)

			createdProgram, err := programQ.CreateProgram(context.Background(), dbProgram.CreateProgramParams{
				Name: "Test Program",
				Capacity: sql.NullInt32{
					Int32: 2,
					Valid: true,
				},
				Level: dbProgram.ProgramProgramLevelAll,
				Type:  dbProgram.ProgramProgramTypeCourse,
			})

			require.NoError(t, err)

			membership, err := membershipQ.CreateMembership(context.Background(), dbMembership.CreateMembershipParams{
				Name:        "Test Membership",
				Description: "Test Description",
				Benefits:    "Test Benefits",
			})

			require.NoError(t, err)

			membershipPlan, err := membershipQ.CreateMembershipPlan(context.Background(), dbMembership.CreateMembershipPlanParams{
				MembershipID: membership.ID,
				Name:         "Test Membership Plan",
				StripeJoiningFeeID: sql.NullString{
					String: "price_123",
					Valid:  true,
				},
				StripePriceID: "price_456",
			})

			require.NoError(t, err)

			// use raw sql to make membership eligible for program
			_, err = dbConn.Exec(`
INSERT INTO program.program_membership (program_id, membership_id, stripe_program_price_id)
VALUES ($1, $2, $3)
`, createdProgram.ID, membership.ID, membershipPlan.StripePriceID)

			require.NoError(t, err)

			// check if membership is eligible for program using raw SQL
			var count int
			err = dbConn.QueryRow(`
SELECT COUNT(*) 
    FROM program.program_membership
        WHERE program_id = $1
        `, createdProgram.ID).Scan(&count)
			require.NoError(t, err)

			assert.Equal(t, 1, count)

			customers := make([]dbidentity.UsersUser, 3)

			for i := 0; i < 3; i++ {
				customer, userErr := identityQ.CreateUser(context.Background(), dbidentity.CreateUserParams{
					CountryAlpha2Code:        "CA",
					Age:                      20,
					HasMarketingEmailConsent: false,
					HasSmsConsent:            false,
					FirstName:                "John",
					LastName:                 "Doe",
				})

				require.NoError(t, userErr)

				enrollmentErr := enrollmentQ.EnrollCustomerInMembershipPlan(context.Background(), dbEnrollment.EnrollCustomerInMembershipPlanParams{
					CustomerID:       customer.ID,
					MembershipPlanID: membershipPlan.ID,
					Status:           dbEnrollment.MembershipMembershipStatusActive,
					StartDate:        time.Now(),
				})

				require.NoError(t, enrollmentErr)

				customers[i] = customer
			}

			assert.Equal(t, 3, len(customers))

			// enroll all ( 3 ) customers from customers list into program events with capacity of 2 concurrently to test write skew
			// this should fail due to predicate locks
			enrolllmentErrs := make(chan *errLib.CommonError, 3)
			for i := 0; i < 3; i++ {
				go func(customerID uuid.UUID) {
					enrolllmentErrs <- enrollmentRepo.EnrollCustomerInProgram(context.Background(), customerID, createdProgram.ID)
				}(customers[i].ID)
			}

			// Collect errors
			var failedEnrollments int
			for i := 0; i < 3; i++ {
				enrollErr := <-enrolllmentErrs
				if enrollErr != nil {
					if enrollErr.HTTPCode == http.StatusConflict {
						assert.Contains(t, enrollErr.Message, "Program is full")
					}
					failedEnrollments++
				}
			}

			log.Println("Failed enrollments for run", run, ":", failedEnrollments)
			assert.GreaterOrEqual(t, failedEnrollments, 1)

			// Verify the number of enrolled customers in the program

			var enrolledCount int
			err = dbConn.QueryRow(`
SELECT COUNT(*) 
    FROM program.customer_enrollment
        WHERE program_id = $1
        `, createdProgram.ID).Scan(&enrolledCount)
			require.NoError(t, err)

			log.Println("Enrolled for run", run, ":", enrolledCount)

			assert.LessOrEqual(t, enrolledCount, 2)
			assert.Equal(t, 3-failedEnrollments, enrolledCount)

		})
	}
}

func TestEnrollCustomerInProgramEvents_ACID_No_race_condition(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	enrollmentQ := dbEnrollment.New(dbConn)
	identityQ := dbidentity.New(dbConn)
	programQ := dbProgram.New(dbConn)
	membershipQ := dbMembership.New(dbConn)

	defer cleanup()

	enrollmentRepo := enrollment.NewEnrollmentRepository(dbConn)

	createdProgram, err := programQ.CreateProgram(context.Background(), dbProgram.CreateProgramParams{
		Name: "Test Program",
		Capacity: sql.NullInt32{
			Int32: 2,
			Valid: true,
		},
		Level: dbProgram.ProgramProgramLevelAll,
		Type:  dbProgram.ProgramProgramTypeCourse,
	})

	require.NoError(t, err)

	membership, err := membershipQ.CreateMembership(context.Background(), dbMembership.CreateMembershipParams{
		Name:        "Test Membership",
		Description: "Test Description",
		Benefits:    "Test Benefits",
	})

	require.NoError(t, err)

	membershipPlan, err := membershipQ.CreateMembershipPlan(context.Background(), dbMembership.CreateMembershipPlanParams{
		MembershipID: membership.ID,
		Name:         "Test Membership Plan",
		StripeJoiningFeeID: sql.NullString{
			String: "price_123",
			Valid:  true,
		},
		StripePriceID: "price_456",
	})

	require.NoError(t, err)

	// use raw sql to make membership eligible for program
	_, err = dbConn.Exec(`
INSERT INTO program.program_membership (program_id, membership_id, stripe_program_price_id)
VALUES ($1, $2, $3)
`, createdProgram.ID, membership.ID, membershipPlan.StripePriceID)

	require.NoError(t, err)

	// check if membership is eligible for program using raw SQL
	var count int
	err = dbConn.QueryRow(`
SELECT COUNT(*) 
    FROM program.program_membership
        WHERE program_id = $1
        `, createdProgram.ID).Scan(&count)
	require.NoError(t, err)

	assert.Equal(t, 1, count)

	customers := make([]dbidentity.UsersUser, 3)

	for i := 0; i < 3; i++ {
		customer, userErr := identityQ.CreateUser(context.Background(), dbidentity.CreateUserParams{
			CountryAlpha2Code:        "CA",
			Age:                      20,
			HasMarketingEmailConsent: false,
			HasSmsConsent:            false,
			FirstName:                "John",
			LastName:                 "Doe",
		})

		require.NoError(t, userErr)

		enrollmentErr := enrollmentQ.EnrollCustomerInMembershipPlan(context.Background(), dbEnrollment.EnrollCustomerInMembershipPlanParams{
			CustomerID:       customer.ID,
			MembershipPlanID: membershipPlan.ID,
			Status:           dbEnrollment.MembershipMembershipStatusActive,
			StartDate:        time.Now(),
		})

		require.NoError(t, enrollmentErr)

		customers[i] = customer
	}

	assert.Equal(t, 3, len(customers))

	// Enroll 3 customers with staggered timing to avoid simultaneous commits
	var errs []*errLib.CommonError
	for i := 0; i < 3; i++ {
		enrollErr := enrollmentRepo.EnrollCustomerInProgram(context.Background(), customers[i].ID, createdProgram.ID)
		if enrollErr != nil {
			errs = append(errs, enrollErr)
		}
	}

	// Verify that at least one error occurred

	assert.GreaterOrEqual(t, len(errs), 1)

	// check the message of the first error
	firstErrPtr := errs[0]
	firstErr := *firstErrPtr

	assert.Contains(t, firstErr.Message, "Program is full")
	assert.Equal(t, firstErr.HTTPCode, http.StatusConflict)

	// Verify the number of enrolled customers in the program

	var enrolledCount int
	err = dbConn.QueryRow(`
SELECT COUNT(*) 
    FROM program.customer_enrollment
        WHERE program_id = $1
        `, createdProgram.ID).Scan(&enrolledCount)
	require.NoError(t, err)

	assert.LessOrEqual(t, enrolledCount, 2)

}
