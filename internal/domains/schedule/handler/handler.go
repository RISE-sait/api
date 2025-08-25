package schedule

import (
	"net/http"

	"api/internal/di"
	eventDto "api/internal/domains/event/dto"
	eventService "api/internal/domains/event/service"
	eventValues "api/internal/domains/event/values"
	gameDto "api/internal/domains/game/dto"
	gameService "api/internal/domains/game/services"
	gameValues "api/internal/domains/game/values"
	practiceDto "api/internal/domains/practice/dto"
	practiceService "api/internal/domains/practice/services"
	practiceValues "api/internal/domains/practice/values"
	responseHandlers "api/internal/libs/responses"
	contextUtils "api/utils/context"
	"github.com/google/uuid"
)

type Handler struct {
	eventSvc    *eventService.Service
	gameSvc     *gameService.Service
	practiceSvc *practiceService.Service
}

func NewHandler(c *di.Container) *Handler {
	return &Handler{
		eventSvc:    eventService.NewEventService(c),
		gameSvc:     gameService.NewService(c),
		practiceSvc: practiceService.NewService(c),
	}
}

type Response struct {
	Events    []eventDto.EventResponseDto `json:"events"`
	Games     []gameDto.ResponseDto       `json:"games"`
	Practices []practiceDto.ResponseDto   `json:"practices"`
}

// GetMySchedule retrieves user's personalized schedule including events, games, and practices.
// @Summary Get my schedule
// @Description Retrieves a consolidated schedule of events, games, and practices based on user role and associations.
// @Tags schedule
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} Response "Schedule retrieved successfully with events, games, and practices"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /secure/schedule [get]
func (h *Handler) GetMySchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	role, err := contextUtils.GetUserRole(ctx)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Events
	var eventRecords []eventValues.ReadEventValues
	if role == contextUtils.RoleAdmin || role == contextUtils.RoleSuperAdmin {
		eventRecords, err = h.eventSvc.GetEvents(ctx, eventValues.GetEventsFilter{})
	} else {
		// For coaches/athletes, get events they're enrolled in or assigned to
		eventRecords, err = h.eventSvc.GetEvents(ctx, eventValues.GetEventsFilter{
			ParticipantID: userID,
		})
	}
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	eventDtos := make([]eventDto.EventResponseDto, len(eventRecords))
	for i, e := range eventRecords {
		eventDtos[i] = eventDto.NewEventResponseDto(e, false) // false = don't include participant details
	}

	// Games
	var gameRecords []gameValues.ReadGameValue
	if role == contextUtils.RoleAdmin || role == contextUtils.RoleSuperAdmin {
		gameRecords, err = h.gameSvc.GetGames(ctx, gameValues.GetGamesFilter{})
	} else {
		gameRecords, err = h.gameSvc.GetUserGames(ctx, userID, role, 1000, 0)
	}
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	gameDtos := make([]gameDto.ResponseDto, len(gameRecords))
	for i, g := range gameRecords {
		gameDtos[i] = gameDto.NewGameResponse(g)
	}

	// Practices
	var practiceRecords []practiceValues.ReadPracticeValue
	if role == contextUtils.RoleAdmin || role == contextUtils.RoleSuperAdmin {
		practiceRecords, err = h.practiceSvc.GetPractices(ctx, uuid.Nil, 1000, 0)
	} else {
		practiceRecords, err = h.practiceSvc.GetUserPractices(ctx, userID, role, 1000, 0)
	}
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	practiceDtos := make([]practiceDto.ResponseDto, len(practiceRecords))
	for i, p := range practiceRecords {
		practiceDtos[i] = practiceDto.NewResponse(p)
	}

	// Results are already sorted by start time within events, games and practices
	resp := Response{
		Events:    eventDtos,
		Games:     gameDtos,
		Practices: practiceDtos,
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}
