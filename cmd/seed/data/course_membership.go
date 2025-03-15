package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"math/rand"
	"time"
)

func GetCourseMemberships(courseIDs, membershipIDs []uuid.UUID) dbSeed.InsertCourseMembershipsEligibilityParams {

	var (
		priceArray      []decimal.Decimal
		isEligibleArray []bool
	)

	randomSource := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(randomSource)

	for range courseIDs {

		isEligible := randomGenerator.Float64() < 0.7

		var pricePerBooking float64
		if isEligible {
			pricePerBooking = 10 + randomGenerator.Float64()*90 // Random price between 10 and 100
		} else {
			pricePerBooking = 0
		}

		priceArray = append(priceArray, decimal.NewFromFloat(pricePerBooking))
		isEligibleArray = append(isEligibleArray, isEligible)
	}

	return dbSeed.InsertCourseMembershipsEligibilityParams{
		CourseIDArray:        courseIDs,
		MembershipIDArray:    membershipIDs,
		IsEligibleArray:      isEligibleArray,
		PricePerBookingArray: priceArray,
	}
}
