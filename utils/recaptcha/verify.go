package recaptcha

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type Response struct {
	Success     bool     `json:"success"`
	ChallengeTs string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

// Verify checks a reCAPTCHA v2 token by calling /siteverify.
func Verify(token string) (bool, error) {
	secret := os.Getenv("RECAPTCHA_SECRET")
	if secret == "" {
		return false, fmt.Errorf("RECAPTCHA_SECRET not set")
	}

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
		"secret":   {secret},
		"response": {token},
	})
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.Success, nil
}
