package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/stripe/stripe-go/v81"

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
	DbConnUrl                        string
	GoogleAuthConfig                 googleAuthConfig
	JwtConfig                        jwtConfig
	HubSpotApiKey                    string
	GmailSmtpPassword                string
	GcpServiceAccountCredentialsJSON string
	StripeSecretKey                  string
	StripeWebhookSecret              string
	ChatBotServiceUrl                string
	FrontendBaseURL                  string // Frontend URL for email verification (supports Universal Links for mobile app)
}

var Env = initConfig()

// initConfig initializes and returns a configuration object containing various environment variables.
// It retrieves sensitive data (such as API keys and database URLs) from environment variables and
// sets them in the config struct. It also configures the Stripe API key.
//
// Returns:
//   - *config: A pointer to the config struct containing all necessary configuration values.
//
// Example usage:
//
//	cfg := initConfig()  // Initializes the configuration with environment variables.
func initConfig() *config {
	stripeApiKey := getEnv("STRIPE_SECRET_KEY")

	stripe.Key = stripeApiKey

	return &config{
		DbConnUrl:     getEnv("DATABASE_URL"),
		HubSpotApiKey: getEnv("HUBSPOT_API_KEY"),
		JwtConfig: jwtConfig{
			Secret: getEnv("JWT_SECRET"),
			Issuer: getEnv("JWT_ISSUER"),
		},
		GmailSmtpPassword:                getEnv("GMAIL_SMTP_PWD"),
		GcpServiceAccountCredentialsJSON: getEnv("GCP_SERVICE_ACCOUNT_CREDENTIALS"),
		StripeSecretKey:                  getEnv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:              getEnv("STRIPE_WEBHOOK_SECRET"),
		ChatBotServiceUrl:                getEnv("CHAT_BOT_SERVICE_URL"),
		FrontendBaseURL:                  getEnv("FRONTEND_BASE_URL"),
	}
}

// getEnv retrieves the value of an environment variable identified by the key.
// If the variable is found, its value is returned. If not, the behavior depends on the calmIfNotExist parameter.
//
// If calmIfNotExist is omitted or set to true, it returns an empty string. If set to false, it panics.
// By default, it doesn't panic.
//
// Parameters:
//   - key: The environment variable to retrieve.
//   - calmIfNotExist: If true (or omitted), returns an empty string when not found. If false, panics.
//
// Returns:
//   - string: The value of the environment variable, or an empty string if not found and calmIfNotExist is true.
//
// Example usage:
//
//	value := getEnv("MY_ENV_VAR")  // Returns an empty string if MY_ENV_VAR is not set.
//	value := getEnv("MY_ENV_VAR", true)  // Same behavior, explicitly returns an empty string.
//	value := getEnv("MY_ENV_VAR", false)  // Panics if MY_ENV_VAR is not set.
func getEnv(key string, calmIfNotExist ...bool) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	// If calmIfNotExist is not provided, len(calmIfNotExist) will be 0
	// which means we should not panic
	if len(calmIfNotExist) > 0 && !calmIfNotExist[0] {
		panic("Environment variable " + key + " not set")
	}

	return ""
}

// GetDBConnection establishes a connection to the database using the connection string from the environment variable.
// It opens a connection to a PostgreSQL database and logs any errors that occur during the process.
//
// Returns:
//   - *sql.DB: A pointer to the PostgreSQL database connection.
//
// Example usage:
//
//	dbConn := GetDBConnection()  // Establishes a connection to the database using the connection string.
func GetDBConnection() *sql.DB {
	connStr := Env.DbConnUrl

	log.Println(connStr)
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return dbConn
}
