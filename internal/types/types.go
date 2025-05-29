package types

// Provider represents an LLM provider
type Provider string

const (
	ProviderGroq      Provider = "groq"
	ProviderGoogle    Provider = "google"
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderMistral   Provider = "mistral"
)

// TokenUsage tracks token usage for a request
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
} 