package agents

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"cgc-lb-and-cdn-backend/internal/models"
)

// ImageOrchestrator implements the OrchestratorAgent interface
type ImageOrchestrator struct {
	name      string
	providers map[string]ImageProvider
	status    map[string]*models.ProviderStatus
	mutex     sync.RWMutex
	random    *rand.Rand
}

// NewImageOrchestrator creates a new orchestrator agent
func NewImageOrchestrator() *ImageOrchestrator {
	return &ImageOrchestrator{
		name:      "ImageOrchestrator",
		providers: make(map[string]ImageProvider),
		status:    make(map[string]*models.ProviderStatus),
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Execute performs image generation with automatic provider selection and fallback
func (o *ImageOrchestrator) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	req, ok := input.(*models.ImageRequest)
	if !ok {
		return nil, fmt.Errorf("invalid input type: expected *models.ImageRequest")
	}

	log.Printf("[ADK] Starting image generation for request: %s", req.RequestID)

	// Select initial provider
	decision, err := o.SelectProvider(ctx, req)
	if err != nil {
		log.Printf("[ADK] Failed to select provider: %v", err)
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}

	log.Printf("[ADK] Selected provider: %s, fallback order: %v", decision.SelectedProvider, decision.FallbackOrder)

	// Try providers in fallback order
	for i, providerName := range decision.FallbackOrder {
		log.Printf("[ADK] Trying provider %d/%d: %s", i+1, len(decision.FallbackOrder), providerName)

		provider, exists := o.providers[providerName]
		if !exists {
			log.Printf("[ADK] Provider %s not found, skipping", providerName)
			continue
		}

		if !provider.IsAvailable() {
			log.Printf("[ADK] Provider %s not available, skipping", providerName)
			continue
		}

		log.Printf("[ADK] Calling provider %s for generation", providerName)
		response, err := provider.Generate(ctx, req)
		if err != nil {
			log.Printf("[ADK] Provider %s failed with error: %v", providerName, err)

			// Handle provider error and check if we should continue
			providerErr := provider.HandleError(err)
			o.updateProviderStatus(providerName, providerErr)

			log.Printf("[ADK] Provider %s error details - Quota: %t, Rate Limited: %t, Retryable: %t",
				providerName, providerErr.IsQuotaHit, providerErr.IsRateLimit, providerErr.Retryable)

			// If this was the last provider, return the error
			if providerName == decision.FallbackOrder[len(decision.FallbackOrder)-1] {
				log.Printf("[ADK] All providers exhausted, returning final error from %s", providerName)
				return nil, fmt.Errorf("all providers failed, last error from %s: %w", providerName, err)
			}

			// Continue to next provider
			log.Printf("[ADK] Falling back to next provider in list")
			continue
		}

		// Success! Update provider status
		log.Printf("[ADK] Provider %s succeeded, generated %d images", providerName, len(response.Images))
		o.updateProviderSuccessStatus(providerName)
		return response, nil
	}

	log.Printf("[ADK] No available providers found")
	return nil, fmt.Errorf("no available providers")
}

// SelectProvider chooses the best provider using random selection with availability filtering
func (o *ImageOrchestrator) SelectProvider(ctx context.Context, req *models.ImageRequest) (*models.AgentDecision, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Get all available providers
	availableProviders := make([]string, 0)
	for name, provider := range o.providers {
		if provider.IsAvailable() {
			availableProviders = append(availableProviders, name)
		}
	}

	if len(availableProviders) == 0 {
		return nil, fmt.Errorf("no available providers")
	}

	// Shuffle for random selection - treat all providers equally until errors occur
	o.random.Shuffle(len(availableProviders), func(i, j int) {
		availableProviders[i], availableProviders[j] = availableProviders[j], availableProviders[i]
	})

	log.Printf("[INFO] Selecting providers: %s", strings.Join(availableProviders, ", "))

	return &models.AgentDecision{
		SelectedProvider: availableProviders[0],
		Reasoning:        "Random selection from available providers",
		FallbackOrder:    availableProviders,
		Confidence:       1.0,
		Metadata: map[string]string{
			"selection_method": "random",
			"total_available":  fmt.Sprintf("%d", len(availableProviders)),
		},
	}, nil
}

// HandleProviderFailure manages fallback when a provider fails
func (o *ImageOrchestrator) HandleProviderFailure(ctx context.Context, provider string, err *models.ProviderError, req *models.ImageRequest) (*models.AgentDecision, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Update provider status
	if status, exists := o.status[provider]; exists {
		status.Available = !err.IsQuotaHit && !err.IsRateLimit
		status.LastError = err.Message
		status.ErrorCount++
		status.QuotaHit = err.IsQuotaHit
		status.RateLimited = err.IsRateLimit
	}

	// Get remaining available providers
	availableProviders := make([]string, 0)
	for name, prov := range o.providers {
		if name != provider && prov.IsAvailable() {
			availableProviders = append(availableProviders, name)
		}
	}

	if len(availableProviders) == 0 {
		return nil, fmt.Errorf("no fallback providers available")
	}

	// Shuffle remaining providers
	o.random.Shuffle(len(availableProviders), func(i, j int) {
		availableProviders[i], availableProviders[j] = availableProviders[j], availableProviders[i]
	})

	return &models.AgentDecision{
		SelectedProvider: availableProviders[0],
		Reasoning:        fmt.Sprintf("Fallback due to %s failure: %s", provider, err.Message),
		FallbackOrder:    availableProviders,
		Confidence:       0.8,
		Metadata: map[string]string{
			"failed_provider":  provider,
			"failure_reason":   err.Message,
			"selection_method": "fallback_random",
		},
	}, nil
}

// RegisterProvider adds a new provider to the orchestrator
func (o *ImageOrchestrator) RegisterProvider(provider ImageProvider) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	name := provider.GetName()
	o.providers[name] = provider
	o.status[name] = &models.ProviderStatus{
		Name:        name,
		Available:   true,
		LastSuccess: time.Now(),
		ErrorCount:  0,
		QuotaHit:    false,
		RateLimited: false,
	}

	return nil
}

