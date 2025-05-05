package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/price"
	_ "github.com/lib/pq"
)

type Plan struct {
	ID            string
	StripePriceID string
}
// This script syncs Stripe pricing data (unit_amount, currency, interval) for all membership plans
// that have a valid stripe_price_id. It fetches each price from Stripe using the API, then updates
// the corresponding row in the membership_plans table with the real Stripe price details.
func main() {
	err := godotenv.Load("config/.env.local")
	if err != nil {
		log.Fatalf("Error loading .env.local: %v", err)
	}
	fmt.Println("Using DB:", os.Getenv("DATABASE_URL"))

	// Load Stripe API Key
	stripe.Key = os.Getenv("STRIPE_API_KEY")

	// Connect to your Postgres DB
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// Fetch all plans with stripe_price_id
	rows, err := db.Query(`
		SELECT id, stripe_price_id 
		FROM membership.membership_plans 
		WHERE stripe_price_id IS NOT NULL
	`)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var plan Plan
		if err := rows.Scan(&plan.ID, &plan.StripePriceID); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		// Call Stripe API
		priceObj, err := price.Get(plan.StripePriceID, nil)
		if err != nil {
			log.Printf("Failed to fetch Stripe price for %s: %v", plan.StripePriceID, err)
			continue
		}

		unitAmount := int(priceObj.UnitAmount)
		currency := string(priceObj.Currency)
		interval := ""
		if priceObj.Recurring != nil {
			interval = string(priceObj.Recurring.Interval)
		}

		// Update DB row
		_, err = db.Exec(`
			UPDATE membership.membership_plans 
			SET unit_amount = $1, currency = $2, interval = $3 
			WHERE id = $4
		`, unitAmount, strings.ToUpper(currency), interval, plan.ID)

		if err != nil {
			log.Printf("Failed to update plan %s: %v", plan.ID, err)
		} else {
			fmt.Printf("Updated %s → $%.2f %s / %s\n", plan.ID, float64(unitAmount)/100, strings.ToUpper(currency), interval)
		}
	}
}
