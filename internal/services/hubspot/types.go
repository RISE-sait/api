package hubspot

import (
	"time"
)

type UsersResponse struct {
	Results []UserResponse `json:"results"`
}

type AssociationInput struct {
	AssociationCategory string `json:"associationCategory"`
	AssociationTypeId   int    `json:"associationTypeId"`
}

type AssociationEndpoint struct {
	ID string `json:"id"`
}

type UserResponse struct {
	HubSpotId    string           `json:"id"`
	Properties   UserProps        `json:"properties"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
	Associations UserAssociations `json:"associations"`
}

type UserProps struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Age       string `json:"age"`
}

type UserAssociations struct {
	Contact UserAssociation `json:"contacts"`
}

type UserAssociation struct {
	Result []UserAssociationResult `json:"results"`
}

type UserAssociationResult struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type UserCreationBody struct {
	Properties UserProps `json:"properties"`
}

type ChildCreationBody struct {
	Properties UserProps `json:"properties"`
	ParentId   string    `json:"parentId"`
}
