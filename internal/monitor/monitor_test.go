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
		expectedChanges []Change
	}{
		{
			name: "new wallet with multiple tokens",
			oldData: map[string]*WalletData{},
			newData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {Balance: 100},
						"TokenB": {Balance: 200},
						"TokenC": {Balance: 300},
					},
				},
			},
			expectedChanges: []Change{
				{
					WalletAddress: "Wallet1",
					ChangeType:    "new_wallet",
					TokenBalances: map[string]uint64{
						"TokenA": 100,
						"TokenB": 200,
						"TokenC": 300,
					},
				},
			},
		},
		{
			name: "multiple balance changes",
			oldData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {Balance: 100},
						"TokenB": {Balance: 200},
					},
				},
			},
			newData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {Balance: 150},
						"TokenB": {Balance: 100},
					},
				},
			},
			expectedChanges: []Change{
				{
					WalletAddress:  "Wallet1",
					TokenMint:      "TokenA",
					ChangeType:     "balance_change",
					OldBalance:     100,
					NewBalance:     150,
					ChangePercent:  0.5,
				},
				{
					WalletAddress:  "Wallet1",
					TokenMint:      "TokenB",
					ChangeType:     "balance_change",
					OldBalance:     200,
					NewBalance:     100,
					ChangePercent:  -0.5,
				},
			},
		},
		{
			name: "new tokens in existing wallet",
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
						"TokenC": {Balance: 300},
					},
				},
			},
			expectedChanges: []Change{
				{
					WalletAddress: "Wallet1",
					TokenMint:     "TokenB",
					ChangeType:    "new_token",
					NewBalance:    200,
				},
				{
					WalletAddress: "Wallet1",
					TokenMint:     "TokenC",
					ChangeType:    "new_token",
					NewBalance:    300,
				},
			},
		},
		{
			name: "multiple wallets with changes",
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
						"TokenA": {Balance: 150},
					},
				},
				"Wallet2": {
					WalletAddress: "Wallet2",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenB": {Balance: 200},
						"TokenC": {Balance: 300},
					},
				},
			},
			expectedChanges: []Change{
				{
					WalletAddress:  "Wallet1",
					TokenMint:      "TokenA",
					ChangeType:     "balance_change",
					OldBalance:     100,
					NewBalance:     150,
					ChangePercent:  0.5,
				},
				{
					WalletAddress: "Wallet2",
					ChangeType:    "new_wallet",
					TokenBalances: map[string]uint64{
						"TokenB": 200,
						"TokenC": 300,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := DetectChanges(tt.oldData, tt.newData)
			
			if len(changes) != len(tt.expectedChanges) {
				t.Errorf("Expected %d changes, got %d", len(tt.expectedChanges), len(changes))
				return
			}

			for i, expected := range tt.expectedChanges {
				actual := changes[i]
				if actual.ChangeType != expected.ChangeType {
					t.Errorf("Change %d: expected type %s, got %s", i, expected.ChangeType, actual.ChangeType)
				}
				
				if actual.ChangeType == "new_wallet" {
					if len(actual.TokenBalances) != len(expected.TokenBalances) {
						t.Errorf("Change %d: expected %d tokens, got %d", i, len(expected.TokenBalances), len(actual.TokenBalances))
					}
					for token, balance := range expected.TokenBalances {
						if actual.TokenBalances[token] != balance {
							t.Errorf("Change %d: token %s expected balance %d, got %d", i, token, balance, actual.TokenBalances[token])
						}
					}
				} else {
					if actual.WalletAddress != expected.WalletAddress {
						t.Errorf("Change %d: expected wallet %s, got %s", i, expected.WalletAddress, actual.WalletAddress)
					}
					if actual.TokenMint != expected.TokenMint {
						t.Errorf("Change %d: expected token %s, got %s", i, expected.TokenMint, actual.TokenMint)
					}
					if actual.NewBalance != expected.NewBalance {
						t.Errorf("Change %d: expected new balance %d, got %d", i, expected.NewBalance, actual.NewBalance)
					}
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