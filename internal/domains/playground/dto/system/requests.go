package playground

import (
	values "api/internal/domains/playground/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type RequestDto struct {
	Name string `json:"name" validate:"required,notwhitespace"`
}

func (dto RequestDto) toValue() (values.CreateSystemValue, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return values.CreateSystemValue{}, err
	}
	return values.CreateSystemValue{Name: dto.Name}, nil
}

func (dto RequestDto) ToCreateValue() (values.CreateSystemValue, *errLib.CommonError) {
	return dto.toValue()
}

func (dto RequestDto) ToUpdateValue(idStr string) (values.UpdateSystemValue, *errLib.CommonError) {
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		return values.UpdateSystemValue{}, err
	}
	if err = validators.ValidateDto(&dto); err != nil {
		return values.UpdateSystemValue{}, err
	}
	return values.UpdateSystemValue{ID: id, Name: dto.Name}, nil
}
