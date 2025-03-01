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

type awsConfig struct {
	AccessKeyId string
	SecretKey   string
}

type config struct {
	DbConnUrl           string
	GoogleAuthConfig    googleAuthConfig
	JwtConfig           jwtConfig
	AwsConfig           awsConfig
	HubSpotApiKey       string
	GmailSmtpPassword   string
	FirebaseCredentials string
}

var Envs = initConfig()

func initConfig() *config {

	return &config{
		DbConnUrl:     getEnv("DATABASE_URL", ""),
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
		AwsConfig: awsConfig{
			AccessKeyId: getEnv("AWS_KEY_ID", ""),
			SecretKey:   getEnv("AWS_SECRET_ACCESS_KEY", ""),
		},
		GmailSmtpPassword:   getEnv("GMAIL_SMTP_PWD", ""),
		FirebaseCredentials: getEnv("firebase_credentials", ""),
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
