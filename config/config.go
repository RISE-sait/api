package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type dbConfig struct {
	host     string
	port     string
	user     string
	password string
	name     string
}

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
	dbConfig          dbConfig
	GoogleAuthConfig  googleAuthConfig
	JwtConfig         jwtConfig
	HubSpotApiKey     string
	GmailSmtpPassword string
}

var Envs = initConfig()

func initConfig() *config {

	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current working directory: %v\n", err)
		return nil
	}

	// Print the current working directory
	log.Printf("Current working directory: %s\n", cwd)

	err = godotenv.Load("C:\\Coding\\Rise\\api\\.env")
	if err != nil {
		log.Printf("Error loading .env file: %v\n", err)
		return nil
	}

	return &config{
		dbConfig: dbConfig{
			host:     getEnv("DB_HOST", "localhost"),
			port:     getEnv("DB_PORT", "5432"),
			user:     getEnv("DB_USER", "postgres"),
			password: getEnv("DB_PASSWORD", "root"),
			name:     getEnv("DB_NAME", "mydatabase"),
		},
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

func getConnectionString() string {
	if os.Getenv("ENV") == "production" {
		return "postgresql://" + Envs.dbConfig.user + ":" + Envs.dbConfig.password + "@" + Envs.dbConfig.host + ":" + Envs.dbConfig.port + "/" + Envs.dbConfig.name + "?sslmode=require"
	}
	return "postgresql://postgres:root@localhost:5432/mydatabase?sslmode=disable"
}

func GetDBConnection() *sql.DB {
	connStr := getConnectionString()
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return dbConn
}
