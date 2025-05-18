package monitor

import (
	"fmt"
	"sort"
	"time"
)

type MockWalletMonitor struct {
	startTime time.Time
	data      map[string]*WalletData
}

func NewMockWalletMonitor() *MockWalletMonitor {
	// Initialize with more realistic test data
	return &MockWalletMonitor{
		startTime: time.Now(),
		data: map[string]*WalletData{
			"TestWallet1": {
				WalletAddress: "TestWallet1",
				TokenAccounts: map[string]TokenAccountInfo{
					"So11111111111111111111111111111111111111112": { // SOL
						Balance:     1000000000, // 1 SOL
						LastUpdated: time.Now(),
					},
					"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": { // USDC
						Balance:     1000000, // 1 USDC
						LastUpdated: time.Now(),
					},
				},
			},
		},
	}
}

func (m *MockWalletMonitor) ScanAllWallets() (map[string]*WalletData, error) {
	results := make(map[string]*WalletData)
	now := time.Now()
	elapsed := now.Sub(m.startTime)

	// Base wallet always present
	baseWallet := &WalletData{
		WalletAddress: "TestWallet1",
		TokenAccounts: map[string]TokenAccountInfo{
			"So11111111111111111111111111111111111111112": { // SOL
				Balance:     1000000000,
				LastUpdated: now,
				Symbol:      "SOL",
				Decimals:    9,
			},
			"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": { // USDC
				Balance:     1000000,
				LastUpdated: now,
				Symbol:      "USDC",
				Decimals:    6,
			},
		},
		LastScanned: now,
	}

	// Apply changes based on time
	if elapsed >= 5*time.Second {
		// Add BONK token after 5 seconds
		baseWallet.TokenAccounts["DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263"] = TokenAccountInfo{
			Balance:     5000000,
			LastUpdated: now,
			Symbol:      "BONK",
			Decimals:    5,
		}
	}

	if elapsed >= 10*time.Second {
		// Increase SOL balance after 10 seconds
		baseWallet.TokenAccounts["So11111111111111111111111111111111111111112"] = TokenAccountInfo{
			Balance:     2000000000,
			LastUpdated: now,
			Symbol:      "SOL",
			Decimals:    9,
		}
	}

	results["TestWallet1"] = baseWallet

	if elapsed >= 15*time.Second {
		// Add second wallet after 15 seconds
		results["TestWallet2"] = &WalletData{
			WalletAddress: "TestWallet2",
			TokenAccounts: map[string]TokenAccountInfo{
				"So11111111111111111111111111111111111111112": {
					Balance:     5000000000,
					LastUpdated: now,
					Symbol:      "SOL",
					Decimals:    9,
				},
			},
			LastScanned: now,
		}
	}

	return results, nil
}

// Add DisplayWalletOverview method to match the real monitor
func (m *MockWalletMonitor) DisplayWalletOverview(walletDataMap map[string]*WalletData) {
	// Terminal color codes
	const (
		colorReset  = "\033[0m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorCyan   = "\033[36m"
		colorBold   = "\033[1m"
	)

	// Symbols
	const (
		walletSymbol = "💼"
		tokenSymbol  = "🔹"
		dollarSymbol = "💲"
		divider      = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	)

	fmt.Println()
	fmt.Printf("%s%s TEST MODE: SIMULATED WALLET DATA %s\n", colorBold, colorYellow, colorReset)
	fmt.Printf("%s%s %s\n\n", colorPurple, divider, colorReset)

	for _, wallet := range walletDataMap {
		fmt.Printf("%s%s %s %s%s\n", colorBold, colorBlue, walletSymbol, wallet.WalletAddress, colorReset)

		// Sort tokens
		type tokenHolding struct {
			mint    string
			symbol  string
			balance uint64
		}
		
		holdings := make([]tokenHolding, 0)
		for mint, info := range wallet.TokenAccounts {
			holdings = append(holdings, tokenHolding{
				mint:    mint,
				symbol:  info.Symbol,
				balance: info.Balance,
			})
		}

		// Sort by balance
		sort.Slice(holdings, func(i, j int) bool {
			return holdings[i].balance > holdings[j].balance
		})

		// Display tokens
		for _, holding := range holdings {
			// Format amount based on token
			var amountStr, tokenName string
			
			if holding.mint == "So11111111111111111111111111111111111111112" {
				// SOL
				amountStr = fmt.Sprintf("%.4f", float64(holding.balance)/1e9)
				tokenName = "SOL"
			} else if holding.mint == "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" {
				// USDC
				amountStr = fmt.Sprintf("%.2f", float64(holding.balance)/1e6)
				tokenName = "USDC"
			} else if holding.mint == "DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263" {
				// BONK
				amountStr = fmt.Sprintf("%.2f", float64(holding.balance)/1e5)
				tokenName = "BONK"
			} else {
				amountStr = fmt.Sprintf("%.2f", float64(holding.balance)/1e9)
				tokenName = holding.symbol
			}
			
			fmt.Printf("   %s %s%-10s%s %15s\n", 
				tokenSymbol,
				colorBold,
				tokenName,
				colorReset,
				amountStr)
		}
		fmt.Println()
	}

	fmt.Printf("%s%s %s\n", colorPurple, divider, colorReset)
	fmt.Printf("%sNote: Test mode simulates activity every 5 seconds%s\n\n", colorYellow, colorReset)
}
