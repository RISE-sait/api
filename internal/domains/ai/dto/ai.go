package dto

type ChatRequest struct {
	Message     string     `json:"message"`
	Context     string     `json:"context"`
	ChatHistory [][]string `json:"chat_history"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}
