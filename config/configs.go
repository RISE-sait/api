package config

import (
	"database/sql"
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
}

var Envs = initConfig()

func initConfig() *config {

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
	}
}

func getEnv(key string, defaultValue string) string {

	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetDBConnection() *sql.DB {
	connStr := Envs.DbConnUrl

	log.Println(connStr)
	dbConn, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}
	return dbConn
}
