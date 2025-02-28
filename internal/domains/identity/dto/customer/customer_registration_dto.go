package customer

import (
	"api/internal/domains/identity/dto/common"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

// RegistrationDto represents the data transfer object for customer registration.
// It includes user information, waivers, Firebase authentication token, and age.
type RegistrationDto struct {
	CustomerWaiversSigningDto []WaiverSigningDto `json:"waivers" validate:"required,dive"`
	identity.UserNecessaryInfoDto
}

// toValueObjectBase validates the DTO and converts waiver signing details into value objects.
// Returns a slice of CustomerWaiverSigning value objects and an error if validation fails.
func (dto RegistrationDto) toValueObjectBase() ([]values.CustomerWaiverSigning, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return nil, err
	}

	waiversVo := make([]values.CustomerWaiverSigning, len(dto.CustomerWaiversSigningDto))
	for i, waiver := range dto.CustomerWaiversSigningDto {
		waiversVo[i] = values.CustomerWaiverSigning{
			IsWaiverSigned: waiver.IsWaiverSigned,
			WaiverUrl:      waiver.WaiverUrl,
		}
	}

	return waiversVo, nil
}

// ToCreateRegularCustomerValueObject converts the DTO into a RegularCustomerRegistrationInfo value object.
// Requires an email address as input. Returns the value object and an error if validation fails.
func (dto RegistrationDto) ToCreateRegularCustomerValueObject(email string) (*values.RegularCustomerRegistrationInfo, *errLib.CommonError) {

	waiversVo, err := dto.toValueObjectBase()

	if err != nil {
		return nil, err
	}

	vo := values.RegularCustomerRegistrationInfo{
		UserNecessaryInfo: values.UserNecessaryInfo{
			Age:       dto.Age,
			FirstName: dto.FirstName,
			LastName:  dto.LastName,
		},
		Email:   email,
		Waivers: waiversVo,
	}

	return &vo, nil
}

// ToCreateChildValueObject converts the DTO into a ChildRegistrationInfo value object.
// Requires a parent email as input. Returns the value object and an error if validation fails.
func (dto RegistrationDto) ToCreateChildValueObject(parentEmail string) (*values.ChildRegistrationInfo, *errLib.CommonError) {

	waiversVo, err := dto.toValueObjectBase()

	if err != nil {
		return nil, err
	}

	vo := values.ChildRegistrationInfo{
		UserNecessaryInfo: values.UserNecessaryInfo{
			Age:       dto.Age,
			FirstName: dto.FirstName,
			LastName:  dto.LastName,
		},
		ParentEmail: parentEmail,
		Waivers:     waiversVo,
	}

	return &vo, nil
}
