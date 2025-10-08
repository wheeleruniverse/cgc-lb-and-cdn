package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"cgc-lb-and-cdn-backend/internal/agents"
	"cgc-lb-and-cdn-backend/internal/models"
	"cgc-lb-and-cdn-backend/internal/storage"
	"cgc-lb-and-cdn-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// prompts contains 100 unique image generation prompts
var prompts = []string{
	// Animals with Jobs
	"A koala wearing a tiny firefighter's helmet, climbing a ladder to rescue a cat from a tree.",
	"An elegant giraffe working as a professional violinist in a concert hall.",
	"A team of squirrels in construction vests, building a miniature skyscraper out of acorns.",
	"A hamster dressed as a mad scientist, running on a wheel that powers a small laser.",
	"A chameleon wearing a detective trench coat, blending into a cluttered bookshelf.",
	"A group of penguins in suits, presenting a quarterly report in a chilly boardroom.",
	"An octopus barista, expertly making lattes with eight arms at a bustling coffee shop.",
	"A wise owl in a professor's cap and gown, teaching a class of baby birds.",
	"A majestic lion working as a librarian, quietly shelving books with a stern but fair expression.",
	"A golden retriever wearing a hard hat and safety goggles, inspecting a construction site.",

	// Fantasy and Mythical
	"A friendly dragon, meticulously tending a garden of glowing, fantastical flowers.",
	"A whimsical gnome architect, designing a house carved from a giant mushroom.",
	"A griffin delivering mail to a tiny floating village in the sky.",
	"An elegant fairy librarian, organizing a library of books with pages made of autumn leaves.",
	"A family of yetis having a picnic on a snowy mountain peak.",
	"A benevolent kraken playing chess against a tiny sailboat on a calm sea.",
	"A unicorn in an enchanted forest, serving tea to woodland creatures.",
	"A phoenix made of flowing molten glass, taking flight from a volcanic crater.",
	"A mischievous satyr playing a pan flute that makes flowers instantly bloom.",
	"A wise wizard using a sparkling wand to bake a cake for a child's birthday.",

	// Sci-Fi and Futuristic
	"A retro-futuristic robot, serving a cup of coffee at a space diner.",
	"A bustling city where all the buildings are giant, glowing crystals.",
	"A friendly alien tourist taking a selfie in front of the Eiffel Tower.",
	"An astronaut in a classic spacesuit, fishing on a distant, peaceful planet.",
	"A hovercraft shaped like a giant loaf of bread, delivering sandwiches.",
	"A futuristic food truck selling \"stardust tacos\" in a neon-lit alleyway.",
	"A cyborg with a heart of gold, building a birdhouse in a lush garden.",
	"A family of robots on a road trip through a galaxy of colorful gas clouds.",
	"A high-tech space port where ships are docked like planes at an airport.",
	"A giant robot, holding a sign that says \"Please Recycle.\"",

	// Nature and Outdoors
	"A friendly-looking squirrel riding a unicycle on a path through an autumn forest.",
	"A family of turtles enjoying a leisurely boat ride on a lily-pad pond.",
	"A whimsical treehouse with a spiral staircase and glowing lanterns.",
	"A vibrant field of sunflowers that turn to face the sun in a synchronized dance.",
	"A calm river flowing through a canyon made of oversized, colorful geodes.",
	"A curious fox peeking out from behind a vibrant, glowing waterfall.",
	"A bustling beehive that looks like a miniature, bustling city.",
	"A peaceful cottage nestled among giant, cloud-like lavender bushes.",
	"A garden where all the plants are made of different types of candy.",
	"A majestic whale with a glowing constellation pattern on its back, swimming in a starry ocean.",

	// Objects with Personality
	"A grumpy old toaster, trying to make the perfect toast.",
	"A friendly, smiling cloud wearing a top hat and a monocle.",
	"A vintage camera with a single, expressive eye, capturing a happy moment.",
	"A pencil and eraser, walking hand-in-hand down a winding road of a sketchbook.",
	"A happy, bouncing red ball, leaving a trail of rainbows.",
	"A wise old teacup, sitting on a shelf, with a small steam cloud that tells stories.",
	"A pair of mismatched socks, finally reunited after a long journey.",
	"A stack of books, happily celebrating the first day of school.",
	"A set of garden tools having a friendly conversation in a shed.",
	"A tiny, glowing lightbulb having a brilliant idea.",

	// Food and Drink
	"A sushi chef, meticulously preparing a plate of sushi on a tiny, detailed stage.",
	"A smiling ice cream cone, melting happily in the summer sun.",
	"A family of pastries, having a tea party in a whimsical kitchen.",
	"A friendly bowl of ramen, with noodles that look like tiny, smiling worms.",
	"A happy, bubbly soda can, playing a video game.",
	"A slice of pizza, wearing a tiny superhero cape, ready to save the day.",
	"A group of vegetables, forming a band and playing instruments made of kitchen utensils.",
	"A cheerful cup of hot chocolate, with marshmallows that look like fluffy clouds.",
	"A tiny, adventurous strawberry, scaling a mountain of whipped cream.",
	"A taco, dressed as a detective, investigating a case of missing salsa.",

	// Transportation and Vehicles
	"A hot air balloon shaped like a giant ice cream sundae, floating over a city.",
	"A whimsical train with a teapot for a boiler, traveling through a teacup landscape.",
	"A tiny submarine, exploring a beautiful coral reef made of gemstones.",
	"A friendly, old-fashioned bicycle, with a flower basket full of sunshine.",
	"A spaceship shaped like a rubber duck, flying through a starry, cosmic bath.",
	"A vintage car with a garden growing in its trunk.",
	"A cheerful sailboat with a sail made of patchwork quilts.",
	"A hot dog vendor cart, being pulled by a team of tiny, happy sausages.",
	"A cheerful, red fire truck with a hose that sprays confetti.",
	"A sleek, futuristic racing car, driving on a track made of light.",

	// Abstract and Surreal
	"A landscape where the sky is a swirling vortex of vibrant, pastel colors.",
	"A whimsical clock with hands that point to feelings instead of hours.",
	"A staircase that leads to a door opening into a sky full of fish.",
	"A single, glowing feather, floating in a room filled with giant, sparkling bubbles.",
	"A tree with roots that are also the branches, creating a perfect circle.",
	"A serene lake that reflects a different, fantastical world.",
	"A quiet room where all the furniture is made of different clouds.",
	"A majestic mountain range made of neatly folded blankets.",
	"A bookshelf where the books are filled with liquid light.",
	"A city skyline where buildings are made of giant, interlocking gears.",

	// Sports and Hobbies
	"A group of teacups, playing a game of miniature golf.",
	"A family of teddy bears, having a grand picnic and playing frisbee.",
	"A happy, colorful robot, painting a masterpiece on an oversized canvas.",
	"A trio of cats, expertly playing an intense game of chess.",
	"A cheerful, bouncing basketball, practicing its free throws.",
	"A group of friendly monsters, having a dance-off in a disco.",
	"A tiny, adventurous snail, hiking up a giant mountain.",
	"A family of garden gnomes, having a friendly race on their tricycles.",
	"A smiling, happy sun, playing hide-and-seek with the moon.",
	"A friendly ghost, learning to play the guitar.",

	// Everyday Life with a Twist
	"A busy city street where the cars are tiny, flying hot dogs.",
	"A serene park bench where a pigeon and a squirrel are reading a newspaper together.",
	"A cozy living room where a dog and a cat are sharing popcorn and watching a movie.",
	"A bustling laundromat where the washing machines are giant, smiling fishbowls.",
	"A family of socks, hanging out on a clothesline and telling jokes.",
	"A happy, bubbling bathtub, full of bubbles shaped like stars.",
	"A quiet library where the books float down to you on a magical breeze.",
	"A busy office where all the computers are powered by tiny, industrious gnomes.",
	"A peaceful night sky where the stars are actually tiny, glowing origami stars.",
	"A sunny day at the beach, where the sandcastles are made of colorful jelly.",
}

