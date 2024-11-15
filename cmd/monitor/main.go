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

func main() {
	testMode := flag.Bool("test", false, "Run in test mode with accelerated scanning")
	flag.Parse()

	scanInterval := time.Minute
	var walletMonitor interface {
		ScanAllWallets() (map[string]*monitor.WalletData, error)
	}

	if *testMode {
		scanInterval = time.Second * 5
		log.Println("Running in test mode with 5-second scan interval")
		walletMonitor = monitor.NewMockWalletMonitor()
	} else {
		cfg := &config.Config{
			NetworkURL: rpc.MainNetBeta_RPC,
			Wallets:    config.Wallets,
		}
		if err := cfg.Validate(); err != nil {
			log.Fatalf("Invalid configuration: %v", err)
		}
		var err error
		walletMonitor, err = monitor.NewWalletMonitor(cfg.NetworkURL, cfg.Wallets)
		if err != nil {
			log.Fatalf("Failed to create wallet monitor: %v", err)
		}
	}

	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	log.Println("Starting Solana Wallet Monitor...")

	storage := storage.New("./data")

	// Load previous data for comparison
	previousData, err := storage.LoadWalletData()
	if err != nil {
		log.Printf("Warning: Could not load previous data: %v", err)
	}

	// Initial scan
	log.Println("Performing initial wallet scan...")
	results, err := walletMonitor.ScanAllWallets()
	if err != nil {
		log.Printf("Initial scan error: %v", err)
	} else {
		log.Printf("Initial scan complete. Found tokens in %d wallets", len(results))
		if err := storage.SaveWalletData(results); err != nil {
			log.Printf("Error saving initial data: %v", err)
		}
		previousData = results
	}

	// Setup interrupt handling
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	log.Println("Monitoring for changes (Ctrl+C to exit)...")

	for {
		select {
		case <-ticker.C:
			newResults, err := walletMonitor.ScanAllWallets()
			if err != nil {
				log.Printf("Error scanning wallets: %v", err)
				continue
			}

			// Use DetectChanges to find changes
			changes := monitor.DetectChanges(previousData, newResults)
			
			// Log any changes
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

			// Save new state and update previous data
			if err := storage.SaveWalletData(newResults); err != nil {
				log.Printf("Error saving data: %v", err)
			}
			previousData = newResults

		case <-interrupt:
			log.Println("Shutting down gracefully...")
			monitor.LogToFile("./data", "Monitor shutting down gracefully")
			return
		}
	}
}
