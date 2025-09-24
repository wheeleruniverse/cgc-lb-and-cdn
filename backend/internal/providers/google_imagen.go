package providers

import (
	"context"
	"fmt"
	"os"
	"time"

	"cgc-image-service/internal/models"

	"google.golang.org/genai"
)

// GoogleImagenProvider implements image generation using Google's Imagen API
type GoogleImagenProvider struct {
	*BaseProvider
	client *genai.Client
}

// NewGoogleImagenProvider creates a new Google Imagen provider
func NewGoogleImagenProvider() *GoogleImagenProvider {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		// Provider will be marked as unavailable
		provider := &GoogleImagenProvider{
			BaseProvider: NewBaseProvider("google-imagen"),
			client:       nil,
		}
		provider.status.Available = false
		provider.status.LastError = "GOOGLE_API_KEY environment variable not set"
		return provider
	}

	// Create client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		provider := &GoogleImagenProvider{
			BaseProvider: NewBaseProvider("google-imagen"),
			client:       nil,
		}
		provider.status.Available = false
		provider.status.LastError = fmt.Sprintf("Failed to create genai client: %v", err)
		return provider
	}

	provider := &GoogleImagenProvider{
		BaseProvider: NewBaseProvider("google-imagen"),
		client:       client,
	}

	return provider
}

// Generate creates images using Google's Imagen API
func (gp *GoogleImagenProvider) Generate(ctx context.Context, req *models.ImageRequest) (*models.ImageResponse, error) {
	startTime := time.Now()

	if !gp.IsAvailable() {
		return nil, fmt.Errorf("google imagen provider is not available: %s", gp.status.LastError)
	}

	if gp.client == nil {
		return nil, fmt.Errorf("genai client not initialized")
	}

	fmt.Printf("[GOOGLE-IMAGEN] Starting generation with prompt: %s, count: %d\n", req.Prompt, req.Count)

	// Default to 4 images if not specified
	count := req.Count
	if count <= 0 {
		count = 4
	}

	// Create generation config
	config := &genai.GenerateImagesConfig{
		NumberOfImages: int32(count),
	}

	// Generate images
	generateImagesResponse, err := gp.client.Models.GenerateImages(
		ctx,
		"imagen-3.0-generate-002",
		req.Prompt,
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate images: %w", err)
	}

	fmt.Printf("[GOOGLE-IMAGEN] API returned %d images (requested %d)\n", len(generateImagesResponse.GeneratedImages), count)

	// Process images
	var images []models.GeneratedImage
	for i, image := range generateImagesResponse.GeneratedImages {
		fmt.Printf("[GOOGLE-IMAGEN] Processing image %d, size: %d bytes\n", i+1, len(image.Image.ImageBytes))
		// Save image bytes directly
		generatedImg, err := gp.saveImageFromBytes(image.Image.ImageBytes, "google-imagen")
		if err != nil {
			return nil, fmt.Errorf("failed to save image %d: %w", i+1, err)
		}

		images = append(images, *generatedImg)
	}

	fmt.Printf("[GOOGLE-IMAGEN] Final response will contain %d images\n", len(images))

	// Update success status
	gp.status.LastSuccess = time.Now()
	gp.status.Available = true
	gp.status.LastError = ""

	return &models.ImageResponse{
		Images:    images,
		Provider:  gp.GetName(),
		Success:   true,
		RequestID: req.RequestID,
		Duration:  time.Since(startTime),
		Metadata: map[string]string{
			"model":       "imagen-3.0-generate-002",
			"api_version": "genai-v0.14.0",
		},
	}, nil
}

// saveImageFromBytes saves image bytes directly using shared BaseProvider method
func (gp *GoogleImagenProvider) saveImageFromBytes(imageBytes []byte, filePrefix string) (*models.GeneratedImage, error) {
	// Check if we got any data
	if len(imageBytes) == 0 {
		return nil, fmt.Errorf("image bytes are empty")
	}

	fmt.Printf("[GOOGLE-IMAGEN] Saving image (size: %d bytes)\n", len(imageBytes))

	// Use shared BaseProvider method
	return gp.BaseProvider.SaveImage(imageBytes, filePrefix)
}
