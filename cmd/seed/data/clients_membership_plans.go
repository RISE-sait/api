package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"github.com/google/uuid"
	"math/rand"
	"time"
)

func GetClientsMembershipPlans(clientIds, membershipPlanIds []uuid.UUID) dbSeed.InsertClientsMembershipPlansParams {

	var (
		clientArray      []uuid.UUID
		planArray        []uuid.UUID
		renewalDateArray []time.Time
		startDateArray   []time.Time
	)

	randomSource := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(randomSource)

	for _, clientID := range clientIds {

		lastStartDate := time.Now()

		// Decide randomly whether the client will have 1 or 2 plans
		numPlans := 1 + randomGenerator.Intn(2) // Randomly 1 or 2 plans

		for planIdx := 0; planIdx < numPlans; planIdx++ {

			// Select a random plan
			randomPlanID := membershipPlanIds[randomGenerator.Intn(len(membershipPlanIds))]

			randomMonths := 2 + randomGenerator.Intn(11)

			// Generate the renewal date for the first plan, or 30 days after the last renewal date for the second plan
			renewalDate := lastStartDate.AddDate(0, randomMonths, 0)

			// Append to the arrays
			clientArray = append(clientArray, clientID)
			planArray = append(planArray, randomPlanID)
			renewalDateArray = append(renewalDateArray, renewalDate)
			startDateArray = append(startDateArray, lastStartDate)

			lastStartDate = renewalDate.AddDate(0, 0, 2)
		}
	}

	return dbSeed.InsertClientsMembershipPlansParams{
		CustomerID:       clientArray,
		PlansArray:       planArray,
		StartDateArray:   startDateArray,
		RenewalDateArray: renewalDateArray,
	}
}
