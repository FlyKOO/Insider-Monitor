package monitor

import (
	"testing"
	"time"
)

func TestMockWalletMonitor(t *testing.T) {
	mock := NewMockWalletMonitor()
	
	// Test cases for different scenarios
	testCases := []struct {
		name     string
		waitTime time.Duration
		check    func(t *testing.T, data map[string]*WalletData)
	}{
		{
			name:     "initial state",
			waitTime: 0,
			check: func(t *testing.T, data map[string]*WalletData) {
				if len(data) != 1 {
					t.Errorf("Expected 1 wallet, got %d", len(data))
				}
				wallet := data["TestWallet1"]
				if len(wallet.TokenAccounts) != 2 {
					t.Errorf("Expected 2 tokens initially, got %d", len(wallet.TokenAccounts))
				}
			},
		},
		{
			name:     "after new token addition",
			waitTime: 6 * time.Second,
			check: func(t *testing.T, data map[string]*WalletData) {
				wallet := data["TestWallet1"]
				if len(wallet.TokenAccounts) != 3 {
					t.Errorf("Expected 3 tokens after addition, got %d", len(wallet.TokenAccounts))
				}
				if _, hasBonk := wallet.TokenAccounts["DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263"]; !hasBonk {
					t.Error("BONK token not found after expected addition")
				}
			},
		},
		{
			name:     "after balance change",
			waitTime: 5 * time.Second,
			check: func(t *testing.T, data map[string]*WalletData) {
				wallet := data["TestWallet1"]
				solBalance := wallet.TokenAccounts["So11111111111111111111111111111111111111112"].Balance
				if solBalance != 2000000000 {
					t.Errorf("Expected SOL balance of 2 SOL, got %d", solBalance)
				}
			},
		},
		{
			name:     "new wallet addition",
			waitTime: 5 * time.Second,
			check: func(t *testing.T, data map[string]*WalletData) {
				if len(data) != 2 {
					t.Errorf("Expected 2 wallets after addition, got %d", len(data))
				}
				if _, hasWallet2 := data["TestWallet2"]; !hasWallet2 {
					t.Error("TestWallet2 not found after expected addition")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			time.Sleep(tc.waitTime)
			data, err := mock.ScanAllWallets()
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}
			tc.check(t, data)
		})
	}
} 