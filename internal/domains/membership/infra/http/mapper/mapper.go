package mapper

import (
	entity "api/internal/domains/membership/entities"
	"api/internal/domains/membership/infra/http/dto"
)

func MapEntityToResponse(membership *entity.Membership) dto.MembershipResponse {
	return dto.MembershipResponse{
		ID:          membership.ID,
		Name:        membership.Name,
		Description: membership.Description,
		StartDate:   membership.StartDate,
		EndDate:     membership.EndDate,
	}
}
