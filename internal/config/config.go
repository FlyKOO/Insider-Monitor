package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	NetworkURL   string        `json:"network_url"`
	Wallets      []string      `json:"wallets"`
	ScanInterval string        `json:"scan_interval"`
	Alerts       AlertConfig   `json:"alerts"`
	Discord      DiscordConfig `json:"discord"`
	Scan         ScanConfig    `json:"scan"`
}

type AlertConfig struct {
	MinimumBalance    uint64   `json:"minimum_balance"`    // Minimum balance to trigger alerts
	SignificantChange float64  `json:"significant_change"` // e.g., 0.20 for 20% change
	IgnoreTokens      []string `json:"ignore_tokens"`      // Tokens to ignore
}

type ScanConfig struct {
	IncludeTokens []string `json:"include_tokens"` // Specific tokens to include (if empty, include all)
	ExcludeTokens []string `json:"exclude_tokens"` // Specific tokens to exclude
	ScanMode      string   `json:"scan_mode"`      // "all", "whitelist", or "blacklist"
}

type DiscordConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
	ChannelID  string `json:"channel_id"`
}

// Public RPC endpoints that have strict rate limits
var publicRPCEndpoints = []string{
	"https://api.mainnet-beta.solana.com",
	"https://api.devnet.solana.com",
	"https://api.testnet.solana.com",
	"https://solana-api.projectserum.com",
}

// Recommended RPC providers with their benefits
var recommendedRPCProviders = map[string]string{
	"Helius":    "100k requests/day free - https://helius.dev",
	"QuickNode": "30M requests/month free - https://quicknode.com",
	"Triton":    "10M requests/month free - https://triton.one",
	"GenesysGo": "Custom limits - https://genesysgo.com",
}

func (c *Config) Validate() error {
	if c.NetworkURL == "" {
		return fmt.Errorf("network URL is required\n\n" +
			"💡 Recommendation: Get a dedicated RPC endpoint for better performance:\n" +
			"   • Helius: 100k requests/day free - https://helius.dev\n" +
			"   • QuickNode: 30M requests/month free - https://quicknode.com\n" +
			"   • Triton: 10M requests/month free - https://triton.one")
	}

	if len(c.Wallets) == 0 {
		return fmt.Errorf("at least one wallet address is required\n\n" +
			"💡 Add wallet addresses to monitor in the 'wallets' array.\n" +
			"   Example: \"CvQk2xkXtiMj2JqqVx1YZkeSqQ7jyQkNqqjeNE1jPTfc\"")
	}

	// Validate wallet addresses format
	for i, wallet := range c.Wallets {
		if len(wallet) < 32 || len(wallet) > 44 {
			return fmt.Errorf("invalid wallet address format at index %d: %s\n\n"+
				"💡 Solana wallet addresses should be base58 encoded strings, typically 32-44 characters long.\n"+
				"   Example: \"CvQk2xkXtiMj2JqqVx1YZkeSqQ7jyQkNqqjeNE1jPTfc\"", i, wallet)
		}
	}

	// Check if using public RPC endpoint
	c.validateRPCEndpoint()

	return nil
}

// validateRPCEndpoint checks if user is using a public RPC and warns them
func (c *Config) validateRPCEndpoint() {
	isPublicRPC := false
	for _, publicURL := range publicRPCEndpoints {
		if strings.EqualFold(c.NetworkURL, publicURL) {
			isPublicRPC = true
			break
		}
	}

	if isPublicRPC {
		log.Printf("\n⚠️  WARNING: You're using a public RPC endpoint (%s)", c.NetworkURL)
		log.Printf("   Public endpoints have strict rate limits and may cause scanning issues.\n")
		log.Printf("🚀 RECOMMENDATION: Get a dedicated RPC endpoint for better performance:")
		for provider, details := range recommendedRPCProviders {
			log.Printf("   • %s: %s", provider, details)
		}
		log.Printf("\n💡 After getting your RPC endpoint, update 'network_url' in your config.json\n")
	} else {
		log.Printf("✅ Using custom RPC endpoint: %s", c.NetworkURL)
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	log.Printf("Loaded config: NetworkURL=%s, Wallets=%d, ScanInterval=%s",
		cfg.NetworkURL, len(cfg.Wallets), cfg.ScanInterval)

	return &cfg, nil
}
