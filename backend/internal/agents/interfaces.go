package agents

import (
	"cgc-lb-and-cdn-backend/internal/models"
	"context"
)

// ImageProvider defines the interface that all image generation providers must implement
type ImageProvider interface {
	// Generate creates images based on the request
	Generate(ctx context.Context, req *models.ImageRequest) (*models.ImageResponse, error)

	// GetStatus returns the current status of this provider
	GetStatus() *models.ProviderStatus

	// GetName returns the name of this provider
	GetName() string

	// IsAvailable checks if the provider is currently available
	IsAvailable() bool

	// HandleError processes an error and updates provider status
	HandleError(err error) *models.ProviderError

	// RefreshQuota updates quota information from the provider's API
	RefreshQuota(ctx context.Context) error
}

// Agent represents a generic agent in the ADK framework
type Agent interface {
	// Execute performs the agent's primary function
	Execute(ctx context.Context, input interface{}) (interface{}, error)

	// GetName returns the agent's name
	GetName() string

	// GetCapabilities returns what this agent can do
	GetCapabilities() []string
}

// OrchestratorAgent manages provider selection and fallback logic
type OrchestratorAgent interface {
	Agent

	// SelectProvider chooses the best provider for a request
	SelectProvider(ctx context.Context, req *models.ImageRequest) (*models.AgentDecision, error)

	// HandleProviderFailure manages fallback when a provider fails
	HandleProviderFailure(ctx context.Context, provider string, err *models.ProviderError, req *models.ImageRequest) (*models.AgentDecision, error)

	// RegisterProvider adds a new provider to the orchestrator
	RegisterProvider(provider ImageProvider) error

	// GetProviderStatus returns status of all registered providers
	GetProviderStatus() map[string]*models.ProviderStatus

	// GetProvider returns a specific provider by name
	GetProvider(name string) (ImageProvider, bool)
}

// ProviderAgent wraps an image provider with agent capabilities
type ProviderAgent interface {
	Agent
	ImageProvider

	// GetProvider returns the underlying provider
	GetProvider() ImageProvider
}
