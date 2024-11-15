package monitor

import (
	"testing"
	"time"
)

func TestMockWalletMonitor(t *testing.T) {
	mock := NewMockWalletMonitor()
	
	// Initial scan
	initial, err := mock.ScanAllWallets()
	if err != nil {
		t.Fatalf("Initial scan failed: %v", err)
	}
	
	if len(initial) != 1 {
		t.Errorf("Expected 1 wallet, got %d", len(initial))
	}
	
	// Check initial tokens
	wallet := initial["TestWallet1"]
	if wallet == nil {
		t.Fatal("Test wallet not found")
	}
 
	// Check for SOL and USDC
	if _, hasSol := wallet.TokenAccounts["So11111111111111111111111111111111111111112"]; !hasSol {
		t.Error("Expected SOL token in initial wallet")
	}
	if _, hasUsdc := wallet.TokenAccounts["EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"]; !hasUsdc {
		t.Error("Expected USDC token in initial wallet")
	}
	
	// Wait for BONK token addition
	time.Sleep(6 * time.Second)
	afterNewToken, _ := mock.ScanAllWallets()
	wallet = afterNewToken["TestWallet1"]
	if _, hasBonk := wallet.TokenAccounts["DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263"]; !hasBonk {
		t.Error("Expected BONK token to be added")
	}
	
	// Wait for SOL balance change
	time.Sleep(5 * time.Second)
	afterBalanceChange, _ := mock.ScanAllWallets()
	wallet = afterBalanceChange["TestWallet1"]
	solBalance := wallet.TokenAccounts["So11111111111111111111111111111111111111112"].Balance
	if solBalance != 2000000000 {
		t.Errorf("Expected SOL balance to be 2 SOL (2000000000), got %d", solBalance)
	}
} 