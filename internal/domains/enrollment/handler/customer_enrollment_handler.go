package enrollment

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"api/internal/di"
	enrollmentService "api/internal/domains/enrollment/service"
	userServices "api/internal/domains/user/services"
	"api/internal/middlewares"
	responseHandlers "api/internal/libs/responses"
	contextUtils "api/utils/context"

	"github.com/google/uuid"
)

// RemoveCustomerRequest represents the optional request body for removing a customer
type RemoveCustomerRequest struct {
	RefundCredits bool   `json:"refund_credits"`
	Reason        string `json:"reason,omitempty"`
}

// RemoveCustomerResponse represents the response for removing a customer
type RemoveCustomerResponse struct {
	Message string        `json:"message"`
	Refund  *RefundInfo   `json:"refund,omitempty"`
}

// RefundInfo holds information about a credit refund
type RefundInfo struct {
	Processed       bool  `json:"processed"`
	CreditsRefunded int32 `json:"credits_refunded"`
}

type CustomerEnrollmentHandler struct {
	Service       *enrollmentService.CustomerEnrollmentService
	CreditService *userServices.CustomerCreditService
	db            *sql.DB
}

func NewCustomerEnrollmentHandler(container *di.Container) *CustomerEnrollmentHandler {
	return &CustomerEnrollmentHandler{
		Service:       enrollmentService.NewCustomerEnrollmentService(container),
		CreditService: userServices.NewCustomerCreditService(container),
		db:            container.DB,
	}
}

// RemoveCustomerFromEvent removes a customer from an event with optional credit refund.
// @Summary Remove a customer from an event
// @Description Removes a customer's enrollment from an event. Optionally refunds credits if the customer paid with credits.
// @Tags event_enrollment
// @Accept json
// @Produce json
// @Param event_id path string true "Event ID" Format(uuid)
// @Param customer_id path string true "Customer ID" Format(uuid)
// @Param body body RemoveCustomerRequest false "Optional refund options"
// @Success 200 {object} RemoveCustomerResponse "Customer successfully removed from event"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Enrollment not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /events/{event_id}/customers/{customer_id} [delete]
func (h *CustomerEnrollmentHandler) RemoveCustomerFromEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseUUIDParam(r, "event_id")
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	customerID, err := parseUUIDParam(r, "customer_id")
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Parse optional request body for refund options
	var request RemoveCustomerRequest
	if r.Body != nil && r.ContentLength > 0 {
		if decodeErr := json.NewDecoder(r.Body).Decode(&request); decodeErr != nil {
			log.Printf("[REMOVE-CUSTOMER] Failed to decode request body: %v", decodeErr)
			// Continue without refund options if body is invalid
		}
	}

	// Prepare response
	response := RemoveCustomerResponse{
		Message: "Customer removed from event",
		Refund: &RefundInfo{
			Processed:       false,
			CreditsRefunded: 0,
		},
	}

	// Process refund if requested
	if request.RefundCredits {
		// Get performer info from context
		performerID, idErr := contextUtils.GetUserID(r.Context())
		if idErr != nil {
			log.Printf("[REMOVE-CUSTOMER] Failed to get performer ID: %v", idErr)
		} else {
			// Get role for audit
			role, _ := contextUtils.GetUserRole(r.Context())
			staffRole := string(role)

			// Get IP address for audit
			ipAddress := middlewares.GetRealIP(r)

			// Get event details for audit snapshot
			eventSnapshot := h.getEventSnapshot(eventID)

			// Process refund
			refundResult, refundErr := h.CreditService.RefundCreditsWithAudit(
				r.Context(),
				eventID,
				customerID,
				performerID,
				staffRole,
				request.Reason,
				ipAddress,
				eventSnapshot,
			)
			if refundErr != nil {
				log.Printf("[REMOVE-CUSTOMER] Credit refund failed: %v", refundErr)
				// Continue with removal even if refund fails
			} else if refundResult != nil {
				response.Refund.Processed = refundResult.Processed
				response.Refund.CreditsRefunded = refundResult.CreditsRefunded
			}
		}
	}

	// Remove customer from event
	if err = h.Service.RemoveCustomerFromEvent(r.Context(), eventID, customerID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// getEventSnapshot retrieves event details for audit logging
func (h *CustomerEnrollmentHandler) getEventSnapshot(eventID uuid.UUID) userServices.EventSnapshot {
	snapshot := userServices.EventSnapshot{}

	query := `
		SELECT
			COALESCE(p.name, t.name, 'Event') AS event_name,
			e.start_at,
			COALESCE(p.name, '') AS program_name,
			COALESCE(l.name, '') AS location_name
		FROM events.events e
		LEFT JOIN program.programs p ON e.program_id = p.id
		LEFT JOIN athletic.teams t ON e.team_id = t.id
		LEFT JOIN location.locations l ON e.location_id = l.id
		WHERE e.id = $1
	`

	var eventName, programName, locationName string
	err := h.db.QueryRow(query, eventID).Scan(&eventName, &snapshot.StartAt, &programName, &locationName)
	if err != nil {
		log.Printf("[REMOVE-CUSTOMER] Failed to get event details for audit: %v", err)
		return snapshot
	}

	snapshot.Name = eventName
	snapshot.ProgramName = programName
	snapshot.LocationName = locationName

	return snapshot
}
