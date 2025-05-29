package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/dskuldeep/gateway/internal/types"
	"net/http"
)

var (
	// LLM request latency
	llmLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_request_latency_seconds",
			Help:    "Latency of LLM requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider", "model"},
	)

	// LLM token usage
	llmTokens = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_tokens_total",
			Help: "Total number of tokens used",
		},
		[]string{"provider", "model", "type"},
	)

	// LLM request errors
	llmErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_request_errors_total",
			Help: "Total number of LLM request errors",
		},
		[]string{"provider", "model"},
	)

	// API request latency
	apiLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_latency_seconds",
			Help:    "Latency of API requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method"},
	)

	// API request errors
	apiErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_request_errors_total",
			Help: "Total number of API request errors",
		},
		[]string{"endpoint", "method", "status"},
	)
)

// Init initializes all metrics
func Init() {
	prometheus.MustRegister(llmLatency)
	prometheus.MustRegister(llmTokens)
	prometheus.MustRegister(llmErrors)
	prometheus.MustRegister(apiLatency)
	prometheus.MustRegister(apiErrors)
}

// Handler returns the Prometheus metrics handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// RecordLLMLatency records the latency of an LLM request
func RecordLLMLatency(provider types.Provider, model string, duration float64) {
	llmLatency.WithLabelValues(string(provider), model).Observe(duration)
}

// RecordTokenUsage records token usage for an LLM request
func RecordTokenUsage(provider types.Provider, model string, usage types.TokenUsage) {
	llmTokens.WithLabelValues(string(provider), model, "prompt").Add(float64(usage.PromptTokens))
	llmTokens.WithLabelValues(string(provider), model, "completion").Add(float64(usage.CompletionTokens))
	llmTokens.WithLabelValues(string(provider), model, "total").Add(float64(usage.TotalTokens))
}

// RecordLLMError records an LLM request error
func RecordLLMError(provider types.Provider, model string) {
	llmErrors.WithLabelValues(string(provider), model).Inc()
}

// RecordAPILatency records the latency of an API request
func RecordAPILatency(endpoint, method string, duration float64) {
	apiLatency.WithLabelValues(endpoint, method).Observe(duration)
}

// RecordAPIError records an API request error
func RecordAPIError(endpoint, method string, status int) {
	apiErrors.WithLabelValues(endpoint, method, string(status)).Inc()
} 