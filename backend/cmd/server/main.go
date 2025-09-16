package main

import (
	"fmt"
	"log"

	"cgc-image-service/internal/agents"
	"cgc-image-service/internal/config"
	"cgc-image-service/internal/handlers"
	"cgc-image-service/internal/providers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create orchestrator agent
	orchestrator := agents.NewImageOrchestrator()

	// Initialize and register providers
	if err := initializeProviders(orchestrator); err != nil {
		log.Fatalf("Failed to initialize providers: %v", err)
	}

	// Create handlers
	imageHandler := handlers.NewImageHandler(orchestrator)

	// Setup Gin router
	router := setupRouter(imageHandler)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Available endpoints:")
	log.Printf("  POST /generate - Generate images")
	log.Printf("  GET /status - Get provider status")
	log.Printf("  GET /health - Health check")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initializeProviders creates and registers all image providers
func initializeProviders(orchestrator *agents.ImageOrchestrator) error {
	// Create providers
	freepikProvider := providers.NewFreepikProvider()
	googleProvider := providers.NewGoogleImagenProvider()
	leonardoProvider := providers.NewLeonardoAIProvider()

	// Register providers with orchestrator
	if err := orchestrator.RegisterProvider(freepikProvider); err != nil {
		return fmt.Errorf("failed to register Freepik provider: %w", err)
	}

	if err := orchestrator.RegisterProvider(googleProvider); err != nil {
		return fmt.Errorf("failed to register Google Imagen provider: %w", err)
	}

	if err := orchestrator.RegisterProvider(leonardoProvider); err != nil {
		return fmt.Errorf("failed to register Leonardo AI provider: %w", err)
	}

	// Log provider status
	status := orchestrator.GetProviderStatus()
	log.Printf("Initialized %d providers:", len(status))
	for name, s := range status {
		availabilityStatus := "✓ Available"
		if !s.Available {
			availabilityStatus = "✗ Unavailable: " + s.LastError
		}
		log.Printf("  %s: %s", name, availabilityStatus)
	}

	return nil
}

// setupRouter configures the Gin router with all routes and middleware
func setupRouter(imageHandler *handlers.ImageHandler) *gin.Engine {
	// Set Gin mode (can be overridden with GIN_MODE env var)
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", imageHandler.HealthCheck)

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/generate", imageHandler.GenerateImage)
		api.GET("/status", imageHandler.GetProviderStatus)
	}

	return router
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
