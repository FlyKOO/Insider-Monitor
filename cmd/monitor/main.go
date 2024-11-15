package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/accursedgalaxy/insider-monitor/internal/alerts"
	"github.com/accursedgalaxy/insider-monitor/internal/config"
	"github.com/accursedgalaxy/insider-monitor/internal/monitor"
	"github.com/accursedgalaxy/insider-monitor/internal/storage"
)

// WalletScanner interface defines the contract for wallet monitoring
type WalletScanner interface {
	ScanAllWallets() (map[string]*monitor.WalletData, error)
}

func main() {
	testMode := flag.Bool("test", false, "Run in test mode with accelerated scanning")
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	var cfg *config.Config
	var err error
	
	if *testMode {
		cfg = config.GetTestConfig()
		log.Println("Running in test mode with 5-second scan interval")
	} else {
		cfg, err = config.LoadConfig(*configPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	// Initialize scanner
	var scanner WalletScanner
	if *testMode {
		scanner = monitor.NewMockWalletMonitor()
	} else {
		scanner, err = monitor.NewWalletMonitor(cfg.NetworkURL, cfg.Wallets)
		if err != nil {
			log.Fatalf("failed to create wallet monitor: %v", err)
		}
	}

	// Initialize alerter
	var alerter alerts.Alerter
	if cfg.Discord.Enabled {
		alerter = alerts.NewDiscordAlerter(cfg.Discord.WebhookURL, cfg.Discord.ChannelID)
		log.Println("Discord alerts enabled")
	} else {
		alerter = &alerts.ConsoleAlerter{}
		log.Println("Console alerts enabled")
	}

	// Parse scan interval
	scanInterval, err := time.ParseDuration(cfg.ScanInterval)
	if err != nil {
		log.Printf("invalid scan interval '%s', using default of 1 minute", cfg.ScanInterval)
		scanInterval = time.Minute
	}

	runMonitor(scanner, alerter, cfg, scanInterval)
}

func runMonitor(scanner WalletScanner, alerter alerts.Alerter, cfg *config.Config, scanInterval time.Duration) {
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
		log.Printf("warning: initial scan had errors: %v", err)
	}
	
	if len(results) > 0 {
		log.Printf("Initial scan complete. Found tokens in %d wallets", len(results))
		if err := storage.SaveWalletData(results); err != nil {
			log.Printf("error saving initial data: %v", err)
		}
		previousData = results
	} else {
		log.Printf("Initial scan failed to find any wallet data, will retry on next scan")
	}

	// Setup monitoring loop
	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	log.Printf("Monitoring %d wallets with %s interval (Ctrl+C to exit)...", len(cfg.Wallets), scanInterval)

	for {
		select {
		case <-ticker.C:
			newResults, err := scanner.ScanAllWallets()
			if err != nil {
				log.Printf("error scanning wallets: %v", err)
				continue
			}

			changes := monitor.DetectChanges(previousData, newResults)
			processChanges(changes, alerter, cfg.Alerts)

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

func processChanges(changes []monitor.Change, alerter alerts.Alerter, alertCfg config.AlertConfig) {
	for _, change := range changes {
		// Skip if balance is below minimum threshold
		if change.NewBalance < alertCfg.MinimumBalance {
			continue
		}

		// Skip if token is in ignore list
		for _, ignoredToken := range alertCfg.IgnoreTokens {
			if change.TokenMint == ignoredToken {
				continue
			}
		}

		var (
			msg   string
			level alerts.AlertLevel
		)

		switch change.ChangeType {
		case "new_wallet":
			msg = fmt.Sprintf("New wallet %s: Token %s with balance %d",
				change.WalletAddress, change.TokenMint, change.NewBalance)
			level = alerts.Info
		case "new_token":
			msg = fmt.Sprintf("New token detected in %s: %s with balance %d",
				change.WalletAddress, change.TokenMint, change.NewBalance)
			level = alerts.Warning
		case "balance_change":
			// Calculate percentage change
			percentChange := float64(change.NewBalance-change.OldBalance) / float64(change.OldBalance)
			if abs(percentChange) < alertCfg.SignificantChange {
				continue
			}

			msg = fmt.Sprintf("Balance change in %s: Token %s from %d to %d (%.2f%%)",
				change.WalletAddress, change.TokenMint, change.OldBalance, change.NewBalance, percentChange*100)
			level = alerts.Warning
			if abs(percentChange) > 0.5 { // 50% change
				level = alerts.Critical
			}
		}

		if msg != "" {
			alert := alerts.Alert{
				Timestamp:     time.Now(),
				WalletAddress: change.WalletAddress,
				TokenMint:     change.TokenMint,
				AlertType:     change.ChangeType,
				Message:       msg,
				Level:        level,
			}

			if err := alerter.SendAlert(alert); err != nil {
				log.Printf("failed to send alert: %v", err)
			}
			monitor.LogToFile("./data", msg)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
