package user

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"api/internal/di"
	athleteDto "api/internal/domains/user/dto/athlete"
	customerDto "api/internal/domains/user/dto/customer"
	contextUtils "api/utils/context"

	customerRepo "api/internal/domains/user/persistence/repository"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type CustomersHandler struct {
	CustomerRepo *customerRepo.CustomerRepository
}

func NewCustomersHandler(container *di.Container) *CustomersHandler {
	return &CustomersHandler{
		CustomerRepo: customerRepo.NewCustomerRepository(container),
	}
}

// UpdateAthleteStats updates statistics based on the provided athlete ID.
// @Tags athletes
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Athlete ID" // Athlete ID to update stats for
// @Param update_body body customerDto.StatsUpdateRequestDto true "Customer stats update data"
// @Success 204 {object} map[string]interface{} "Customer stats updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /athletes/{id}/stats [patch]
func (h *CustomersHandler) UpdateAthleteStats(w http.ResponseWriter, r *http.Request) {
	athleteIdStr := chi.URLParam(r, "id")

	var requestDto customerDto.StatsUpdateRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	details, err := requestDto.ToUpdateValue(athleteIdStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.CustomerRepo.UpdateStats(r.Context(), details); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// UpdateAthletesTeam updates the team of an athlete.
// @Tags athletes
// @Accept json
// @Produce json
// @Param athlete_id path string true "Athlete ID"
// @Param team_id path string true "Team ID"
// @Security Bearer
// @Success 204 "No Content: Team updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Athlete or team not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /athletes/{athlete_id}/team/{team_id} [put]
func (h *CustomersHandler) UpdateAthletesTeam(w http.ResponseWriter, r *http.Request) {
	athleteIdStr := chi.URLParam(r, "athlete_id")
	teamIdStr := chi.URLParam(r, "team_id")

	athleteID, err := validators.ParseUUID(athleteIdStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	teamID, err := validators.ParseUUID(teamIdStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.CustomerRepo.UpdateAthleteTeam(r.Context(), athleteID, teamID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// UpdateAthleteProfile updates athlete profile information like photo URL.
// @Tags athletes
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Athlete ID" // Athlete ID to update profile for
// @Param update_body body customerDto.AthleteProfileUpdateRequestDto true "Athlete profile update data including photo_url"
// @Success 204 {object} map[string]interface{} "Athlete profile updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Not Found: Athlete not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /athletes/{id}/profile [patch]
func (h *CustomersHandler) UpdateAthleteProfile(w http.ResponseWriter, r *http.Request) {
	athleteIdStr := chi.URLParam(r, "id")

	var requestDto customerDto.AthleteProfileUpdateRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	details, err := requestDto.ToUpdateValue(athleteIdStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Security check: Only admins or the athlete themselves can update the profile
	userRole, roleErr := contextUtils.GetUserRole(r.Context())
	if roleErr != nil {
		responseHandlers.RespondWithError(w, roleErr)
		return
	}

	// If not admin, check if user is updating their own profile
	if userRole != contextUtils.RoleAdmin && userRole != contextUtils.RoleSuperAdmin {
		currentUserID, userErr := contextUtils.GetUserID(r.Context())
		if userErr != nil {
			responseHandlers.RespondWithError(w, userErr)
			return
		}

		if currentUserID != details.ID {
			responseHandlers.RespondWithError(w, errLib.New("You can only update your own profile", http.StatusForbidden))
			return
		}
	}

	if err = h.CustomerRepo.UpdateAthleteProfile(r.Context(), details); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// RemoveAthleteFromTeam removes an athlete from a team.
// @Tags athletes
// @Accept json
// @Produce json
// @Param athlete_id path string true "Athlete ID"
// @Security Bearer
// @Success 204 "No Content: Athlete removed from team"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Athlete not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /athletes/{athlete_id}/team [delete]
func (h *CustomersHandler) RemoveAthleteFromTeam(w http.ResponseWriter, r *http.Request) {
	athleteIdStr := chi.URLParam(r, "athlete_id")

	athleteID, err := validators.ParseUUID(athleteIdStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.CustomerRepo.UpdateAthleteTeam(r.Context(), athleteID, uuid.Nil); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// GetCustomers retrieves a list of customers with optional filtering and pagination.
// @Summary Get customers
// @Description Retrieves a list of customers, optionally filtered by fields like parent ID, name, email, phone, and ID, with pagination support.
// @Tags customers
// @Accept json
// @Produce json
// @Param limit query int false "Number of customers to retrieve (default: 20, max: 20)"
// @Param offset query int false "Number of customers to skip (default: 0)"
// @Param search query string false "Search term to filter customers"
// @Param parent_id query string false "Parent ID to filter customers (example: 123e4567-e89b-12d3-a456-426614174000)"
// @Success 200 {array} customer.Response "List of customers"
// @Failure 400 "Bad Request: Invalid parameters"
// @Failure 500 "Internal Server Error"
// @Router /customers [get]
func (h *CustomersHandler) GetCustomers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limit := 20
	page := 1
	offset := 0
	maxLimit := 20

	// Parse limit
	if limitStr := query.Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			responseHandlers.RespondWithError(w, errLib.New("Invalid 'limit' value", http.StatusBadRequest))
			return
		}
		if parsedLimit > maxLimit {
			responseHandlers.RespondWithError(w, errLib.New(fmt.Sprintf("Max limit is %d", maxLimit), http.StatusBadRequest))
			return
		}
		limit = parsedLimit
	}

	// Parse page (preferred) or offset (fallback)
	if pageStr := query.Get("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
			offset = (page - 1) * limit
		}
	} else if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
			page = (offset / limit) + 1
		} else {
			responseHandlers.RespondWithError(w, errLib.New("Offset must be at least 0", http.StatusBadRequest))
			return
		}
	}

	// Optional filters
	var parentID uuid.UUID
	if parentIdStr := query.Get("parent_id"); parentIdStr != "" {
		id, err := validators.ParseUUID(parentIdStr)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
		parentID = id
	}

	searchTerm := query.Get("search")
	log.Printf("Search term: %s", searchTerm)

	// Fetch paginated data
	dbCustomers, err := h.CustomerRepo.GetCustomers(r.Context(), int32(limit), int32(offset), parentID, searchTerm)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Fetch total count for pagination
	totalCount, err := h.CustomerRepo.CountCustomers(r.Context(), parentID, searchTerm)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to count customers: "+err.Error(), http.StatusInternalServerError))
		return
	}

	// Map results
	result := make([]customerDto.Response, len(dbCustomers))
	for i, customer := range dbCustomers {
		result[i] = customerDto.UserReadValueToResponse(customer)
	}

	// Compose response with pagination metadata
	response := map[string]interface{}{
		"data":  result,
		"page":  page,
		"limit": limit,
		"total": totalCount,
		"pages": int((totalCount + int64(limit) - 1) / int64(limit)), // ceil(total / limit)
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetCustomerByID retrieves a customer by ID.
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} customer.Response "The customer"
// @Failure 400 "Bad Request: Invalid parameters"
// @Failure 500 "Internal Server Error"
// @Router /customers/id/{id} [get]
func (h *CustomersHandler) GetCustomerByID(w http.ResponseWriter, r *http.Request) {
	var id uuid.UUID

	if idStr := chi.URLParam(r, "id"); idStr != "" {
		tempId, err := validators.ParseUUID(idStr)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
		id = tempId
	}

	customer, err := h.CustomerRepo.GetCustomer(r.Context(), id, "")
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := customerDto.UserReadValueToResponse(customer)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetCustomerByEmail retrieves a customer by email
// @Tags customers
// @Accept json
// @Produce json
// @Param email path string true "Customer Email"
// @Success 200 {object} customer.Response "The customer"
// @Failure 400 "Bad Request: Invalid parameters"
// @Failure 500 "Internal Server Error"
// @Router /customers/email/{email} [get]
func (h *CustomersHandler) GetCustomerByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	customer, err := h.CustomerRepo.GetCustomer(r.Context(), uuid.Nil, email)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := customerDto.UserReadValueToResponse(customer)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetAthletes returns a list of all athletes with profile info and stats
// @Summary Get all athletes
// @Description Retrieves a paginated list of athletes with profile details and stats.
// @Tags athletes
// @Accept json
// @Produce json
// @Param limit query int false "Number of athletes to retrieve (default: 10)"
// @Param offset query int false "Number of athletes to skip (default: 0)"
// @Success 200 {array} athleteDto.ResponseAthlete
// @Failure 500 "Internal Server Error"
// @Router /athletes [get]
func (h *CustomersHandler) GetAthletes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit == 0 {
		limit = 10
	}

	athletes, err := h.CustomerRepo.ListAthletes(ctx, int32(limit), int32(offset))
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responses := make([]athleteDto.ResponseAthlete, len(athletes))
	for i, athlete := range athletes {
		responses[i] = athleteDto.FromReadValue(athlete)
	}
	responseHandlers.RespondWithSuccess(w, responses, http.StatusOK)
}

// CheckinCustomer verifies active membership for access scanning.
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} map[string]interface{} "Membership info"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/checkin/{id} [get]
func (h *CustomersHandler) CheckinCustomer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	customerID, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	membership, err := h.CustomerRepo.GetActiveMembershipInfo(r.Context(), customerID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	customer, err := h.CustomerRepo.GetCustomer(r.Context(), customerID, "")
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"customer_id":           customer.ID,
		"first_name":            customer.FirstName,
		"last_name":             customer.LastName,
		"membership_name":       membership.MembershipName,
		"membership_plan_name":  membership.MembershipPlanName,
		"membership_start_date": membership.StartDate,
	}

	if membership.PhotoURL != nil {
		response["photo_url"] = *membership.PhotoURL
	} else if customer.AthleteInfo != nil && customer.AthleteInfo.PhotoURL != nil {
		response["photo_url"] = *customer.AthleteInfo.PhotoURL
	}

	if membership.RenewalDate != nil {
		response["membership_renewal_date"] = membership.RenewalDate
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetMembershipHistory returns membership history for a customer.
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {array} customer.MembershipHistoryResponse "Membership history"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{id}/memberships [get]
func (h *CustomersHandler) GetMembershipHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	customerID, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	history, err := h.CustomerRepo.ListMembershipHistory(r.Context(), customerID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responses := make([]customerDto.MembershipHistoryResponse, len(history))
	for i, hst := range history {
		responses[i] = customerDto.MembershipHistoryValueToResponse(hst)
	}

	responseHandlers.RespondWithSuccess(w, responses, http.StatusOK)
}

// GetMyMembershipHistory returns membership history for the logged-in customer.
// @Tags customers
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} customer.MembershipHistoryResponse "Membership history"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /secure/customers/memberships [get]
func (h *CustomersHandler) GetUserMembershipHistory(w http.ResponseWriter, r *http.Request) {
	customerID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	history, err := h.CustomerRepo.ListMembershipHistory(r.Context(), customerID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responses := make([]customerDto.MembershipHistoryResponse, len(history))
	for i, hst := range history {
		responses[i] = customerDto.MembershipHistoryValueToResponse(hst)
	}

	responseHandlers.RespondWithSuccess(w, responses, http.StatusOK)
}

// ArchiveCustomer archives a customer by ID.
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 204 "No Content: Customer archived"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{id}/archive [post]
func (h *CustomersHandler) ArchiveCustomer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	customerID, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	if err = h.CustomerRepo.ArchiveCustomer(r.Context(), customerID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// UnarchiveCustomer unarchives a customer by ID.
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 204 "No Content: Customer unarchived"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{id}/unarchive [post]
func (h *CustomersHandler) UnarchiveCustomer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	customerID, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	if err = h.CustomerRepo.UnarchiveCustomer(r.Context(), customerID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// ListArchivedCustomers returns archived customers.
// @Tags customers
// @Accept json
// @Produce json
// @Param limit query int false "Number of customers to retrieve (default: 20, max: 20)"
// @Param offset query int false "Number of customers to skip (default: 0)"
// @Success 200 {array} customer.Response "List of archived customers"
// @Failure 400 "Bad Request"
// @Failure 500 "Internal Server Error"
// @Router /customers/archived [get]
func (h *CustomersHandler) ListArchivedCustomers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit := 20
	offset := 0
	if limitStr := query.Get("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
			if limit > 20 {
				limit = 20
			}
		} else {
			responseHandlers.RespondWithError(w, errLib.New("invalid limit", http.StatusBadRequest))
			return
		}
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		} else {
			responseHandlers.RespondWithError(w, errLib.New("invalid offset", http.StatusBadRequest))
			return
		}
	}
	customers, err := h.CustomerRepo.ListArchivedCustomers(r.Context(), int32(limit), int32(offset))
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responses := make([]customerDto.Response, len(customers))
	for i, c := range customers {
		responses[i] = customerDto.UserReadValueToResponse(c)
	}
	responseHandlers.RespondWithSuccess(w, responses, http.StatusOK)
}
