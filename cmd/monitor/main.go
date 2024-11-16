package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
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
	
	// Create buffered channels for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Track connection state
	var lastSuccessfulScan time.Time
	var connectionLost bool
	
	// Define maximum allowed time between successful scans
	maxTimeBetweenScans := scanInterval * 3

	// Perform initial scan immediately
	log.Println("Performing initial wallet scan...")
	initialResults, err := scanner.ScanAllWallets()
	if err != nil {
		log.Printf("Warning: initial scan had errors: %v", err)
	} else {
		if err := storage.SaveWalletData(initialResults); err != nil {
			log.Printf("Error saving initial data: %v", err)
		}
		lastSuccessfulScan = time.Now()
		log.Printf("Initial scan complete. Found data for %d wallets", len(initialResults))
	}
	
	// Start monitoring in a separate goroutine
	go func() {
		ticker := time.NewTicker(scanInterval)
		defer ticker.Stop()

		// Load previous data for comparison
		previousData, err := storage.LoadWalletData()
		if err != nil {
			log.Printf("Warning: could not load previous data: %v", err)
			previousData = make(map[string]*monitor.WalletData)
		}

		log.Printf("Starting monitoring loop with %v interval...", scanInterval)

		for {
			select {
			case <-ticker.C:
				// Check if we've exceeded the maximum time between scans
				if time.Since(lastSuccessfulScan) > maxTimeBetweenScans && !connectionLost {
					connectionLost = true
					log.Printf("No successful scan in %v, marking connection as lost", maxTimeBetweenScans)
					continue
				}

				newResults, err := scanner.ScanAllWallets()
				if err != nil {
					log.Printf("Error scanning wallets: %v", err)
					if !connectionLost {
						connectionLost = true
						log.Println("Connection appears to be lost, will suppress alerts until restored")
					}
					continue
				}

				// Connection restored check
				if connectionLost {
					connectionLost = false
					log.Println("Connection restored, loading previous data to prevent false alerts")
					if savedData, err := storage.LoadWalletData(); err == nil {
						previousData = savedData
					}
					lastSuccessfulScan = time.Now()
					continue
				}

				// Update last successful scan time
				lastSuccessfulScan = time.Now()

				// Process changes only if we have previous data and connection is stable
				if len(previousData) > 0 {
					changes := monitor.DetectChanges(previousData, newResults, cfg.Alerts.SignificantChange)
					processChanges(changes, alerter, cfg.Alerts)
				}

				// Save new results
				if err := storage.SaveWalletData(newResults); err != nil {
					log.Printf("Error saving data: %v", err)
				}
				previousData = newResults

			case <-done:
				log.Println("Monitoring loop stopped")
				return
			}
		}
	}()

	// Wait for interrupt signal
	<-interrupt
	log.Println("Shutting down gracefully...")
	monitor.LogToFile("./data", "Monitor shutting down gracefully")
	done <- true
	time.Sleep(time.Second) // Give a moment for final cleanup
}

func processChanges(changes []monitor.Change, alerter alerts.Alerter, alertCfg config.AlertConfig) {
	for _, change := range changes {
		var msg string
		var level alerts.AlertLevel
		var alertData map[string]interface{}

		switch change.ChangeType {
		case "new_wallet":
			// Create a consolidated message for all tokens
			var tokenDetails []string
			tokenData := make(map[string]uint64)
			tokenDecimals := make(map[string]uint8)
			for mint, balance := range change.TokenBalances {
				tokenDetails = append(tokenDetails, fmt.Sprintf("%s: %d", mint, balance))
				tokenData[mint] = balance
				tokenDecimals[mint] = 9 // Default decimals, adjust if you have actual decimals
			}
			msg = fmt.Sprintf("New wallet %s detected with %d tokens:\n%s",
				change.WalletAddress,
				len(change.TokenBalances),
				strings.Join(tokenDetails, "\n"))
			level = alerts.Warning
			alertData = map[string]interface{}{
				"token_balances": tokenData,
				"token_decimals": tokenDecimals,
			}

		case "new_token":
			msg = fmt.Sprintf("New token %s (%s) detected in wallet with initial balance %d",
				change.TokenSymbol, change.TokenMint, change.NewBalance)
			level = alerts.Warning
			alertData = map[string]interface{}{
				"balance":  change.NewBalance,
				"decimals": change.TokenDecimals,
				"symbol":   change.TokenSymbol,
			}

		case "balance_change":
			msg = fmt.Sprintf("Balance change for %s (%s): from %d to %d (%.2f%%)",
				change.TokenSymbol, change.TokenMint,
				change.OldBalance, change.NewBalance, change.ChangePercent)
			
			absChange := abs(change.ChangePercent)
			switch {
			case absChange >= (alertCfg.SignificantChange * 5):
				level = alerts.Critical
			case absChange >= (alertCfg.SignificantChange * 2):
				level = alerts.Warning
			default:
				level = alerts.Info
			}
			
			alertData = map[string]interface{}{
				"old_balance":    change.OldBalance,
				"new_balance":    change.NewBalance,
				"decimals":       change.TokenDecimals,
				"symbol":         change.TokenSymbol,
				"change_percent": change.ChangePercent,
			}
		}

		if level >= alerts.Warning {
			alert := alerts.Alert{
				Timestamp:     time.Now(),
				WalletAddress: change.WalletAddress,
				TokenMint:     change.TokenMint,
				AlertType:     change.ChangeType,
				Message:       msg,
				Level:        level,
				Data:         alertData,
			}

			if err := alerter.SendAlert(alert); err != nil {
				log.Printf("failed to send alert: %v", err)
			}
			monitor.LogToFile("./data", msg)
		} else {
			monitor.LogToFile("./data", fmt.Sprintf("INFO: %s", msg))
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
