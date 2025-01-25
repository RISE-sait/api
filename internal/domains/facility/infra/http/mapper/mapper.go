package mapper

import (
	entity "api/internal/domains/facility/entities"
	"api/internal/domains/facility/infra/http/dto"
)

// MapEntityToResponse maps Facility entity to FacilityResponse DTO.
func MapEntityToResponse(facility *entity.Facility) dto.FacilityResponse {
	return dto.FacilityResponse{
		ID:             facility.ID,
		Name:           facility.Name,
		Location:       facility.Location,
		FacilityTypeID: facility.FacilityTypeID,
	}
}
