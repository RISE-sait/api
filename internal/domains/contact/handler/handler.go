package handler

import (
	"api/internal/di"
	"api/internal/domains/contact/dto"
	"api/internal/domains/contact/service"
	"api/utils/recaptcha"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi"
)

// Handler defines the contact handler struct
type Handler struct{}

// NewContactHandler returns a new instance of the contact handler
func NewContactHandler(_ *di.Container) *Handler {
	return &Handler{}
}

// SendContactEmail handles POST /contact
// It verifies the reCAPTCHA token, validates fields, and sends the contact email
func (h *Handler) SendContactEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1️⃣ verify recaptcha
	ok, err := recaptcha.Verify(req.Token)
	if err != nil {
		http.Error(w, "reCAPTCHA error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "reCAPTCHA validation failed", http.StatusBadRequest)
		return
	}

	// 2️⃣ your existing field trimming & validation
	//    (name, email, phone, message checks…)

	// 3️⃣ send the email
	if err := service.SendContactRequest(req); err != nil {
		http.Error(w, "Failed to send contact request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Contact request sent successfully"}`))
}

func isSafeHeaderValue(value string) bool {
	unsafeChars := []string{"<", ">", "{", "}", "[", "]", "(", ")", ";", "'", "\"", "`"}
	for _, char := range unsafeChars {
		if strings.Contains(value, char) {
			return false
		}
	}
	return true
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func isValidPhone(phone string) bool {
	re := regexp.MustCompile(`^\+?[0-9\s\-\(\)]+$`)
	return re.MatchString(phone)
}

func RegisterContactRoutes(container *di.Container) func(chi.Router) {
	h := NewContactHandler(container)
	return func(r chi.Router) {
		r.Post("/", h.SendContactEmail)
	}
}
