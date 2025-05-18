package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
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

func (c *Config) Validate() error {
	if c.NetworkURL == "" {
		return errors.New("network URL is required")
	}
	if len(c.Wallets) == 0 {
		return errors.New("at least one wallet address is required")
	}
	return nil
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
