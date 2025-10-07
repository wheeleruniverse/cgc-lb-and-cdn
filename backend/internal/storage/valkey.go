package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// ValkeyClient wraps the Redis client for vote persistence
type ValkeyClient struct {
	client *redis.Client
}

// Vote represents a user vote
type Vote struct {
	PairID    string    `json:"pair_id"`
	Winner    string    `json:"winner"`
	LeftID    string    `json:"left_id"`
	RightID   string    `json:"right_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ProviderStats represents aggregated statistics for a provider
type ProviderStats struct {
	Provider   string `json:"provider"`
	Wins       int64  `json:"wins"`
	Losses     int64  `json:"losses"`
	TotalVotes int64  `json:"total_votes"`
	WinRate    float64 `json:"win_rate"`
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
		TLSConfig: nil, // Valkey uses TLS by default on DO
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

	// Increment provider stats (winner)
	if err := v.client.HIncrBy(ctx, "provider:wins", vote.Winner, 1).Err(); err != nil {
		return fmt.Errorf("failed to increment wins: %w", err)
	}

	// Increment total votes for both providers
	if err := v.client.HIncrBy(ctx, "provider:total", vote.LeftID, 1).Err(); err != nil {
		return fmt.Errorf("failed to increment total: %w", err)
	}
	if err := v.client.HIncrBy(ctx, "provider:total", vote.RightID, 1).Err(); err != nil {
		return fmt.Errorf("failed to increment total: %w", err)
	}

	// Increment losses for loser
	loser := "left"
	if vote.Winner == "left" {
		loser = "right"
	}
	if err := v.client.HIncrBy(ctx, "provider:losses", loser, 1).Err(); err != nil {
		return fmt.Errorf("failed to increment losses: %w", err)
	}

	return nil
}

// GetProviderStats retrieves statistics for all providers
func (v *ValkeyClient) GetProviderStats(ctx context.Context) (map[string]*ProviderStats, error) {
	wins, err := v.client.HGetAll(ctx, "provider:wins").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get wins: %w", err)
	}

	losses, err := v.client.HGetAll(ctx, "provider:losses").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get losses: %w", err)
	}

	totals, err := v.client.HGetAll(ctx, "provider:total").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get totals: %w", err)
	}

	stats := make(map[string]*ProviderStats)

	// Combine stats from all providers
	allProviders := make(map[string]bool)
	for provider := range wins {
		allProviders[provider] = true
	}
	for provider := range losses {
		allProviders[provider] = true
	}
	for provider := range totals {
		allProviders[provider] = true
	}

	for provider := range allProviders {
		winsCount := int64(0)
		if w, ok := wins[provider]; ok {
			fmt.Sscanf(w, "%d", &winsCount)
		}

		lossesCount := int64(0)
		if l, ok := losses[provider]; ok {
			fmt.Sscanf(l, "%d", &lossesCount)
		}

		totalCount := int64(0)
		if t, ok := totals[provider]; ok {
			fmt.Sscanf(t, "%d", &totalCount)
		}

		winRate := 0.0
		if totalCount > 0 {
			winRate = float64(winsCount) / float64(totalCount) * 100
		}

		stats[provider] = &ProviderStats{
			Provider:   provider,
			Wins:       winsCount,
			Losses:     lossesCount,
			TotalVotes: totalCount,
			WinRate:    winRate,
		}
	}

	return stats, nil
}

// GetTotalVotes returns the total number of votes recorded
func (v *ValkeyClient) GetTotalVotes(ctx context.Context) (int64, error) {
	count, err := v.client.LLen(ctx, "votes:all").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get total votes: %w", err)
	}
	return count, nil
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
