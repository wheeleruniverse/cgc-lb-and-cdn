package storage

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// ValkeyClient wraps the Redis client for vote persistence
type ValkeyClient struct {
	client *redis.Client
}

// Vote represents a user vote
// Simplified structure: pair-id is the atomic unit, both images share the same provider
type Vote struct {
	PairID    string    `json:"pair_id"`
	Winner    string    `json:"winner"`   // "left" or "right"
	Provider  string    `json:"provider"` // The provider that generated this pair
	Prompt    string    `json:"prompt"`
	Timestamp time.Time `json:"timestamp"`
}

// ImagePair represents a pair of images generated from the same prompt by the same provider
// New simplified structure: uses pair-id as the primary identifier
// Images are stored in Spaces at: images/<provider>/<pair-id>/<side>.png
// Both images are from the same provider for apples-to-apples comparison
type ImagePair struct {
	PairID    string    `json:"pair_id"`
	Prompt    string    `json:"prompt"`
	Provider  string    `json:"provider"`  // Single provider for both images
	LeftURL   string    `json:"left_url"`  // CDN URL for left image
	RightURL  string    `json:"right_url"` // CDN URL for right image
	Timestamp time.Time `json:"timestamp"`
}

// NewValkeyClient creates a new Valkey client
func NewValkeyClient() (*ValkeyClient, error) {
	host := os.Getenv("DO_VALKEY_HOST")
	port := os.Getenv("DO_VALKEY_PORT")
	password := os.Getenv("DO_VALKEY_PASSWORD")

	if host == "" || port == "" {
		return nil, fmt.Errorf("valkey configuration missing (DO_VALKEY_HOST or DO_VALKEY_PORT)")
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Valkey: %w", err)
	}

	return &ValkeyClient{client: client}, nil
}

// RecordVote stores a vote in Valkey
func (v *ValkeyClient) RecordVote(ctx context.Context, vote *Vote) error {
	vote.Timestamp = time.Now()

	// Store vote with unique key
	voteKey := fmt.Sprintf("vote:%s", vote.PairID)
	voteJSON, err := json.Marshal(vote)
	if err != nil {
		return fmt.Errorf("failed to marshal vote: %w", err)
	}

	// Store vote with 30-day expiration
	if err := v.client.Set(ctx, voteKey, voteJSON, 30*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store vote: %w", err)
	}

	// Add to votes list for analytics
	if err := v.client.LPush(ctx, "votes:all", voteJSON).Err(); err != nil {
		return fmt.Errorf("failed to add to votes list: %w", err)
	}

	// Trim votes list to last 10,000 votes
	if err := v.client.LTrim(ctx, "votes:all", 0, 9999).Err(); err != nil {
		return fmt.Errorf("failed to trim votes list: %w", err)
	}

	// Track side preference (which side users tend to choose overall)
	// This is useful for detecting position bias
	if err := v.client.HIncrBy(ctx, "side:wins", vote.Winner, 1).Err(); err != nil {
		return fmt.Errorf("failed to increment side wins: %w", err)
	}

	return nil
}

// GetTotalVotes returns the total number of votes recorded
func (v *ValkeyClient) GetTotalVotes(ctx context.Context) (int64, error) {
	count, err := v.client.LLen(ctx, "votes:all").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get total votes: %w", err)
	}
	return count, nil
}

// GetSideWins returns the vote counts for left and right sides
func (v *ValkeyClient) GetSideWins(ctx context.Context) (map[string]int64, error) {
	sideWins, err := v.client.HGetAll(ctx, "side:wins").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get side wins: %w", err)
	}

	result := map[string]int64{
		"left":  0,
		"right": 0,
	}

	for side, countStr := range sideWins {
		var count int64
		fmt.Sscanf(countStr, "%d", &count)
		result[side] = count
	}

	return result, nil
}

// GetRecentVotes retrieves the most recent votes
func (v *ValkeyClient) GetRecentVotes(ctx context.Context, limit int64) ([]*Vote, error) {
	voteStrings, err := v.client.LRange(ctx, "votes:all", 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent votes: %w", err)
	}

	votes := make([]*Vote, 0, len(voteStrings))
	for _, voteStr := range voteStrings {
		var vote Vote
		if err := json.Unmarshal([]byte(voteStr), &vote); err != nil {
			continue // Skip malformed votes
		}
		votes = append(votes, &vote)
	}

	return votes, nil
}

// Close closes the Valkey client connection
func (v *ValkeyClient) Close() error {
	return v.client.Close()
}

