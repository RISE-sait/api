package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"github.com/shopspring/decimal"
)

func GetMembershipPlans() dbSeed.InsertMembershipPlansParams {
	var (
		nameArray             []string
		priceArray            []decimal.Decimal
		joiningFeeArray       []decimal.Decimal
		autoRenewArray        []bool
		membershipNameArray   []string
		paymentFrequencyArray []dbSeed.PaymentFrequency
		amtPeriodsArray       []int32
	)

	for _, membership := range Memberships {
		for _, plan := range membership.MembershipPlans {
			// Calculate periods
			periods := int32(0)
			if plan.PaymentFrequency.HasEndDate.Value {
				periods = int32(plan.PaymentFrequency.HasEndDate.NoOfPeriods)
			}

			// Convert price
			price := decimal.NewFromFloat(plan.PaymentFrequency.Price)

			// Determine payment frequency
			var freq dbSeed.PaymentFrequency
			switch {
			case plan.PaymentFrequency.RecurringPaymentInterval == 2 &&
				plan.PaymentFrequency.PaymentFrequency == "week":
				freq = dbSeed.PaymentFrequencyBiweekly
			default:
				freq = dbSeed.PaymentFrequency(plan.PaymentFrequency.RecurringPeriod)
			}

			// Append all values
			nameArray = append(nameArray, plan.PlanName)
			priceArray = append(priceArray, price)
			joiningFeeArray = append(joiningFeeArray,
				decimal.NewFromFloat(plan.PaymentFrequency.JoiningFee))
			autoRenewArray = append(autoRenewArray,
				plan.PaymentFrequency.HasEndDate.WillPlanAutoRenew)
			membershipNameArray = append(membershipNameArray, membership.Name)
			paymentFrequencyArray = append(paymentFrequencyArray, freq)
			amtPeriodsArray = append(amtPeriodsArray, periods)
		}
	}

	return dbSeed.InsertMembershipPlansParams{
		NameArray:             nameArray,
		PriceArray:            priceArray,
		JoiningFeeArray:       joiningFeeArray,
		AutoRenewArray:        autoRenewArray,
		MembershipNameArray:   membershipNameArray,
		PaymentFrequencyArray: paymentFrequencyArray,
		AmtPeriodsArray:       amtPeriodsArray,
	}
}
