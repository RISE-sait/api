package service

import (
	"api/config"
	errLib "api/internal/libs/errors"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type Service struct {
	Client  *http.Client
	BaseURL string
}

func NewService() *Service {
	return &Service{
		Client:  &http.Client{Timeout: 30 * time.Second},
		BaseURL: config.Env.ChatBotServiceUrl,
	}
}

func (s *Service) Chat(message, context string, chatHistory [][]string) (string, *errLib.CommonError) {
	if s.BaseURL == "" {
		log.Println("[Chat Debug] ChatBotServiceUrl is empty")
		return "", errLib.New("chatbot service URL not configured", http.StatusInternalServerError)
	}

	log.Printf("[Chat Debug] Using chatbot URL: %s", s.BaseURL)

	reqBody := map[string]interface{}{
		"query":        message,
		"chat_history": chatHistory,
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[Chat Debug] Failed to marshal request body: %v", err)
		return "", errLib.New("failed to marshal request", http.StatusInternalServerError)
	}

	log.Printf("[Chat Debug] Request Body: %s", string(buf))

	httpReq, err := http.NewRequest(http.MethodPost, s.BaseURL, bytes.NewBuffer(buf))
	if err != nil {
		log.Printf("[Chat Debug] Failed to create HTTP request: %v", err)
		return "", errLib.New("failed to create request", http.StatusInternalServerError)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(httpReq)
	if err != nil {
		log.Printf("[Chat Debug] HTTP request failed: %v", err)
		return "", errLib.New("failed to contact chat service", http.StatusBadGateway)
	}
	defer resp.Body.Close()

	log.Printf("[Chat Debug] Response Status Code: %d", resp.StatusCode)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Chat Debug] Error Response Body: %s", string(body))
		return "", errLib.New("chat service error", http.StatusBadGateway)
	}

	var res struct {
		Answer string `json:"answer"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Printf("[Chat Debug] Failed to decode response: %v", err)
		return "", errLib.New("invalid chat response", http.StatusInternalServerError)
	}

	log.Printf("[Chat Debug] Chatbot Answer: %s", res.Answer)
	return res.Answer, nil
}
