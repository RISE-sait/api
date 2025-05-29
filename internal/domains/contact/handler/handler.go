package handler

import (
	"encoding/json"
	"html"
	"log"
	"net/http"
	"regexp"
	"strings"

	"api/internal/di"
	"api/internal/domains/contact/dto"
	"api/internal/domains/contact/service"
	"api/utils/recaptcha"

)

type Handler struct{}

func NewContactHandler(_ *di.Container) *Handler {
	return &Handler{}
}

func (h *Handler) SendContactEmail(w http.ResponseWriter, r *http.Request) {
	// 1️⃣ Cap request body to 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// 2️⃣ Decode JSON
	var req dto.ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3️⃣ Trim whitespace
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Message = strings.TrimSpace(req.Message)
	// (req.Token is a one-time value; no trim)

	// 4️⃣ Field validation & blacklists
	if req.Name == "" || len(req.Name) > 100 || !isSafeHeaderValue(req.Name) {
		http.Error(w, "Invalid name (required, max 100 chars, no special symbols)", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) || !isSafeHeaderValue(req.Email) {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}
	// … after trimming…
	if req.Phone != "" {
		// 1️⃣ Regex‐based sanity check
		if !isValidPhone(req.Phone) || !isSafeHeaderValue(req.Phone) {
			http.Error(w, "Invalid phone number format", http.StatusBadRequest)
			return
		}

		// 2️⃣ Count digits and enforce a minimum (e.g. 7 digits)
		digits := regexp.MustCompile(`\d`).FindAllString(req.Phone, -1)
		if len(digits) < 10 {
			http.Error(w, "Phone number must contain at least 10 digits", http.StatusBadRequest)
			return
		}
	}

	if req.Message == "" || len(req.Message) > 5000 {
		http.Error(w, "Message is required (max 5000 chars)", http.StatusBadRequest)
		return
	}

	// 5️⃣ Escape the message body
	req.Message = html.EscapeString(req.Message)

	// 6️⃣ Verify reCAPTCHA
	ok, err := recaptcha.Verify(req.Token)
	if err != nil {
		log.Printf("reCAPTCHA error: %v", err)
		http.Error(w, "Unable to verify reCAPTCHA", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "reCAPTCHA validation failed", http.StatusBadRequest)
		return
	}

	// 7️⃣ Send the email
	if err := service.SendContactRequest(req); err != nil {
		log.Printf("SendContactRequest error: %v", err)
		http.Error(w, "Failed to send contact request", http.StatusInternalServerError)
		return
	}

	// 8️⃣ Success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Contact request sent successfully"}`))
}

func isSafeHeaderValue(value string) bool {
	// reject CR/LF to prevent header injection
	if strings.ContainsAny(value, "\r\n") {
		return false
	}
	// blacklist other dangerous punctuation
	for _, ch := range []string{"<", ">", "{", "}", "[", "]", "(", ")", ";", "'", "\"", "`"} {
		if strings.Contains(value, ch) {
			return false
		}
	}
	return true
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	return re.MatchString(email)
}

func isValidPhone(phone string) bool {
	re := regexp.MustCompile(`^\+?[0-9\s\-\(\)]+$`)
	return re.MatchString(phone)
}


