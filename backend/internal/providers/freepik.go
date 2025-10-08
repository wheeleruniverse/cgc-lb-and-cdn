package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"cgc-lb-and-cdn-backend/internal/models"
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
	Base64 string `json:"base64"` // base64 encoded image
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

	fmt.Printf("[FREEPIK] Starting generation with prompt: %s, count: %d\n", req.Prompt, ImageCount)

	// Prepare request
	freepikReq := FreepikRequest{
		Prompt:      req.Prompt,
		NumImages:   ImageCount,
		AspectRatio: "square_1_1", // Square aspect ratio by default (correct format)
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
		if i >= ImageCount {
			break // Limit to ImageCount
		}

		// Debug: Check if we have base64 data
		if img.Base64 == "" {
			return nil, fmt.Errorf("image %d has empty base64 data", i+1)
		}

		generatedImg, err := fp.saveImageFromBase64(img.Base64, "freepik", req.PairID, req.Prompt, i)
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
			"aspect_ratio": "square_1_1",
			"api_version":  "v1",
		},
	}, nil
}

// saveImageFromBase64 saves a base64 encoded image using shared BaseProvider method
func (fp *FreepikProvider) saveImageFromBase64(base64Data, provider, pairID, prompt string, index int) (*models.GeneratedImage, error) {
	// Handle empty base64 data
	if base64Data == "" {
		return nil, fmt.Errorf("empty base64 data received")
	}

	// Remove data URL prefix if present (e.g., "data:image/png;base64,")
	if strings.Contains(base64Data, ",") {
		parts := strings.Split(base64Data, ",")
		if len(parts) > 1 {
			base64Data = parts[1]
		}
	}

	// Decode base64
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image (length: %d): %w", len(base64Data), err)
	}

	// Check if we got any data
	if len(imageData) == 0 {
		return nil, fmt.Errorf("decoded image data is empty")
	}

	// Use shared BaseProvider method with new simplified API
	return fp.BaseProvider.SaveImage(imageData, provider, pairID, prompt, index)
}
