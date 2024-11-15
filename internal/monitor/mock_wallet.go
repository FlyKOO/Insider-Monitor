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
	now := time.Now()
	elapsed := now.Sub(m.startTime)
	
	// Create a copy of the current data
	result := make(map[string]*WalletData)
	for k, v := range m.data {
		walletCopy := &WalletData{
			WalletAddress: v.WalletAddress,
			TokenAccounts: make(map[string]TokenAccountInfo),
			LastScanned:  now,
		}
		for tk, tv := range v.TokenAccounts {
			walletCopy.TokenAccounts[tk] = tv
		}
		result[k] = walletCopy
	}
	
	// Simulate different changes at different intervals
	switch {
	case elapsed >= 5*time.Second && elapsed < 10*time.Second:
		// Add a new token after 5 seconds (simulating a new token acquisition)
		result["TestWallet1"].TokenAccounts["DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263"] = TokenAccountInfo{
			Balance:     5000000, // 5 BONK
			LastUpdated: now,
		}
	
	case elapsed >= 10*time.Second && elapsed < 15*time.Second:
		// Significant increase in SOL balance after 10 seconds
		result["TestWallet1"].TokenAccounts["So11111111111111111111111111111111111111112"] = TokenAccountInfo{
			Balance:     2000000000, // 2 SOL (100% increase)
			LastUpdated: now,
		}
	
	case elapsed >= 15*time.Second && elapsed < 20*time.Second:
		// Critical decrease in USDC balance after 15 seconds
		result["TestWallet1"].TokenAccounts["EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"] = TokenAccountInfo{
			Balance:     100000, // 0.1 USDC (90% decrease)
			LastUpdated: now,
		}
	
	case elapsed >= 20*time.Second:
		// Add a new wallet after 20 seconds
		result["TestWallet2"] = &WalletData{
			WalletAddress: "TestWallet2",
			TokenAccounts: map[string]TokenAccountInfo{
				"So11111111111111111111111111111111111111112": {
					Balance:     5000000000, // 5 SOL
					LastUpdated: now,
				},
			},
			LastScanned: now,
		}
	}
	
	// Update internal state
	m.data = result
	return result, nil
} 