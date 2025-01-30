package controller

// import (
// 	"api/internal/domains/identity/customer_registration/dto"
// 	errLib "api/internal/libs/errors"
// 	response_handlers "api/internal/libs/responses"
// 	credentials_dto "api/internal/shared/accounts"
// 	"encoding/json"
// 	"io"
// 	"net/http"
// )

// type ChildAccountConfirmationController struct {
// 	ChildAccountRegistrationService *customer_registration.ChildAccountRegistrationService
// }

// func NewChildAccountConfirmationController(childAccountRegistrationService *customer_registration.ChildAccountRegistrationService) *ChildAccountConfirmationController {
// 	return &ChildAccountConfirmationController{
// 		ChildAccountRegistrationService: childAccountRegistrationService,
// 	}
// }

// func (c *ChildAccountConfirmationController) CreatePendingChildAccount(w http.ResponseWriter, r *http.Request) {

// 	body, ioErr := io.ReadAll(r.Body)

// 	if ioErr != nil {
// 		response_handlers.RespondWithError(w, errLib.New("Failed to read request body", http.StatusBadRequest))
// 		return
// 	}

// 	var credentials credentials_dto.Credentials
// 	var customerWaiverDto dto.CustomerWaiverCreateDto
// 	var childAccountDto dto.CreateChildAccountDto

// 	if ioErr := json.Unmarshal(body, &credentials); ioErr != nil {
// 		response_handlers.RespondWithError(w, errLib.New("Invalid JSON format for credentials", http.StatusBadRequest))
// 		return
// 	}

// 	if ioErr := json.Unmarshal(body, &customerWaiverDto); ioErr != nil {
// 		response_handlers.RespondWithError(w, errLib.New("Invalid JSON format for waiver data", http.StatusBadRequest))
// 		return
// 	}

// 	if ioErr := json.Unmarshal(body, &childAccountDto); ioErr != nil {
// 		response_handlers.RespondWithError(w, errLib.New("Invalid format for parent email", http.StatusBadRequest))
// 		return
// 	}

// 	credentialsCreate := credentials_dto.NewCredentials(credentials.Email, credentials.Password)
// 	waiverCreate := dto.NewCustomerWaiverCreateDto(customerWaiverDto.WaiverUrl, customerWaiverDto.IsWaiverSigned)
// 	childAccountCreate := dto.NewChildAccountCreateDto(childAccountDto.ParentEmail)

// 	// Step 2: Call the service to create the account
// 	err := c.ChildAccountRegistrationService.CreatePendingAccount(r.Context(), credentialsCreate, waiverCreate, childAccountCreate)
// 	if err != nil {
// 		response_handlers.RespondWithError(w, err)
// 		return
// 	}

// 	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
// }
