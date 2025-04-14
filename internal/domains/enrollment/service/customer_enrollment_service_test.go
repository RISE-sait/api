package enrollment

import (
	"api/internal/di"
	errLib "api/internal/libs/errors"
	dbTestUtils "api/utils/test_utils"
	"context"
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	dbEvent "api/internal/domains/event/persistence/sqlc/generated"
	dbIdentity "api/internal/domains/identity/persistence/sqlc/generated"
	dbMembership "api/internal/domains/membership/persistence/sqlc/generated"
	dbProgram "api/internal/domains/program/persistence/sqlc/generated"
)

func TestCustomerReserveProgram_ACID_Serializable(t *testing.T) {

	for run := 1; run < 21; run++ {
		t.Run(fmt.Sprintf("Run %d", run), func(t *testing.T) {

			testDb, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../db/migrations")

			defer cleanup()

			container := di.Container{
				DB: testDb,
				Queries: &di.QueriesType{
					EnrollmentDb: dbEnrollment.New(testDb),
					EventDb:      dbEvent.New(testDb),
					IdentityDb:   dbIdentity.New(testDb),
					MembershipDb: dbMembership.New(testDb),
					ProgramDb:    dbProgram.New(testDb),
				},
			}

			enrollmentService := NewCustomerEnrollmentService(&container)

			createdProgram, err := container.Queries.ProgramDb.CreateProgram(context.Background(), dbProgram.CreateProgramParams{
				Name: "Test Program",
				Capacity: sql.NullInt32{
					Int32: 2,
					Valid: true,
				},
				Level: dbProgram.ProgramProgramLevelAll,
				Type:  dbProgram.ProgramProgramTypeCourse,
			})

			require.NoError(t, err)

			membership, err := container.Queries.MembershipDb.CreateMembership(context.Background(), dbMembership.CreateMembershipParams{
				Name:        "Test Membership",
				Description: "Test Description",
				Benefits:    "Test Benefits",
			})

			require.NoError(t, err)

			membershipPlan, err := container.Queries.MembershipDb.CreateMembershipPlan(context.Background(), dbMembership.CreateMembershipPlanParams{
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
			_, err = container.DB.Exec(`
INSERT INTO program.fees (program_id, membership_id, stripe_price_id)
VALUES ($1, $2, $3)
`, createdProgram.ID, membership.ID, membershipPlan.StripePriceID)

			require.NoError(t, err)

			// check if membership is eligible for program using raw SQL
			var count int

			err = container.DB.QueryRow(`
SELECT COUNT(*) 
    FROM program.fees
        WHERE program_id = $1
        `, createdProgram.ID).Scan(&count)
			require.NoError(t, err)

			assert.Equal(t, 1, count)

			customers := make([]dbIdentity.UsersUser, 3)

			for i := 0; i < 3; i++ {
				customer, userErr := container.Queries.IdentityDb.CreateUser(context.Background(), dbIdentity.CreateUserParams{
					CountryAlpha2Code:        "CA",
					Age:                      20,
					HasMarketingEmailConsent: false,
					HasSmsConsent:            false,
					FirstName:                "John",
					LastName:                 "Doe",
				})

				require.NoError(t, userErr)

				customers[i] = customer
			}

			assert.Equal(t, 3, len(customers))
			// enroll all ( 3 ) customers from customers list into program events with capacity of 2 concurrently to test write skew
			// this should fail due to predicate locks
			enrollmentErrs := make(chan *errLib.CommonError, 3)
			for i := 0; i < 3; i++ {
				go func() {
					enrollmentErrs <- enrollmentService.ReserveSeatInProgram(context.Background(), createdProgram.ID, customers[i].ID)
				}()
			}

			// Collect errors
			var failedEnrollments int
			for i := 0; i < 3; i++ {
				enrollErr := <-enrollmentErrs
				if enrollErr != nil {
					assert.Equal(t, http.StatusConflict, enrollErr.HTTPCode)
					assert.True(t, strings.Contains(enrollErr.Message, "Too many people enrolled at the same time. Please try again.") ||
						strings.Contains(enrollErr.Message, "Too many people enrolled at the same time. Please try again."))
					failedEnrollments++
				}
			}

			// Verify the number of enrolled customers in the program

			var enrolledCount int
			err = container.DB.QueryRow(`
SELECT COUNT(*) 
    FROM program.customer_enrollment
        WHERE program_id = $1
        `, createdProgram.ID).Scan(&enrolledCount)
			require.NoError(t, err)

			assert.LessOrEqual(t, failedEnrollments, 2)
			assert.GreaterOrEqual(t, enrolledCount, 1)

			assert.Equal(t, 3-failedEnrollments, enrolledCount)

			log.Printf("Enrolled Count: %d", enrolledCount)
			log.Printf("Failed Enrollments: %d", failedEnrollments)

		})
	}
}

func TestEnrollCustomerInProgramEvents_ACID_No_race_condition(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../db/migrations")

	enrollmentQ := dbEnrollment.New(dbConn)
	identityQ := dbIdentity.New(dbConn)
	programQ := dbProgram.New(dbConn)
	membershipQ := dbMembership.New(dbConn)

	defer cleanup()

	container := di.Container{
		DB: dbConn,
		Queries: &di.QueriesType{
			EnrollmentDb: dbEnrollment.New(dbConn),
			EventDb:      dbEvent.New(dbConn),
			IdentityDb:   dbIdentity.New(dbConn),
			MembershipDb: dbMembership.New(dbConn),
			ProgramDb:    dbProgram.New(dbConn),
		},
	}

	enrollmentService := NewCustomerEnrollmentService(&container)

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
INSERT INTO program.fees (program_id, membership_id, stripe_price_id)
VALUES ($1, $2, $3)
`, createdProgram.ID, membership.ID, membershipPlan.StripePriceID)

	require.NoError(t, err)

	// check if membership is eligible for program using raw SQL
	var count int
	err = dbConn.QueryRow(`
SELECT COUNT(*) 
    FROM program.fees
        WHERE program_id = $1
        `, createdProgram.ID).Scan(&count)
	require.NoError(t, err)

	assert.Equal(t, 1, count)

	customers := make([]dbIdentity.UsersUser, 3)

	for i := 0; i < 3; i++ {
		customer, userErr := identityQ.CreateUser(context.Background(), dbIdentity.CreateUserParams{
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

	var errs []*errLib.CommonError
	for i := 0; i < 3; i++ {
		enrollErr := enrollmentService.ReserveSeatInProgram(context.Background(), createdProgram.ID, customers[i].ID)
		if enrollErr != nil {
			errs = append(errs, enrollErr)
		}
	}

	var enrolledCount int
	err = dbConn.QueryRow(`
SELECT COUNT(*) 
    FROM program.customer_enrollment
        WHERE program_id = $1
        `, createdProgram.ID).Scan(&enrolledCount)
	require.NoError(t, err)

	assert.Equal(t, 2, enrolledCount)

	assert.Equal(t, 1, len(errs))

	// check the message of the first error
	firstErrPtr := errs[0]
	firstErr := *firstErrPtr

	assert.Contains(t, firstErr.Message, "Program is full")
	assert.Equal(t, firstErr.HTTPCode, http.StatusConflict)
}
