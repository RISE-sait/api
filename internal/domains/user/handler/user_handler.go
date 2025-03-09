package user

import (
	"api/internal/di"
	enrollmentRepo "api/internal/domains/enrollment/persistence"
	enrollmentService "api/internal/domains/enrollment/service"
	dto "api/internal/domains/user/dto/user"
	userRepo "api/internal/domains/user/persistence/repository/user"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/services/hubspot"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"strconv"
)

type UsersHandler struct {
	HubSpotService    *hubspot.Service
	UserRepo          userRepo.RepositoryInterface
	EnrollmentService *enrollmentService.Service
}

func NewUsersHandler(container *di.Container) *UsersHandler {
	return &UsersHandler{
		HubSpotService: container.HubspotService,
		EnrollmentService: enrollmentService.NewEnrollmentService(
			enrollmentRepo.NewEnrollmentRepository(container.Queries.EnrollmentDb),
		),
		UserRepo: userRepo.NewUserRepository(container.Queries.UserDb),
	}
}

// GetUserByEmail retrieves a repository by email.
// @Summary Get a repository by email
// @Description Retrieves a repository using their email address
// @Tags users
// @Accept json
// @Produce json
// @Param email path string true "Email"
// @Success 200 {object} dto.Response "Customer retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid Email"
// @Failure 404 {object} map[string]interface{} "Not Found: Customer not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{email} [get]
func (h *UsersHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	user, err := h.HubSpotService.GetUserByEmail(email)

	if err != nil && err.HTTPCode == http.StatusInternalServerError {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var customerInfo dto.Response

	if user != nil {

		userId, err := h.UserRepo.GetUserIDByHubSpotId(r.Context(), user.HubSpotId)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		customerInfo = dto.Response{
			HubspotId: user.HubSpotId,
			Email:     &user.Properties.Email,
			FirstName: user.Properties.FirstName,
			LastName:  user.Properties.LastName,
			UserID:    *userId,
		}
	}

	responseHandlers.RespondWithSuccess(w, customerInfo, http.StatusOK)
}

// GetChildrenByParentEmail retrieves a repository's children using the parent's email.
// @Summary Get a repository's children by parent email
// @Description Retrieves a repository's children using the parent's email address
// @Tags users
// @Accept json
// @Produce json
// @Param email path string true "Parent Email"
// @Success 200 {array} hubspot.UserResponse "Customer's children retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid Email"
// @Failure 404 {object} map[string]interface{} "Not Found: Parent or children not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{email}/children [get]
func (h *UsersHandler) GetChildrenByParentEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	// Fetch repository's children from HubSpot
	customer, err := h.HubSpotService.GetUserByEmail(email)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	contacts := customer.Associations.Contact.Result

	var childrenIds []string

	// Map HubSpot response to DTO
	for _, contact := range contacts {

		if contact.Type == "child_parent" {

			childrenIds = append(childrenIds, contact.ID)
		}
	}

	hubspotChildren, err := h.HubSpotService.GetUsersByIds(childrenIds)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var children []dto.ChildResponse

	for _, child := range hubspotChildren {

		age, err := strconv.Atoi(child.Properties.Age)
		log.Println(age)
		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("Invalid age format", http.StatusInternalServerError))
			return
		}

		children = append(children, dto.ChildResponse{
			HubspotId: child.HubSpotId,
			FirstName: child.Properties.FirstName,
			LastName:  child.Properties.LastName,
			Age:       age,
		})
	}

	responseHandlers.RespondWithSuccess(w, children, http.StatusOK)
}
