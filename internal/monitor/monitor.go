package monitor

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/accursedgalaxy/insider-monitor/internal/config"
	"github.com/accursedgalaxy/insider-monitor/internal/price"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

type WalletMonitor struct {
	client       *rpc.Client
	wallets      []solana.PublicKey
	networkURL   string
	isConnected  bool
	scanConfig   *config.ScanConfig
	priceService *price.JupiterPrice
}

func NewWalletMonitor(networkURL string, wallets []string, scanConfig *config.ScanConfig) (*WalletMonitor, error) {
	client := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
		networkURL,
		4,
		1,
	))

	// 将钱包地址转换为 PublicKey
	pubKeys := make([]solana.PublicKey, len(wallets))
	for i, addr := range wallets {
		pubKey, err := solana.PublicKeyFromBase58(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid wallet address %s: %v", addr, err)
		}
		pubKeys[i] = pubKey
	}

	return &WalletMonitor{
		client:       client,
		wallets:      pubKeys,
		networkURL:   networkURL,
		scanConfig:   scanConfig,
		priceService: price.NewJupiterPrice(),
	}, nil
}

// 简化的 TokenAccountInfo
type TokenAccountInfo struct {
	Balance         uint64    `json:"balance"`
	LastUpdated     time.Time `json:"last_updated"`
	Symbol          string    `json:"symbol"`
	Decimals        uint8     `json:"decimals"`
	USDPrice        float64   `json:"usd_price"`
	USDValue        float64   `json:"usd_value"`
	ConfidenceLevel string    `json:"confidence_level"`
}

// 简化的 WalletData
type WalletData struct {
	WalletAddress string                      `json:"wallet_address"`
	TokenAccounts map[string]TokenAccountInfo `json:"token_accounts"` // mint -> 信息
	LastScanned   time.Time                   `json:"last_scanned"`
}

// 以下常量用于重试配置
const (
	maxRetries     = 5
	initialBackoff = 5 * time.Second
	maxBackoff     = 30 * time.Second
)

