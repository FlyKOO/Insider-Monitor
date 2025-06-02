# Solana Insider Monitor

A tool for monitoring Solana wallet activities, detecting balance changes, and receiving real-time alerts.

## Community

Join our Discord community to:
- Get help with setup and configuration
- Share feedback and suggestions
- Connect with other users
- Stay updated on new features and releases
- Discuss Solana development

👉 [Join the Discord Server](https://discord.gg/7vY9ZBPdya)

## Features

- 🔍 Monitor multiple Solana wallets simultaneously
- 💰 Track token balance changes
- ⚡ Real-time alerts for significant changes
- 🔔 Discord integration for notifications
- 💾 Persistent storage of wallet data
- 🛡️ Graceful handling of network interruptions

---

## ⚠️ Important: RPC Endpoint Setup

**The most common setup issue is using the default public RPC endpoint**, which has strict rate limits and will cause scanning failures. Follow this guide to get a proper RPC endpoint:

### 🚀 Recommended RPC Providers (Free Tiers Available)

| Provider | Free Tier | Speed | Setup |
|----------|-----------|-------|-------|
| **Helius** | 100k requests/day | ⚡⚡⚡ | [Get Free Account](https://helius.dev) |
| **QuickNode** | 30M requests/month | ⚡⚡⚡ | [Get Free Account](https://quicknode.com) |
| **Triton** | 10M requests/month | ⚡⚡ | [Get Free Account](https://triton.one) |
| **GenesysGo** | Custom limits | ⚡⚡ | [Get Account](https://genesysgo.com) |

### ❌ Avoid These (Rate Limited)
```
❌ https://api.mainnet-beta.solana.com (default - gets rate limited)
❌ https://api.devnet.solana.com (only for development)
❌ https://solana-api.projectserum.com (rate limited)
```

### ✅ How to Set Up Your RPC

1. **Sign up** for any provider above (they're free!)
2. **Get your RPC URL** from the dashboard
3. **Update your config.json**:
   ```json
   {
     "network_url": "https://your-custom-rpc-endpoint.com",
     ...
   }
   ```

---

## Quick Start

### Prerequisites

- Go 1.23.2 or later
- **A dedicated Solana RPC endpoint** (see [RPC Setup](#️-important-rpc-endpoint-setup) above - this is crucial!)

### Installation

```bash
# Clone the repository
git clone https://github.com/accursedgalaxy/insider-monitor
cd insider-monitor

# Install dependencies
go mod download
```

### Configuration

1. Copy the example configuration:
```bash
cp config.example.json config.json
```

2. **⚠️ IMPORTANT**: Edit `config.json` and replace the RPC endpoint:
```json
{
    "network_url": "YOUR_DEDICATED_RPC_URL_HERE",
    "wallets": [
        "YOUR_WALLET_ADDRESS_1",
        "YOUR_WALLET_ADDRESS_2"
    ],
    "scan_interval": "1m",
    "alerts": {
        "minimum_balance": 1000,
        "significant_change": 0.20,
        "ignore_tokens": []
    },
    "discord": {
        "enabled": false,
        "webhook_url": "",
        "channel_id": ""
    }
}
```

3. **Get your RPC endpoint** from the [providers listed above](#-recommended-rpc-providers-free-tiers-available) and update `network_url`

### Configuration Options

- `network_url`: **Your dedicated RPC endpoint URL** (see RPC setup section above)
- `wallets`: Array of Solana wallet addresses to monitor
- `scan_interval`: Time between scans (e.g., "30s", "1m", "5m")
- `alerts`:
  - `minimum_balance`: Minimum token balance to trigger alerts
  - `significant_change`: Percentage change to trigger alerts (0.20 = 20%)
  - `ignore_tokens`: Array of token addresses to ignore
- `discord`:
  - `enabled`: Set to true to enable Discord notifications
  - `webhook_url`: Discord webhook URL
  - `channel_id`: Discord channel ID
- `scan`:
  - `scan_mode`: Token scanning mode
    - `"all"`: Monitor all tokens (default)
    - `"whitelist"`: Only monitor tokens in `include_tokens`
    - `"blacklist"`: Monitor all tokens except those in `exclude_tokens`
  - `include_tokens`: Array of token addresses to specifically monitor (used with `whitelist` mode)
  - `exclude_tokens`: Array of token addresses to ignore (used with `blacklist` mode)

### Scan Mode Examples

Here are examples of different scan configurations:

1. Monitor all tokens:
```json
{
    "scan": {
        "scan_mode": "all",
        "include_tokens": [],
        "exclude_tokens": []
    }
}
```

2. Monitor only specific tokens (whitelist):
```json
{
    "scan": {
        "scan_mode": "whitelist",
        "include_tokens": [
            "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",  // USDC
            "So11111111111111111111111111111111111111112"     // SOL
        ],
        "exclude_tokens": []
    }
}
```

3. Monitor all tokens except specific ones (blacklist):
```json
{
    "scan": {
        "scan_mode": "blacklist",
        "include_tokens": [],
        "exclude_tokens": [
            "TokenAddressToIgnore1",
            "TokenAddressToIgnore2"
        ]
    }
}
```

### Running the Monitor

```bash
go run cmd/monitor/main.go
```

#### Custom Config File
```bash
go run cmd/monitor/main.go -config path/to/config.json
```

### Alert Levels

The monitor uses three alert levels based on the configured `significant_change`:
- 🔴 **Critical**: Changes >= 5x the threshold
- 🟡 **Warning**: Changes >= 2x the threshold
- 🟢 **Info**: Changes below 2x the threshold

### Data Storage

The monitor stores wallet data in the `./data` directory to:
- Prevent false alerts after restarts
- Track historical changes
- Handle network interruptions gracefully

### Building from Source

```bash
make build
```

The binary will be available in the `bin` directory.

## 🔧 Troubleshooting

### Common Issues & Solutions

#### ❌ "Rate limit exceeded" / "Too Many Requests" Error
**Problem**: Using the default public RPC endpoint which has strict rate limits
```
❌ Rate limit exceeded after 5 retries
```

**Solution**: 
1. Get a free RPC endpoint from [one of the providers above](#-recommended-rpc-providers-free-tiers-available)
2. Update your `config.json` with the new endpoint:
   ```json
   {
     "network_url": "https://your-custom-rpc-endpoint.com",
     ...
   }
   ```

#### ❌ "Invalid wallet address format" Error
**Problem**: Incorrect wallet address format in config.json
```
❌ invalid wallet address format at index 0: abc123
```

**Solution**: Ensure wallet addresses are valid Solana base58 encoded addresses (32-44 characters)
```json
{
  "wallets": [
    "CvQk2xkXtiMj2JqqVx1YZkeSqQ7jyQkNqqjeNE1jPTfc"  ✅ Valid format
  ]
}
```

#### ❌ "Configuration file not found" Error
**Problem**: config.json doesn't exist
```
❌ Configuration file not found: config.json
```

**Solution**: 
```bash
cp config.example.json config.json
```

#### ❌ "Connection check failed" Error
**Problem**: Network or RPC endpoint issues

**Solution**:
1. Check your internet connection
2. Verify your RPC endpoint URL is correct
3. Try a different RPC provider
4. Test your RPC endpoint manually:
   ```bash
   curl -X POST -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","id":1,"method":"getSlot"}' \
     YOUR_RPC_ENDPOINT_URL
   ```

### Getting Help

If you're still having issues:
1. Check our [Discord community](https://discord.gg/7vY9ZBPdya) for help
2. Review the logs for specific error messages
3. Ensure you have the latest version of the monitor
4. Try the troubleshooting steps above

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
