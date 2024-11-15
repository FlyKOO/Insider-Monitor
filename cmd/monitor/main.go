package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/accursedgalaxy/insider-monitor/internal/config"
	"github.com/accursedgalaxy/insider-monitor/internal/monitor"
	"github.com/accursedgalaxy/insider-monitor/internal/storage"
	"github.com/gagliardetto/solana-go/rpc"
)

// WalletScanner interface defines the contract for wallet monitoring
type WalletScanner interface {
	ScanAllWallets() (map[string]*monitor.WalletData, error)
}

func main() {
	testMode := flag.Bool("test", false, "Run in test mode with accelerated scanning")
	flag.Parse()

	// Initialize monitor based on mode
	var scanner WalletScanner
	scanInterval := time.Minute

	if *testMode {
		scanInterval = time.Second * 5
		log.Println("Running in test mode with 5-second scan interval")
		scanner = monitor.NewMockWalletMonitor()
	} else {
		cfg := &config.Config{
			NetworkURL: rpc.MainNetBeta_RPC,
			Wallets:   config.Wallets,
		}
		if err := cfg.Validate(); err != nil {
			log.Fatalf("invalid configuration: %v", err)
		}

		var err error
		scanner, err = monitor.NewWalletMonitor(cfg.NetworkURL, cfg.Wallets)
		if err != nil {
			log.Fatalf("failed to create wallet monitor: %v", err)
		}
	}

	runMonitor(scanner, scanInterval)
}

func runMonitor(scanner WalletScanner, scanInterval time.Duration) {
	storage := storage.New("./data")
	
	// Load previous data for comparison
	previousData, err := storage.LoadWalletData()
	if err != nil {
		log.Printf("warning: could not load previous data: %v", err)
	}

	// Perform initial scan
	log.Println("Performing initial wallet scan...")
	results, err := scanner.ScanAllWallets()
	if err != nil {
		log.Printf("initial scan error: %v", err)
	} else {
		log.Printf("Initial scan complete. Found tokens in %d wallets", len(results))
		if err := storage.SaveWalletData(results); err != nil {
			log.Printf("error saving initial data: %v", err)
		}
		previousData = results
	}

	// Setup monitoring loop
	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	log.Println("Monitoring for changes (Ctrl+C to exit)...")

	for {
		select {
		case <-ticker.C:
			newResults, err := scanner.ScanAllWallets()
			if err != nil {
				log.Printf("error scanning wallets: %v", err)
				continue
			}

			changes := monitor.DetectChanges(previousData, newResults)
			logChanges(changes)

			if err := storage.SaveWalletData(newResults); err != nil {
				log.Printf("error saving data: %v", err)
			}
			previousData = newResults

		case <-interrupt:
			log.Println("Shutting down gracefully...")
			monitor.LogToFile("./data", "Monitor shutting down gracefully")
			return
		}
	}
}

func logChanges(changes []monitor.Change) {
	for _, change := range changes {
		var msg string
		switch change.ChangeType {
		case "new_wallet":
			msg = fmt.Sprintf("New wallet %s: Token %s with balance %d",
				change.WalletAddress, change.TokenMint, change.NewBalance)
		case "new_token":
			msg = fmt.Sprintf("New token detected in %s: %s with balance %d",
				change.WalletAddress, change.TokenMint, change.NewBalance)
		case "balance_change":
			msg = fmt.Sprintf("Balance change in %s: Token %s from %d to %d",
				change.WalletAddress, change.TokenMint, change.OldBalance, change.NewBalance)
		}
		
		if msg != "" {
			log.Print(msg)
			monitor.LogToFile("./data", msg)
		}
	}
}
