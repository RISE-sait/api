package discount

import (
	values "api/internal/domains/discount/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"time"
)

type RequestDto struct {
	Name            string `json:"name" validate:"required,notwhitespace"`
	Description     string `json:"description"`
	DiscountPercent int    `json:"discount_percent" validate:"gte=0,lte=100"`
	IsUseUnlimited  bool   `json:"is_use_unlimited"`
	UsePerClient    int    `json:"use_per_client" validate:"omitempty,gt=0"`
	IsActive        bool   `json:"is_active"`
	ValidFrom       string `json:"valid_from" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	ValidTo         string `json:"valid_to" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

// ApplyRequestDto is used to apply a discount code for a customer
type ApplyRequestDto struct {
	Name             string  `json:"name" validate:"required,notwhitespace"`
	MembershipPlanID *string `json:"membership_plan_id,omitempty" validate:"omitempty,uuid"`
}

func (dto *ApplyRequestDto) Validate() *errLib.CommonError {
	return validators.ValidateDto(dto)
}

func (dto *RequestDto) validate() *errLib.CommonError {
	return validators.ValidateDto(dto)
}

func (dto *RequestDto) toValues() (values.CreateValues, *errLib.CommonError) {
	if err := dto.validate(); err != nil {
		return values.CreateValues{}, err
	}

	validFrom, err := time.Parse(time.RFC3339, dto.ValidFrom)
	if err != nil {
		return values.CreateValues{}, errLib.New("Invalid valid_from format. Expected RFC3339", 400)
	}

	var validTo time.Time
	if dto.ValidTo != "" {
		validTo, err = time.Parse(time.RFC3339, dto.ValidTo)
		if err != nil {
			return values.CreateValues{}, errLib.New("Invalid valid_to format. Expected RFC3339", 400)
		}
	}

	return values.CreateValues{
		Name:            dto.Name,
		Description:     dto.Description,
		DiscountPercent: dto.DiscountPercent,
		IsUseUnlimited:  dto.IsUseUnlimited,
		UsePerClient:    dto.UsePerClient,
		IsActive:        dto.IsActive,
		ValidFrom:       validFrom,
		ValidTo:         validTo,
	}, nil
}

func (dto *RequestDto) ToCreateValues() (values.CreateValues, *errLib.CommonError) {
	return dto.toValues()
}

func (dto *RequestDto) ToUpdateValues(idStr string) (values.UpdateValues, *errLib.CommonError) {
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		return values.UpdateValues{}, err
	}
	val, err2 := dto.toValues()
	if err2 != nil {
		return values.UpdateValues{}, err2
	}
	return values.UpdateValues{ID: id, CreateValues: val}, nil
}
