package config

import (
	"database/sql"
	"github.com/stripe/stripe-go/v81"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type googleAuthConfig struct {
	ClientId          string
	ClientSecret      string
	GoogleRedirectUrl string
}

type jwtConfig struct {
	Secret string
	Issuer string
}

type config struct {
	DbConnUrl                    string
	GoogleAuthConfig             googleAuthConfig
	JwtConfig                    jwtConfig
	HubSpotApiKey                string
	GmailSmtpPassword            string
	GcpServiceAccountCredentials string
	SquareAccessToken            string
	StripeSecretKey              string
	StripeWebhookSecret          string
}

var Env = initConfig()

func initConfig() *config {

	stripeApiKey := os.Getenv("STRIPE_API_KEY")

	stripe.Key = stripeApiKey

	return &config{
		DbConnUrl:     getEnv("DATABASE_URL", ""),
		HubSpotApiKey: getEnv("HUBSPOT_API_KEY", ""),
		JwtConfig: jwtConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Issuer: getEnv("JWT_ISSUER", ""),
		},
		GmailSmtpPassword:            getEnv("GMAIL_SMTP_PWD", ""),
		GcpServiceAccountCredentials: getEnv("GCP_SERVICE_ACCOUNT_CREDENTIALS", ""),
		SquareAccessToken:            getEnv("SQUARE_ACCESS_TOKEN", ""),
		StripeSecretKey:              getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:          getEnv("STRIPE_WEBHOOK_SECRET", ""),
	}
}

func getEnv(key string, defaultValue string) string {

	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetDBConnection() *sql.DB {
	connStr := Env.DbConnUrl

	log.Println(connStr)
	dbConn, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}
	return dbConn
}
