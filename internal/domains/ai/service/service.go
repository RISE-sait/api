package service

import (
	"api/config"
	errLib "api/internal/libs/errors"
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type Service struct {
	Client  *http.Client
	BaseURL string
}

func NewService() *Service {
	return &Service{
		Client:  &http.Client{Timeout: 15 * time.Second},
		BaseURL: config.Env.ChatBotServiceUrl,
	}
}

func (s *Service) Chat(message, context string, chatHistory [][]string) (string, *errLib.CommonError) {
	if s.BaseURL == "" {
		return "", errLib.New("chatbot service URL not configured", http.StatusInternalServerError)
	}

	reqBody := map[string]interface{}{
		"query":        message,
		"chat_history": chatHistory,
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", errLib.New("failed to marshal request", http.StatusInternalServerError)
	}

	httpReq, err := http.NewRequest(http.MethodPost, s.BaseURL, bytes.NewBuffer(buf))
	if err != nil {
		return "", errLib.New("failed to create request", http.StatusInternalServerError)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return "", errLib.New("failed to contact chat service", http.StatusBadGateway)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", errLib.New("chat service error", http.StatusBadGateway)
	}

	var res struct {
		Answer string `json:"answer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", errLib.New("invalid chat response", http.StatusInternalServerError)
	}

	return res.Answer, nil
}
