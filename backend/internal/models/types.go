package models

import (
	"time"
)

// ImageRequest represents a request to generate images
type ImageRequest struct {
	Prompt    string            `json:"prompt" binding:"required"`
	Count     int               `json:"count,omitempty"`
	Width     int               `json:"width,omitempty"`
	Height    int               `json:"height,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
}

// ImageResponse represents the response from image generation
type ImageResponse struct {
	Images    []GeneratedImage  `json:"images"`
	Provider  string            `json:"provider"`
	Success   bool              `json:"success"`
	Message   string            `json:"message,omitempty"`
	RequestID string            `json:"request_id"`
	Duration  time.Duration     `json:"duration"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// GeneratedImage represents a single generated image
type GeneratedImage struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
	URL      string `json:"url,omitempty"`
	Size     int64  `json:"size"`
}

// ProviderError represents errors from image generation providers
type ProviderError struct {
	Provider    string `json:"provider"`
	Code        string `json:"code"`
	Message     string `json:"message"`
	IsQuotaHit  bool   `json:"is_quota_hit"`
	IsRateLimit bool   `json:"is_rate_limit"`
	Retryable   bool   `json:"retryable"`
}

func (e *ProviderError) Error() string {
	return e.Message
}

// ProviderStatus represents the current status of a provider
type ProviderStatus struct {
	Name        string    `json:"name"`
	Available   bool      `json:"available"`
	LastError   string    `json:"last_error,omitempty"`
	LastSuccess time.Time `json:"last_success"`
	ErrorCount  int       `json:"error_count"`
	QuotaHit    bool      `json:"quota_hit"`
	RateLimited bool      `json:"rate_limited"`
}

// AgentDecision represents a decision made by the orchestrator agent
type AgentDecision struct {
	SelectedProvider string            `json:"selected_provider"`
	Reasoning        string            `json:"reasoning"`
	FallbackOrder    []string          `json:"fallback_order"`
	Confidence       float64           `json:"confidence"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}
