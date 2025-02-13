package hubspot

import (
	"time"
)

type HubSpotCustomersResponse struct {
	Results []HubSpotCustomerResponse `json:"results"`
}

type AssociationInput struct {
	AssociationCategory string `json:"associationCategory"`
	AssociationTypeId   int    `json:"associationTypeId"`
}

type AssociationEndpoint struct {
	ID string `json:"id"`
}

type HubSpotCustomerResponse struct {
	ID           string               `json:"id"`
	Properties   HubSpotCustomerProps `json:"properties"`
	CreatedAt    time.Time            `json:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt"`
	Associations HubSpotAssociation   `json:"associations"`
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
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"results"`
}

type HubSpotCustomerCreateBody struct {
	Properties HubSpotCustomerProps `json:"properties"`
}
