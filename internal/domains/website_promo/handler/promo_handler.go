package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"api/internal/di"
	"api/internal/domains/website_promo/dto"
	db "api/internal/domains/website_promo/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	responseHandlers "api/internal/libs/responses"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type WebsitePromoHandler struct {
	queries *db.Queries
}

func NewWebsitePromoHandler(container *di.Container) *WebsitePromoHandler {
	return &WebsitePromoHandler{
		queries: db.New(container.DB),
	}
}

// ============ Hero Promos ============

// GetAllHeroPromos returns all hero promos (admin)
func (h *WebsitePromoHandler) GetAllHeroPromos(w http.ResponseWriter, r *http.Request) {
	promos, err := h.queries.GetAllHeroPromos(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to get hero promos", http.StatusInternalServerError))
		return
	}

	response := make([]dto.HeroPromoResponse, len(promos))
	for i, p := range promos {
		response[i] = mapHeroPromoToResponse(p)
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetActiveHeroPromos returns only active hero promos (public)
func (h *WebsitePromoHandler) GetActiveHeroPromos(w http.ResponseWriter, r *http.Request) {
	promos, err := h.queries.GetActiveHeroPromos(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to get hero promos", http.StatusInternalServerError))
		return
	}

	response := make([]dto.HeroPromoResponse, len(promos))
	for i, p := range promos {
		response[i] = mapHeroPromoToResponse(p)
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetHeroPromoById returns a single hero promo
func (h *WebsitePromoHandler) GetHeroPromoById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid promo ID", http.StatusBadRequest))
		return
	}

	promo, err := h.queries.GetHeroPromoById(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Hero promo not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to get hero promo", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapHeroPromoToResponse(promo), http.StatusOK)
}

// CreateHeroPromo creates a new hero promo
func (h *WebsitePromoHandler) CreateHeroPromo(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateHeroPromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	userID, _ := contextUtils.GetUserID(r.Context())

	durationSeconds := req.DurationSeconds
	if durationSeconds == 0 {
		durationSeconds = 5
	}

	mediaType := req.MediaType
	if mediaType == "" {
		mediaType = "image"
	}

	promo, err := h.queries.CreateHeroPromo(r.Context(), db.CreateHeroPromoParams{
		Title:           req.Title,
		Subtitle:        sql.NullString{String: req.Subtitle, Valid: req.Subtitle != ""},
		Description:     sql.NullString{String: req.Description, Valid: req.Description != ""},
		MediaUrl:        req.MediaURL,
		MediaType:       mediaType,
		ThumbnailUrl:    sql.NullString{String: req.ThumbnailURL, Valid: req.ThumbnailURL != ""},
		ButtonText:      sql.NullString{String: req.ButtonText, Valid: req.ButtonText != ""},
		ButtonLink:      sql.NullString{String: req.ButtonLink, Valid: req.ButtonLink != ""},
		DisplayOrder:    int32(req.DisplayOrder),
		DurationSeconds: int32(durationSeconds),
		IsActive:        req.IsActive,
		StartDate:       toNullTime(req.StartDate),
		EndDate:         toNullTime(req.EndDate),
		CreatedBy:       uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil},
	})
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to create hero promo", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapHeroPromoToResponse(promo), http.StatusCreated)
}

// UpdateHeroPromo updates an existing hero promo
func (h *WebsitePromoHandler) UpdateHeroPromo(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid promo ID", http.StatusBadRequest))
		return
	}

	var req dto.UpdateHeroPromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	userID, _ := contextUtils.GetUserID(r.Context())

	durationSeconds := req.DurationSeconds
	if durationSeconds == 0 {
		durationSeconds = 5
	}

	mediaType := req.MediaType
	if mediaType == "" {
		mediaType = "image"
	}

	promo, err := h.queries.UpdateHeroPromo(r.Context(), db.UpdateHeroPromoParams{
		ID:              id,
		Title:           req.Title,
		Subtitle:        sql.NullString{String: req.Subtitle, Valid: req.Subtitle != ""},
		Description:     sql.NullString{String: req.Description, Valid: req.Description != ""},
		MediaUrl:        req.MediaURL,
		MediaType:       mediaType,
		ThumbnailUrl:    sql.NullString{String: req.ThumbnailURL, Valid: req.ThumbnailURL != ""},
		ButtonText:      sql.NullString{String: req.ButtonText, Valid: req.ButtonText != ""},
		ButtonLink:      sql.NullString{String: req.ButtonLink, Valid: req.ButtonLink != ""},
		DisplayOrder:    int32(req.DisplayOrder),
		DurationSeconds: int32(durationSeconds),
		IsActive:        req.IsActive,
		StartDate:       toNullTime(req.StartDate),
		EndDate:         toNullTime(req.EndDate),
		UpdatedBy:       uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Hero promo not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to update hero promo", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapHeroPromoToResponse(promo), http.StatusOK)
}

// DeleteHeroPromo deletes a hero promo
func (h *WebsitePromoHandler) DeleteHeroPromo(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid promo ID", http.StatusBadRequest))
		return
	}

	rowsAffected, err := h.queries.DeleteHeroPromo(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to delete hero promo", http.StatusInternalServerError))
		return
	}

	if rowsAffected == 0 {
		responseHandlers.RespondWithError(w, errLib.New("Hero promo not found", http.StatusNotFound))
		return
	}

	responseHandlers.RespondWithSuccess(w, map[string]string{"message": "Hero promo deleted successfully"}, http.StatusOK)
}

// ============ Feature Cards ============

// GetAllFeatureCards returns all feature cards (admin)
func (h *WebsitePromoHandler) GetAllFeatureCards(w http.ResponseWriter, r *http.Request) {
	cards, err := h.queries.GetAllFeatureCards(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to get feature cards", http.StatusInternalServerError))
		return
	}

	response := make([]dto.FeatureCardResponse, len(cards))
	for i, c := range cards {
		response[i] = mapFeatureCardToResponse(c)
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetActiveFeatureCards returns only active feature cards (public)
func (h *WebsitePromoHandler) GetActiveFeatureCards(w http.ResponseWriter, r *http.Request) {
	cards, err := h.queries.GetActiveFeatureCards(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to get feature cards", http.StatusInternalServerError))
		return
	}

	response := make([]dto.FeatureCardResponse, len(cards))
	for i, c := range cards {
		response[i] = mapFeatureCardToResponse(c)
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetFeatureCardById returns a single feature card
func (h *WebsitePromoHandler) GetFeatureCardById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid card ID", http.StatusBadRequest))
		return
	}

	card, err := h.queries.GetFeatureCardById(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Feature card not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to get feature card", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapFeatureCardToResponse(card), http.StatusOK)
}

// CreateFeatureCard creates a new feature card
func (h *WebsitePromoHandler) CreateFeatureCard(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateFeatureCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	userID, _ := contextUtils.GetUserID(r.Context())

	card, err := h.queries.CreateFeatureCard(r.Context(), db.CreateFeatureCardParams{
		Title:        req.Title,
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
		ImageUrl:     req.ImageURL,
		ButtonText:   sql.NullString{String: req.ButtonText, Valid: req.ButtonText != ""},
		ButtonLink:   sql.NullString{String: req.ButtonLink, Valid: req.ButtonLink != ""},
		DisplayOrder: int32(req.DisplayOrder),
		IsActive:     req.IsActive,
		StartDate:    toNullTime(req.StartDate),
		EndDate:      toNullTime(req.EndDate),
		CreatedBy:    uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil},
	})
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to create feature card", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapFeatureCardToResponse(card), http.StatusCreated)
}

// UpdateFeatureCard updates an existing feature card
func (h *WebsitePromoHandler) UpdateFeatureCard(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid card ID", http.StatusBadRequest))
		return
	}

	var req dto.UpdateFeatureCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	userID, _ := contextUtils.GetUserID(r.Context())

	card, err := h.queries.UpdateFeatureCard(r.Context(), db.UpdateFeatureCardParams{
		ID:           id,
		Title:        req.Title,
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
		ImageUrl:     req.ImageURL,
		ButtonText:   sql.NullString{String: req.ButtonText, Valid: req.ButtonText != ""},
		ButtonLink:   sql.NullString{String: req.ButtonLink, Valid: req.ButtonLink != ""},
		DisplayOrder: int32(req.DisplayOrder),
		IsActive:     req.IsActive,
		StartDate:    toNullTime(req.StartDate),
		EndDate:      toNullTime(req.EndDate),
		UpdatedBy:    uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Feature card not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to update feature card", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapFeatureCardToResponse(card), http.StatusOK)
}

// DeleteFeatureCard deletes a feature card
func (h *WebsitePromoHandler) DeleteFeatureCard(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid card ID", http.StatusBadRequest))
		return
	}

	rowsAffected, err := h.queries.DeleteFeatureCard(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to delete feature card", http.StatusInternalServerError))
		return
	}

	if rowsAffected == 0 {
		responseHandlers.RespondWithError(w, errLib.New("Feature card not found", http.StatusNotFound))
		return
	}

	responseHandlers.RespondWithSuccess(w, map[string]string{"message": "Feature card deleted successfully"}, http.StatusOK)
}

// ============ Promo Videos ============

// GetAllPromoVideos returns all promo videos (admin)
func (h *WebsitePromoHandler) GetAllPromoVideos(w http.ResponseWriter, r *http.Request) {
	videos, err := h.queries.GetAllPromoVideos(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to get promo videos", http.StatusInternalServerError))
		return
	}

	response := make([]dto.PromoVideoResponse, len(videos))
	for i, v := range videos {
		response[i] = mapPromoVideoToResponse(v)
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetActivePromoVideos returns only active promo videos (public)
func (h *WebsitePromoHandler) GetActivePromoVideos(w http.ResponseWriter, r *http.Request) {
	// Check if category filter is provided
	category := r.URL.Query().Get("category")

	var videos []db.WebsitePromoVideo
	var err error

	if category != "" {
		videos, err = h.queries.GetActivePromoVideosByCategory(r.Context(), sql.NullString{String: category, Valid: true})
	} else {
		videos, err = h.queries.GetActivePromoVideos(r.Context())
	}

	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to get promo videos", http.StatusInternalServerError))
		return
	}

	response := make([]dto.PromoVideoResponse, len(videos))
	for i, v := range videos {
		response[i] = mapPromoVideoToResponse(v)
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetPromoVideoById returns a single promo video
func (h *WebsitePromoHandler) GetPromoVideoById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid video ID", http.StatusBadRequest))
		return
	}

	video, err := h.queries.GetPromoVideoById(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Promo video not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to get promo video", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapPromoVideoToResponse(video), http.StatusOK)
}

// CreatePromoVideo creates a new promo video
func (h *WebsitePromoHandler) CreatePromoVideo(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePromoVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	userID, _ := contextUtils.GetUserID(r.Context())

	category := req.Category
	if category == "" {
		category = "highlight"
	}

	video, err := h.queries.CreatePromoVideo(r.Context(), db.CreatePromoVideoParams{
		Title:        req.Title,
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
		VideoUrl:     req.VideoURL,
		ThumbnailUrl: req.ThumbnailURL,
		Category:     sql.NullString{String: category, Valid: true},
		DisplayOrder: int32(req.DisplayOrder),
		IsActive:     req.IsActive,
		StartDate:    toNullTime(req.StartDate),
		EndDate:      toNullTime(req.EndDate),
		CreatedBy:    uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil},
	})
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to create promo video", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapPromoVideoToResponse(video), http.StatusCreated)
}

// UpdatePromoVideo updates an existing promo video
func (h *WebsitePromoHandler) UpdatePromoVideo(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid video ID", http.StatusBadRequest))
		return
	}

	var req dto.UpdatePromoVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	userID, _ := contextUtils.GetUserID(r.Context())

	category := req.Category
	if category == "" {
		category = "highlight"
	}

	video, err := h.queries.UpdatePromoVideo(r.Context(), db.UpdatePromoVideoParams{
		ID:           id,
		Title:        req.Title,
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
		VideoUrl:     req.VideoURL,
		ThumbnailUrl: req.ThumbnailURL,
		Category:     sql.NullString{String: category, Valid: true},
		DisplayOrder: int32(req.DisplayOrder),
		IsActive:     req.IsActive,
		StartDate:    toNullTime(req.StartDate),
		EndDate:      toNullTime(req.EndDate),
		UpdatedBy:    uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Promo video not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to update promo video", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapPromoVideoToResponse(video), http.StatusOK)
}

// DeletePromoVideo deletes a promo video
func (h *WebsitePromoHandler) DeletePromoVideo(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid video ID", http.StatusBadRequest))
		return
	}

	rowsAffected, err := h.queries.DeletePromoVideo(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to delete promo video", http.StatusInternalServerError))
		return
	}

	if rowsAffected == 0 {
		responseHandlers.RespondWithError(w, errLib.New("Promo video not found", http.StatusNotFound))
		return
	}

	responseHandlers.RespondWithSuccess(w, map[string]string{"message": "Promo video deleted successfully"}, http.StatusOK)
}

// ============ Helper Functions ============

func mapHeroPromoToResponse(p db.WebsiteHeroPromo) dto.HeroPromoResponse {
	return dto.HeroPromoResponse{
		ID:              p.ID,
		Title:           p.Title,
		Subtitle:        nullStringToPtr(p.Subtitle),
		Description:     nullStringToPtr(p.Description),
		MediaURL:        p.MediaUrl,
		MediaType:       p.MediaType,
		ThumbnailURL:    nullStringToPtr(p.ThumbnailUrl),
		ButtonText:      nullStringToPtr(p.ButtonText),
		ButtonLink:      nullStringToPtr(p.ButtonLink),
		DisplayOrder:    int(p.DisplayOrder),
		DurationSeconds: int(p.DurationSeconds),
		IsActive:        p.IsActive,
		StartDate:       nullTimeToPtr(p.StartDate),
		EndDate:         nullTimeToPtr(p.EndDate),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func mapFeatureCardToResponse(c db.WebsiteFeatureCard) dto.FeatureCardResponse {
	return dto.FeatureCardResponse{
		ID:           c.ID,
		Title:        c.Title,
		Description:  nullStringToPtr(c.Description),
		ImageURL:     c.ImageUrl,
		ButtonText:   nullStringToPtr(c.ButtonText),
		ButtonLink:   nullStringToPtr(c.ButtonLink),
		DisplayOrder: int(c.DisplayOrder),
		IsActive:     c.IsActive,
		StartDate:    nullTimeToPtr(c.StartDate),
		EndDate:      nullTimeToPtr(c.EndDate),
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

func mapPromoVideoToResponse(v db.WebsitePromoVideo) dto.PromoVideoResponse {
	return dto.PromoVideoResponse{
		ID:           v.ID,
		Title:        v.Title,
		Description:  nullStringToPtr(v.Description),
		VideoURL:     v.VideoUrl,
		ThumbnailURL: v.ThumbnailUrl,
		Category:     nullStringToPtr(v.Category),
		DisplayOrder: int(v.DisplayOrder),
		IsActive:     v.IsActive,
		StartDate:    nullTimeToPtr(v.StartDate),
		EndDate:      nullTimeToPtr(v.EndDate),
		CreatedAt:    v.CreatedAt,
		UpdatedAt:    v.UpdatedAt,
	}
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func toNullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{Valid: false}
}
