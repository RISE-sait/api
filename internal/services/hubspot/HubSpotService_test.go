package hubspot

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestGetUserByID(t *testing.T) {

	loadEnv(t)

	apiKey := os.Getenv("HUBSPOT_API_KEY")
	require.NotEmpty(t, apiKey, "HUBSPOT_API_KEY must be set")

	hubspotService := GetHubSpotService(&apiKey)

	user, hubspotErr := hubspotService.GetUserById("102789928823")

	require.Nil(t, hubspotErr)

	email := user.Properties.Email

	// Assert course data
	require.Equal(t, "klintlee1@gmail.com", email)
}

func TestGetUserByNonExistentID(t *testing.T) {

	loadEnv(t)

	apiKey := os.Getenv("HUBSPOT_API_KEY")
	require.NotEmpty(t, apiKey, "HUBSPOT_API_KEY must be set")

	hubspotService := GetHubSpotService(&apiKey)

	user, hubspotErr := hubspotService.GetUserById("abcdnwfwefwefewee")

	require.NotNil(t, hubspotErr)

	require.Contains(t, hubspotErr.Message, "Error getting customer with id abcdnwfwefwefewee")

	require.Nil(t, user)

}

func loadEnv(t *testing.T) {

	if err := godotenv.Load("../../../config/.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
}