func (w *WalletMonitor) getTokenAccountsWithRetry(wallet solana.PublicKey) (*rpc.GetTokenAccountsResult, error) {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		accounts, err := w.client.GetTokenAccountsByOwner(
			context.Background(),
			wallet,
			&rpc.GetTokenAccountsConfig{
				ProgramId: solana.TokenProgramID.ToPointer(),
			},
			&rpc.GetTokenAccountsOpts{
				Encoding: solana.EncodingBase64,
			},
		)

		if err == nil {
			return accounts, nil
		}

		lastErr = err
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too Many Requests") {
			log.Printf("⚠️  Rate limited on attempt %d for wallet %s, waiting %v before retry",
				attempt+1, wallet.String(), backoff)

			// 在首次触发速率限制时显示提示信息
			if attempt == 0 {
				log.Printf("💡 Rate limit detected. This usually happens when using public RPC endpoints.")
				log.Printf("   Consider upgrading to a dedicated RPC provider:")
				log.Printf("   • Helius: 100k requests/day free - https://helius.dev")
				log.Printf("   • QuickNode: 30M requests/month free - https://quicknode.com")
				log.Printf("   • Triton: 10M requests/month free - https://triton.one")
			}

			time.Sleep(backoff)

			// 指数回退并设置最大值
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// 处理其他常见错误并提供提示
		if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "timeout") {
			return nil, fmt.Errorf("connection error: %w\n\n"+
				"💡 This might be due to:\n"+
				"   • Network connectivity issues\n"+
				"   • RPC endpoint is down or overloaded\n"+
				"   • Try a different RPC provider from the list above", err)
		}

		// 若不是速率限制或连接错误，则立即返回
		return nil, fmt.Errorf("RPC request failed: %w\n\n"+
			"💡 If this error persists, try:\n"+
			"   • Check your RPC endpoint URL in config.json\n"+
			"   • Verify your network connection\n"+
			"   • Consider switching to a more reliable RPC provider", err)
	}

	// 提供带有解决方案的增强错误信息
	if strings.Contains(lastErr.Error(), "429") || strings.Contains(lastErr.Error(), "Too Many Requests") {
		return nil, fmt.Errorf("❌ Rate limit exceeded after %d retries\n\n"+
			"🔧 SOLUTION: You're likely using a public RPC endpoint with strict limits.\n"+
			"   Update your config.json with a dedicated RPC endpoint:\n\n"+
			"   {\n"+
			"     \"network_url\": \"YOUR_DEDICATED_RPC_URL_HERE\",\n"+
			"     ...\n"+
			"   }\n\n"+
			"🚀 Get a free RPC endpoint from:\n"+
			"   • Helius: https://helius.dev (100k requests/day)\n"+
			"   • QuickNode: https://quicknode.com (30M requests/month)\n"+
			"   • Triton: https://triton.one (10M requests/month)\n\n"+
			"Original error: %w", maxRetries, lastErr)
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// shouldIncludeToken 根据扫描配置判断是否包含某个代币
func (w *WalletMonitor) shouldIncludeToken(mint string) bool {
	if w.scanConfig == nil {
		return true // 若无扫描配置，则包含全部
	}

	switch w.scanConfig.ScanMode {
	case "whitelist":
		// 仅包含 IncludeTokens 列表中的代币
		for _, token := range w.scanConfig.IncludeTokens {
			if strings.EqualFold(token, mint) {
				return true
			}
		}
		return false

	case "blacklist":
		// 包含所有代币，但排除 ExcludeTokens 列表中的
		for _, token := range w.scanConfig.ExcludeTokens {
			if strings.EqualFold(token, mint) {
				return false
			}
		}
		return true

	default: // "all" 或其他值
		return true
	}
}

func (w *WalletMonitor) GetWalletData(wallet solana.PublicKey) (*WalletData, error) {
	walletData := &WalletData{
		WalletAddress: wallet.String(),
		TokenAccounts: make(map[string]TokenAccountInfo),
		LastScanned:   time.Now(),
	}

	// 使用带重试的版本
	accounts, err := w.getTokenAccountsWithRetry(wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to get token accounts for wallet %s: %w", wallet.String(), err)
	}

	// 处理代币账户
	for _, acc := range accounts.Value {
		var tokenAccount token.Account
		err = bin.NewBinDecoder(acc.Account.Data.GetBinary()).Decode(&tokenAccount)
		if err != nil {
			log.Printf("⚠️  Warning: failed to decode token account (this is usually normal): %v", err)
			continue
		}

		// 仅包含余额为正且通过筛选的账户
		if tokenAccount.Amount > 0 {
			mint := tokenAccount.Mint.String()
			if w.shouldIncludeToken(mint) {
				walletData.TokenAccounts[mint] = TokenAccountInfo{
					Balance:     tokenAccount.Amount,
					LastUpdated: time.Now(),
					Symbol:      mint[:8] + "...",
					Decimals:    9,
				}
			}
		}
	}

	log.Printf("✅ Wallet %s: found %d token accounts (after filtering)", wallet.String(), len(walletData.TokenAccounts))
	return walletData, nil
}

// 添加以下类型定义
type Change struct {
	WalletAddress string
	TokenMint     string
	TokenSymbol   string // 代币符号
	TokenDecimals uint8  // 代币小数位
	ChangeType    string
	OldBalance    uint64
	NewBalance    uint64
	ChangePercent float64
	TokenBalances map[string]uint64 `json:",omitempty"`
}

func calculatePercentageChange(old, new uint64) float64 {
	if old == 0 {
		return 100.0 // 对于新增代币返回 100%
	}

	// 在除法前转换为 float64 以保持精度
	oldFloat := float64(old)
	newFloat := float64(new)

	// 计算百分比变化
	change := ((newFloat - oldFloat) / oldFloat) * 100.0

	// 四舍五入保留两位小数以避免浮点精度问题
	change = float64(int64(change*100)) / 100

	return change
}

// 计算绝对值的辅助函数
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func (w *WalletMonitor) checkConnection() error {
	// 尝试获取 slot 号作为简单的连接测试
	_, err := w.client.GetSlot(context.Background(), rpc.CommitmentFinalized)
	w.isConnected = err == nil

	if err != nil {
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too Many Requests") {
			return fmt.Errorf("RPC rate limit exceeded during connection check\n\n"+
				"💡 This indicates you're using a public RPC endpoint with strict limits.\n"+
				"   Consider upgrading to a dedicated RPC provider for reliable monitoring.\n\n"+
				"Original error: %w", err)
		}

		return fmt.Errorf("connection check failed: %w\n\n"+
			"💡 Troubleshooting steps:\n"+
			"   1. Check your network connection\n"+
			"   2. Verify your RPC endpoint URL in config.json\n"+
			"   3. Try a different RPC provider if the issue persists", err)
	}

	return nil
}

