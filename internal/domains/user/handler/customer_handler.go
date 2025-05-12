package user

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	
	"api/internal/di"
	athleteDto "api/internal/domains/user/dto/athlete"
	customerDto "api/internal/domains/user/dto/customer"

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

	maxLimit, offset := 20, 0

	var parentID uuid.UUID

	if limitStr := query.Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New("Invalid 'limit' value", http.StatusBadRequest))
			return
		}
		if parsedLimit <= 0 {
			responseHandlers.RespondWithError(w, errLib.New("Limit must be greater than 0", http.StatusBadRequest))
			return
		}
		if parsedLimit > maxLimit {
			responseHandlers.RespondWithError(w, errLib.New(fmt.Sprintf("max limit is %d", maxLimit), http.StatusBadRequest))
			return
		}
		maxLimit = parsedLimit
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New(fmt.Sprintf("Error encountered parsing offset: %s", err.Error()), http.StatusBadRequest))
			return
		}
		if parsedOffset < 0 {
			responseHandlers.RespondWithError(w, errLib.New("Offset must be at least 0", http.StatusBadRequest))
			return
		}
		offset = parsedOffset
	}

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

	dbCustomers, err := h.CustomerRepo.GetCustomers(r.Context(), int32(maxLimit), int32(offset), parentID, searchTerm)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	log.Println("DB Customers: ", len(dbCustomers))

	result := make([]customerDto.Response, len(dbCustomers))

	for i, customer := range dbCustomers {
		response := customerDto.UserReadValueToResponse(customer)

		result[i] = response
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
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
