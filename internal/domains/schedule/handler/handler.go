package schedule

import (
	"net/http"

	"api/internal/di"
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
	gameSvc     *gameService.Service
	practiceSvc *practiceService.Service
}

func NewHandler(c *di.Container) *Handler {
	return &Handler{
		gameSvc:     gameService.NewService(c),
		practiceSvc: practiceService.NewService(c),
	}
}

type Response struct {
	Games     []gameDto.ResponseDto     `json:"games"`
	Practices []practiceDto.ResponseDto `json:"practices"`
}

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

	// Results are already sorted by start time within games and practices
	resp := Response{
		Games:     gameDtos,
		Practices: practiceDtos,
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}
