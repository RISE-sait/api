package data

type Practice struct {
	Name                   string
	Description            string
	Schedules              []Schedule
	Capacity               int
	MembershipsEligibility []MembershipsEligibility
}

type MembershipsEligibility struct {
	Name string
	*EligibleMembership
}

type EligibleMembership struct {
	Price int
}

type Schedule struct {
	ProgramStartDate string
	ProgramEndDate   string
	Day              string
	EventStartTime   string
	EventEndTime     string
	Location         string
	TrainerNames     []string
}

var Practices = []Practice{
	{
		Name:        "Monday- Shooting Class",
		Description: "Shooting Class, All ages",
		Capacity:    30,
		MembershipsEligibility: []MembershipsEligibility{
			//{Name: "PAYG", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "1. Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "2. Jr.Rise Elite Hooper (Ages 5-8)"},
			{Name: "2025 Spring Club Membership", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "3. Seasonal Membership- Winter Rise League", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "4. High School Pro Club"},
			{Name: "5. Gym Membership"},
			{Name: "Jr. Rise Seasonal (3 Months)"},
			{Name: "Open Gym- Strength Room and Courts"},
			{Name: "PAYMENT PLAN 2025 SPRING CLUB", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "Rise Full Year Family Member Guided Strength Gym Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "Seasonal member - Rise WINTER LEAGUE", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "SPRING RISE LEAGUE 2025"},
			{Name: "Strength Room Unlimited Membership"},
		},
		Schedules: []Schedule{
			{
				Day:            "Monday",
				EventStartTime: "17:30",
				EventEndTime:   "18:30",
				//Location:         "Check out Tryout location via website",
				TrainerNames: []string{"Test_Trainer"},
			},
			{
				Day:            "Monday",
				EventStartTime: "18:30",
				EventEndTime:   "19:30",
				//Location:         "Check out Tryout location via website",
				TrainerNames: []string{"Test_Trainer"},
			},
		},
	},
	{
		Name:        "OPEN GYM/DROP IN-Select Courts",
		Description: "Enjoy full access to one of our open courts for basketball during designated times. Be sure to check availability and schedule. See front desk for available court. Courts may be subject to change.",
		Capacity:    1000,
		MembershipsEligibility: []MembershipsEligibility{
			//{Name: "PAYG", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "1. Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "2. Jr.Rise Elite Hooper (Ages 5-8)", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "2025 Spring Club Membership", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "3. Seasonal Membership- Winter Rise League", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "4. High School Pro Club", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "5. Gym Membership"},
			{Name: "Jr. Rise Seasonal (3 Months)"},
			{Name: "Open Gym- Strength Room and Courts"},
			{Name: "PAYMENT PLAN 2025 SPRING CLUB", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "Rise Full Year Family Member Guided Strength Gym Membership"},
			{Name: "Seasonal member - Rise WINTER LEAGUE", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "SPRING RISE LEAGUE 2025"},
			{Name: "Strength Room Unlimited Membership"},
		},
		Schedules: []Schedule{
			{
				Day:            "Wednesday",
				EventStartTime: "16:00",
				EventEndTime:   "23:00",
				Location:       "Rise Facility- Calgary Central Sportsplex",
				TrainerNames:   []string{"Test_Trainer"},
			},
			{
				Day:            "Saturday",
				EventStartTime: "13:00",
				EventEndTime:   "23:00",
				Location:       "Rise Facility- Calgary Central Sportsplex",
				TrainerNames:   []string{"Test_Trainer"},
			},
		},
	},
	{
		Name:        "Thursday Strength",
		Description: "Strength",
		Capacity:    15,
		MembershipsEligibility: []MembershipsEligibility{
			//{Name: "PAYG", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "1. Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "2. Jr.Rise Elite Hooper (Ages 5-8)"},
			{Name: "2025 Spring Club Membership", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "3. Seasonal Membership- Winter Rise League", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "4. High School Pro Club"},
			{Name: "5. Gym Membership"},
			{Name: "Jr. Rise Seasonal (3 Months)"},
			{Name: "Open Gym- Strength Room and Courts", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "PAYMENT PLAN 2025 SPRING CLUB", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "Rise Full Year Family Member Guided Strength Gym Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "Seasonal member - Rise WINTER LEAGUE", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "SPRING RISE LEAGUE 2025"},
			{Name: "Strength Room Unlimited Membership"},
		},
		Schedules: []Schedule{
			{
				Day:            "Thursday",
				EventStartTime: "17:30",
				EventEndTime:   "18:30",
				//Location:         "Check out Tryout location via website",
				TrainerNames: []string{"Test_Trainer"},
			},
			{
				Day:            "Thursday",
				EventStartTime: "18:30",
				EventEndTime:   "19:30",
				//Location:         "Check out Tryout location via website",
				TrainerNames: []string{"Test_Trainer"},
			},
		},
	},
	{
		Name:        "Tuesday-Ball Handling/Skills",
		Description: "Ball Handling/ Skills",
		Capacity:    30,
		MembershipsEligibility: []MembershipsEligibility{
			//{Name: "PAYG", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "1. Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "2. Jr.Rise Elite Hooper (Ages 5-8)"},
			{Name: "2025 Spring Club Membership", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "3. Seasonal Membership- Winter Rise League", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "4. High School Pro Club"},
			{Name: "5. Gym Membership"},
			{Name: "Jr. Rise Seasonal (3 Months)"},
			{Name: "Open Gym- Strength Room and Courts", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "PAYMENT PLAN 2025 SPRING CLUB", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "Rise Full Year Family Member Guided Strength Gym Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "Seasonal member - Rise WINTER LEAGUE", EligibleMembership: &EligibleMembership{Price: 15}},
			{Name: "SPRING RISE LEAGUE 2025"},
			{Name: "Strength Room Unlimited Membership"},
		},
		Schedules: []Schedule{
			{
				Day:            "Tuesday",
				EventStartTime: "17:30",
				EventEndTime:   "18:30",
				//Location:         "Check out Tryout location via website",
				TrainerNames: []string{"Test_Trainer"},
			},
			{
				Day:            "Tuesday",
				EventStartTime: "18:30",
				EventEndTime:   "19:30",
				//Location:         "Check out Tryout location via website",
				TrainerNames: []string{"Test_Trainer"},
			},
		},
	},
	{
		Name:        "All Hs Girls (Gr. 10-12s) Spring Club Tryouts // March 10th, 2025",
		Description: "March 10th, 2025  6PM - 8M  Non-refundable",
		Capacity:    300,
		MembershipsEligibility: []MembershipsEligibility{
			//{Name: "PAYG", EligibleMembership: &EligibleMembership{Price: 25}},
			{Name: "1. Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "2. Jr.Rise Elite Hooper (Ages 5-8)"},
			{Name: "2025 Spring Club Membership"},
			{Name: "3. Seasonal Membership- Winter Rise League", EligibleMembership: &EligibleMembership{Price: 25}},
			{Name: "4. High School Pro Club"},
			{Name: "5. Gym Membership"},
			{Name: "Jr. Rise Seasonal (3 Months)"},
			{Name: "Open Gym- Strength Room and Courts"},
			{Name: "PAYMENT PLAN 2025 SPRING CLUB"},
			{Name: "Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 0}},
			{Name: "Rise Full Year Family Member Guided Strength Gym Membership"},
			{Name: "Seasonal member - Rise WINTER LEAGUE", EligibleMembership: &EligibleMembership{Price: 25}},
			{Name: "SPRING RISE LEAGUE 2025"},
			{Name: "Strength Room Unlimited Membership"},
		},
		Schedules: []Schedule{
			{
				ProgramStartDate: "2025-03-10",
				ProgramEndDate:   "2025-03-10",
				Day:              "Monday",
				EventStartTime:   "18:00",
				EventEndTime:     "20:00",
				Location:         "Rise Facility- Calgary Central Sportsplex",
				TrainerNames:     []string{"Test_Trainer"},
			},
		},
	},
	//{
	//	Name:        "APRIL Spring Break Camp",
	//	Description: "Join us for skills, drills and fun on the Court! DATES: April 22, 23, 24, 25 TIMES: 10AM-3:30PM Please bring indoor shoes, a ball, water bottles, Lunch and Snacks",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    300,
	//},
	//{
	//	Name:        "BOYS U12/U13 Spring Club Tryouts",
	//	Description: "JANUARY 10, 2025 8:00 - 9:45PM Court 1, 2 and 3 AND JANUARY 12, 2025 11:00 - 12:45PM Court 1, 2 and 3 Address: Rise Facility Non-refundable",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    500,
	//},
	//{
	//	Name:        "BOYS U11 Spring Club Tryouts",
	//	Description: "January 09, 2025 6:00PM - 8:15PM Court 1 and 2 January 10, 2025 6:00 - 7:30 PM Court 2 Address: Rise Facility Non-refundable",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    500,
	//},
	//{
	//	Name:        "BOYS U14/U15 Spring Club Tryouts",
	//	Description: "January 12, 2025 7:15 - 9:30PM Court 1, 2 and 3 January 13, 2025  8:00 - 9:45PM Court 1, 2 and 3 Address: Rise Facility Non-refundable",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    483,
	//},
	//{
	//	Name:        "Drop In",
	//	Description: "Drop in access to Rise 3 courts",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    100,
	//},
	//{
	//	Name:        "GIRLS U11 Spring Club Tryouts",
	//	Description: "January 10, 2025 6:00PM - 7:30PM Address: Rise Facility Court 1 Non-refundable",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    100,
	//},
	//{
	//	Name:        "GIRLS U12/U13 Spring Club Tryouts",
	//	Description: "January 10, 2025 5:30PM - 7:30PM Court 3 & Surge Court January 12, 2025 1:00 - 2:30PM Court 1&2 Non-refundable Address: Rise Facility",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    500,
	//},
	//{
	//	Name:        "GIRLS U14/U15 Spring Club Tryouts",
	//	Description: "January 12, 2025 5:00 -7:00PM Court 1, 2 and 3 January 13, 2025 5:30 - 7:30PM Court 1, 2 and 3 Address: Rise Facility",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    500,
	//},
	//{
	//	Name:        "Holiday Hoops Academy (Winter Camp 2024) Ages 9-17",
	//	Description: "Join us for 4 days filled with Skills, Games, Fundamentals and Competitive drills all with a holiday twist! DEC 21 1PM-7PM DEC 22 130PM-730PM DEC 23 9AM-4PM",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    300,
	//},
	//{
	//	Name:        "MARCH Spring Break Camp",
	//	Description: "Join us for skills, drills and fun on the court. DATES: March 24, 25, 26, 27, 28 TIME: 10AM-3:30PM Please bring indoor shoes, a ball, water bottles, Lunch and Snacks",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    300,
	//},
	//{
	//	Name:        "Pro Rise Club",
	//	Description: "WARNING ** SERIOUS ATHLETES ONLY** 4x per week Strength & Conditioning to get you prepared and stronger. Designed to build endurance and resilience for the season ahead. $650+gst for non-members",
	//	Location:    "Check out Tryout location via website",
	//	Capacity:    400,
	//},
	//{
	//	Name:        "Rise & Honor Memorial Cup",
	//	Description: "Available Age Groups: U11 Boys & Girls U13 Boys & Girls U15 Boys & Girls U17 Boys & Girls U18 Boys & Girls",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    100,
	//},
	//{
	//	Name:        "Rising Stars Camp (Winter Camp 2024) Ages 10-13",
	//	Description: "Join us for 3 days filled with Skills, Games, Fundamentals and Competitive drills all with a holiday twist! DEC 20 6PM-8PM, DEC 21 9AM-12PM, DEC 22 10AM-1PM",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    300,
	//},
	//{
	//	Name:        "U11 CO-ED Winter League Assessments (Tier 3)",
	//	Description: "January 17, 2025 6:00-7:45PM Court 1 and 2 Address: Rise Facility Non-Refundable",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    500,
	//},
	//{
	//	Name:        "U13 BOYS Winter Rise League Assessments (Tier 3)",
	//	Description: "January 17, 2025 8:00 -9:30 PM Court 1 and 2 Address: Rise Facility Non-Refundable",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    500,
	//},
	//{
	//	Name:        "U13/U15 GIRLS Winter Rise League Assessments (Tier 3)",
	//	Description: "January 19, 2025 12:30 - 2:30PM Court 1, 2 and 3 Address: Rise Facility Non-Refundable",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    500,
	//},
	//{
	//	Name:        "U15 BOYS Winter League Assessments (Tier 3)",
	//	Description: "January 19, 2025 10:00 - 12:00PM Court 1, 2 and 3 Address: Rise Facility Non-Refundable",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    500,
	//},
	//{
	//	Name:        "U16 Boys (Gr. 10) Spring Club Tryouts // March 10th",
	//	Description: "March 10th, 2025 8PM - 10PM",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    300,
	//},
	//{
	//	Name:        "U17/U18 Boys (Gr. 11 & 12) Spring Club Tryouts // March 11th",
	//	Description: "March 11th, 2025 6PM - 8PM Non-refundable",
	//	Location:    "Check out Tryout location via website",
	//	Capacity:    300,
	//},
	//{
	//	Name:        "U17/U18 Girls (Gr. 11 & 12) Spring Club Tryouts // March 10th",
	//	Description: "March 10th, 2025 6PM - 8PM Non-refundable",
	//	Location:    "Rise Facility- Calgary Central Sportsplex",
	//	Capacity:    300,
	//},
}
