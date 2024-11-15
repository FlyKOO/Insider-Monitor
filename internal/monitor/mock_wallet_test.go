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
	
	// Wait for first change
	time.Sleep(11 * time.Second)
	
	after10s, _ := mock.ScanAllWallets()
	wallet := after10s["TestWallet1"]
	if wallet == nil {
		t.Fatal("Test wallet not found")
	}
 
	if _, hasTokenB := wallet.TokenAccounts["TokenB"]; !hasTokenB {
		t.Error("Expected TokenB to be added after 10 seconds")
	}
	
	// Wait for second change
	time.Sleep(5 * time.Second)
	
	after15s, _ := mock.ScanAllWallets()
	wallet = after15s["TestWallet1"]
	
	tokenA := wallet.TokenAccounts["TokenA"]
	if tokenA.Balance != 200 {
		t.Errorf("Expected TokenA balance to be 200, got %d", tokenA.Balance)
	}
} 