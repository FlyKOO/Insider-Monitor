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
	MinimumBalance    uint64   `json:"minimum_balance"`    // 触发告警的最小余额
	SignificantChange float64  `json:"significant_change"` // 例如 0.20 表示 20% 变化
	IgnoreTokens      []string `json:"ignore_tokens"`      // 需要忽略的代币
}

type ScanConfig struct {
	IncludeTokens []string `json:"include_tokens"` // 指定需要包含的代币（为空则包含全部）
	ExcludeTokens []string `json:"exclude_tokens"` // 指定需要排除的代币
	ScanMode      string   `json:"scan_mode"`      // "all"、"whitelist" 或 "blacklist"
}

type DiscordConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
	ChannelID  string `json:"channel_id"`
}

// 具有严格速率限制的公共 RPC 端点
var publicRPCEndpoints = []string{
	"https://api.mainnet-beta.solana.com",
	"https://api.devnet.solana.com",
	"https://api.testnet.solana.com",
	"https://solana-api.projectserum.com",
}

// 推荐的 RPC 提供商及其优势
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

	// 校验钱包地址格式
	for i, wallet := range c.Wallets {
		if len(wallet) < 32 || len(wallet) > 44 {
			return fmt.Errorf("invalid wallet address format at index %d: %s\n\n"+
				"💡 Solana wallet addresses should be base58 encoded strings, typically 32-44 characters long.\n"+
				"   Example: \"CvQk2xkXtiMj2JqqVx1YZkeSqQ7jyQkNqqjeNE1jPTfc\"", i, wallet)
		}
	}

	// 检查是否使用公共 RPC 端点
	c.validateRPCEndpoint()

	return nil
}

// validateRPCEndpoint 检查用户是否使用公共 RPC 并给出警告
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

// LoadConfig 从 JSON 文件加载配置
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
