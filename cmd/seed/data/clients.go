package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Client struct to hold client data
type Client struct {
	FirstName        string
	LastName         string
	Age              int
	CountryAlpha2    string
	Email            string
	Phone            string
	Gender           string
	CreditsRemaining int
	StudioWaiver     bool
	EmailConsent     bool
	SMSConsent       bool
}

// Function to parse integers safely
func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

// Function to parse booleans safely
func parseBool(s string) bool {
	return s == "true" || s == "false"
}

// GetClients : extract clients data from the CSV file
func GetClients() ([]Client, error) {

	// Open the CSV file
	file, err := os.Open("cmd/seed/clients.csv")
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}

	// Initialize a slice to hold the clients
	var clients []Client

	// Loop through the records (skipping the header row)
	for i, record := range records {
		if i == 0 {
			// Skip the header row
			continue
		}

		// Parse booleans
		studioWaiver := parseBool(record[23])
		emailConsent := parseBool(record[24])
		smsConsent := parseBool(record[25])

		var gender string

		genderStr := strings.ToLower(record[5])

		switch genderStr {
		case "male":
			{
				gender = "M"
			}
		case "female":
			gender = "F"
		default:
			gender = "N"
		}

		ageStr := record[6]

		age, err := strconv.Atoi(ageStr)

		if err != nil {
			age = 0
		}

		// Create a new Client struct
		client := Client{
			FirstName:        record[1],
			LastName:         record[2],
			Email:            record[3],
			Phone:            record[4],
			Age:              age,
			Gender:           gender,
			CountryAlpha2:    record[8],
			CreditsRemaining: parseInt(record[19]),
			StudioWaiver:     studioWaiver,
			EmailConsent:     emailConsent,
			SMSConsent:       smsConsent,
		}

		// Append the client to the slice
		clients = append(clients, client)
	}

	// Return the list of clients
	return clients, nil
}
