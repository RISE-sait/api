package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Client struct to hold client data
type Client struct {
	FirstName        string
	LastName         string
	DOB              time.Time
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

// convertCountryToAlpha2 converts country names to ISO Alpha-2 codes.
func convertCountryToAlpha2(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "canada":
		return "CA"
	case "united states", "usa":
		return "US"
	case "mexico":
		return "MX"
	case "":
		return "XX" // Unknown
	default:
		return "XX" // Fallback for unsupported
	}
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
		studioWaiver := parseBool(record[22])
		emailConsent := parseBool(record[23])
		smsConsent := parseBool(record[23])

		var gender string

		genderStr := strings.ToLower(record[5])

		switch genderStr {
		case "male":
			gender = "M"
		case "female":
			gender = "F"
		default:
			gender = "N"
		}

		dobOriginal := record[6]

		if dobOriginal == "" {
			dobOriginal = "2000-01-01T00:00:00Z" // Default date of birth
		}

		dob, err := time.Parse(time.RFC3339, dobOriginal)
		if err != nil {
			return nil, fmt.Errorf("error parsing date of birth: %w", err)
		}

		// Create a new Client struct
		client := Client{
			FirstName:        record[1],
			LastName:         record[2],
			Email:            record[3],
			Phone:            record[4],
			DOB:              dob,
			Gender:           gender,
			CountryAlpha2:    convertCountryToAlpha2(record[10]),
			CreditsRemaining: parseInt(record[18]),
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
