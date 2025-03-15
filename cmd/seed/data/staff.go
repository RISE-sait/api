package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"strings"
)

type Staff struct {
	FirstName   string
	LastName    string
	Role        string
	PhoneNumber string `json:"ph_no"`
	Email       string
}

func GetStaffs() dbSeed.InsertStaffParams {

	var (
		firstNameArray []string
		lastNameArray  []string
		roleArray      []string
		phoneArray     []string
		emailArray     []string
		isActiveArray  []bool // Add this line
	)

	// Create a slice of Staff instances
	var staff = []Staff{
		{
			FirstName:   "Gurkaran",
			LastName:    "Sihota",
			Role:        "Receptionist",
			PhoneNumber: "",
			Email:       "gurkaran@risesportscomplex.com",
		},
		{
			FirstName:   "Jenelle",
			LastName:    "Hawrys",
			Role:        "SuperAdmin",
			PhoneNumber: "604-813-7589",
			Email:       "jenelle.hawrys@treewalk.com",
		},
		{
			FirstName:   "Kelvin",
			LastName:    "Dela pena",
			Role:        "SuperAdmin",
			PhoneNumber: "403-479-2965",
			Email:       "info@riseup-hoops.com",
		},
		{
			FirstName:   "Rise Front Desk",
			LastName:    "Champ Corner",
			Role:        "Receptionist",
			PhoneNumber: "",
			Email:       "bre@risesportscomplex.com",
		},
		{
			FirstName:   "Sait",
			LastName:    "Developer",
			Role:        "SuperAdmin",
			PhoneNumber: "",
			Email:       "rise.development@outlook.com",
		},
		{
			FirstName:   "Steve",
			LastName:    "S",
			Role:        "SuperAdmin",
			PhoneNumber: "",
			Email:       "ssnider236@gmail.com",
		},
		{
			FirstName:   "Steve",
			LastName:    "Snider",
			Role:        "Receptionist",
			PhoneNumber: "587-834-5823",
			Email:       "s75snider@gmail.com",
		},
		{
			FirstName:   "Sunny",
			LastName:    "Sarpal",
			Role:        "Admin",
			PhoneNumber: "",
			Email:       "sunny@risesportscomplex.com",
		},
		{
			FirstName:   "Test",
			LastName:    "Trainer",
			Role:        "Coach",
			PhoneNumber: "",
			Email:       "viktor.djurasic+1@abcfitness.com",
		},
	}

	for _, s := range staff {
		firstNameArray = append(firstNameArray, s.FirstName)
		lastNameArray = append(lastNameArray, s.LastName)
		roleArray = append(roleArray, strings.ToLower(s.Role))
		phoneArray = append(phoneArray, s.PhoneNumber)
		emailArray = append(emailArray, s.Email)
		isActiveArray = append(isActiveArray, true) // Add this line
	}

	// Return staff insert parameters
	return dbSeed.InsertStaffParams{
		Emails:        emailArray,
		IsActiveArray: isActiveArray,
		RoleNameArray: roleArray,
	}
}