// 更新 ScanAllWallets 以处理批量
func (w *WalletMonitor) ScanAllWallets() (map[string]*WalletData, error) {
	// 先检查连接
	if err := w.checkConnection(); err != nil {
		return nil, err
	}

	results := make(map[string]*WalletData)
	batchSize := 2

	for i := 0; i < len(w.wallets); i += batchSize {
		end := i + batchSize
		if end > len(w.wallets) {
			end = len(w.wallets)
		}

		log.Printf("📊 Processing wallets %d-%d of %d", i+1, end, len(w.wallets))

		// 处理一批钱包
		for _, wallet := range w.wallets[i:end] {
			data, err := w.GetWalletData(wallet)
			if err != nil {
				log.Printf("❌ Error scanning wallet %s: %v", wallet.String(), err)
				// 返回错误以传递增强的错误信息
				return nil, fmt.Errorf("failed to scan wallet %s: %w", wallet.String(), err)
			}
			results[wallet.String()] = data
		}

		// 批次之间短暂延迟以照顾 RPC
		if end < len(w.wallets) {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return results, nil
}

func DetectChanges(oldData, newData map[string]*WalletData, significantChange float64) []Change {
	var changes []Change

	// 检查现有钱包的变化
	for walletAddr, newWalletData := range newData {
		oldWalletData, existed := oldData[walletAddr]

		if !existed {
			continue // 暂时跳过新钱包检测
		}

		// 检查现有钱包的变化
		for mint, newInfo := range newWalletData.TokenAccounts {
			oldInfo, existed := oldWalletData.TokenAccounts[mint]

			if !existed {
				// 检测到新代币
				changes = append(changes, Change{
					WalletAddress: walletAddr,
					TokenMint:     mint,
					TokenSymbol:   newInfo.Symbol,
					TokenDecimals: newInfo.Decimals,
					ChangeType:    "new_token",
					NewBalance:    newInfo.Balance,
				})
				continue
			}

			// 检查显著的余额变化
			pctChange := calculatePercentageChange(oldInfo.Balance, newInfo.Balance)
			absChange := abs(pctChange)

			if absChange >= significantChange {
				changes = append(changes, Change{
					WalletAddress: walletAddr,
					TokenMint:     mint,
					TokenSymbol:   newInfo.Symbol,
					TokenDecimals: newInfo.Decimals,
					ChangeType:    "balance_change",
					OldBalance:    oldInfo.Balance,
					NewBalance:    newInfo.Balance,
					ChangePercent: pctChange,
				})
			}
		}
	}

	return changes
}

// 添加此辅助函数
func formatTokenAmount(amount uint64, decimals uint8) string {
	if decimals == 0 {
		return fmt.Sprintf("%d", amount)
	}

	// 转换为 float64 并除以 10^decimals
	divisor := math.Pow(10, float64(decimals))
	value := float64(amount) / divisor

	// 根据数值大小格式化小数位
	switch {
	case value >= 5000:
		return fmt.Sprintf("%.2fM", value/1000)
	case value >= 5:
		return fmt.Sprintf("%.2fK", value)
	default:
		return fmt.Sprintf("%.4f", value)
	}
}

// FormatWalletOverview 返回钱包持仓的简洁表示
func FormatWalletOverview(data map[string]*WalletData) string {
	var overview strings.Builder
	overview.WriteString("\nWallet Holdings Overview:\n")
	overview.WriteString("------------------------\n")

	for _, wallet := range data {
		overview.WriteString(fmt.Sprintf("📍 %s\n", wallet.WalletAddress))
		if len(wallet.TokenAccounts) == 0 {
			overview.WriteString("   No tokens found\n")
			continue
		}

		// 将映射转换为切片以便排序
		type tokenHolding struct {
			symbol   string
			balance  uint64
			decimals uint8
		}
		holdings := make([]tokenHolding, 0, len(wallet.TokenAccounts))
		for _, info := range wallet.TokenAccounts {
			holdings = append(holdings, tokenHolding{
				symbol:   info.Symbol,
				balance:  info.Balance,
				decimals: info.Decimals,
			})
		}

		// 按余额排序（从高到低）
		sort.Slice(holdings, func(i, j int) bool {
			return holdings[i].balance > holdings[j].balance
		})

		// 显示前五大持仓
		maxDisplay := 5
		if len(holdings) < maxDisplay {
			maxDisplay = len(holdings)
		}
		for i := 0; i < maxDisplay; i++ {
			balance := formatTokenAmount(holdings[i].balance, holdings[i].decimals)
			overview.WriteString(fmt.Sprintf("   • %s: %s\n", holdings[i].symbol, balance))
		}

		// 如有更多代币则显示数量
		remaining := len(holdings) - maxDisplay
		if remaining > 0 {
			overview.WriteString(fmt.Sprintf("   ... and %d more tokens\n", remaining))
		}
		overview.WriteString("\n")
	}
	return overview.String()
}

// 更新 FormatWalletOverview 以包含置信度指示器
func formatTokenValue(value float64, confidence string) string {
	var indicator string
	switch strings.ToLower(confidence) {
	case "high":
		indicator = "✅"
	case "medium":
		indicator = "⚠️"
	default:
		indicator = "❓"
	}

	if value >= 1000000 {
		return fmt.Sprintf(" ($%.2fM) %s", value/1000000, indicator)
	} else if value >= 1000 {
		return fmt.Sprintf(" ($%.2fK) %s", value/1000, indicator)
	}
	return fmt.Sprintf(" ($%.2f) %s", value, indicator)
}

// 添加结构体以存储带有美元价值的代币数据
type tokenHolding struct {
	Mint     string
	Amount   float64
	USDValue float64
	Symbol   string
}

// 更新 DisplayWalletOverview 函数以提供更美观的输出
func (m *WalletMonitor) DisplayWalletOverview(walletDataMap map[string]*WalletData) {
	// 终端颜色代码
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorCyan   = "\033[36m"
		colorWhite  = "\033[37m"
		colorBold   = "\033[1m"
	)

	// 符号
	const (
		walletSymbol = "💼"
		tokenSymbol  = "🔹"
		dollarSymbol = "💲"
		moreSymbol   = "..."
		divider      = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	)

	fmt.Println()
	fmt.Printf("%s%s SOLANA WALLET MONITOR %s\n", colorBold, colorPurple, colorReset)
	fmt.Printf("%s%s %s\n\n", colorPurple, divider, colorReset)

	// 收集所有唯一的铸币地址
	mints := make([]string, 0)
	for _, walletData := range walletDataMap {
		for mint := range walletData.TokenAccounts {
			mints = append(mints, mint)
		}
	}

	// 更新所有代币的价格
	if err := m.priceService.UpdatePrices(mints); err != nil {
		log.Printf("Error updating prices: %v", err)
	}

	// 总价值计数器
	totalPortfolioValue := 0.0

	for _, wallet := range m.wallets {
		fmt.Printf("%s%s %s %s%s\n", colorBold, colorBlue, walletSymbol, wallet.String(), colorReset)
		walletData, exists := walletDataMap[wallet.String()]
		if !exists {
			fmt.Printf("   %sNo data available%s\n\n", colorYellow, colorReset)
			continue
		}

		// 将代币持仓转换为切片以便排序
		holdings := make([]tokenHolding, 0)
		walletTotalValue := 0.0

		for mint, info := range walletData.TokenAccounts {
			// 从 Jupiter 获取价格数据
			priceData, exists := m.priceService.GetPrice(mint)

			usdValue := 0.0
			if exists {
				// 考虑小数位将余额转换为浮点数
				actualAmount := float64(info.Balance) / math.Pow(10, float64(info.Decimals))
				usdValue = actualAmount * priceData.Price
				walletTotalValue += usdValue
			}

			// 尝试查找常见代币地址以获得更好的名称
			symbol := info.Symbol
			if tokenName, found := getKnownTokenName(mint); found {
				symbol = tokenName
			}

			holdings = append(holdings, tokenHolding{
				Mint:     mint,
				Amount:   float64(info.Balance),
				USDValue: usdValue,
				Symbol:   symbol,
			})
		}

		totalPortfolioValue += walletTotalValue

		// 按美元价值降序排序
		sort.Slice(holdings, func(i, j int) bool {
			return holdings[i].USDValue > holdings[j].USDValue
		})

		// 显示钱包总价值
		if walletTotalValue > 0 {
			// 根据数值大小格式化
			valueStr := ""
			if walletTotalValue >= 1000000 {
				valueStr = fmt.Sprintf("$%.2fM", walletTotalValue/1000000)
			} else if walletTotalValue >= 1000 {
				valueStr = fmt.Sprintf("$%.2fK", walletTotalValue/1000)
			} else {
				valueStr = fmt.Sprintf("$%.2f", walletTotalValue)
			}
			fmt.Printf("   %s%sTotal Value: %s%s\n", colorBold, colorGreen, valueStr, colorReset)
		}

		// 以更好的格式显示前五大持仓
		for i := 0; i < min(5, len(holdings)); i++ {
			holding := holdings[i]

			// 获取用于展示的缩写或符号
			displayName := holding.Symbol
			if displayName == holding.Mint[:8]+"..." {
				// 如果仍然只是缩写，则检查是否为常见代币
				if tokenName, found := getKnownTokenName(holding.Mint); found {
					displayName = tokenName
				}
			}

			// 格式化数量
			actualAmount := holding.Amount / math.Pow(10, float64(9)) // 假设 9 位小数
			amountStr := ""
			if actualAmount >= 1000000 {
				amountStr = fmt.Sprintf("%.2fM", actualAmount/1000000)
			} else if actualAmount >= 1000 {
				amountStr = fmt.Sprintf("%.2fK", actualAmount/1000)
			} else {
				amountStr = fmt.Sprintf("%.4f", actualAmount)
			}

			// 根据价值选择颜色
			valueColor := colorWhite
			if holding.USDValue > 1000 {
				valueColor = colorGreen
			} else if holding.USDValue > 100 {
				valueColor = colorCyan
			}

			if holding.USDValue > 0 {
				fmt.Printf("   %s %s%-15s%s %12s %s%s($%.2f)%s\n",
					tokenSymbol,
					colorBold,
					displayName,
					colorReset,
					amountStr,
					valueColor,
					dollarSymbol,
					holding.USDValue,
					colorReset)
			} else {
				fmt.Printf("   %s %s%-15s%s %12s\n",
					tokenSymbol,
					colorBold,
					displayName,
					colorReset,
					amountStr)
			}
		}

		if len(holdings) > 5 {
			fmt.Printf("   %s %s%d more tokens%s\n", moreSymbol, colorYellow, len(holdings)-5, colorReset)
		}
		fmt.Println()
	}

	// 显示投资组合总价值
	if totalPortfolioValue > 0 {
		fmt.Printf("%s%s %s\n", colorPurple, divider, colorReset)
		if totalPortfolioValue >= 1000000 {
			fmt.Printf("%s%sTOTAL PORTFOLIO VALUE: $%.2fM%s\n", colorBold, colorGreen, totalPortfolioValue/1000000, colorReset)
		} else if totalPortfolioValue >= 1000 {
			fmt.Printf("%s%sTOTAL PORTFOLIO VALUE: $%.2fK%s\n", colorBold, colorGreen, totalPortfolioValue/1000, colorReset)
		} else {
			fmt.Printf("%s%sTOTAL PORTFOLIO VALUE: $%.2f%s\n", colorBold, colorGreen, totalPortfolioValue, colorReset)
		}
	}

	fmt.Printf("%s%s %s\n", colorPurple, divider, colorReset)
	fmt.Printf("%sLast updated: %s%s\n\n", colorYellow, time.Now().Format("2006-01-02 15:04:05"), colorReset)
}

// 查找常见代币名称的辅助函数
func getKnownTokenName(mint string) (string, bool) {
	// 常见代币铸币地址到符号的映射
	knownTokens := map[string]string{
		"So11111111111111111111111111111111111111112":  "SOL",
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": "USDC",
		"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB": "USDT",
		"DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263": "BONK",
		"7dHbWXmci3dT8UFYWYZweBLXgycu7Y3iL6trKn1Y7ARj": "stSOL",
		"mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So":  "mSOL",
		"kinXdEcpDQeHPEuQnqmUgtYykqKGVFq6CeVX5iAHJq6":  "KIN",
		"JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN":  "JUP",
	}

	symbol, found := knownTokens[mint]
	return symbol, found
}

// 计算最小值的辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
