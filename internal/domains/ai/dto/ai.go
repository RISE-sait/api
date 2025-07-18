package dto

type ChatRequest struct {
	Message string `json:"message"`
	Context string `json:"context"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}