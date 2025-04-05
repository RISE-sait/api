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
		Name:        "Rise Basketball Full Year Membership",
		Description: "Our Full Year Membership offers monthly fees for athletes in Tier 1, 2, and 3, with a one-time annual fee of $100, ensuring convenience and affordability. Please see our website for full details.",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "GIRLS U11 Full Year Membership",
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
			{
				PlanName:           "BOYS U11 Full Year Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1R9siLAB1pU7Ebknc2rVHUmO",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "BOYS U13 Full Year Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1R9sjIAB1pU7EbknMxpxBoTt",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "BOYS U15 Full Year Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1R9sjgAB1pU7Ebkn6jduGXQT",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "BOYS HIGH SCHOOL Full Year Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1R9slVAB1pU7Ebkn3BKYPZY0",
				NoOfPeriods:        toPointer(26),
			},
		},
	},

	{
		Name:        "Jr.Rise Elite Hooper (Ages 5-8)",
		Description: "Jr. Rise for ages 5 to 8, where young athletes learn fundamental skills, build confidence, and develop a love for basketball in a fun & supportive environment. Please see our website for full details.",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "FULL YEAR MEMBERSHIP Jr.Rise Hooper",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1RA7MAAB1pU7EbknpkvwLmyp",
				StripePriceID:      "price_1RA7KUAB1pU7EbknnE6ANxAa",
				NoOfPeriods:        toPointer(12),
			},
		},
	},
	{
		Name:        "2025 Spring Club Membership",
		Description: "Club 2025 Jan-July",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "U10 GIRLS 2025 Spring Club Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RA7TjAB1pU7Ebknc334Mtaj",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U11 GIRLS 2025 Spring Club Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RA7NgAB1pU7EbknhCPQR6nR",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U12 GIRLS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7O6AB1pU7EbknONkh4lwX",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U13 GIRLS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7OVAB1pU7EbknH5VYGEGJ",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U14 GIRLS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7OqAB1pU7Ebkn1bcK4T89",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U15 GIRLS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7PKAB1pU7Ebkno2zqMvM3",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U16 GIRLS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7PgAB1pU7EbknW3IV8mF4",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U17 GIRLS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7Q1AB1pU7EbknyGWuDEV6",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U18 GIRLS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7QIAB1pU7Ebkn41LTt7Sv",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U10 BOYS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7QdAB1pU7Ebkn9T0ougLk",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U11 BOYS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7R3AB1pU7Ebknz8dyea78",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U12 BOYS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7RvAB1pU7EbkngDttGRlA",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U13 BOYS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7RaAB1pU7Ebkndf1SwC2N",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U14 BOYS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7SCAB1pU7EbknLFnVzfYk",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U15 BOYS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7SWAB1pU7EbknSmtzwjpR",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U16 BOYS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7SmAB1pU7EbknsU6UJThh",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U17 BOYS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7T4AB1pU7EbknIYZStQAN",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "U18 BOYS 2025 Spring Club Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RA7TMAB1pU7Ebkn8osBcIUr",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
		},
	},
	{
		Name:        "Seasonal Membership- Winter Rise League",
		Description: "Join our Winter Rise League (3 months) and experience competitive basketball with 2-3 weekly sessions, available for beginners and club-level athletes. See our website for full details.",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "GIRLS U11 WINTER LEAGUE- Seasonal 3 months",
				Type:               "Membership",
				StripePriceID:      "price_1RA7jpAB1pU7EbknIHJNkVKd",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "GIRLS U13  WINTER LEAGUE- Seasonal 3 months",
				Type:               "Membership",
				StripePriceID:      "price_1RA7kBAB1pU7EbknKOPhoXX2",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "GIRLS U15  WINTER LEAGUE- Seasonal 3 months",
				Type:               "Membership",
				StripePriceID:      "price_1RA7kaAB1pU7EbknEhbFTd5G",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "GIRLS HIGH SCHOOL WINTER LEAGUE- Seasonal 3 months",
				Type:               "Membership",
				StripePriceID:      "price_1RA7ktAB1pU7EbknBO4536kb",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "BOYS U11  WINTER LEAGUE- Seasonal 3 months",
				Type:               "Membership",
				StripePriceID:      "price_1RA7lGAB1pU7EbknZa4SDeXW",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "BOYS U13  WINTER LEAGUE- Seasonal 3 months",
				Type:               "Membership",
				StripePriceID:      "price_1RA7lYAB1pU7Ebkn6WRSms5R",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "BOYS U15  WINTER LEAGUE- Seasonal 3 months",
				Type:               "Membership",
				StripePriceID:      "price_1RA7lxAB1pU7EbknvL2Q2jzV",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "BOYS HIGH SCHOOL  WINTER LEAGUE- Seasonal 3 months",
				Type:               "Membership",
				StripePriceID:      "price_1RA7nWAB1pU7EbknVY0SKEvP",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
		},
	},

	{
		Name:        "High School Pro Club",
		Description: "WARNING ** SERIOUS ATHLETES ONLY** 4x per week Strength & Conditioning to get you prepared and stronger. Designed to build endurance and resilience for the season ahead. $650+gst for non-members (3 installment payments)",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "NON-MEMBER High School Pro Club- 3 Month",
				Type:               "Membership",
				StripePriceID:      "price_1RA7urAB1pU7EbknqPyrRvGg",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "FULL YEAR MEMBER High School Pro Club",
				Type:               "Membership",
				StripePriceID:      "price_1RA7vkAB1pU7EbknqAxA1bjh",
				StripeJoiningFeeID: "",
			},
		},
	},
	{
		Name:        "Gym Membership",
		Description: "Access to weight room and drop-in hours on the court.",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "Gym Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RA82pAB1pU7EbknQZybnhEn",
				NoOfPeriods:        nil,
			},
		},
	},

	{
		Name:        "Jr. Rise Seasonal (3 Months)",
		Description: ".",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "Jr. Rise Seasonal (3 Months)",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RA850AB1pU7EbknAY9J8gRV",
				NoOfPeriods:        nil,
			},
		},
	},
	{
		Name:        "Open Gym- Strength Room and Courts",
		Description: "**ONLY $10 FOR THE MONTH OF JANUARY** For those who know their way around the weight room or looking to get shots up on our courts, this membership is for you! Utilize the strength room and/or courts during our open gym hours. **Credits will be added for promo**",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "Open Gym- Strength Room and Courts",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RA8BhAB1pU7Ebkn9LvG61HY",
				NoOfPeriods:        nil,
			},
		},
	},
	{
		Name:        "PAYMENT PLAN 2025 SPRING CLUB",
		Description: "5 PAYMENTS OVER 5 MONTHS",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "PAYMENT PLAN 2025 SPRING CLUB",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RA8D2AB1pU7EbknrZLaoIBg",
				NoOfPeriods:        toPointer(5),
			},
		},
	},

	{
		Name:        "Rise Basketball Full Year Membership No.2",
		Description: "Our Full Year Membership offers monthly fees for athletes in Tier 1, 2, and 3, with a one-time annual fee of $100, ensuring convenience and affordability. Please see our website for full details.",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "Boys Rise Basketball Full Year Membership Ages 11 - 13 - Beginners (Plan 1)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8IgAB1pU7Ebkn8fuIs8LK",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Boys Rise Basketball Full Year Membership Ages 11 - 13 - Beginners (Plan 2)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8J8AB1pU7Ebkn9NcKuEFy",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Boys Rise Basketball Full Year Membership Ages 13 - 15 - Beginners (Plan 3)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8KLAB1pU7EbknDTzl265z",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Boys Rise Basketball Full Year Membership Ages 13 - 15 Advanced (Plan 4)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8KtAB1pU7EbknHczLLyGF",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Boys Rise Basketball Full Year Membership Ages 15 - 18 Beginners (Plan 5)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8LJAB1pU7EbkneEdaXJwb",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Boys Rise Basketball Full Year Membership Ages 15 - 18 Beginners (Plan 6)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8LnAB1pU7EbknPJ7Xhxna",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Girls Rise Basketball Full Year Membership Ages 11 - 13 Beginners (Plan 7)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8MBAB1pU7Ebknrics9Pnu",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Girls Rise Basketball Full Year Membership Ages 11 - 13 Advanced (Plan 8)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8MhAB1pU7EbknHqDfXyGe",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Girls Rise Basketball Full Year Membership Ages 13 - 15 Beginners (Plan 9)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8N7AB1pU7EbknmwoKygX2",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Girls Rise Basketball Full Year Membership Ages 13 - 15 Advanced (Plan 10)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8NeAB1pU7EbknAhRCxGd8",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Girls Rise Basketball Full Year Membership 15- 18 Beginners (Plan 11)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8O8AB1pU7Ebkn9abMHKcc",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Girls Rise Basketball Full Year Membership 15- 18 Advanced (Plan 12)",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8ObAB1pU7EbknOeCOXhY4",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Boys Rise Basketball Full Year Membership Ages 9 - 11 - All Levels",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8P0AB1pU7Ebkn02tF5vWy",
				NoOfPeriods:        toPointer(26),
			},
			{
				PlanName:           "Girls Rise Basketball Full Year Membership Ages 9 - 11 - All Levels",
				Type:               "Membership",
				StripeJoiningFeeID: "price_1R9tAJAB1pU7EbknWTKZTWbD",
				StripePriceID:      "price_1RA8PgAB1pU7EbknHKP1R59U",
				NoOfPeriods:        toPointer(26),
			},
		},
	},

	{
		Name:        "Rise Full Year Family Member Guided Strength Gym Membership",
		Description: "**For family members who have a child in the RISE FULL YEAR program** While your child is training, you can do the same and focus on your health and strength. These sessions are unlimited and guided by a certified strength trainer. *Members will be cross-referenced with athletes in programs*",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "Rise Family Member Guided Strength Gym Membership",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RA8jxAB1pU7EbknQFhcG6cX",
				NoOfPeriods:        nil,
			},
		},
	},
	{
		Name:        "Seasonal member - Rise WINTER LEAGUE",
		Description: "Join our Winter Rise League for Non-Members and experience competitive basketball with 2-3 weekly sessions, available for beginners and club-level athletes. See our website for full details.",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "Boys Seasonal Member - Rise League Beginners Ages 11-13 (Winter) (Plan 1)",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RAIdfAB1pU7EbknH8nFXK7m",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "Boys Seasonal Member - Rise League Advanced Ages 11-13 (Winter) (Plan 2)",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RAIdrAB1pU7Ebkn56UEvb7d",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "Boys Seasonal Member - Rise League Beginners Ages 13 - 15 (Winter) (Plan 3)",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RAIe2AB1pU7Ebkn0pNxW7lj",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "Boys Seasonal Member - Rise League Advanced Ages 13 - 15 (Winter) (Plan 4)",
				Type:               "Membership",
				StripeJoiningFeeID: "",
				StripePriceID:      "price_1RAIeDAB1pU7EbknbIqw3El0",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "Boys Seasonal Member - Rise League Beginners Ages 15 - 17 (Winter) (Plan 5)",
				Type:               "Membership",
				StripePriceID:      "price_1RAIeQAB1pU7EbknKD5hXuqd",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "Boys Seasonal Member - Rise League Club Level Ages 15 - 17(Winter) (Plan 6)",
				Type:               "Membership",
				StripePriceID:      "price_1RAIegAB1pU7Ebkn2UopH5dM",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "Girls Seasonal Member - Rise League Beginners 11 - 13 (Winter) (Plan 1)",
				Type:               "Membership",
				StripePriceID:      "price_1RAIevAB1pU7EbknTVgAXAFg",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "Girls Seasonal Member - Rise League Club Level 11 - 13 (Winter) (Plan 2)",
				Type:               "Membership",
				StripePriceID:      "price_1RAIf5AB1pU7Ebkn7QWKfmxe",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "Girls Seasonal Member - Rise League Beginners 13 - 15 (Winter) (Plan 3)",
				Type:               "Membership",
				StripePriceID:      "price_1RAIfGAB1pU7EbknHspoOUse",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "Girls Seasonal Member - Rise League Club level 13 - 15 (Winter) (Plan 4)",
				Type:               "Membership",
				StripePriceID:      "price_1RAIfSAB1pU7EbknhMpN0SWK",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "HS Boys Seasonal Member - Winter Rise League Beginners Grade.10-12",
				Type:               "Membership",
				StripePriceID:      "price_1RAIfgAB1pU7EbknQhB9wyQO",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "HS Girls Seasonal Member - Winter Rise League Club level Grade.10-12",
				Type:               "Membership",
				StripePriceID:      "price_1RAIfqAB1pU7EbknuzSb4UeI",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "3. Seasonal member - Rise (Winter Program)  9-11 BOYS (Plan 13)",
				Type:               "Membership",
				StripePriceID:      "price_1RAIg2AB1pU7EbknHp411qgR",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "3. Seasonal member - Rise (Winter Program) 9-11 GIRLS (Plan 14)",
				Type:               "Membership",
				StripePriceID:      "price_1RAIgEAB1pU7EbknNn8yjrHp",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
		},
	},

	{
		Name:        "SPRING RISE LEAGUE 2025",
		Description: "RISE SPRING LEAGUE IS HERE! Didn’t make your club team? Don’t stop now—keep grinding! Join RISE SPRING LEAGUE and take your game to the next level. ✅ Elite Training & Development ✅ Competitive Games ✅ State-of-the-Art Facility ✅ A Community That Supports Your Growth This is your chance to improve, compete, and get ready for the next level. Spots are limited—don’t miss out!",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "SPRING RISE LEAGUE 2025 U11 CO-ED DIV",
				Type:               "Membership",
				StripePriceID:      "price_1RAJAbAB1pU7EbknaS4eyk06",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "SPRING RISE LEAGUE 2025 U13/U15 GIRLS DIV",
				Type:               "Membership",
				StripePriceID:      "price_1RAJAoAB1pU7EbknSPwfkkSj",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "SPRING RISE LEAGUE 2025 U13 BOYS DIV",
				Type:               "Membership",
				StripePriceID:      "price_1RAJB1AB1pU7Ebkn2x1DZYkN",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
			{
				PlanName:           "SPRING RISE LEAGUE 2025 U15 BOYS DIV",
				Type:               "Membership",
				StripePriceID:      "price_1RAJBIAB1pU7Ebknyr1EIyY2",
				StripeJoiningFeeID: "",
				NoOfPeriods:        toPointer(3),
			},
		},
	},

	{
		Name:        "Strength Room Unlimited Membership",
		Description: "Push yourself further with unlimited access to our strength weight room with sessions guided by certified strength trainers.",
		MembershipPlans: []MembershipPlan{
			{
				PlanName:           "Strength Room Unlimited Membership",
				Type:               "Membership",
				StripePriceID:      "price_1RAJEOAB1pU7EbknIH4e3bBu",
				StripeJoiningFeeID: "",
				NoOfPeriods:        nil,
			},
		},
	},
}
