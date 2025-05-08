package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Staff struct {
	FirstName   string
	LastName    string
	Role        string
	PhoneNumber string
	Email       string
	Country     string
}

var staff = []Staff{
	{
		FirstName:   "Gurkaran",
		LastName:    "Sihota",
		Role:        "Receptionist",
		PhoneNumber: "",
		Email:       "gurkaran@risesportscomplex.com",
		Country:     "CA",
	},
	{
		FirstName:   "Jenelle",
		LastName:    "Hawrys",
		Role:        "SuperAdmin",
		PhoneNumber: "604-813-7589",
		Email:       "jenelle.hawrys@treewalk.com",
		Country:     "CA",
	},
	{
		FirstName:   "Kelvin",
		LastName:    "Dela pena",
		Role:        "SuperAdmin",
		PhoneNumber: "403-479-2965",
		Email:       "info@riseup-hoops.com",
		Country:     "CA",
	},
	{
		FirstName:   "Rise Front Desk",
		LastName:    "Champ Corner",
		Role:        "Receptionist",
		PhoneNumber: "",
		Email:       "bre@risesportscomplex.com",
		Country:     "CA",
	},
	{
		FirstName:   "Sait",
		LastName:    "Developer",
		Role:        "SuperAdmin",
		PhoneNumber: "",
		Email:       "rise.development@outlook.com",
		Country:     "CA",
	},
	{
		FirstName:   "Steve",
		LastName:    "S",
		Role:        "SuperAdmin",
		PhoneNumber: "",
		Email:       "ssnider236@gmail.com",
		Country:     "CA",
	},
	{
		FirstName:   "Steve",
		LastName:    "Snider",
		Role:        "Receptionist",
		PhoneNumber: "587-834-5823",
		Email:       "s75snider@gmail.com",
		Country:     "CA",
	},
	{
		FirstName:   "Sunny",
		LastName:    "Sarpal",
		Role:        "Admin",
		PhoneNumber: "",
		Email:       "sunny@risesportscomplex.com",
		Country:     "CA",
	},
	{
		FirstName:   "Test",
		LastName:    "Trainer",
		Role:        "Coach",
		PhoneNumber: "",
		Email:       "viktor.djurasic+1@abcfitness.com",
		Country:     "CA",
	},
	{
		FirstName:   "test",
		LastName:    "admin",
		Role:        "SuperAdmin",
		PhoneNumber: "403 466 1009",
		Email:       "testadmin@rise.com",
		Country:     "CA",
	},
	{
		FirstName:   "Sukhdeep",
		LastName:    "Singh",
		Role:        "Admin",
		PhoneNumber: "",
		Email:       "sukhdeepboparai2005@gmail.com",
		Country:     "CA",
	},
	{
		FirstName:   "Coach",
		LastName:    "Mike",
		Role:        "Coach",
		PhoneNumber: "",
		Email:       "coach@test.com",
		Country:     "IR",
	},
	{
		FirstName:   "Barber",
		LastName:    "Dan",
		Role:        "Barber",
		PhoneNumber: "",
		Email:       "barber@test.com",
		Country:     "IQ",
	},
	{
		FirstName:   "Barber",
		LastName:    "Anthony",
		Role:        "Barber",
		PhoneNumber: "",
		Email:       "barber.anthony@test.com",
		Country:     "IN",
	},
	{
		FirstName:   "Instructor",
		LastName:    "Mike",
		Role:        "Instructor",
		PhoneNumber: "+15875656565",
		Email:       "instructor@test.com",
		Country:     "US",
	},
	{
		FirstName:   "Denis",
		LastName:    "Torjai",
		Role:        "Admin",
		PhoneNumber: "",
		Email:       "denistorjai@gmail.com",
		Country:     "CA",
	},
	{
		FirstName:   "Mostapha",
		LastName:    "Alahmair",
		Email:       "mostapha4567@gmail.com",
		PhoneNumber: "+9645876648922",
		Country:     "IQ",
		Role:        "Coach",
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

func GetStaffsAsClients() dbSeed.InsertUsersParams {

	var (
		firstNameArray           []string
		lastNameArray            []string
		emailArray               []string
		countryArray             []string
		genderArray              []string
		phoneArray               []string
		dobArray                 []time.Time
		hasMarketingConsentArray []bool
		hasSMSConsentArray       []bool
		parentIDArray            []uuid.UUID
	)

	defaultDOB, err := time.Parse("2006-01-02", "2000-01-01")
	if err != nil {
		panic("Failed to parse default date of birth: " + err.Error())
	}

	for _, s := range staff {
		firstNameArray = append(firstNameArray, s.FirstName)
		lastNameArray = append(lastNameArray, s.LastName)
		countryArray = append(countryArray, s.Country)
		genderArray = append(genderArray, "N")
		phoneArray = append(phoneArray, s.PhoneNumber)
		dobArray = append(dobArray, defaultDOB)
		emailArray = append(emailArray, s.Email)
		hasMarketingConsentArray = append(hasMarketingConsentArray, false)
		hasSMSConsentArray = append(hasSMSConsentArray, false)
		parentIDArray = append(parentIDArray, uuid.Nil)
	}

	// Return staff insert parameters
	return dbSeed.InsertUsersParams{
		CountryAlpha2CodeArray:        countryArray,
		FirstNameArray:                firstNameArray,
		LastNameArray:                 lastNameArray,
		DobArray:                      dobArray,
		GenderArray:                   genderArray,
		ParentIDArray:                 parentIDArray,
		PhoneArray:                    phoneArray,
		EmailArray:                    emailArray,
		HasMarketingEmailConsentArray: hasMarketingConsentArray,
		HasSmsConsentArray:            hasSMSConsentArray,
	}
}
