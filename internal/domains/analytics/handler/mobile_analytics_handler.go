package handler

import (
	"net/http"
	"strconv"
	"time"

	"api/internal/di"
	repo "api/internal/domains/analytics/persistence/repository"
	responseHandlers "api/internal/libs/responses"
	contextUtils "api/utils/context"
)

type MobileAnalyticsHandler struct {
	repo *repo.MobileAnalyticsRepository
}

func NewMobileAnalyticsHandler(container *di.Container) *MobileAnalyticsHandler {
	return &MobileAnalyticsHandler{
		repo: repo.NewMobileAnalyticsRepository(container),
	}
}

// RecordMobileSession records that a user has logged into the mobile app
// @Summary Record mobile app session
// @Description Called by the mobile app after successful login to track usage
// @Tags mobile
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Session recorded successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /mobile/session [post]
func (h *MobileAnalyticsHandler) RecordMobileSession(w http.ResponseWriter, r *http.Request) {
	userID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.repo.RecordMobileLogin(r.Context(), userID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"message":   "Session recorded",
		"timestamp": time.Now().UTC(),
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetMobileUsageStats returns mobile app usage statistics for admin dashboard
// @Summary Get mobile usage statistics
// @Description Returns DAU, WAU, MAU and other mobile app usage metrics
// @Tags admin,analytics
// @Produce json
// @Security Bearer
// @Success 200 {object} repo.MobileUsageStats "Mobile usage statistics"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin only"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/analytics/mobile [get]
func (h *MobileAnalyticsHandler) GetMobileUsageStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetMobileUsageStats(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, stats, http.StatusOK)
}

// GetRecentMobileLogins returns a list of users who recently logged into the mobile app
// @Summary Get recent mobile logins
// @Description Returns a paginated list of users who recently used the mobile app
// @Tags admin,analytics
// @Produce json
// @Param limit query int false "Number of results (default 50)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of recent mobile logins"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin only"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/analytics/mobile/logins [get]
func (h *MobileAnalyticsHandler) GetRecentMobileLogins(w http.ResponseWriter, r *http.Request) {
	limit := int32(50)
	offset := int32(0)

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	logins, err := h.repo.GetRecentMobileLogins(r.Context(), limit, offset)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"logins": logins,
		"limit":  limit,
		"offset": offset,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetMobileLoginTrends returns daily mobile login counts for the specified period
// @Summary Get mobile login trends
// @Description Returns daily unique mobile logins for trend analysis
// @Tags admin,analytics
// @Produce json
// @Param days query int false "Number of days to look back (default 30)"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Daily mobile login counts"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin only"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/analytics/mobile/trends [get]
func (h *MobileAnalyticsHandler) GetMobileLoginTrends(w http.ResponseWriter, r *http.Request) {
	days := 30

	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	endDate := time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -days)

	trends, err := h.repo.GetMobileLoginsByDateRange(r.Context(), startDate, endDate)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"trends":     trends,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"days":       days,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}
