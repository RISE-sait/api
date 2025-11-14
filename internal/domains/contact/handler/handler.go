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
	hupspotService "api/internal/services/hubspot"
	"api/utils/recaptcha"
)

type Handler struct{}

func NewContactHandler(_ *di.Container) *Handler {
	return &Handler{}
}

// SendContactEmail handles POST /contact
// @Summary      Send a contact request
// @Description  Verifies reCAPTCHA, sanitizes input, and emails the contact form
// @Tags         contact
// @Accept       json
// @Produce      json
// @Param        payload  body    dto.ContactRequest  true  "Contact form data"
// @Success      200      {object} map[string]string  "success message"
// @Failure      400      {object} map[string]string  "validation or recaptcha error"
// @Failure      429      {object} map[string]string  "rate limit exceeded"
// @Failure      500      {object} map[string]string  "internal server error"
// @Router       /contact [post]
func (h *Handler) SendContactEmail(w http.ResponseWriter, r *http.Request) {
	//Cap request body to 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	//Decode JSON
	var req dto.ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//Trim whitespace
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Message = strings.TrimSpace(req.Message)
	// (req.Token is a one-time value; no trim)

	//Field validation & blacklists
	if req.Name == "" || len(req.Name) > 100 || !isSafeHeaderValue(req.Name) {
		http.Error(w, "Invalid name (required, max 100 chars, no special symbols)", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) || !isSafeHeaderValue(req.Email) {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}
	if req.Phone != "" {
		//Regex‐based sanity check
		if !isValidPhone(req.Phone) || !isSafeHeaderValue(req.Phone) {
			http.Error(w, "Invalid phone number format", http.StatusBadRequest)
			return
		}

		//Count digits and enforce a minimum (e.g. 10 digits)
		// This allows for international formats like +1 (123) 456-7890
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

	//Escape the message body
	req.Message = html.EscapeString(req.Message)

	//Verify reCAPTCHA
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

	//Send the email
	if err := service.SendContactRequest(req); err != nil {
		log.Printf("SendContactRequest error: %v", err)
		http.Error(w, "Failed to send contact request", http.StatusInternalServerError)
		return
	}

	// Send to HubSpot as a lead
	hubspotSvc := hupspotService.GetHubSpotService(nil)
	if _, err := hubspotSvc.CreateContactLead(req.Name, req.Email, req.Phone, req.Message); err != nil {
		// Log the error but don't fail the request since email was sent successfully
		log.Printf("Failed to create HubSpot lead: %v", err)
	}

	//Success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Contact request sent successfully"}`))
}

// isSafeHeaderValue checks if a header value is safe to use.
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

// isValidEmail checks if the email address is in a valid format.
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	return re.MatchString(email)
}

// isValidPhone checks if the phone number is in a valid format.
func isValidPhone(phone string) bool {
	re := regexp.MustCompile(`^\+?[0-9\s\-\(\)]+$`)
	return re.MatchString(phone)
}

// SubscribeNewsletter handles POST /newsletter
// @Summary      Subscribe to newsletter
// @Description  Adds or updates a contact with a HubSpot newsletter tag
// @Tags         newsletter
// @Accept       json
// @Produce      json
// @Param        payload  body    dto.NewsletterRequest  true  "Email and tag"
// @Success      200      {object} map[string]string  "Subscription confirmation"
// @Failure      400      {object} map[string]string  "Validation error"
// @Failure      500      {object} map[string]string  "Internal server error"
// @Router       /newsletter [post]
func (h *Handler) SubscribeNewsletter(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req dto.NewsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Tag = strings.TrimSpace(req.Tag)

	if !isValidEmail(req.Email) || !isSafeHeaderValue(req.Email) {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	hubspotSvc := hupspotService.GetHubSpotService(nil)
	message, err := hubspotSvc.SubscribeToNewsletter(req.Email, req.Tag)
	if err != nil {
		log.Printf("SubscribeToNewsletter error: %v", err)
		http.Error(w, err.Error(), err.HTTPCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": message,
	})
}
