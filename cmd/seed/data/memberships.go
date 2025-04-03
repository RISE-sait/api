package data

type MembershipPlan struct {
	PlanName           string
	Type               string
	StripeJoiningFeeID string
	StripePriceID      string
	NoOfPeriods        *int
}

type Membership struct {
	Name            string
	Description     string
	MembershipPlans []MembershipPlan
}

func toPointer(i int) *int {
	return &i
}

var Memberships = []Membership{
	{
		Name:        "1. Rise Basketball Full Year Membership",
		Description: "Our Full Year Membership offers monthly fees for athletes in Tier 1, 2, and 3, with a one-time annual fee of $100, ensuring convenience and affordability. Please see our website for full details.",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "1. GIRLS U11 Full Year Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1R6i0fAB1pU7EbkntynBbgif",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "GIRLS U13 Full Year Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1R9sfHAB1pU7Ebkng79qqUxh",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "GIRLS U15 Full Year Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1R9shTAB1pU7EbknRm9gCSzM",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "GIRLS HIGH SCHOOL Full Year Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1R9si0AB1pU7EbknlWMWJ5xA",
				NoOfPeriods:        toPointer(26),
			},
		},
	},
}

//		// Additional plans...
//	},
//	{
//		Name:        "2. Jr.Rise Elite Hooper (Ages 5-8)",
//		Description: "Jr. Rise for ages 5 to 8, where young athletes learn fundamental skills, build confidence, and develop a love for basketball in a fun & supportive environment. Please see our website for full details.",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "FULL YEAR MEMBERSHIP Jr.Rise Hooper",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       12,
//						WillPlanAutoRenew: true,
//					},
//					Price:                    135,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 75,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "2025 Spring Club Membership",
//		Description: "Club 2025 Jan-July",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "U11 GIRLS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U12 GIRLS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U13 GIRLS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U14 GIRLS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U15 GIRLS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U16 GIRLS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U17 GIRLS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U18 GIRLS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U10 BOYS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U11 BOYS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U12 BOYS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U13 BOYS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U14 BOYS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U15 BOYS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U16 BOYS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U17 BOYS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "U18 BOYS 2025 Spring Club Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    560,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "3. Seasonal Membership- Winter Rise League",
//		Description: "Join our Winter Rise League (3 months) and experience competitive basketball with 2-3 weekly sessions, available for beginners and club-level athletes. See our website for full details.",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "GIRLS U11 WINTER LEAGUE- Seasonal 3 months",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "GIRLS U13 WINTER LEAGUE- Seasonal 3 months",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "GIRLS U15 WINTER LEAGUE- Seasonal 3 months",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "GIRLS HIGH SCHOOL WINTER LEAGUE- Seasonal 3 months",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "BOYS U11 WINTER LEAGUE- Seasonal 3 months",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "BOYS U13 WINTER LEAGUE- Seasonal 3 months",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "BOYS U15 WINTER LEAGUE- Seasonal 3 months",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "BOYS HIGH SCHOOL WINTER LEAGUE- Seasonal 3 months",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "4. High School Pro Club",
//		Description: "WARNING ** SERIOUS ATHLETES ONLY** 4x per week Strength & Conditioning to get you prepared and stronger. Designed to build endurance and resilience for the season ahead. $650+gst for non-members (3 installment payments)",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "NON-MEMBER High School Pro Club- 3 Month",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    216.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "FULL YEAR MEMBER High School Pro Club",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:         "single",
//					ChargeOn1stOfEveryMonth:  false,
//					Price:                    89,
//					RecurringPaymentInterval: 3,
//					RecurringPeriod:          "month",
//					ServiceProvidedWithPlan:  "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "5. Gym Membership",
//		Description: "Access to weight room and drop-in hours on the court.",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "Gym Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value: false,
//					},
//					Price:                    47,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "Jr. Rise Seasonal (3 Months)",
//		Description: ".",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "Jr. Rise Seasonal (3 Months)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value: false,
//					},
//					Price:                    141.66,
//					RecurringPaymentInterval: 3,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "Open Gym- Strength Room and Courts",
//		Description: "**ONLY $10 FOR THE MONTH OF JANUARY** For those who know their way around the weight room or looking to get shots up on our courts, this membership is for you! Utilize the strength room and/or courts during our open gym hours. **Credits will be added for promo**",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "Open Gym- Strength Room and Courts",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value: false,
//					},
//					Price:                    47,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "PAYMENT PLAN 2025 SPRING CLUB",
//		Description: "5 PAYMENTS OVER 5 MONTHS",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "PAYMENT PLAN 2025 SPRING CLUB",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       5,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    336,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "Rise Basketball Full Year Membership",
//		Description: "Our Full Year Membership offers monthly fees for athletes in Tier 1, 2, and 3, with a one-time annual fee of $100, ensuring convenience and affordability. Please see our website for full details.",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "Boys Rise Basketball Full Year Membership Ages 11 - 13 - Beginners (Plan 1)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       26,
//						WillPlanAutoRenew: true,
//					},
//					Price:                    103.85,
//					RecurringPaymentInterval: 2,
//					RecurringPeriod:          "week", JoiningFee: 100,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "Boys Rise Basketball Full Year Membership Ages 11 - 13 - Advanced (Plan 2)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       26,
//						WillPlanAutoRenew: true,
//					},
//					Price:                    103.85,
//					RecurringPaymentInterval: 2,
//					RecurringPeriod:          "week", JoiningFee: 100,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			// Add similar plans for the other membership plans (Plan 3 to Plan 14) in the same structure as above
//		},
//	},
//	{
//		Name:        "Rise Full Year Family Member Guided Strength Gym Membership",
//		Description: "**For family members who have a child in the RISE FULL YEAR program** While your child is training, you can do the same and focus on your health and strength. These sessions are unlimited and guided by a certified strength trainer. *Members will be cross-referenced with athletes in programs*",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "Rise Family Member Guided Strength Gym Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             false,
//						NoOfPeriods:       0,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    85,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "Seasonal member - Rise WINTER LEAGUE",
//		Description: "Join our Winter Rise League for Non-Members and experience competitive basketball with 2-3 weekly sessions, available for beginners and club-level athletes. See our website for full details.",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "Boys Seasonal Member - Rise League Beginners Ages 11-13 (Winter) (Plan 1)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "Boys Seasonal Member - Rise League Advanced Ages 11-13 (Winter) (Plan 2)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "Boys Seasonal Member - Rise League Beginners Ages 13 - 15 (Winter) (Plan 3)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "Boys Seasonal Member - Rise League Advanced Ages 13 - 15 (Winter) (Plan 4)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "Boys Seasonal Member - Rise League Beginners Ages 15 - 17 (Winter) (Plan 5)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "Boys Seasonal Member - Rise League Club Level Ages 15 - 17(Winter) (Plan 6)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "Girls Seasonal Member - Rise League Beginners 11 - 13 (Winter) (Plan 1)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "Girls Seasonal Member - Rise League Club Level 11 - 13 (Winter) (Plan 2)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "Girls Seasonal Member - Rise League Beginners 13 - 15 (Winter) (Plan 3)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "Girls Seasonal Member - Rise League Club level 13 - 15 (Winter) (Plan 4)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "HS Boys Seasonal Member - Winter Rise League Beginners Grade.10-12",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "HS Girls Seasonal Member - Winter Rise League Club level Grade.10-12",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "3. Seasonal member - Rise (Winter Program)  9-11 BOYS (Plan 13)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "3. Seasonal member - Rise (Winter Program) 9-11 GIRLS (Plan 14)",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226.66,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "SPRING RISE LEAGUE 2025",
//		Description: "RISE SPRING LEAGUE IS HERE! Didn’t make your club team? Don’t stop now—keep grinding! Join RISE SPRING LEAGUE and take your game to the next level. ✅ Elite Training & Development ✅ Competitive Games ✅ State-of-the-Art Facility ✅ A Community That Supports Your Growth This is your chance to improve, compete, and get ready for the next level. Spots are limited—don’t miss out!",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "SPRING RISE LEAGUE 2025 U11 CO-ED DIV",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "SPRING RISE LEAGUE 2025 U13/U15 GIRLS DIV",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "SPRING RISE LEAGUE 2025 U13 BOYS DIV",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//			{
//				PlanName: "SPRING RISE LEAGUE 2025 U15 BOYS DIV",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value:             true,
//						NoOfPeriods:       3,
//						WillPlanAutoRenew: false,
//					},
//					Price:                    226,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//	{
//		Name:        "Strength Room Unlimited Membership",
//		Description: "Push yourself further with unlimited access to our strength weight room with sessions guided by certified strength trainers.",
//		MembershipPlans: []MembershipPlan{
//			{
//				PlanName: "Strength Room Unlimited Membership",
//				Type:     "Membership",
//				PaymentFrequency: PaymentFrequency{
//					PaymentFrequency:        "recurring",
//					ChargeOn1stOfEveryMonth: false,
//					HasEndDate: struct {
//						Value             bool
//						NoOfPeriods       int
//						WillPlanAutoRenew bool
//					}{
//						Value: false,
//					},
//					Price:                    155,
//					RecurringPaymentInterval: 1,
//					RecurringPeriod:          "month", JoiningFee: 0,
//					ServiceProvidedWithPlan: "unlimited",
//				},
//			},
//		},
//	},
//}
