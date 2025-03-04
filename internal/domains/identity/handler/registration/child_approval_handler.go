package registration

//
//import (
//	"api/internal/di"
//	"api/internal/domains/identity/service/registration"
//	"api/internal/domains/identity/values"
//	errLib "api/internal/libs/errors"
//	responsehandlers "api/internal/libs/responses"
//	"net/http"
//)
//
//type ChildAccountRegistrationHandlers struct {
//	RegistrationService *registration.ChildRegistrationService
//}
//
//func NewChildAccountRegistrationHandlers(container *di.Container) *ChildAccountRegistrationHandlers {
//	return &ChildAccountRegistrationHandlers{
//		RegistrationService: registration.NewChildAccountRegistrationService(container),
//	}
//}
//
//// ApproveChild approves a child account after receiving approval from the parent.
//// @Summary Approve a child's account
//// @Description Approves a pending child account by linking it to the parent's account.
//// @Tags identity
//// @Accept json
//// @Produce json
//// @Param child query string true "Child's email address"
//// @Success 201 {object} map[string]interface{} "Child account approved successfully"
//// @Failure 400 {object} map[string]interface{} "Bad Request: Missing required query parameters"
//// @Failure 500 {object} map[string]interface{} "Internal Server Error"
//// @Router /register/child/approve [post]
//func (h *ChildAccountRegistrationHandlers) RegisterChild(w http.ResponseWriter, r *http.Request) {
//
//	parentEmail := r.URL.Query().Get("parentEmail")
//
//	args := values.ChildRegistrationInfo{
//		UserNecessaryInfoRequestDto:    values.UserNecessaryInfoRequestDto{},
//		ParentEmail: parentEmail,
//		Waivers:     nil,
//	}
//
//	// Step 2: Call the service to create the account
//	err := h.RegistrationService.CreateChildAccount(r.Context(), childEmail)
//	if err != nil {
//		responsehandlers.RespondWithError(w, err)
//		return
//	}
//
//	responsehandlers.RespondWithSuccess(w, nil, http.StatusCreated)
//}
