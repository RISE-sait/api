package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Client struct to hold client data
type Client struct {
	FirstName        string
	LastName         string
	Email            string
	Phone            string
	Gender           string
	DateOfBirth      time.Time
	Street           string
	State            string
	City             string
	Country          string
	ZipCode          string
	Source           string
	LastContacted    time.Time
	TotalBookings    int
	LastBooking      time.Time
	TotalAttendances int
	MembershipName   string
	MembershipPlan   string
	MembershipExpiry time.Time
	CreditsRemaining int
	StudioWaiver     bool
	EmailConsent     bool
	SMSConsent       bool
}

var clients []Client

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
	return s == "true" || s == "Yes"
}

// Function to extract data from the CSV file
func extract() ([]Client, error) {
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

		// Parse date fields
		dateOfBirth, _ := time.Parse("2006-01-02", record[6])
		lastContacted, _ := time.Parse("2006-01-02", record[12])
		lastBooking, _ := time.Parse("2006-01-02", record[15])
		membershipExpiry, _ := time.Parse("2006-01-02", record[18])

		// Parse booleans
		studioWaiver := parseBool(record[19])
		emailConsent := parseBool(record[20])
		smsConsent := parseBool(record[21])

		// Create a new Client struct
		client := Client{
			FirstName:        record[1],
			LastName:         record[2],
			Email:            record[3],
			Phone:            record[4],
			Gender:           record[5],
			DateOfBirth:      dateOfBirth,
			Street:           record[7],
			State:            record[8],
			City:             record[9],
			Country:          record[10],
			ZipCode:          record[11],
			Source:           record[12],
			LastContacted:    lastContacted,
			TotalBookings:    parseInt(record[13]),
			LastBooking:      lastBooking,
			TotalAttendances: parseInt(record[16]),
			MembershipName:   record[17],
			MembershipPlan:   record[18],
			MembershipExpiry: membershipExpiry,
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
