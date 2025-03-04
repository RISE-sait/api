package customer

import (
	"api/internal/domains/identity/dto/common"
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

// RegistrationRequestDto represents the data transfer object for customer registration.
// It includes user information, waivers, Firebase authentication token, and age.
type RegistrationRequestDto struct {
	CustomerWaiversSigningDto []WaiverSigningRequestDto `json:"waivers" validate:"required"`
	identity.UserNecessaryInfoRequestDto
}

// toValueObjectBase validates the DTO and converts waiver signing details into value objects.
// Returns a slice of CustomerWaiverSigning value objects and an error if validation fails.
func (dto RegistrationRequestDto) toValueObjectBase() ([]values.CustomerWaiverSigning, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return nil, err
	}

	waiversVo := make([]values.CustomerWaiverSigning, len(dto.CustomerWaiversSigningDto))
	for i, waiver := range dto.CustomerWaiversSigningDto {
		vo, err := waiver.ToValueObjects()

		if err != nil {
			return nil, err
		}

		waiversVo[i] = values.CustomerWaiverSigning{
			IsWaiverSigned: vo.IsWaiverSigned,
			WaiverUrl:      vo.WaiverUrl,
		}
	}

	return waiversVo, nil
}

// ToCreateRegularCustomerValueObject converts the DTO into a RegularCustomerRegistrationRequestInfo value object.
// Requires an email address as input. Returns the value object and an error if validation fails.
func (dto RegistrationRequestDto) ToCreateRegularCustomerValueObject(email string) (*values.RegularCustomerRegistrationRequestInfo, *errLib.CommonError) {

	waiversVo, err := dto.toValueObjectBase()

	if err != nil {
		return nil, err
	}

	vo := values.RegularCustomerRegistrationRequestInfo{
		UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
			Age:       dto.Age,
			FirstName: dto.FirstName,
			LastName:  dto.LastName,
		},
		Email:   email,
		Waivers: waiversVo,
	}

	return &vo, nil
}

// ToCreateChildValueObject converts the DTO into a ChildRegistrationRequestInfo value object.
// Requires a parent email as input. Returns the value object and an error if validation fails.
func (dto RegistrationRequestDto) ToCreateChildValueObject(parentEmail string) (*values.ChildRegistrationRequestInfo, *errLib.CommonError) {

	waiversVo, err := dto.toValueObjectBase()

	if err != nil {
		return nil, err
	}

	vo := values.ChildRegistrationRequestInfo{
		UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
			Age:       dto.Age,
			FirstName: dto.FirstName,
			LastName:  dto.LastName,
		},
		ParentEmail: parentEmail,
		Waivers:     waiversVo,
	}

	return &vo, nil
}
