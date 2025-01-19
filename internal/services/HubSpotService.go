package services

import (
	"api/config"
	"net/http"
	"time"
)

// HubSpotService handles integration with HubSpot API.
type HubSpotService struct {
	Client  *http.Client
	ApiKey  string
	BaseURL string
}

func GetHubSpotService() *HubSpotService {
	apiKey := config.Envs.HubSpotApiKey

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &HubSpotService{
		Client:  httpClient,
		ApiKey:  apiKey,
		BaseURL: "https://api.hubapi.com/",
	}
}
