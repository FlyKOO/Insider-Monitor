package monitor

import (
	"testing"
	"time"
)

func TestDetectChanges(t *testing.T) {
	tests := []struct {
		name          string
		oldData       map[string]*WalletData
		newData       map[string]*WalletData
		expectedCount int
		expectedTypes []string
	}{
		{
			name: "new wallet detection",
			oldData: map[string]*WalletData{},
			newData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {Balance: 100},
					},
				},
			},
			expectedCount: 1,
			expectedTypes: []string{"new_wallet"},
		},
		{
			name: "balance change detection",
			oldData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {Balance: 100},
					},
				},
			},
			newData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {Balance: 200},
					},
				},
			},
			expectedCount: 1,
			expectedTypes: []string{"balance_change"},
		},
		{
			name: "new token detection",
			oldData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {Balance: 100},
					},
				},
			},
			newData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {Balance: 100},
						"TokenB": {Balance: 200},
					},
				},
			},
			expectedCount: 1,
			expectedTypes: []string{"new_token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := DetectChanges(tt.oldData, tt.newData)
			
			if len(changes) != tt.expectedCount {
				t.Errorf("expected %d changes, got %d", tt.expectedCount, len(changes))
			}
			
			for i, expectedType := range tt.expectedTypes {
				if i >= len(changes) {
					t.Errorf("missing expected change type: %s", expectedType)
					continue
				}
				if changes[i].ChangeType != expectedType {
					t.Errorf("expected change type %s, got %s", expectedType, changes[i].ChangeType)
				}
			}
		})
	}
}

func TestMockMonitorIntegration(t *testing.T) {
	mock := NewMockWalletMonitor()
	
	// Initial scan
	initial, err := mock.ScanAllWallets()
	if err != nil {
		t.Fatalf("Initial scan failed: %v", err)
	}
	
	if len(initial) != 1 {
		t.Errorf("Expected 1 wallet initially, got %d", len(initial))
	}
	
	// Check initial tokens
	wallet := initial["TestWallet1"]
	if wallet == nil {
		t.Fatal("TestWallet1 not found")
	}
	
	if len(wallet.TokenAccounts) != 2 {
		t.Errorf("Expected 2 initial tokens, got %d", len(wallet.TokenAccounts))
	}
	
	// Test changes over time
	time.Sleep(6 * time.Second)
	afterNewToken, _ := mock.ScanAllWallets()
	if len(afterNewToken["TestWallet1"].TokenAccounts) != 3 {
		t.Error("New token not added after 5 seconds")
	}
	
	time.Sleep(5 * time.Second)
	afterBalanceChange, _ := mock.ScanAllWallets()
	solBalance := afterBalanceChange["TestWallet1"].TokenAccounts["So11111111111111111111111111111111111111112"].Balance
	if solBalance != 2000000000 {
		t.Errorf("Expected SOL balance change to 2 SOL, got %d", solBalance)
	}
} 