package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cgc-image-service/internal/agents"
	"cgc-image-service/internal/models"
	"cgc-image-service/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ImageHandler handles image generation requests
type ImageHandler struct {
	orchestrator agents.OrchestratorAgent
}

// NewImageHandler creates a new image handler
func NewImageHandler(orchestrator agents.OrchestratorAgent) *ImageHandler {
	return &ImageHandler{
		orchestrator: orchestrator,
	}
}

// GenerateImage handles POST /generate requests
func (h *ImageHandler) GenerateImage(c *gin.Context) {
	var req models.ImageRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", map[string]string{
			"validation_error": err.Error(),
		})
		return
	}

	// Set request metadata
	req.RequestID = uuid.New().String()
	req.Timestamp = time.Now()

	// Validate request
	if req.Prompt == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "Prompt is required", "MISSING_PROMPT", nil)
		return
	}

	// Set defaults
	if req.Count <= 0 {
		req.Count = 4
	} else if req.Count > 4 {
		req.Count = 4 // Limit to prevent abuse
	}

	// Execute through orchestrator
	result, err := h.orchestrator.Execute(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Image generation failed", "GENERATION_FAILED", map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Type assert the result
	response, ok := result.(*models.ImageResponse)
	if !ok {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid response type from orchestrator", "INVALID_RESPONSE", nil)
		return
	}

	// Return success response
	utils.RespondWithSuccess(c, response, "Images generated successfully", map[string]string{
		"provider":   response.Provider,
		"count":      string(rune(len(response.Images))),
		"duration":   response.Duration.String(),
		"request_id": response.RequestID,
	})
}

// GetProviderStatus handles GET /status requests
func (h *ImageHandler) GetProviderStatus(c *gin.Context) {
	// Check if quota refresh is requested
	refreshQuota := c.Query("refresh_quota") == "true"

	if refreshQuota {
		h.refreshAllProviderQuotas(c.Request.Context())
	}

	status := h.orchestrator.GetProviderStatus()

	utils.RespondWithSuccess(c, status, "Provider status retrieved", map[string]string{
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"quota_refreshed": fmt.Sprintf("%t", refreshQuota),
	})
}

// refreshAllProviderQuotas refreshes quota information for all providers
func (h *ImageHandler) refreshAllProviderQuotas(ctx context.Context) {
	providerStatus := h.orchestrator.GetProviderStatus()

	for providerName := range providerStatus {
		fmt.Printf("[STATUS] Refreshing quota for provider: %s\n", providerName)

		provider, exists := h.orchestrator.GetProvider(providerName)
		if !exists {
			fmt.Printf("[STATUS] Provider %s not found\n", providerName)
			continue
		}

		if err := provider.RefreshQuota(ctx); err != nil {
			fmt.Printf("[STATUS] Failed to refresh quota for %s: %v\n", providerName, err)
		} else {
			fmt.Printf("[STATUS] Successfully refreshed quota for %s\n", providerName)
		}
	}
}

// HealthCheck handles GET /health requests
func (h *ImageHandler) HealthCheck(c *gin.Context) {
	status := h.orchestrator.GetProviderStatus()

	// Check if at least one provider is available
	availableCount := 0
	totalCount := len(status)

	for _, providerStatus := range status {
		if providerStatus.Available {
			availableCount++
		}
	}

	healthStatus := "healthy"
	statusCode := http.StatusOK

	if availableCount == 0 {
		healthStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	} else if availableCount < totalCount {
		healthStatus = "degraded"
	}

	c.JSON(statusCode, gin.H{
		"status":              healthStatus,
		"available_providers": availableCount,
		"total_providers":     totalCount,
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
		"providers":           status,
	})
}

// GetImagePair handles GET /images/pair requests
func (h *ImageHandler) GetImagePair(c *gin.Context) {
	images, err := h.getRandomImages(2)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get random images", "IMAGES_UNAVAILABLE", map[string]string{
			"error": err.Error(),
		})
		return
	}

	if len(images) < 2 {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "Not enough images available", "INSUFFICIENT_IMAGES", map[string]string{
			"available": fmt.Sprintf("%d", len(images)),
			"required":  "2",
		})
		return
	}

	pairID := uuid.New().String()
	response := models.ImagePairResponse{
		PairID: pairID,
		Left:   images[0],
		Right:  images[1],
	}

	utils.RespondWithSuccess(c, response, "Image pair retrieved successfully", map[string]string{
		"pair_id":  pairID,
		"left_id":  images[0].ID,
		"right_id": images[1].ID,
	})
}

// SubmitRating handles POST /images/rate requests
func (h *ImageHandler) SubmitRating(c *gin.Context) {
	var req models.ComparisonRatingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", map[string]string{
			"validation_error": err.Error(),
		})
		return
	}

	// Validate winner value
	if req.Winner != "left" && req.Winner != "right" {
		utils.RespondWithError(c, http.StatusBadRequest, "Winner must be 'left' or 'right'", "INVALID_WINNER", nil)
		return
	}

	// Here you could store the rating in a database
	// For now, we'll just log it and return success
	fmt.Printf("[RATING] Pair: %s, Winner: %s, Left: %s, Right: %s\n",
		req.PairID, req.Winner, req.LeftID, req.RightID)

	response := models.ComparisonRatingResponse{
		Success:   true,
		PairID:    req.PairID,
		Winner:    req.Winner,
		Message:   "Rating submitted successfully",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	utils.RespondWithSuccess(c, response, "Rating submitted successfully", map[string]string{
		"pair_id": req.PairID,
		"winner":  req.Winner,
	})
}

// getRandomImages returns a slice of random images from the images directory
func (h *ImageHandler) getRandomImages(count int) ([]models.ImageInfo, error) {
	imagesDir := "images"

	files, err := os.ReadDir(imagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read images directory: %w", err)
	}

	var imageFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && isImageFile(file.Name()) {
			imageFiles = append(imageFiles, file)
		}
	}

	if len(imageFiles) == 0 {
		return nil, fmt.Errorf("no image files found")
	}

	// Shuffle the slice
	rand.Shuffle(len(imageFiles), func(i, j int) {
		imageFiles[i], imageFiles[j] = imageFiles[j], imageFiles[i]
	})

	// Take the requested count
	if count > len(imageFiles) {
		count = len(imageFiles)
	}

	var images []models.ImageInfo
	for i := 0; i < count; i++ {
		file := imageFiles[i]
		info, err := file.Info()
		if err != nil {
			continue
		}

		// Extract provider from filename (e.g., "freepik-uuid.png" -> "freepik")
		provider := extractProviderFromFilename(file.Name())

		imageInfo := models.ImageInfo{
			ID:       strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())),
			Filename: file.Name(),
			Path:     filepath.Join(imagesDir, file.Name()),
			URL:      fmt.Sprintf("/images/%s", file.Name()),
			Provider: provider,
			Size:     info.Size(),
		}
		images = append(images, imageInfo)
	}

	return images, nil
}

// isImageFile checks if a file is an image based on its extension
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"
}

// extractProviderFromFilename extracts the provider name from the filename
func extractProviderFromFilename(filename string) string {
	parts := strings.Split(filename, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}
