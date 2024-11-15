package monitor

import (
	"time"
)

type MockWalletMonitor struct {
	startTime time.Time
	data      map[string]*WalletData
}

func NewMockWalletMonitor() *MockWalletMonitor {
	return &MockWalletMonitor{
		startTime: time.Now(),
		data: map[string]*WalletData{
			"TestWallet1": {
				WalletAddress: "TestWallet1",
				TokenAccounts: map[string]TokenAccountInfo{
					"TokenA": {Balance: 100, LastUpdated: time.Now()},
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
	
	// Simulate changes based on elapsed time
	if elapsed >= 10*time.Second && elapsed < 15*time.Second {
		// Add TokenB after 10 seconds
		result["TestWallet1"].TokenAccounts["TokenB"] = TokenAccountInfo{
			Balance:     500,
			LastUpdated: now,
		}
	} else if elapsed >= 15*time.Second {
		// Change TokenA balance after 15 seconds
		result["TestWallet1"].TokenAccounts["TokenA"] = TokenAccountInfo{
			Balance:     200,
			LastUpdated: now,
		}
		result["TestWallet1"].TokenAccounts["TokenB"] = TokenAccountInfo{
			Balance:     500,
			LastUpdated: now,
		}
	}
	
	// Update internal state
	m.data = result
	return result, nil
} 