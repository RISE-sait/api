package data

type Location struct {
	Name     string
	Address  string
	IsPublic bool
}

var Locations = []Location{
	{
		Name:     "Check out Tryout Locations via website",
		Address:  "Check website for Venue Address",
		IsPublic: true,
	},
	{
		Name:     "Prolific Sports House North",
		Address:  "292212 Wagon Wheel Blvd, Rocky View, AB T4A 0T5",
		IsPublic: true,
	},
	{
		Name:     "Rise Facility- Calgary Central Sportsplex",
		Address:  "401, 33 St. NE, Calgary AB Entrance #2",
		IsPublic: true,
	},
	{
		Name:     "The Genesis Centre",
		Address:  "7555 Falconridge Blvd NE #10, Calgary, AB T3J 0C9",
		IsPublic: true,
	},
}
