package configs

import (
	"database/sql"
	"log"
	"os"
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
	dbConnUrl         string
	GoogleAuthConfig  googleAuthConfig
	JwtConfig         jwtConfig
	HubSpotApiKey     string
	GmailSmtpPassword string
}

var Envs = initConfig()

func initConfig() *config {

	return &config{
		dbConnUrl:     getEnv("DATABASE_URL", ""),
		HubSpotApiKey: getEnv("HUBSPOT_API_KEY", ""),
		GoogleAuthConfig: googleAuthConfig{
			ClientId:          getEnv("GOOGLE_AUTH_CLIENT_ID", ""),
			ClientSecret:      getEnv("GOOGLE_AUTH_CLIENT_SECRET", ""),
			GoogleRedirectUrl: getEnv("GOOGLE_REDIRECT_URL", ""),
		},
		JwtConfig: jwtConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Issuer: getEnv("JWT_ISSUER", ""),
		},
		GmailSmtpPassword: getEnv("GMAIL_SMTP_PWD", ""),
	}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetDBConnection() *sql.DB {
	connStr := Envs.dbConnUrl
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return dbConn
}
