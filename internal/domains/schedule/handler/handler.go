package schedule

import (
	"net/http"

	"api/internal/di"
	eventDto "api/internal/domains/event/dto"
	eventService "api/internal/domains/event/service"
	eventValues "api/internal/domains/event/values"
	familyService "api/internal/domains/family/service"
	gameDto "api/internal/domains/game/dto"
	gameService "api/internal/domains/game/services"
	gameValues "api/internal/domains/game/values"
	practiceDto "api/internal/domains/practice/dto"
	practiceService "api/internal/domains/practice/services"
	practiceValues "api/internal/domains/practice/values"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"
	"github.com/google/uuid"
)

type Handler struct {
	eventSvc    *eventService.Service
	gameSvc     *gameService.Service
	practiceSvc *practiceService.Service
	familySvc   *familyService.Service
}

func NewHandler(c *di.Container) *Handler {
	return &Handler{
		eventSvc:    eventService.NewEventService(c),
		gameSvc:     gameService.NewService(c),
		practiceSvc: practiceService.NewService(c),
		familySvc:   familyService.NewService(c),
	}
}

type Response struct {
	Events    []eventDto.EventResponseDto `json:"events"`
	Games     []gameDto.ResponseDto       `json:"games"`
	Practices []practiceDto.ResponseDto   `json:"practices"`
}

// GetMySchedule retrieves user's personalized schedule including events, games, and practices.
// Parents can view their child's schedule by passing the child_id query parameter.
// @Summary Get my schedule
// @Description Retrieves a consolidated schedule of events, games, and practices based on user role and associations.
// @Tags schedule
// @Accept json
// @Produce json
// @Security Bearer
// @Param child_id query string false "Child user ID (for parent viewing child's schedule)" format(uuid)
// @Success 200 {object} Response "Schedule retrieved successfully with events, games, and practices"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Not authorized to view child's schedule"
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

	// Check if parent is requesting child's schedule
	targetUserID := userID
	targetRole := role
	viewingChild := false
	if childIDStr := r.URL.Query().Get("child_id"); childIDStr != "" {
		childID, parseErr := validators.ParseUUID(childIDStr)
		if parseErr != nil {
			responseHandlers.RespondWithError(w, parseErr)
			return
		}

		// Verify parent has access to this child
		if verifyErr := h.familySvc.VerifyParentChildAccess(ctx, userID, childID); verifyErr != nil {
			responseHandlers.RespondWithError(w, verifyErr)
			return
		}

		targetUserID = childID
		targetRole = contextUtils.RoleAthlete // Children with games/practices are athletes
		viewingChild = true
	}

	// Events
	var eventRecords []eventValues.ReadEventValues
	if !viewingChild && (role == contextUtils.RoleAdmin || role == contextUtils.RoleSuperAdmin || role == contextUtils.RoleIT || role == contextUtils.RoleReceptionist) {
		eventRecords, err = h.eventSvc.GetEvents(ctx, eventValues.GetEventsFilter{})
	} else {
		// For coaches/athletes/children, get events they're enrolled in or assigned to
		eventRecords, err = h.eventSvc.GetEvents(ctx, eventValues.GetEventsFilter{
			ParticipantID: targetUserID,
		})
	}
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	eventDtos := make([]eventDto.EventResponseDto, len(eventRecords))
	for i, e := range eventRecords {
		eventDtos[i] = eventDto.NewEventResponseDto(e, false, false) // don't include participant details or contact info
	}

	// Games
	var gameRecords []gameValues.ReadGameValue
	if !viewingChild && (role == contextUtils.RoleAdmin || role == contextUtils.RoleSuperAdmin || role == contextUtils.RoleIT || role == contextUtils.RoleReceptionist) {
		gameRecords, err = h.gameSvc.GetGames(ctx, gameValues.GetGamesFilter{
			Limit:  1000,
			Offset: 0,
		})
	} else {
		gameRecords, err = h.gameSvc.GetUserGames(ctx, targetUserID, targetRole, 1000, 0)
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
	if !viewingChild && (role == contextUtils.RoleAdmin || role == contextUtils.RoleSuperAdmin || role == contextUtils.RoleIT || role == contextUtils.RoleReceptionist) {
		practiceRecords, err = h.practiceSvc.GetPractices(ctx, uuid.Nil, 1000, 0)
	} else {
		practiceRecords, err = h.practiceSvc.GetUserPractices(ctx, targetUserID, targetRole, 1000, 0)
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
