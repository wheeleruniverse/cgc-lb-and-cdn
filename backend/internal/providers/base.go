package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cgc-image-service/internal/models"
)

// BaseProvider provides common functionality for all image generation providers
type BaseProvider struct {
	name       string
	status     *models.ProviderStatus
	httpClient *http.Client
	imageDir   string
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(name string) *BaseProvider {
	return &BaseProvider{
		name: name,
		status: &models.ProviderStatus{
			Name:        name,
			Available:   true,
			LastSuccess: time.Now(),
			ErrorCount:  0,
			QuotaHit:    false,
			RateLimited: false,
			QuotaInfo: &models.ProviderQuota{
				Supported:   false,
				LastUpdated: time.Now(),
			},
		},
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		imageDir: "images",
	}
}

// GetName returns the provider name
func (bp *BaseProvider) GetName() string {
	return bp.name
}

// GetStatus returns the current status
func (bp *BaseProvider) GetStatus() *models.ProviderStatus {
	return bp.status
}

// IsAvailable checks if the provider is available
func (bp *BaseProvider) IsAvailable() bool {
	return bp.status.Available && !bp.status.QuotaHit && !bp.status.RateLimited
}

// HandleError processes errors and determines their type
func (bp *BaseProvider) HandleError(err error) *models.ProviderError {
	errMsg := err.Error()
	providerErr := &models.ProviderError{
		Provider:    bp.name,
		Message:     errMsg,
		IsQuotaHit:  false,
		IsRateLimit: false,
		Retryable:   true,
	}

	// Check for common quota/rate limit indicators
	errLower := strings.ToLower(errMsg)

	if strings.Contains(errLower, "quota") || strings.Contains(errLower, "limit exceeded") ||
		strings.Contains(errLower, "insufficient") || strings.Contains(errLower, "usage limit") {
		providerErr.IsQuotaHit = true
		providerErr.Code = "QUOTA_EXCEEDED"
		providerErr.Retryable = false
	} else if strings.Contains(errLower, "rate limit") || strings.Contains(errLower, "too many requests") ||
		strings.Contains(errLower, "429") {
		providerErr.IsRateLimit = true
		providerErr.Code = "RATE_LIMITED"
		providerErr.Retryable = true
	} else if strings.Contains(errLower, "unauthorized") || strings.Contains(errLower, "403") ||
		strings.Contains(errLower, "invalid key") {
		providerErr.Code = "UNAUTHORIZED"
		providerErr.Retryable = false
	} else {
		providerErr.Code = "UNKNOWN_ERROR"
	}

	// Update status
	bp.status.LastError = errMsg
	bp.status.ErrorCount++
	bp.status.QuotaHit = providerErr.IsQuotaHit
	bp.status.RateLimited = providerErr.IsRateLimit
	bp.status.Available = !(providerErr.IsQuotaHit || providerErr.IsRateLimit)

	return providerErr
}

// RefreshQuota provides a default implementation (no quota support)
func (bp *BaseProvider) RefreshQuota(ctx context.Context) error {
	// Default implementation - no quota support
	if bp.status.QuotaInfo == nil {
		bp.status.QuotaInfo = &models.ProviderQuota{
			Supported:   false,
			LastUpdated: time.Now(),
		}
	}
	return nil
}

// MakeHTTPRequest is a helper for making HTTP requests with error handling
func (bp *BaseProvider) MakeHTTPRequest(method, url string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := bp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// ParseJSONResponse parses a JSON response into the provided interface
func (bp *BaseProvider) ParseJSONResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}
