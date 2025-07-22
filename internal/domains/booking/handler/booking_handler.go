package booking

import (
	"api/internal/di"
	hairDto "api/internal/domains/haircut/event/dto"
	hairRepo "api/internal/domains/haircut/event/persistence"
	playgroundDto "api/internal/domains/playground/dto/session"
	playgroundService "api/internal/domains/playground/services"
	responseHandlers "api/internal/libs/responses"
	contextUtils "api/utils/context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Handler aggregates bookings from multiple domains.
type Handler struct {
	HaircutRepo       *hairRepo.Repository
	PlaygroundService *playgroundService.Service
}

// NewHandler creates a new Handler instance.
func NewHandler(container *di.Container) *Handler {
	return &Handler{
		HaircutRepo:       hairRepo.NewEventsRepository(container),
		PlaygroundService: playgroundService.NewService(container),
	}
}

// UpcomingBookingsResponse represents combined upcoming bookings.
type UpcomingBookingsResponse struct {
	Haircuts   []hairDto.EventResponseDto  `json:"haircuts"`
	Playground []playgroundDto.ResponseDto `json:"playground"`
}

// GetMyUpcomingBookings returns upcoming haircut and playground bookings for the logged-in customer.
// @Tags bookings
// @Produce json
// @Security Bearer
// @Success 200 {object} booking.UpcomingBookingsResponse "Upcoming bookings"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /bookings/upcoming [get]
func (h *Handler) GetMyUpcomingBookings(w http.ResponseWriter, r *http.Request) {
	customerID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	events, err := h.HaircutRepo.GetEvents(r.Context(), uuid.Nil, customerID, time.Time{}, time.Time{})
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	now := time.Now()
	var haircutBookings []hairDto.EventResponseDto
	for _, e := range events {
		if e.BeginDateTime.After(now) {
			haircutBookings = append(haircutBookings, hairDto.NewEventResponse(e))
		}
	}

	sessions, err := h.PlaygroundService.GetSessions(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	var playgroundBookings []playgroundDto.ResponseDto
	for _, s := range sessions {
		if s.CustomerID == customerID && s.StartTime.After(now) {
			playgroundBookings = append(playgroundBookings, playgroundDto.NewResponse(s))
		}
	}

	resp := UpcomingBookingsResponse{
		Haircuts:   haircutBookings,
		Playground: playgroundBookings,
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}
