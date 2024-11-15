package main

import (
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
	log.Println("Starting Solana Wallet Monitor...")

	cfg := &config.Config{
		NetworkURL: rpc.MainNetBeta_RPC,
		Wallets:    config.Wallets,
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Monitoring %d wallets on Solana mainnet", len(cfg.Wallets))

	walletMonitor, err := monitor.NewWalletMonitor(cfg.NetworkURL, cfg.Wallets)
	if err != nil {
		log.Fatalf("Failed to create wallet monitor: %v", err)
	}

	storage := storage.New("./data")

	// Initial scan
	log.Println("Performing initial wallet scan...")
	results, err := walletMonitor.ScanAllWallets()
	if err != nil {
		log.Printf("Initial scan error: %v", err)
	} else {
		log.Printf("Initial scan complete. Found tokens in %d wallets", len(results))
		// Save initial state
		if err := storage.SaveWalletData(results); err != nil {
			log.Printf("Error saving initial data: %v", err)
		}
	}

	// Setup interrupt handling
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	log.Println("Monitoring for changes (Ctrl+C to exit)...")

	for {
		select {
		case <-ticker.C:
			newResults, err := walletMonitor.ScanAllWallets()
			if err != nil {
				log.Printf("Error scanning wallets: %v", err)
				continue
			}

			// Compare and log only the changes
			for wallet, data := range newResults {
				log.Printf("Wallet %s: Found %d token(s)", 
					wallet, len(data.TokenAccounts))
				
				for mint, token := range data.TokenAccounts {
					log.Printf("\tToken %s: Balance %d", 
						mint, token.Balance)
				}
			}

			if err := storage.SaveWalletData(newResults); err != nil {
				log.Printf("Error saving data: %v", err)
			}

		case <-interrupt:
			log.Println("Shutting down gracefully...")
			return
		}
	}
}
