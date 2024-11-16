package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gagliardetto/solana-go/rpc"
)

type Config struct {
    NetworkURL    string        `json:"network_url"`
    Wallets      []string      `json:"wallets"`
    ScanInterval string        `json:"scan_interval"`
    Alerts       AlertConfig   `json:"alerts"`
    Discord      DiscordConfig `json:"discord"`
}

type AlertConfig struct {
    MinimumBalance    uint64  `json:"minimum_balance"`    // Minimum balance to trigger alerts
    SignificantChange float64 `json:"significant_change"` // e.g., 0.20 for 20% change
    IgnoreTokens      []string `json:"ignore_tokens"`     // Tokens to ignore
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

var TestWallets = []string{
    "55kBY9yxqQzj2zxZqRkqENYq6R8PkXmn5GKyQN9YeVFr", // Known test wallet
    "DWuopnuSqYdBhCXqxfqjqzPGibnhkj6SQqFvgC4jkvjF", // Another test wallet
}

// Test configuration
func GetTestConfig() *Config {
    return &Config{
        NetworkURL:    rpc.DevNet_RPC,
        Wallets:      TestWallets,
        ScanInterval: "5s",
        Alerts: AlertConfig{
            MinimumBalance:    1000,
            SignificantChange: 0.05,
            IgnoreTokens:      []string{},
        },
        Discord: DiscordConfig{
            Enabled:    false,
            WebhookURL: "",
            ChannelID:  "",
        },
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
