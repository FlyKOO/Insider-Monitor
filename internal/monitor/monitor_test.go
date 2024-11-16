package monitor

import (
	"testing"
	"time"
)

func TestDetectChanges(t *testing.T) {
	const testSignificantChange = 5.0 // 5% threshold for testing

	tests := []struct {
		name            string
		oldData         map[string]*WalletData
		newData         map[string]*WalletData
		expectedChanges []Change
	}{
		{
			name: "balance changes",
			oldData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {
							Balance:  100,
							Symbol:   "TKNA",
							Decimals: 6,
						},
					},
				},
			},
			newData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {
							Balance:  150,
							Symbol:   "TKNA",
							Decimals: 6,
						},
					},
				},
			},
			expectedChanges: []Change{
				{
					WalletAddress: "Wallet1",
					TokenMint:     "TokenA",
					TokenSymbol:   "TKNA",
					TokenDecimals: 6,
					ChangeType:    "balance_change",
					OldBalance:    100,
					NewBalance:    150,
					ChangePercent: 50.0,
				},
			},
		},
		{
			name: "new tokens in existing wallet",
			oldData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {
							Balance:  100,
							Symbol:   "TKNA",
							Decimals: 6,
						},
					},
				},
			},
			newData: map[string]*WalletData{
				"Wallet1": {
					WalletAddress: "Wallet1",
					TokenAccounts: map[string]TokenAccountInfo{
						"TokenA": {
							Balance:  100,
							Symbol:   "TKNA",
							Decimals: 6,
						},
						"TokenB": {
							Balance:  200,
							Symbol:   "TKNB",
							Decimals: 9,
						},
					},
				},
			},
			expectedChanges: []Change{
				{
					WalletAddress: "Wallet1",
					TokenMint:     "TokenB",
					TokenSymbol:   "TKNB",
					TokenDecimals: 9,
					ChangeType:    "new_token",
					NewBalance:    200,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := DetectChanges(tt.oldData, tt.newData, testSignificantChange)

			if len(changes) != len(tt.expectedChanges) {
				t.Errorf("Expected %d changes, got %d", len(tt.expectedChanges), len(changes))
				return
			}

			for i, expected := range tt.expectedChanges {
				actual := changes[i]
				if actual.ChangeType != expected.ChangeType {
					t.Errorf("Change %d: expected type %s, got %s", i, expected.ChangeType, actual.ChangeType)
				}
				if actual.WalletAddress != expected.WalletAddress {
					t.Errorf("Change %d: expected wallet %s, got %s", i, expected.WalletAddress, actual.WalletAddress)
				}
				if actual.TokenMint != expected.TokenMint {
					t.Errorf("Change %d: expected token %s, got %s", i, expected.TokenMint, actual.TokenMint)
				}
				if actual.TokenSymbol != expected.TokenSymbol {
					t.Errorf("Change %d: expected symbol %s, got %s", i, expected.TokenSymbol, actual.TokenSymbol)
				}
				if actual.TokenDecimals != expected.TokenDecimals {
					t.Errorf("Change %d: expected decimals %d, got %d", i, expected.TokenDecimals, actual.TokenDecimals)
				}
				if actual.NewBalance != expected.NewBalance {
					t.Errorf("Change %d: expected new balance %d, got %d", i, expected.NewBalance, actual.NewBalance)
				}
				if actual.ChangeType == "balance_change" {
					if actual.OldBalance != expected.OldBalance {
						t.Errorf("Change %d: expected old balance %d, got %d", i, expected.OldBalance, actual.OldBalance)
					}
					if actual.ChangePercent != expected.ChangePercent {
						t.Errorf("Change %d: expected change percent %.2f, got %.2f", i, expected.ChangePercent, actual.ChangePercent)
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
