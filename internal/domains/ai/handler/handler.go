package handler

import (
	"api/internal/domains/ai/dto"
	"api/internal/domains/ai/service"
	errLib "api/internal/libs/errors"
	responses "api/internal/libs/responses"
	"api/internal/libs/validators"
	"html"
	"log"
	"net/http"
	"strings"
)

type Handler struct {
	svc *service.Service
}

func NewHandler() *Handler {
	return &Handler{svc: service.NewService()}
}

// ProxyMessage proxies a chat message to the AI model.
// @Summary Proxy chat message to AI model
// @Tags ai
// @Accept json
// @Produce json
// @Param payload body dto.ChatRequest true "Chat message"
// @Success 200 {object} dto.ChatResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /ai/chat [post]
func (h *Handler) ProxyMessage(w http.ResponseWriter, r *http.Request) {
	var req dto.ChatRequest
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responses.RespondWithError(w, err)
		return
	}

	req.Query = strings.TrimSpace(req.Query)
	req.Context = strings.TrimSpace(req.Context)
	req.Query = html.EscapeString(req.Query)
	if req.Query == "" {
		responses.RespondWithError(w, errLib.New("query is required", http.StatusBadRequest))
		return
	}

	log.Printf("AI chat request: %s", req.Query)

	reply, err := h.svc.Chat(req.Query, req.Context, req.ChatHistory)
	if err != nil {
		responses.RespondWithError(w, err)
		return
	}

	responses.RespondWithSuccess(w, dto.ChatResponse{Reply: reply}, http.StatusOK)
}
