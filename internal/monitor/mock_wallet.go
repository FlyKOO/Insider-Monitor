package monitor

import (
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
				Symbol:     "SOL",
				Decimals:   9,
			},
			"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": { // USDC
				Balance:     1000000,
				LastUpdated: now,
				Symbol:     "USDC",
				Decimals:   6,
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
			Symbol:     "BONK",
			Decimals:   5,
		}
	}

	if elapsed >= 10*time.Second {
		// Increase SOL balance after 10 seconds
		baseWallet.TokenAccounts["So11111111111111111111111111111111111111112"] = TokenAccountInfo{
			Balance:     2000000000,
			LastUpdated: now,
			Symbol:     "SOL",
			Decimals:   9,
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
					Symbol:     "SOL",
					Decimals:   9,
				},
			},
			LastScanned: now,
		}
	}

	return results, nil
}
