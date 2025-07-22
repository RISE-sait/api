package dto

type ChatRequest struct {
	Query       string     `json:"query"`
	Context     string     `json:"context"`
	ChatHistory [][]string `json:"chat_history"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}
