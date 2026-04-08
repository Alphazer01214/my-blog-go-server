package response

type StandardAiResponse struct {
	AgentId          uint   `json:"agent_id"`
	ReasoningContent string `json:"reasoning_content"`
	Content          string `json:"content"`
	Status           bool   `json:"status"`
	Message          string `json:"message"`
}
