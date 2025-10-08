package models

import (
	"time"
)

// ImageRequest represents a request to generate images
type ImageRequest struct {
	Prompt    string    `json:"prompt" binding:"required"`
	RequestID string    `json:"request_id,omitempty"`
	PairID    string    `json:"pair_id,omitempty"` // Unique identifier for this image pair
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// ImageResponse represents the response from image generation
type ImageResponse struct {
	Images    []GeneratedImage  `json:"images"`
	Provider  string            `json:"provider"`
	Success   bool              `json:"success"`
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
	Name        string         `json:"name"`
	Available   bool           `json:"available"`
	LastError   string         `json:"last_error,omitempty"`
	LastSuccess time.Time      `json:"last_success"`
	ErrorCount  int            `json:"error_count"`
	QuotaHit    bool           `json:"quota_hit"`
	RateLimited bool           `json:"rate_limited"`
	QuotaInfo   *ProviderQuota `json:"quota_info,omitempty"`
}

// ProviderQuota represents quota information for a provider
type ProviderQuota struct {
	Remaining          int       `json:"remaining,omitempty"`
	Total              int       `json:"total,omitempty"`
	RenewalDate        time.Time `json:"renewal_date,omitempty"`
	APITokens          int       `json:"api_tokens,omitempty"`
	SubscriptionTokens int       `json:"subscription_tokens,omitempty"`
	PaidTokens         int       `json:"paid_tokens,omitempty"`
	ConcurrencySlots   int       `json:"concurrency_slots,omitempty"`
	LastUpdated        time.Time `json:"last_updated"`
	Supported          bool      `json:"supported"`
}

// AgentDecision represents a decision made by the orchestrator agent
type AgentDecision struct {
	SelectedProvider string            `json:"selected_provider"`
	Reasoning        string            `json:"reasoning"`
	FallbackOrder    []string          `json:"fallback_order"`
	Confidence       float64           `json:"confidence"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// ImageInfo represents metadata about an existing image
type ImageInfo struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
	URL      string `json:"url"`
	Provider string `json:"provider"`
	Size     int64  `json:"size"`
}

// ImagePairResponse represents a pair of images for comparison
type ImagePairResponse struct {
	PairID string    `json:"pair_id"`
	Prompt string    `json:"prompt"`
	Left   ImageInfo `json:"left"`
	Right  ImageInfo `json:"right"`
}

// ComparisonRatingRequest represents a rating submission for image comparison
// Simplified: pair-id is sufficient since both images are from the same provider
type ComparisonRatingRequest struct {
	PairID string `json:"pair_id" binding:"required"`
	Winner string `json:"winner" binding:"required"` // "left" or "right"
}

// ComparisonRatingResponse represents the response to a rating submission
type ComparisonRatingResponse struct {
	Success   bool   `json:"success"`
	PairID    string `json:"pair_id"`
	Winner    string `json:"winner"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}
