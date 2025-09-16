package handlers

import (
	"net/http"
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
	}
	if req.Count > 10 {
		req.Count = 10 // Limit to prevent abuse
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
	status := h.orchestrator.GetProviderStatus()

	utils.RespondWithSuccess(c, status, "Provider status retrieved", map[string]string{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
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
