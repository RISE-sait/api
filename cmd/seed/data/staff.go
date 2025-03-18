package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"github.com/google/uuid"
	"strings"
)

type Staff struct {
	FirstName   string
	LastName    string
	Role        string
	PhoneNumber string
	Email       string
}

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
	{
		FirstName:   "Coach",
		LastName:    "John",
		Role:        "Coach",
		PhoneNumber: "403 466 1009",
		Email:       "coach@test.com",
	},
}

func GetStaffs() dbSeed.InsertStaffParams {

	var (
		roleArray     []string
		emailArray    []string
		isActiveArray []bool // Add this line
	)

	for _, s := range staff {
		roleArray = append(roleArray, strings.ToLower(s.Role))
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

func GetStaffsAsClients() dbSeed.InsertClientsParams {

	var (
		firstNameArray           []string
		lastNameArray            []string
		emailArray               []string
		countryArray             []string
		genderArray              []string
		phoneArray               []string
		ageArray                 []int32
		hasMarketingConsentArray []bool
		hasSMSConsentArray       []bool
		parentIDArray            []uuid.UUID
	)

	for _, s := range staff {
		firstNameArray = append(firstNameArray, s.FirstName)
		lastNameArray = append(lastNameArray, s.LastName)
		countryArray = append(countryArray, "CA")
		genderArray = append(genderArray, "N")
		phoneArray = append(phoneArray, s.PhoneNumber)
		ageArray = append(ageArray, 1)
		emailArray = append(emailArray, s.Email)
		hasMarketingConsentArray = append(hasMarketingConsentArray, false)
		hasSMSConsentArray = append(hasSMSConsentArray, false)
		parentIDArray = append(parentIDArray, uuid.Nil)
	}

	// Return staff insert parameters
	return dbSeed.InsertClientsParams{
		CountryAlpha2CodeArray:        countryArray,
		FirstNameArray:                firstNameArray,
		LastNameArray:                 lastNameArray,
		AgeArray:                      ageArray,
		GenderArray:                   genderArray,
		ParentIDArray:                 parentIDArray,
		PhoneArray:                    phoneArray,
		EmailArray:                    emailArray,
		HasMarketingEmailConsentArray: hasMarketingConsentArray,
		HasSmsConsentArray:            hasSMSConsentArray,
	}
}