// GetProviderStatus returns status of all registered providers
func (o *ImageOrchestrator) GetProviderStatus() map[string]*models.ProviderStatus {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Return a copy to prevent external modification
	statusCopy := make(map[string]*models.ProviderStatus)
	for name, status := range o.status {
		statusCopy[name] = &models.ProviderStatus{
			Name:        status.Name,
			Available:   status.Available,
			LastError:   status.LastError,
			LastSuccess: status.LastSuccess,
			ErrorCount:  status.ErrorCount,
			QuotaHit:    status.QuotaHit,
			RateLimited: status.RateLimited,
		}
	}

	return statusCopy
}

// GetProvider returns a specific provider by name
func (o *ImageOrchestrator) GetProvider(name string) (ImageProvider, bool) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	provider, exists := o.providers[name]
	return provider, exists
}

// GetName returns the agent's name
func (o *ImageOrchestrator) GetName() string {
	return o.name
}

// GetCapabilities returns what this agent can do
func (o *ImageOrchestrator) GetCapabilities() []string {
	return []string{
		"provider_selection",
		"automatic_fallback",
		"quota_management",
		"random_load_balancing",
	}
}

// updateProviderStatus updates status after an error
func (o *ImageOrchestrator) updateProviderStatus(providerName string, err *models.ProviderError) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if status, exists := o.status[providerName]; exists {
		status.Available = !err.IsQuotaHit && !err.IsRateLimit
		status.LastError = err.Message
		status.ErrorCount++
		status.QuotaHit = err.IsQuotaHit
		status.RateLimited = err.IsRateLimit
	}
}

// updateProviderSuccessStatus updates status after a successful generation
func (o *ImageOrchestrator) updateProviderSuccessStatus(providerName string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if status, exists := o.status[providerName]; exists {
		status.Available = true
		status.LastSuccess = time.Now()
		status.LastError = ""
		// Don't reset error count completely, but reduce it
		if status.ErrorCount > 0 {
			status.ErrorCount--
		}
		status.QuotaHit = false
		status.RateLimited = false
	}
}
