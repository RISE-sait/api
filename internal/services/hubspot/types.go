package hubspot

import (
	"time"
)

type HubSpotCustomersResponse struct {
	Results []HubSpotCustomerResponse `json:"results"`
}

type HubSpotCustomerResponse struct {
	ID           string               `json:"id"`
	Properties   HubSpotCustomerProps `json:"properties"`
	UpdatedAt    time.Time            `json:"updatedAt"`
	Associations HubSpotAssociation   `json:"associations"`
}

type HubSpotCustomerCreateBody struct {
	Properties HubSpotCustomerProps `json:"properties"`
}

type HubSpotCustomerProps struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
}

type HubSpotAssociation struct {
	Contact HubSpotCustomerAssociation `json:"contacts"`
}

type HubSpotCustomerAssociation struct {
	Result []struct {
		Id   string `json:"id"`
		Type string `json:"type"`
	} `json:"results"`
}
