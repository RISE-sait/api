package mapper

import (
	"api/internal/domains/facility/dto"
	entity "api/internal/domains/facility/entities"
)

// MapCreateRequestToEntity maps CreateFacilityRequest DTO to Facility entity.
func MapCreateRequestToEntity(req dto.CreateFacilityRequest) entity.Facility {
	return entity.Facility{
		Name:           req.Name,
		Location:       req.Location,
		FacilityTypeID: req.FacilityTypeID,
	}
}

// MapUpdateRequestToEntity maps UpdateFacilityRequest DTO to Facility entity.
func MapUpdateRequestToEntity(req dto.UpdateFacilityRequest) entity.Facility {
	return entity.Facility{
		ID:             req.ID,
		Name:           req.Name,
		Location:       req.Location,
		FacilityTypeID: req.FacilityTypeID,
	}
}

// MapEntityToResponse maps Facility entity to FacilityResponse DTO.
func MapEntityToResponse(facility *entity.Facility) dto.FacilityResponse {
	return dto.FacilityResponse{
		ID:             facility.ID,
		Name:           facility.Name,
		Location:       facility.Location,
		FacilityTypeID: facility.FacilityTypeID,
	}
}
