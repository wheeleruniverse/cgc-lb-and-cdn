package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"cgc-image-service/internal/models"
)

// LeonardoAIProvider implements image generation using Leonardo AI's API
type LeonardoAIProvider struct {
	*BaseProvider
	apiKey  string
	baseURL string
	modelID string
}

// LeonardoGenerationRequest represents the request structure for Leonardo AI
type LeonardoGenerationRequest struct {
	Height            int    `json:"height"`
	ModelID           string `json:"modelId"`
	Prompt            string `json:"prompt"`
	Width             int    `json:"width"`
	NumImages         int    `json:"num_images"`
	GuidanceScale     int    `json:"guidance_scale,omitempty"`
	NumInferenceSteps int    `json:"num_inference_steps,omitempty"`
}

// LeonardoGenerationResponse represents the initial response from Leonardo AI
type LeonardoGenerationResponse struct {
	SDGenerationJob LeonardoGenerationJob `json:"sdGenerationJob"`
}

// LeonardoGenerationJob represents the generation job
type LeonardoGenerationJob struct {
	GenerationID string `json:"generationId"`
}

// LeonardoStatusResponse represents the status check response
type LeonardoStatusResponse struct {
	GenerationsByPK LeonardoGeneration `json:"generations_by_pk"`
}

// LeonardoGeneration represents a generation with its images
type LeonardoGeneration struct {
	Status          string          `json:"status"`
	GeneratedImages []LeonardoImage `json:"generated_images"`
}

// LeonardoImage represents a generated image
type LeonardoImage struct {
	URL string `json:"url"`
	ID  string `json:"id"`
}

// NewLeonardoAIProvider creates a new Leonardo AI provider
func NewLeonardoAIProvider() *LeonardoAIProvider {
	apiKey := os.Getenv("LEONARDO_API_KEY")
	if apiKey == "" {
		// Provider will be marked as unavailable
	}

	provider := &LeonardoAIProvider{
		BaseProvider: NewBaseProvider("leonardo-ai"),
		apiKey:       apiKey,
		baseURL:      "https://cloud.leonardo.ai/api/rest/v1",
		modelID:      "6bef9f1b-29cb-40c7-b9df-32b51c1f67d3", // Leonardo Creative model
	}

	// Mark as unavailable if no API key
	if apiKey == "" {
		provider.status.Available = false
		provider.status.LastError = "LEONARDO_API_KEY environment variable not set"
	}

	return provider
}

// Generate creates images using Leonardo AI's API
func (lp *LeonardoAIProvider) Generate(ctx context.Context, req *models.ImageRequest) (*models.ImageResponse, error) {
	startTime := time.Now()

	if !lp.IsAvailable() {
		return nil, fmt.Errorf("leonardo ai provider is not available: %s", lp.status.LastError)
	}

	fmt.Printf("[LEONARDO-AI] Starting generation with prompt: %s, count: %d\n", req.Prompt, req.Count)

	// Default to 4 images if not specified
	count := req.Count
	if count <= 0 {
		count = 4
	}

	// Step 1: Start generation
	generationID, err := lp.startGeneration(req.Prompt, count)
	if err != nil {
		return nil, fmt.Errorf("failed to start generation: %w", err)
	}

	// Step 2: Poll for completion
	images, err := lp.pollForCompletion(ctx, generationID, req.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for completion: %w", err)
	}

	// Update success status
	lp.status.LastSuccess = time.Now()
	lp.status.Available = true
	lp.status.LastError = ""

	return &models.ImageResponse{
		Images:    images,
		Provider:  lp.GetName(),
		Success:   true,
		RequestID: req.RequestID,
		Duration:  time.Since(startTime),
		Metadata: map[string]string{
			"model_id":      lp.modelID,
			"generation_id": generationID,
			"api_version":   "v1",
		},
	}, nil
}

