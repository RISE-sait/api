package payment

import (
	"api/internal/di"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"io"
	"log"
	"net/http"
)

type WebhookHandlers struct{}

func NewWebhookHandlers(container *di.Container) *WebhookHandlers {
	return &WebhookHandlers{}
}

// HandleSquareWebhook accepts incoming Square webhook events.
// @Tags payments
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Webhook processed successfully"
// @Router /webhooks/square [post]
func (h *WebhookHandlers) HandleSquareWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Error reading request body", http.StatusBadRequest))
		return
	}

	log.Println(">>> Incoming Square webhook")
	log.Println("Payload:", string(payload))

	// If verification or signature checking is needed, it can be added here later.

	w.WriteHeader(http.StatusOK)
}