// StoreImagePair stores an image pair in Valkey
func (v *ValkeyClient) StoreImagePair(ctx context.Context, pair *ImagePair) error {
	// Store pair with unique key
	pairKey := fmt.Sprintf("pair:%s", pair.PairID)
	pairJSON, err := json.Marshal(pair)
	if err != nil {
		return fmt.Errorf("failed to marshal pair: %w", err)
	}

	// Store pair (no expiration - we want to keep all pairs)
	if err := v.client.Set(ctx, pairKey, pairJSON, 0).Err(); err != nil {
		return fmt.Errorf("failed to store pair: %w", err)
	}

	// Add pair ID to the list of all pairs for random selection
	if err := v.client.LPush(ctx, "pairs:all", pair.PairID).Err(); err != nil {
		return fmt.Errorf("failed to add to pairs list: %w", err)
	}

	return nil
}

// GetImagePairByID retrieves a specific image pair by its ID
func (v *ValkeyClient) GetImagePairByID(ctx context.Context, pairID string) (*ImagePair, error) {
	pairKey := fmt.Sprintf("pair:%s", pairID)
	pairJSON, err := v.client.Get(ctx, pairKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("pair not found: %s", pairID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pair: %w", err)
	}

	var pair ImagePair
	if err := json.Unmarshal([]byte(pairJSON), &pair); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pair: %w", err)
	}

	return &pair, nil
}

// GetRandomImagePair retrieves a random image pair from Valkey
// excludedPairIDs allows filtering out already-voted pairs
func (v *ValkeyClient) GetRandomImagePair(ctx context.Context, excludedPairIDs []string) (*ImagePair, error) {
	// Get all pair IDs
	allPairIDs, err := v.client.LRange(ctx, "pairs:all", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pairs list: %w", err)
	}

	if len(allPairIDs) == 0 {
		return nil, fmt.Errorf("no pairs available")
	}

	// Filter out excluded pairs
	excludedMap := make(map[string]bool)
	for _, id := range excludedPairIDs {
		excludedMap[id] = true
	}

	availablePairs := make([]string, 0, len(allPairIDs))
	for _, pairID := range allPairIDs {
		if !excludedMap[pairID] {
			availablePairs = append(availablePairs, pairID)
		}
	}

	if len(availablePairs) == 0 {
		return nil, fmt.Errorf("no unvoted pairs available")
	}

	// Get random pair ID from available pairs
	randomIndex := rand.Intn(len(availablePairs))
	pairID := availablePairs[randomIndex]

	// Retrieve the pair
	pairKey := fmt.Sprintf("pair:%s", pairID)
	pairJSON, err := v.client.Get(ctx, pairKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("pair not found: %s", pairID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pair: %w", err)
	}

	var pair ImagePair
	if err := json.Unmarshal([]byte(pairJSON), &pair); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pair: %w", err)
	}

	return &pair, nil
}

// WinningImagePair extends ImagePair with vote count information
type WinningImagePair struct {
	ImagePair
	VoteCount int64 `json:"vote_count"` // Number of votes this pair won with
}

// GetWinningImages retrieves all images that won their battles for the specified side
// Returns pairs sorted by vote count (descending)
// side parameter should be "left" or "right"
func (v *ValkeyClient) GetWinningImages(ctx context.Context, side string) ([]WinningImagePair, error) {
	// Validate side parameter
	if side != "left" && side != "right" {
		return nil, fmt.Errorf("invalid side parameter: must be 'left' or 'right'")
	}

	// Get all votes
	voteStrings, err := v.client.LRange(ctx, "votes:all", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get votes: %w", err)
	}

	// Track winning images by pair ID with vote counts
	winningPairVotes := make(map[string]int64)
	for _, voteStr := range voteStrings {
		var vote Vote
		if err := json.Unmarshal([]byte(voteStr), &vote); err != nil {
			continue // Skip malformed votes
		}

		// Only include winners for the specified side
		if vote.Winner == side {
			winningPairVotes[vote.PairID]++
		}
	}

	// Retrieve the winning pairs with vote counts
	var winningPairs []WinningImagePair
	for pairID, voteCount := range winningPairVotes {
		pairKey := fmt.Sprintf("pair:%s", pairID)
		pairJSON, err := v.client.Get(ctx, pairKey).Result()
		if err == redis.Nil {
			continue // Pair no longer exists
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get pair %s: %w", pairID, err)
		}

		var pair ImagePair
		if err := json.Unmarshal([]byte(pairJSON), &pair); err != nil {
			continue // Skip malformed pairs
		}

		winningPairs = append(winningPairs, WinningImagePair{
			ImagePair: pair,
			VoteCount: voteCount,
		})
	}

	// Sort by vote count descending (most votes first)
	for i := 0; i < len(winningPairs)-1; i++ {
		for j := i + 1; j < len(winningPairs); j++ {
			if winningPairs[j].VoteCount > winningPairs[i].VoteCount {
				winningPairs[i], winningPairs[j] = winningPairs[j], winningPairs[i]
			}
		}
	}

	return winningPairs, nil
}