// ImageHandler handles image generation requests
type ImageHandler struct {
	orchestrator agents.OrchestratorAgent
	valkeyClient *storage.ValkeyClient
}

// NewImageHandler creates a new image handler
func NewImageHandler(orchestrator agents.OrchestratorAgent, valkeyClient *storage.ValkeyClient) *ImageHandler {
	return &ImageHandler{
		orchestrator: orchestrator,
		valkeyClient: valkeyClient,
	}
}

// getRandomPrompt returns a random prompt from the prompts list
func getRandomPrompt() string {
	return prompts[rand.Intn(len(prompts))]
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
	requestID := uuid.New().String()
	pairID := uuid.New().String()
	req.RequestID = requestID
	req.PairID = pairID
	req.Timestamp = time.Now()

	// Use random prompt
	req.Prompt = getRandomPrompt()
	fmt.Printf("[INFO] Using random prompt: %s, pair_id: %s\n", req.Prompt, pairID)

	// Generate image pair (2 images: left and right)
	result, err := h.orchestrator.Execute(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Image generation failed", "GENERATION_FAILED", map[string]string{
			"error": err.Error(),
		})
		return
	}

	response, ok := result.(*models.ImageResponse)
	if !ok || len(response.Images) < 2 {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid response - need 2 images", "INVALID_RESPONSE", nil)
		return
	}

	// First image is "left" (index 0), second is "right" (index 1)
	leftImage := response.Images[0]
	rightImage := response.Images[1]
	timestamp := time.Now()

	// Store the pair in Valkey (simplified structure with pair-id only)
	if h.valkeyClient != nil {
		pair := &storage.ImagePair{
			PairID:    pairID,
			Prompt:    req.Prompt,
			Provider:  response.Provider,
			LeftURL:   leftImage.URL,
			RightURL:  rightImage.URL,
			Timestamp: timestamp,
		}

		if err := h.valkeyClient.StoreImagePair(c.Request.Context(), pair); err != nil {
			fmt.Printf("[ERROR] Failed to store image pair: %v\n", err)
			// Continue anyway - don't fail the request
		} else {
			fmt.Printf("[PAIR] Stored in Valkey - Pair: %s, Prompt: %s, Provider: %s\n", pairID, req.Prompt, response.Provider)
		}
	}

	// Return success response with both images
	utils.RespondWithSuccess(c, gin.H{
		"pair_id":     pairID,
		"prompt":      req.Prompt,
		"provider":    response.Provider,
		"left_image":  leftImage,
		"right_image": rightImage,
		"timestamp":   timestamp.Format(time.RFC3339),
	}, "Image pair generated successfully", map[string]string{
		"pair_id":    pairID,
		"request_id": requestID,
		"provider":   response.Provider,
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
// Supports optional "exclude" query parameter with comma-separated pair IDs
func (h *ImageHandler) GetImagePair(c *gin.Context) {
	if h.valkeyClient == nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "Image pairs unavailable", "VALKEY_UNAVAILABLE", nil)
		return
	}

	// Parse excluded pair IDs from query parameter
	excludedPairIDs := []string{}
	if excludeParam := c.Query("exclude"); excludeParam != "" {
		excludedPairIDs = strings.Split(excludeParam, ",")
		// Trim whitespace from each ID
		for i, id := range excludedPairIDs {
			excludedPairIDs[i] = strings.TrimSpace(id)
		}
	}

	// Get random pair from Valkey, excluding already-voted pairs
	pair, err := h.valkeyClient.GetRandomImagePair(c.Request.Context(), excludedPairIDs)
	if err != nil {
		// Check if it's an empty database (no pairs available yet)
		if strings.Contains(err.Error(), "no pairs available") {
			utils.RespondWithError(c, http.StatusNotFound, "No image pairs available yet. Images are being generated in the background - please check back in a few moments!", "NO_PAIRS_YET", map[string]string{
				"suggestion": "Try generating a new pair or wait for automatic generation",
			})
			return
		}

		// Check if all pairs have been voted on
		if strings.Contains(err.Error(), "no unvoted pairs available") {
			utils.RespondWithError(c, http.StatusNotFound, "You've voted on all available pairs! Great job! 🎉", "ALL_PAIRS_VOTED", map[string]string{
				"suggestion": "Check back later for new images",
			})
			return
		}

		// Other errors
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get random image pair", "PAIR_UNAVAILABLE", map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Simplified response: no duplicate data
	response := models.ImagePairResponse{
		PairID:   pair.PairID,
		Prompt:   pair.Prompt,
		Provider: pair.Provider,
		LeftURL:  pair.LeftURL,
		RightURL: pair.RightURL,
	}

	utils.RespondWithSuccess(c, response, "Image pair retrieved successfully", nil)
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
		// Fetch the image pair to get provider and prompt information
		pair, err := h.valkeyClient.GetImagePairByID(c.Request.Context(), req.PairID)
		if err != nil {
			// Pair not found - still record the vote but without provider/prompt
			fmt.Printf("[WARN] Could not fetch pair %s for vote metadata: %v\n", req.PairID, err)
		}

		vote := &storage.Vote{
			PairID:   req.PairID,
			Winner:   req.Winner,
			Provider: "",
			Prompt:   "",
		}

		// Add provider and prompt if we found the pair
		if pair != nil {
			vote.Provider = pair.Provider
			vote.Prompt = pair.Prompt
		}

		if err := h.valkeyClient.RecordVote(c.Request.Context(), vote); err != nil {
			fmt.Printf("[ERROR] Failed to record vote in Valkey: %v\n", err)
			// Continue anyway - don't fail the request if Valkey is down
		} else {
			fmt.Printf("[VOTE] Recorded in Valkey - Pair: %s, Winner: %s, Provider: %s\n", req.PairID, req.Winner, vote.Provider)
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

// GetWinners handles GET /images/winners requests
func (h *ImageHandler) GetWinners(c *gin.Context) {
	if h.valkeyClient == nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "Winners unavailable", "VALKEY_UNAVAILABLE", nil)
		return
	}

	// Get side parameter (default to "left")
	side := c.DefaultQuery("side", "left")
	if side != "left" && side != "right" {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid side parameter", "INVALID_SIDE", map[string]string{
			"side":    side,
			"allowed": "left, right",
		})
		return
	}

	winningPairs, err := h.valkeyClient.GetWinningImages(c.Request.Context(), side)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get winners", "WINNERS_ERROR", map[string]string{
			"error": err.Error(),
			"side":  side,
		})
		return
	}

	// Transform to response format
	type WinnerImage struct {
		ImageURL  string `json:"image_url"`
		Prompt    string `json:"prompt"`
		Provider  string `json:"provider"`
		PairID    string `json:"pair_id"`
		Timestamp string `json:"timestamp"`
	}

	var winners []WinnerImage
	for _, pair := range winningPairs {
		imageURL := pair.LeftURL
		if side == "right" {
			imageURL = pair.RightURL
		}

		winners = append(winners, WinnerImage{
			ImageURL:  imageURL,
			Prompt:    pair.Prompt,
			Provider:  pair.Provider,
			PairID:    pair.PairID,
			Timestamp: pair.Timestamp.Format(time.RFC3339),
		})
	}

	utils.RespondWithSuccess(c, gin.H{
		"winners":   winners,
		"count":     len(winners),
		"side":      side,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, fmt.Sprintf("%s winners retrieved successfully", strings.Title(side)), nil)
}
