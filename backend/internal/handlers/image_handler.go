package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"cgc-image-service/internal/agents"
	"cgc-image-service/internal/models"
	"cgc-image-service/internal/storage"
	"cgc-image-service/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ImageHandler handles image generation requests
type ImageHandler struct {
	orchestrator  agents.OrchestratorAgent
	valkeyClient  *storage.ValkeyClient
	cdnEndpoint   string
	spacesBucket  string
}

// NewImageHandler creates a new image handler
func NewImageHandler(orchestrator agents.OrchestratorAgent, valkeyClient *storage.ValkeyClient) *ImageHandler {
	cdnEndpoint := os.Getenv("DO_SPACES_ENDPOINT")
	if cdnEndpoint == "" {
		cdnEndpoint = "nyc3.digitaloceanspaces.com"
	}

	spacesBucket := os.Getenv("DO_SPACES_BUCKET")
	if spacesBucket == "" {
		spacesBucket = "cgc-lb-and-cdn-content"
	}

	return &ImageHandler{
		orchestrator: orchestrator,
		valkeyClient: valkeyClient,
		cdnEndpoint:  cdnEndpoint,
		spacesBucket: spacesBucket,
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

	// Store vote in Valkey
	if h.valkeyClient != nil {
		vote := &storage.Vote{
			PairID:  req.PairID,
			Winner:  req.Winner,
			LeftID:  req.LeftID,
			RightID: req.RightID,
		}

		if err := h.valkeyClient.RecordVote(c.Request.Context(), vote); err != nil {
			fmt.Printf("[ERROR] Failed to record vote in Valkey: %v\n", err)
			// Continue anyway - don't fail the request if Valkey is down
		} else {
			fmt.Printf("[VOTE] Recorded in Valkey - Pair: %s, Winner: %s\n", req.PairID, req.Winner)
		}
	}

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

// getRandomImages returns a slice of random images from the CDN
func (h *ImageHandler) getRandomImages(count int) ([]models.ImageInfo, error) {
	// Hardcoded list of pre-generated images
	// In production, this could come from a database or Spaces bucket listing
	imageFiles := []string{
		"leonardo-ai-7b9f4e11-3725-4723-b696-96da7a6cdf26.png",
		"leonardo-ai-b1b7a068-3ce3-4df0-a164-42a33db1e556.png",
		"leonardo-ai-e0f437af-3b6a-4d17-8cca-a5cc220e6442.png",
		"leonardo-ai-eb3fa3cd-c254-453c-a475-a3480c2b1ef6.png",
		"leonardo-ai-61028ab2-605f-4670-9a52-c4f1a34dbcbc.png",
		"leonardo-ai-d12e8720-ef07-4e7f-adb8-1b3fff3f1683.png",
		"leonardo-ai-7e9eb824-2d63-46fb-a8ca-954e008789f6.png",
		"leonardo-ai-79243f3b-5271-4769-a7e5-9fe89311a4c4.png",
		"leonardo-ai-053cb0ad-f7b6-465b-b286-ef2e1244685b.png",
		"leonardo-ai-bb9fbaf3-6bf6-4cb2-ba55-3b61e4bf15cb.png",
		"freepik-f1152da5-e707-4d84-9b9d-4c7d584187a2.png",
		"freepik-7f8bd76c-2eaa-4133-84f9-cc1f640146d8.png",
		"freepik-16b577b0-8fff-486c-9171-141a9fe035c7.png",
		"freepik-e3ba8750-4328-4f7e-8e64-6119b9004065.png",
		"leonardo-ai-ca5dee0f-ada7-42d2-877f-0d531acd8b95.png",
		"leonardo-ai-9a5edfcc-d443-4038-8b3c-01ffa843d672.png",
		"google-imagen-61edb395-7c9d-49c6-9129-5e92678df1c8.png",
		"freepik-eceff6ab-e8dc-4697-9fe5-fa3cb4708935.png",
		"freepik-bed5064a-8b5b-43b3-8c6c-8564be77c2da.png",
		"google-imagen-52ec5670-57a3-4808-925b-9be6db2c24bb.png",
		"google-imagen-a98440c6-76a7-4e9c-a9ee-9aaddd00a1a6.png",
		"freepik-11053ba3-4bc3-45b8-b868-74e02f8c7d46.png",
		"freepik-228a3550-ce0b-4668-804f-6437d52c91a4.png",
	}

	if len(imageFiles) == 0 {
		return nil, fmt.Errorf("no image files available")
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
		filename := imageFiles[i]

		// Extract provider from filename (e.g., "freepik-uuid.png" -> "freepik")
		provider := extractProviderFromFilename(filename)

		// Construct CDN URL
		cdnURL := fmt.Sprintf("https://%s.%s/%s", h.spacesBucket, h.cdnEndpoint, filename)

		imageInfo := models.ImageInfo{
			ID:       strings.TrimSuffix(filename, ".png"),
			Filename: filename,
			Path:     cdnURL,
			URL:      cdnURL,
			Provider: provider,
			Size:     0, // Size not available without fetching from CDN
		}
		images = append(images, imageInfo)
	}

	return images, nil
}


// GetLeaderboard handles GET /leaderboard requests
func (h *ImageHandler) GetLeaderboard(c *gin.Context) {
	if h.valkeyClient == nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "Leaderboard unavailable", "VALKEY_UNAVAILABLE", nil)
		return
	}

	stats, err := h.valkeyClient.GetProviderStats(c.Request.Context())
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get leaderboard", "LEADERBOARD_ERROR", map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Convert map to sorted slice
	type LeaderboardEntry struct {
		Provider   string  `json:"provider"`
		Wins       int64   `json:"wins"`
		Losses     int64   `json:"losses"`
		TotalVotes int64   `json:"total_votes"`
		WinRate    float64 `json:"win_rate"`
	}

	leaderboard := make([]LeaderboardEntry, 0, len(stats))
	for _, stat := range stats {
		leaderboard = append(leaderboard, LeaderboardEntry{
			Provider:   stat.Provider,
			Wins:       stat.Wins,
			Losses:     stat.Losses,
			TotalVotes: stat.TotalVotes,
			WinRate:    stat.WinRate,
		})
	}

	// Sort by wins descending
	for i := 0; i < len(leaderboard)-1; i++ {
		for j := i + 1; j < len(leaderboard); j++ {
			if leaderboard[j].Wins > leaderboard[i].Wins {
				leaderboard[i], leaderboard[j] = leaderboard[j], leaderboard[i]
			}
		}
	}

	utils.RespondWithSuccess(c, gin.H{
		"leaderboard": leaderboard,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}, "Leaderboard retrieved successfully", nil)
}

// GetStatistics handles GET /statistics requests
func (h *ImageHandler) GetStatistics(c *gin.Context) {
	if h.valkeyClient == nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "Statistics unavailable", "VALKEY_UNAVAILABLE", nil)
		return
	}

	stats, err := h.valkeyClient.GetProviderStats(c.Request.Context())
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get statistics", "STATISTICS_ERROR", map[string]string{
			"error": err.Error(),
		})
		return
	}

	totalVotes, err := h.valkeyClient.GetTotalVotes(c.Request.Context())
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get total votes", "STATISTICS_ERROR", map[string]string{
			"error": err.Error(),
		})
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"providers":   stats,
		"total_votes": totalVotes,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}, "Statistics retrieved successfully", nil)
}

// extractProviderFromFilename extracts the provider name from the filename
func extractProviderFromFilename(filename string) string {
	parts := strings.Split(filename, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}
