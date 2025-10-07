package providers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"cgc-image-service/internal/models"

	"github.com/google/uuid"
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

// SaveToSpaces uploads image data to DigitalOcean Spaces using direct HTTP
func (bp *BaseProvider) SaveToSpaces(imageData []byte, filename, imageID string) (*models.GeneratedImage, error) {
	// Get DO Spaces configuration
	bucketName := os.Getenv("DO_SPACES_BUCKET")
	endpoint := os.Getenv("DO_SPACES_ENDPOINT")
	accessKey := os.Getenv("DO_SPACES_ACCESS_KEY")
	secretKey := os.Getenv("DO_SPACES_SECRET_KEY")

	if bucketName == "" || endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("missing DO Spaces configuration (DO_SPACES_BUCKET, DO_SPACES_ENDPOINT, DO_SPACES_ACCESS_KEY, DO_SPACES_SECRET_KEY)")
	}

	// Construct URL for upload
	url := fmt.Sprintf("https://%s.%s/%s", bucketName, endpoint, filename)

	// Create HTTP request
	req, err := http.NewRequest("PUT", url, bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "image/png")
	req.Header.Set("x-amz-acl", "public-read")

	// Create signature for authentication
	date := time.Now().UTC().Format(time.RFC1123)
	req.Header.Set("Date", date)

	// Create string to sign
	stringToSign := fmt.Sprintf("PUT\n\nimage/png\n%s\nx-amz-acl:public-read\n/%s/%s", date, bucketName, filename)

	// Create signature
	h := hmac.New(sha1.New, []byte(secretKey))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Set authorization header
	req.Header.Set("Authorization", fmt.Sprintf("AWS %s:%s", accessKey, signature))

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to DO Spaces: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to upload to DO Spaces: HTTP %d - %s", resp.StatusCode, string(body))
	}

	// Return CDN URL instead of direct Spaces URL
	cdnURL := fmt.Sprintf("https://%s.%s/%s", bucketName, endpoint, filename)

	return &models.GeneratedImage{
		ID:       imageID,
		Filename: filename,
		Path:     cdnURL,
		Size:     int64(len(imageData)),
	}, nil
}

// SaveImage saves image data to DigitalOcean Spaces CDN (production only)
func (bp *BaseProvider) SaveImage(imageData []byte, filePrefix string) (*models.GeneratedImage, error) {
	// Generate unique identifiers
	imageID := uuid.New().String()
	filename := fmt.Sprintf("%s-%s.png", filePrefix, imageID)

	// Always use DO Spaces (no local storage fallback)
	return bp.SaveToSpaces(imageData, filename, imageID)
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
