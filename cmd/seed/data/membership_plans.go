package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
)

func GetMembershipPlans() dbSeed.InsertMembershipPlansParams {
	var (
		nameArray               []string
		stripeJoiningFeeIDArray []string
		stripePriceIDArray      []string
		membershipNameArray     []string
		amtPeriodsArray         []int32
	)

	for _, membership := range Memberships {
		for _, plan := range membership.MembershipPlans {
			// Calculate periods
			periods := int32(0)
			if plan.NoOfPeriods != nil {
				periods = int32(*plan.NoOfPeriods)
			}

			// Append all values
			nameArray = append(nameArray, plan.PlanName)
			stripeJoiningFeeIDArray = append(stripeJoiningFeeIDArray, plan.StripeJoiningFeeID)
			stripePriceIDArray = append(stripePriceIDArray, plan.StripePriceID)
			membershipNameArray = append(membershipNameArray, membership.Name)
			amtPeriodsArray = append(amtPeriodsArray, periods)
		}
	}

	return dbSeed.InsertMembershipPlansParams{
		NameArray:               nameArray,
		StripeJoiningFeeIDArray: stripeJoiningFeeIDArray,
		StripePriceIDArray:      stripePriceIDArray,
		MembershipNameArray:     membershipNameArray,
		AmtPeriodsArray:         amtPeriodsArray,
	}
}
