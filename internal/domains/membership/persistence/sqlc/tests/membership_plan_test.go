package tests_test

import (
	dbTestUtils "api/utils/test_utils"
	"github.com/lib/pq"

	membershipDb "api/internal/domains/membership/persistence/sqlc/generated"

	"database/sql"

	"context"
	"github.com/google/uuid"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateMembershipPlan(t *testing.T) {

	_, _, _, _, _, queries, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	// Create a user to be the creator of the event
	params := membershipDb.CreateMembershipParams{
		Name:        "Go Basics Practice",
		Description: "Learn Go programming",
		Benefits:    "Free access to all events",
	}

	createdMembership, err := queries.CreateMembership(context.Background(), params)

	require.NoError(t, err)

	createPlanParams := membershipDb.CreateMembershipPlanParams{
		MembershipID: createdMembership.ID,
		Name:         "Go Basics Practice Plan",
		StripeJoiningFeeID: sql.NullString{
			String: "price_123",
			Valid:  true,
		},
		StripePriceID: "price_456",
	}

	createdPlan, err := queries.CreateMembershipPlan(context.Background(), createPlanParams)

	require.NoError(t, err)
	require.Equal(t, createPlanParams.MembershipID, createdPlan.MembershipID)
	require.Equal(t, createPlanParams.Name, createdPlan.Name)
	require.Equal(t, createPlanParams.StripeJoiningFeeID.String, createdPlan.StripeJoiningFeeID.String)
	require.Equal(t, createPlanParams.StripePriceID, createdPlan.StripePriceID)
}

func TestCreateExistingMembershipPlan(t *testing.T) {

	_, _, _, _, _, queries, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	// Create a user to be the creator of the event
	params := membershipDb.CreateMembershipParams{
		Name:        "Go Basics Practice",
		Description: "Learn Go programming",
		Benefits:    "Free access to all events",
	}

	createdMembership, err := queries.CreateMembership(context.Background(), params)

	require.NoError(t, err)

	createPlanParams := membershipDb.CreateMembershipPlanParams{
		MembershipID: createdMembership.ID,
		Name:         "Go Basics Practice Plan",
		StripeJoiningFeeID: sql.NullString{
			String: "price_123",
			Valid:  true,
		},
		StripePriceID: "price_456",
	}

	_, err = queries.CreateMembershipPlan(context.Background(), createPlanParams)

	require.NoError(t, err)

	_, err = queries.CreateMembershipPlan(context.Background(), createPlanParams)

	require.Error(t, err)

	// Check if the error's constraint is unique_membership_table_name
	var pqErr *pq.Error

	require.ErrorAs(t, err, &pqErr)
	require.Equal(t, "unique_membership_name", pqErr.Constraint)
}

func TestUpdateMembershipPlan(t *testing.T) {

	_, _, _, _, _, queries, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	params := membershipDb.CreateMembershipParams{
		Name:        "Go Basics Practice",
		Description: "Learn Go programming",
		Benefits:    "Free access to all events",
	}

	createdMembership, err := queries.CreateMembership(context.Background(), params)

	require.NoError(t, err)

	createPlanParams := membershipDb.CreateMembershipPlanParams{
		MembershipID: createdMembership.ID,
		Name:         "Go Basics Practice Plan",
		StripeJoiningFeeID: sql.NullString{
			String: "price_123",
			Valid:  true,
		},
		StripePriceID: "price_456",
	}

	createdMembershipPlan, err := queries.CreateMembershipPlan(context.Background(), createPlanParams)

	require.NoError(t, err)

	updateMembershipPlanParams := membershipDb.UpdateMembershipPlanParams{
		Name:          "Updated Membership Name",
		StripePriceID: "price_789",
		StripeJoiningFeeID: sql.NullString{
			String: "price_456",
			Valid:  true,
		},
		AmtPeriods: sql.NullInt32{
			Int32: 12,
			Valid: true,
		},
		MembershipID: createdMembership.ID,
		ID:           createdMembershipPlan.ID,
	}

	updatedMembershipPlan, err := queries.UpdateMembershipPlan(context.Background(), updateMembershipPlanParams)
	require.NoError(t, err)

	require.Equal(t, updateMembershipPlanParams.Name, updatedMembershipPlan.Name)
	require.Equal(t, updateMembershipPlanParams.StripePriceID, updatedMembershipPlan.StripePriceID)
	require.Equal(t, updateMembershipPlanParams.StripeJoiningFeeID.String, updatedMembershipPlan.StripeJoiningFeeID.String)
	require.Equal(t, updateMembershipPlanParams.AmtPeriods.Int32, updatedMembershipPlan.AmtPeriods.Int32)
	require.Equal(t, updateMembershipPlanParams.MembershipID, updatedMembershipPlan.MembershipID)
	require.Equal(t, updateMembershipPlanParams.ID, updatedMembershipPlan.ID)
	require.Equal(t, createdMembershipPlan.CreatedAt, updatedMembershipPlan.CreatedAt)
	require.NotEqual(t, createdMembershipPlan.UpdatedAt, updatedMembershipPlan.UpdatedAt)
}

func TestUpdateMembershipPlanNonExistingMembership(t *testing.T) {

	_, _, _, _, _, queries, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	params := membershipDb.CreateMembershipParams{
		Name:        "Go Basics Practice",
		Description: "Learn Go programming",
		Benefits:    "Free access to all events",
	}

	createdMembership, err := queries.CreateMembership(context.Background(), params)

	require.NoError(t, err)

	createPlanParams := membershipDb.CreateMembershipPlanParams{
		MembershipID: createdMembership.ID,
		Name:         "Go Basics Practice Plan",
		StripeJoiningFeeID: sql.NullString{
			String: "price_123",
			Valid:  true,
		},
		StripePriceID: "price_456",
	}

	createdMembershipPlan, err := queries.CreateMembershipPlan(context.Background(), createPlanParams)

	require.NoError(t, err)

	updateMembershipPlanParams := membershipDb.UpdateMembershipPlanParams{
		Name:          "Updated Membership Name",
		StripePriceID: "price_789",
		StripeJoiningFeeID: sql.NullString{
			String: "price_456",
			Valid:  true,
		},
		AmtPeriods: sql.NullInt32{
			Int32: 12,
			Valid: true,
		},
		MembershipID: uuid.New(),
		ID:           createdMembershipPlan.ID,
	}

	_, err = queries.UpdateMembershipPlan(context.Background(), updateMembershipPlanParams)

	var pqErr *pq.Error

	require.Error(t, err)

	require.ErrorAs(t, err, &pqErr)
	require.Equal(t, "fk_membership", pqErr.Constraint)
}

func TestUpdateNonExistingMembershipPlan(t *testing.T) {

	_, _, _, _, _, queries, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	params := membershipDb.UpdateMembershipPlanParams{
		ID:            uuid.New(),
		Name:          "Updated Membership Name",
		StripePriceID: "price_789",
		StripeJoiningFeeID: sql.NullString{
			String: "price_456",
			Valid:  true,
		},
		AmtPeriods: sql.NullInt32{
			Int32: 12,
			Valid: true,
		},
		MembershipID: uuid.New(),
	}

	_, err := queries.UpdateMembershipPlan(context.Background(), params)
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteMembershipPlan(t *testing.T) {

	_, _, _, _, _, queries, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	params := membershipDb.CreateMembershipParams{
		Name:        "Go Basics Practice",
		Description: "Learn Go programming",
		Benefits:    "Free access to all events",
	}

	createdMembership, err := queries.CreateMembership(context.Background(), params)

	require.NoError(t, err)

	createPlanParams := membershipDb.CreateMembershipPlanParams{
		MembershipID: createdMembership.ID,
		Name:         "Go Basics Practice Plan",
		StripeJoiningFeeID: sql.NullString{
			String: "price_123",
			Valid:  true,
		},
		StripePriceID: "price_456",
	}

	createdMembershipPlan, err := queries.CreateMembershipPlan(context.Background(), createPlanParams)

	require.NoError(t, err)

	impactedRows, err := queries.DeleteMembershipPlan(context.Background(), createdMembershipPlan.ID)

	require.NoError(t, err)

	require.Equal(t, int64(1), impactedRows)
}
