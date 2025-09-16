package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cgc-image-service/internal/models"
)

// FreepikProvider implements image generation using Freepik's API
type FreepikProvider struct {
	*BaseProvider
	apiKey  string
	baseURL string
}

// FreepikRequest represents the request structure for Freepik API
type FreepikRequest struct {
	Prompt      string `json:"prompt"`
	NumImages   int    `json:"num_images"`
	AspectRatio string `json:"aspect_ratio,omitempty"`
}

// FreepikResponse represents the response from Freepik API
type FreepikResponse struct {
	Data []FreepikImage `json:"data"`
}

// FreepikImage represents a single image in Freepik response
type FreepikImage struct {
	Image string `json:"image"` // base64 encoded image
}

// NewFreepikProvider creates a new Freepik provider
func NewFreepikProvider() *FreepikProvider {
	apiKey := os.Getenv("FREEPIK_API_KEY")
	if apiKey == "" {
		// Provider will be marked as unavailable
	}

	provider := &FreepikProvider{
		BaseProvider: NewBaseProvider("freepik"),
		apiKey:       apiKey,
		baseURL:      "https://api.freepik.com",
	}

	// Mark as unavailable if no API key
	if apiKey == "" {
		provider.status.Available = false
		provider.status.LastError = "FREEPIK_API_KEY environment variable not set"
	}

	return provider
}

// Generate creates images using Freepik's Classic Fast API
func (fp *FreepikProvider) Generate(ctx context.Context, req *models.ImageRequest) (*models.ImageResponse, error) {
	startTime := time.Now()

	if !fp.IsAvailable() {
		return nil, fmt.Errorf("freepik provider is not available: %s", fp.status.LastError)
	}

	// Default to 4 images if not specified
	count := req.Count
	if count <= 0 {
		count = 4
	}

	// Prepare request
	freepikReq := FreepikRequest{
		Prompt:      req.Prompt,
		NumImages:   count,
		AspectRatio: "1:1", // Square aspect ratio by default
	}

	jsonData, err := json.Marshal(freepikReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	headers := map[string]string{
		"Content-Type":      "application/json",
		"x-freepik-api-key": fp.apiKey,
	}

	resp, err := fp.MakeHTTPRequest("POST", fp.baseURL+"/v1/ai/text-to-image", headers, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Parse response
	var freepikResp FreepikResponse
	if err := fp.ParseJSONResponse(resp, &freepikResp); err != nil {
		return nil, err
	}

	// Debug: Check what we got
	if len(freepikResp.Data) == 0 {
		return nil, fmt.Errorf("no images returned from Freepik API")
	}

	// Process images
	var images []models.GeneratedImage
	for i, img := range freepikResp.Data {
		if i >= count {
			break // Limit to requested count
		}

		// Debug: Check if we have base64 data
		if img.Image == "" {
			return nil, fmt.Errorf("image %d has empty base64 data", i+1)
		}

		generatedImg, err := fp.SaveImageFromBase64(img.Image, "freepik")
		if err != nil {
			return nil, fmt.Errorf("failed to save image %d: %w", i+1, err)
		}

		images = append(images, *generatedImg)
	}

	// Update success status
	fp.status.LastSuccess = time.Now()
	fp.status.Available = true
	fp.status.LastError = ""

	return &models.ImageResponse{
		Images:    images,
		Provider:  fp.GetName(),
		Success:   true,
		RequestID: req.RequestID,
		Duration:  time.Since(startTime),
		Metadata: map[string]string{
			"model":        "classic-fast",
			"aspect_ratio": "1:1",
			"api_version":  "v1",
		},
	}, nil
}
