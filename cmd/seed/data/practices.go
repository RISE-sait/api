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
	Price float64
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
			{Name: "PAYG", EligibleMembership: &EligibleMembership{Price: 15}},
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
				ProgramStartDate: "2025-03-03",
				Day:              "Monday",
				EventStartTime:   "17:30",
				EventEndTime:     "18:30",
				Location:         "Check out Tryout Locations via website",
				TrainerNames:     []string{"Test_Trainer"},
			},
			{
				Day:              "Monday",
				ProgramStartDate: "2025-03-03",
				EventStartTime:   "18:31",
				EventEndTime:     "19:30",
				Location:         "Check out Tryout Locations via website",
				TrainerNames:     []string{"Test_Trainer"},
			},
		},
	},
	{
		Name:        "OPEN GYM/DROP IN-Select Courts",
		Description: "Enjoy full access to one of our open courts for basketball during designated times. Be sure to check availability and schedule. See front desk for available court. Courts may be subject to change.",
		Capacity:    1000,
		MembershipsEligibility: []MembershipsEligibility{
			{Name: "PAYG", EligibleMembership: &EligibleMembership{Price: 15}},
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
				Day:              "Saturday",
				ProgramStartDate: "2024-12-21",
				ProgramEndDate:   "2025-07-05",
				EventStartTime:   "13:00",
				EventEndTime:     "23:00",
				Location:         "Rise Facility- Calgary Central Sportsplex",
				TrainerNames:     []string{"Test_Trainer"},
			},
		},
	},
	{
		Name:        "Saturday Strength",
		Description: "Strength",
		Capacity:    15,
		MembershipsEligibility: []MembershipsEligibility{
			{Name: "PAYG", EligibleMembership: &EligibleMembership{Price: 15}},
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
				Day:            "Saturday",
				EventStartTime: "09:00",
				EventEndTime:   "10:00",
				Location:       "Check out Tryout Locations via website",
				TrainerNames:   []string{"Test_Trainer"},
			},
		},
	},
	{
		Name:        "APRIL Spring Break Camp",
		Description: "Join us for skills, drills and fun on the Court! DATES: April 22, 23, 24, 25TIMES: 10AM-3:30PMPlease bring indoor shoes, a ball, water bottles, Lunch and Snacks",
		Capacity:    300,
		MembershipsEligibility: []MembershipsEligibility{
			{Name: "PAYG", EligibleMembership: &EligibleMembership{Price: 252}},
			{Name: "Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 226.8}},
			{Name: "Jr.Rise Elite Hooper (Ages 5-8)"},
			{Name: "2025 Spring Club Membership", EligibleMembership: &EligibleMembership{Price: 252}},
			{Name: "Seasonal Membership- Winter Rise League", EligibleMembership: &EligibleMembership{Price: 252}},
			{Name: "High School Pro Club"},
			{Name: "Gym Membership"},
			{Name: "Jr. Rise Seasonal (3 Months)"},
			{Name: "Open Gym- Strength Room and Courts"},
			{Name: "PAYMENT PLAN 2025 SPRING CLUB", EligibleMembership: &EligibleMembership{Price: 252}},
			{Name: "Rise Basketball Full Year Membership", EligibleMembership: &EligibleMembership{Price: 226.8}},
			{Name: "Rise Full Year Family Member Guided Strength Gym Membership"},
			{Name: "Seasonal member - Rise WINTER LEAGUE", EligibleMembership: &EligibleMembership{Price: 252}},
			{Name: "SPRING RISE LEAGUE 2025"},
			{Name: "Strength Room Unlimited Membership"},
		},
		Schedules: []Schedule{
			{
				ProgramStartDate: "2025-04-22",
				ProgramEndDate:   "2025-04-22",
				Day:              "Tuesday",
				EventStartTime:   "10:00",
				EventEndTime:     "15:30",
				Location:         "Rise Facility- Calgary Central Sportsplex",
				TrainerNames:     []string{"Test_Trainer"},
			},
		},
	},
	{
		Name:        "Rise & Honor Memorial Cup",
		Description: "Available Age Groups: U11 Boys & Girls U13 Boys & Girls U15 Boys & Girls U17 Boys & Girls U18 Boys & Girls",
		Capacity:    100,
		MembershipsEligibility: []MembershipsEligibility{
			{Name: "PAYG", EligibleMembership: &EligibleMembership{Price: 650}},
			{Name: "Clients", EligibleMembership: &EligibleMembership{Price: 650}},
		},
		Schedules: []Schedule{
			{
				ProgramStartDate: "2025-05-30",
				ProgramEndDate:   "2025-06-01",
				Day:              "Friday",
				EventStartTime:   "10:00",
				EventEndTime:     "20:00",
				Location:         "Rise Facility- Calgary Central Sportsplex",
				TrainerNames:     []string{"Test_Trainer"},
			},
		},
	},
}
