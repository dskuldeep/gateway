package intelligence

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dskuldeep/gateway/internal/llms"
)

// Evaluator defines the interface for intelligence layer evaluators
type Evaluator interface {
	// Evaluate evaluates an LLM response
	Evaluate(ctx context.Context, req llms.Request, resp llms.Response) (*Evaluation, error)
}

// Evaluation represents the result of an intelligence layer evaluation
type Evaluation struct {
	Score       float64           `json:"score"`       // 0-1 score of response quality
	Reason      string            `json:"reason"`      // Explanation of the score
	Suggestions []string          `json:"suggestions"` // Suggestions for improvement
	Metadata    map[string]string `json:"metadata"`    // Additional metadata
	Timestamp   time.Time         `json:"timestamp"`   // When the evaluation was performed
}

// Config represents the configuration for an evaluator
type Config struct {
	Enabled      bool              `json:"enabled"`
	SamplingRate float64           `json:"sampling_rate"` // 0-1, percentage of requests to evaluate
	Keywords     []string          `json:"keywords"`      // Keywords that trigger evaluation
	MinScore     float64           `json:"min_score"`     // Minimum acceptable score
	MaxLatency   time.Duration     `json:"max_latency"`   // Maximum acceptable latency
	CustomRules  map[string]string `json:"custom_rules"`  // Custom evaluation rules
}

// DefaultEvaluator is a sample implementation that uses OpenAI to evaluate responses
type DefaultEvaluator struct {
	config    Config
	llmClient llms.LLMClient
}

// NewDefaultEvaluator creates a new default evaluator
func NewDefaultEvaluator(config Config, llmClient llms.LLMClient) *DefaultEvaluator {
	return &DefaultEvaluator{
		config:    config,
		llmClient: llmClient,
	}
}

// Evaluate implements the Evaluator interface
func (e *DefaultEvaluator) Evaluate(ctx context.Context, req llms.Request, resp llms.Response) (*Evaluation, error) {
	// Check if evaluation is enabled
	if !e.config.Enabled {
		return nil, nil
	}

	// Check sampling rate
	if e.config.SamplingRate < 1.0 && time.Now().UnixNano()%100 > int64(e.config.SamplingRate*100) {
		return nil, nil
	}

	// Check keywords
	shouldEvaluate := false
	for _, keyword := range e.config.Keywords {
		if strings.Contains(strings.ToLower(resp.Text), strings.ToLower(keyword)) {
			shouldEvaluate = true
			break
		}
	}

	if !shouldEvaluate {
		return nil, nil
	}

	// Prepare evaluation prompt
	prompt := fmt.Sprintf(`Evaluate the following LLM response:

Request: %s
Response: %s

Evaluate the response based on:
1. Accuracy and relevance
2. Completeness
3. Clarity and coherence
4. Safety and appropriateness

Provide a score from 0-1 and explain your reasoning.`, req.Prompt, resp.Text)

	// Make evaluation request
	evalReq := llms.Request{
		Model:       "gpt-4",
		Prompt:      prompt,
		Temperature: 0.3,
	}

	evalResp, err := e.llmClient.Query(ctx, evalReq, "")
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate response: %w", err)
	}

	// Parse evaluation
	var evaluation Evaluation
	if err := json.Unmarshal([]byte(evalResp.Text), &evaluation); err != nil {
		// If JSON parsing fails, try to extract score from text
		evaluation = Evaluation{
			Score:     extractScore(evalResp.Text),
			Reason:    evalResp.Text,
			Timestamp: time.Now(),
		}
	}

	// Add metadata
	evaluation.Metadata = map[string]string{
		"model":         resp.Model,
		"latency":       resp.Latency.String(),
		"token_usage":   fmt.Sprintf("%d", resp.Usage.TotalTokens),
		"finish_reason": resp.FinishReason,
	}

	return &evaluation, nil
}

// extractScore attempts to extract a score from text
func extractScore(text string) float64 {
	// Simple implementation - look for numbers between 0 and 1
	// This could be made more sophisticated
	words := strings.Fields(text)
	for _, word := range words {
		if strings.HasPrefix(word, "0.") || word == "1" {
			var score float64
			if _, err := fmt.Sscanf(word, "%f", &score); err == nil {
				return score
			}
		}
	}
	return 0.5 // Default score if none found
}

// LoadConfig loads evaluator configuration from environment
func LoadConfig() Config {
	return Config{
		Enabled:      getEnvBool("INTELLIGENCE_ENABLED", true),
		SamplingRate: getEnvFloat("INTELLIGENCE_SAMPLING_RATE", 0.1),
		Keywords:     getEnvStrings("INTELLIGENCE_KEYWORDS", []string{"error", "sorry", "cannot"}),
		MinScore:     getEnvFloat("INTELLIGENCE_MIN_SCORE", 0.7),
		MaxLatency:   getEnvDuration("INTELLIGENCE_MAX_LATENCY", 5*time.Second),
		CustomRules:  getEnvMap("INTELLIGENCE_CUSTOM_RULES", map[string]string{}),
	}
}

// Helper functions for environment variables
func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true"
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		var f float64
		if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
			return f
		}
	}
	return defaultVal
}

func getEnvStrings(key string, defaultVal []string) []string {
	if val := os.Getenv(key); val != "" {
		return strings.Split(val, ",")
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

func getEnvMap(key string, defaultVal map[string]string) map[string]string {
	if val := os.Getenv(key); val != "" {
		m := make(map[string]string)
		if err := json.Unmarshal([]byte(val), &m); err == nil {
			return m
		}
	}
	return defaultVal
}
