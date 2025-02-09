package staff

// import (
// 	"api/internal/domains/staff/dto"

// 	response_handlers "api/internal/libs/responses"
// 	"api/internal/libs/validators"
// 	"net/http"
// )

// type StaffController struct {
// 	StaffService *StaffService
// }

// func NewStaffController(accountRegistrationService *StaffService) *StaffController {
// 	return &StaffController{
// 		StaffService: accountRegistrationService,
// 	}
// }

// func (c *StaffController) CreateStaff(w http.ResponseWriter, r *http.Request) {

// 	var staffDto dto.StaffRequestDto

// 	if err := validators.ParseJSON(r.Body, &staffDto); err != nil {
// 		response_handlers.RespondWithError(w, err)
// 		return
// 	}

// 	staffCreate, err := staffDto.ToCreateValueObjects()

// 	if err != nil {
// 		response_handlers.RespondWithError(w, err)
// 		return
// 	}

// 	// Step 2: Call the service to create the account
// 	err = c.StaffService.CreateAccount(r.Context(), staffCreate)
// 	if err != nil {
// 		response_handlers.RespondWithError(w, err)
// 		return
// 	}

// 	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
// }
