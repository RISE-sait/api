package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"math/rand"
	"time"
)

func GetMembershipCoursesEligibility(membershipIds, courseIds []uuid.UUID) dbSeed.InsertCourseMembershipsEligibilityParams {

	var (
		membershipArray  []uuid.UUID
		courseArray      []uuid.UUID
		eligibilityArray []bool
		pricePerBooking  []decimal.Decimal
	)

	randomSource := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(randomSource)

	for _, membershipID := range membershipIds {

		for _, courseID := range courseIds {

			isEligible := randomGenerator.Intn(3) != 0 // 66% chance of being eligible

			var price decimal.Decimal
			if isEligible {
				// Generate a random price between 10 and 100
				price = decimal.NewFromFloat(10 + randomGenerator.Float64()*90).Round(2)
			}

			courseArray = append(courseArray, courseID)
			membershipArray = append(membershipArray, membershipID)
			pricePerBooking = append(pricePerBooking, price)
			eligibilityArray = append(eligibilityArray, isEligible)
		}
	}

	return dbSeed.InsertCourseMembershipsEligibilityParams{
		CourseIDArray:        courseArray,
		MembershipIDArray:    membershipArray,
		IsEligibleArray:      eligibilityArray,
		PricePerBookingArray: pricePerBooking,
	}
}

func GetMembershipPracticesEligibility(membershipIds, practiceIds []uuid.UUID) dbSeed.InsertPracticeMembershipsEligibilityParams {

	var (
		membershipArray  []uuid.UUID
		practiceArray    []uuid.UUID
		eligibilityArray []bool
		pricePerBooking  []decimal.Decimal
	)

	randomSource := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(randomSource)

	for _, membershipID := range membershipIds {

		for _, practiceID := range practiceIds {

			isEligible := randomGenerator.Intn(3) != 0 // 66% chance of being eligible

			var price decimal.Decimal
			if isEligible {
				// Generate a random price between 10 and 100
				price = decimal.NewFromFloat(10 + randomGenerator.Float64()*90)
			}

			practiceArray = append(practiceArray, practiceID)
			membershipArray = append(membershipArray, membershipID)
			pricePerBooking = append(pricePerBooking, price)
			eligibilityArray = append(eligibilityArray, isEligible)
		}
	}

	return dbSeed.InsertPracticeMembershipsEligibilityParams{
		PracticeIDArray:      practiceArray,
		MembershipIDArray:    membershipArray,
		IsEligibleArray:      eligibilityArray,
		PricePerBookingArray: pricePerBooking,
	}
}