// startGeneration initiates the image generation process
func (lp *LeonardoAIProvider) startGeneration(prompt string, count int) (string, error) {
	// Prepare request
	leonardoReq := LeonardoGenerationRequest{
		Height:            1024,
		ModelID:           lp.modelID,
		Prompt:            prompt,
		Width:             1024,
		NumImages:         count,
		GuidanceScale:     7,
		NumInferenceSteps: 15,
	}

	jsonData, err := json.Marshal(leonardoReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + lp.apiKey,
	}

	resp, err := lp.MakeHTTPRequest("POST", lp.baseURL+"/generations", headers, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// Parse response
	var leonardoResp LeonardoGenerationResponse
	if err := lp.ParseJSONResponse(resp, &leonardoResp); err != nil {
		return "", err
	}

	return leonardoResp.SDGenerationJob.GenerationID, nil
}

// pollForCompletion polls the API until generation is complete
func (lp *LeonardoAIProvider) pollForCompletion(ctx context.Context, generationID, bucketName string) ([]models.GeneratedImage, error) {
	maxAttempts := 24 // 2 minutes with 5-second intervals

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Check status
		headers := map[string]string{
			"Authorization": "Bearer " + lp.apiKey,
		}

		url := fmt.Sprintf("%s/generations/%s", lp.baseURL, generationID)
		resp, err := lp.MakeHTTPRequest("GET", url, headers, nil)
		if err != nil {
			return nil, err
		}

		var statusResp LeonardoStatusResponse
		if err := lp.ParseJSONResponse(resp, &statusResp); err != nil {
			return nil, err
		}

		generation := statusResp.GenerationsByPK

		switch generation.Status {
		case "COMPLETE":
			// Download and save images
			var images []models.GeneratedImage
			for i, img := range generation.GeneratedImages {
				generatedImg, err := lp.saveImageFromURL(img.URL, "leonardo-ai", bucketName)
				if err != nil {
					return nil, fmt.Errorf("failed to save image %d: %w", i+1, err)
				}
				images = append(images, *generatedImg)
			}
			return images, nil

		case "FAILED":
			return nil, fmt.Errorf("generation failed")

		case "PENDING":
			// Continue polling
			time.Sleep(5 * time.Second)

		default:
			// Unknown status, continue polling
			time.Sleep(5 * time.Second)
		}
	}

	return nil, fmt.Errorf("generation timed out after %d attempts", maxAttempts)
}

// saveImageFromURL downloads and saves an image from a URL using shared BaseProvider method
func (lp *LeonardoAIProvider) saveImageFromURL(imageURL, filePrefix, bucketName string) (*models.GeneratedImage, error) {
	// Download image
	resp, err := lp.httpClient.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	// Read image data into memory
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// Use shared BaseProvider method
	return lp.BaseProvider.SaveImage(imageData, filePrefix, bucketName)
}

// LeonardoUserResponse represents the response from Leonardo AI /me endpoint
type LeonardoUserResponse struct {
	UserDetails []LeonardoUserDetail `json:"user_details"`
}

// LeonardoUserDetail represents user details including quota information
type LeonardoUserDetail struct {
	User                    LeonardoUser `json:"user"`
	TokenRenewalDate        *string      `json:"tokenRenewalDate"`
	PaidTokens              int          `json:"paidTokens"`
	SubscriptionTokens      int          `json:"subscriptionTokens"`
	SubscriptionGptTokens   int          `json:"subscriptionGptTokens"`
	SubscriptionModelTokens int          `json:"subscriptionModelTokens"`
	APIConcurrencySlots     int          `json:"apiConcurrencySlots"`
	APIPaidTokens           *int         `json:"apiPaidTokens"`
	APISubscriptionTokens   int          `json:"apiSubscriptionTokens"`
	APIPlanTokenRenewalDate string       `json:"apiPlanTokenRenewalDate"`
}

// LeonardoUser represents basic user information
type LeonardoUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// RefreshQuota updates quota information from Leonardo AI's /me endpoint
func (lp *LeonardoAIProvider) RefreshQuota(ctx context.Context) error {
	if !lp.IsAvailable() {
		return fmt.Errorf("provider not available")
	}

	fmt.Printf("[LEONARDO-AI] Refreshing quota information\n")

	// Make API request to /me endpoint
	headers := map[string]string{
		"Authorization": "Bearer " + lp.apiKey,
	}

	resp, err := lp.MakeHTTPRequest("GET", lp.baseURL+"/me", headers, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch user info: %w", err)
	}

	var userResp LeonardoUserResponse
	if err := lp.ParseJSONResponse(resp, &userResp); err != nil {
		return fmt.Errorf("failed to parse user response: %w", err)
	}

	if len(userResp.UserDetails) == 0 {
		return fmt.Errorf("no user details in response")
	}

	userDetail := userResp.UserDetails[0]

	// Parse renewal date
	var renewalDate time.Time
	if userDetail.APIPlanTokenRenewalDate != "" {
		if parsed, err := time.Parse(time.RFC3339, userDetail.APIPlanTokenRenewalDate); err == nil {
			renewalDate = parsed
		}
	}

	// Calculate total available tokens
	totalTokens := userDetail.APISubscriptionTokens
	if userDetail.APIPaidTokens != nil {
		totalTokens += *userDetail.APIPaidTokens
	}

	// Update quota information
	lp.status.QuotaInfo = &models.ProviderQuota{
		APITokens:          userDetail.APISubscriptionTokens,
		PaidTokens:         userDetail.PaidTokens,
		SubscriptionTokens: userDetail.SubscriptionTokens,
		ConcurrencySlots:   userDetail.APIConcurrencySlots,
		Total:              totalTokens,
		Remaining:          userDetail.APISubscriptionTokens, // API tokens are the main quota
		RenewalDate:        renewalDate,
		LastUpdated:        time.Now(),
		Supported:          true,
	}

	fmt.Printf("[LEONARDO-AI] Quota updated - API Tokens: %d, Renewal: %s\n",
		userDetail.APISubscriptionTokens, renewalDate.Format("2006-01-02"))

	return nil
}
