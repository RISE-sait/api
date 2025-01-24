package mapper

import (
	"api/internal/domains/membership/dto"
	entity "api/internal/domains/membership/entities"
)

func MapCreateRequestToEntity(body dto.CreateMembershipRequest) entity.Membership {
	return entity.Membership{
		Name:        body.Name,
		Description: body.Description,
		StartDate:   body.StartDate,
		EndDate:     body.EndDate,
	}
}

func MapUpdateRequestToEntity(body dto.UpdateMembershipRequest) entity.Membership {
	return entity.Membership{
		ID:          body.ID,
		Name:        body.Name,
		Description: body.Description,
		StartDate:   body.StartDate,
		EndDate:     body.EndDate,
	}
}

func MapEntityToResponse(membership *entity.Membership) dto.MembershipResponse {
	return dto.MembershipResponse{
		ID:          membership.ID,
		Name:        membership.Name,
		Description: membership.Description,
		StartDate:   membership.StartDate,
		EndDate:     membership.EndDate,
	}
}
