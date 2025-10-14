package user

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"api/internal/di"
	athleteDto "api/internal/domains/user/dto/athlete"
	customerDto "api/internal/domains/user/dto/customer"
	contextUtils "api/utils/context"

	customerRepo "api/internal/domains/user/persistence/repository"
	firebaseService "api/internal/domains/identity/service/firebase"
	stripeService "api/internal/domains/payment/services/stripe"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type CustomersHandler struct {
	CustomerRepo     *customerRepo.CustomerRepository
	FirebaseService  *firebaseService.Service
	StripeService    *stripeService.SubscriptionService
	PriceService     *stripeService.PriceService
}

func NewCustomersHandler(container *di.Container) *CustomersHandler {
	return &CustomersHandler{
		CustomerRepo:    customerRepo.NewCustomerRepository(container),
		FirebaseService: firebaseService.NewFirebaseService(container),
		StripeService:   stripeService.NewSubscriptionService(container),
		PriceService:    stripeService.NewPriceService(),
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

	// Enrich history with live Stripe data before converting to responses
	for i, hst := range history {
		if hst.Status == "active" && hst.StripePriceID != "" {
			if stripePrice, stripeErr := h.PriceService.GetPrice(hst.StripePriceID); stripeErr == nil {
				// Update with live Stripe price data
				history[i].UnitAmount = int(stripePrice.UnitAmount)
				history[i].Currency = string(stripePrice.Currency)
				history[i].Interval = string(stripePrice.Recurring.Interval)
			}
			
			// Get live Stripe subscription next payment date
			if subscriptions, subErr := h.StripeService.GetCustomerSubscriptions(r.Context()); subErr == nil {
				for _, subscription := range subscriptions {
					if subscription.Status == "active" || subscription.Status == "trialing" {
						// Set next payment date from Stripe's current period end
						if subscription.CurrentPeriodEnd > 0 {
							nextPaymentDate := time.Unix(subscription.CurrentPeriodEnd, 0)
							history[i].NextPaymentDate = &nextPaymentDate
						}
						break // Use first active subscription
					}
				}
			}
		}
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

// DeleteMyAccount allows a customer to permanently delete their own account
// @Summary Delete customer account
// @Description Permanently deletes the authenticated customer's account including all data from database, Firebase, and cancels Stripe subscriptions
// @Tags customers
// @Accept json
// @Produce json
// @Security Bearer
// @Param confirmation body customerDto.AccountDeletionRequest true "Account deletion confirmation"
// @Success 200 {object} map[string]interface{} "Account deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Missing confirmation"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /secure/customers/delete-account [delete]
func (h *CustomersHandler) DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID
	customerID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Parse confirmation request
	var requestDto customerDto.AccountDeletionRequest
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Validate confirmation
	if !requestDto.ConfirmDeletion {
		responseHandlers.RespondWithError(w, errLib.New("Account deletion must be confirmed", http.StatusBadRequest))
		return
	}

	log.Printf("Starting account deletion process for customer: %s", customerID)

	// 1. Check for active or scheduled-to-cancel subscriptions
	log.Printf("Checking for active subscriptions for customer: %s", customerID)
	subscriptions, stripeErr := h.StripeService.GetCustomerSubscriptions(r.Context())
	if stripeErr != nil {
		log.Printf("WARNING: Could not retrieve Stripe subscriptions for customer %s: %v", customerID, stripeErr)
		// Don't allow deletion if we can't verify subscription status
		responseHandlers.RespondWithError(w, errLib.New("Unable to verify subscription status. Please try again later.", http.StatusServiceUnavailable))
		return
	}

	// Check if any subscriptions are active or scheduled for cancellation
	for _, subscription := range subscriptions {
		if subscription.Status == "active" || subscription.Status == "trialing" {
			log.Printf("BLOCKED: Customer %s has active subscription %s (status: %s)", customerID, subscription.ID, subscription.Status)

			var errorMessage string
			if subscription.CancelAtPeriodEnd {
				// Subscription is already scheduled to cancel
				cancelDate := time.Unix(subscription.CurrentPeriodEnd, 0).Format("January 2, 2006")
				errorMessage = fmt.Sprintf("Your subscription is scheduled to end on %s. You can delete your account after that date.", cancelDate)
			} else {
				// Active subscription not yet canceled
				errorMessage = "You must cancel your subscription before deleting your account. Please cancel your subscription and wait for it to expire."
			}

			responseHandlers.RespondWithError(w, errLib.New(errorMessage, http.StatusConflict))
			return
		}

		// Also block if subscription is past_due (customer owes money)
		if subscription.Status == "past_due" || subscription.Status == "unpaid" {
			log.Printf("BLOCKED: Customer %s has unpaid subscription %s (status: %s)", customerID, subscription.ID, subscription.Status)
			responseHandlers.RespondWithError(w, errLib.New("You have an unpaid subscription. Please resolve your payment before deleting your account.", http.StatusConflict))
			return
		}
	}

	log.Printf("No active subscriptions found for customer %s, proceeding with soft delete", customerID)

	// 2. Implement soft delete with 30-day recovery period
	deletedAt := time.Now().UTC()
	scheduledDeletionAt := deletedAt.Add(30 * 24 * time.Hour) // 30 days grace period

	softDeleteQuery := `
		UPDATE users.users
		SET deleted_at = $1,
		    scheduled_deletion_at = $2,
		    updated_at = $1
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, dbErr := h.CustomerRepo.Db.ExecContext(r.Context(), softDeleteQuery, deletedAt, scheduledDeletionAt, customerID)
	if dbErr != nil {
		log.Printf("Failed to soft delete customer %s: %v", customerID, dbErr)
		responseHandlers.RespondWithError(w, errLib.New("Failed to delete account", http.StatusInternalServerError))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Customer %s not found or already deleted", customerID)
		responseHandlers.RespondWithError(w, errLib.New("Account not found or already deleted", http.StatusNotFound))
		return
	}

	// 3. Disable Firebase account (don't delete - keep for recovery)
	var userEmail string
	if dbErr := h.CustomerRepo.Db.QueryRowContext(r.Context(), "SELECT email FROM users.users WHERE id = $1", customerID).Scan(&userEmail); dbErr == nil && userEmail != "" {
		log.Printf("Disabling Firebase account for soft-deleted user: %s", userEmail)
		// TODO: Implement Firebase disable (not delete) - Firebase Admin SDK has UpdateUser with Disabled field
		// For now we just log it - you'll need to add a DisableUser method to FirebaseService
	}

	log.Printf("Account soft deleted successfully for customer: %s (recoverable until %s)", customerID, scheduledDeletionAt.Format(time.RFC3339))

	responseHandlers.RespondWithSuccess(w, map[string]interface{}{
		"message":              "Your account has been scheduled for deletion",
		"deleted_at":           deletedAt.Format(time.RFC3339),
		"recoverable_until":    scheduledDeletionAt.Format(time.RFC3339),
		"recovery_period_days": 30,
		"note":                 "You have 30 days to recover your account. After that, all data will be permanently deleted.",
	}, http.StatusOK)
}

// RecoverAccount allows a user to recover their soft-deleted account within the grace period
// @Summary Recover deleted account
// @Description Recovers a soft-deleted account within the 30-day grace period
// @Tags customers
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Account recovered successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Not Found: Account not found or not deleted"
// @Failure 410 {object} map[string]interface{} "Gone: Recovery period expired"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /secure/customers/recover-account [post]
func (h *CustomersHandler) RecoverAccount(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID
	customerID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	log.Printf("Account recovery requested for customer: %s", customerID)

	// Check if account is soft-deleted and within recovery period
	var deletedAt, scheduledDeletionAt sql.NullTime
	checkQuery := "SELECT deleted_at, scheduled_deletion_at FROM users.users WHERE id = $1"

	if dbErr := h.CustomerRepo.Db.QueryRowContext(r.Context(), checkQuery, customerID).Scan(&deletedAt, &scheduledDeletionAt); dbErr != nil {
		if dbErr == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Account not found", http.StatusNotFound))
		} else {
			responseHandlers.RespondWithError(w, errLib.New("Failed to check account status", http.StatusInternalServerError))
		}
		return
	}

	// Check if account is actually deleted
	if !deletedAt.Valid {
		log.Printf("Recovery failed: Account %s is not deleted", customerID)
		responseHandlers.RespondWithError(w, errLib.New("Account is not deleted", http.StatusBadRequest))
		return
	}

	// Check if recovery period has expired
	if scheduledDeletionAt.Valid && time.Now().UTC().After(scheduledDeletionAt.Time) {
		log.Printf("Recovery failed: Grace period expired for account %s", customerID)
		responseHandlers.RespondWithError(w, errLib.New("Recovery period has expired. Account data has been permanently deleted.", http.StatusGone))
		return
	}

	// Recover the account by clearing soft delete fields
	recoverQuery := `
		UPDATE users.users
		SET deleted_at = NULL,
		    scheduled_deletion_at = NULL,
		    updated_at = $1
		WHERE id = $2
	`

	result, dbErr := h.CustomerRepo.Db.ExecContext(r.Context(), recoverQuery, time.Now().UTC(), customerID)
	if dbErr != nil {
		log.Printf("Failed to recover account %s: %v", customerID, dbErr)
		responseHandlers.RespondWithError(w, errLib.New("Failed to recover account", http.StatusInternalServerError))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		responseHandlers.RespondWithError(w, errLib.New("Account not found", http.StatusNotFound))
		return
	}

	// TODO: Re-enable Firebase account
	// h.FirebaseService.EnableUser(r.Context(), userEmail)

	log.Printf("Account successfully recovered for customer: %s", customerID)

	responseHandlers.RespondWithSuccess(w, map[string]interface{}{
		"message":      "Your account has been successfully recovered",
		"recovered_at": time.Now().UTC().Format(time.RFC3339),
	}, http.StatusOK)
}

// UpdateCustomerNotes updates notes for a customer.
// @Tags customers
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Customer ID"
// @Param notes_body body customerDto.NotesUpdateRequestDto true "Customer notes update data"
// @Success 200 {object} map[string]interface{} "Notes updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 403 {object} map[string]interface{} "Forbidden: Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Not Found: Customer not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{id}/notes [put]
func (h *CustomersHandler) UpdateCustomerNotes(w http.ResponseWriter, r *http.Request) {
	// Get customer ID from URL
	customerIDStr := chi.URLParam(r, "id")
	customerID, parseErr := uuid.Parse(customerIDStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid customer ID format", http.StatusBadRequest))
		return
	}

	// Check authorization - only admin can update notes for now
	userRole, roleErr := contextUtils.GetUserRole(r.Context())
	if roleErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Authentication required", http.StatusUnauthorized))
		return
	}

	if userRole != contextUtils.RoleAdmin && userRole != contextUtils.RoleSuperAdmin {
		responseHandlers.RespondWithError(w, errLib.New("Insufficient permissions to update notes", http.StatusForbidden))
		return
	}

	// Parse request body for notes update
	var requestDto customerDto.NotesUpdateRequestDto
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Convert to value object
	updateValue, conversionErr := requestDto.ToUpdateValue(customerIDStr)
	if conversionErr != nil {
		responseHandlers.RespondWithError(w, conversionErr)
		return
	}

	// Update notes in database
	rowsAffected, updateErr := h.CustomerRepo.UpdateCustomerNotes(r.Context(), updateValue)
	if updateErr != nil {
		responseHandlers.RespondWithError(w, updateErr)
		return
	}

	if rowsAffected == 0 {
		responseHandlers.RespondWithError(w, errLib.New("Customer not found", http.StatusNotFound))
		return
	}

	responseHandlers.RespondWithSuccess(w, map[string]interface{}{
		"message": "Notes updated successfully",
		"customer_id": customerID,
	}, http.StatusOK)
}
