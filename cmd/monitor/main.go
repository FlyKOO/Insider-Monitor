package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/accursedgalaxy/insider-monitor/internal/alerts"
	"github.com/accursedgalaxy/insider-monitor/internal/config"
	"github.com/accursedgalaxy/insider-monitor/internal/monitor"
	"github.com/accursedgalaxy/insider-monitor/internal/storage"
	"github.com/accursedgalaxy/insider-monitor/internal/utils"
)

// WalletScanner 接口定义了钱包监控的约定
type WalletScanner interface {
	ScanAllWallets() (map[string]*monitor.WalletData, error)
	DisplayWalletOverview(walletDataMap map[string]*monitor.WalletData)
}

func main() {
	// 创建自定义日志记录器
	logger := utils.NewLogger(false)

	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// 打印欢迎信息
	fmt.Printf("\n%s%s SOLANA INSIDER MONITOR %s\n", utils.ColorBold, utils.ColorPurple, utils.ColorReset)
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n\n", utils.ColorPurple, utils.ColorReset)

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Fatal("Configuration file not found: %v\n\n"+
				"💡 Quick fix:\n"+
				"   1. Copy the example: cp config.example.json config.json\n"+
				"   2. Edit config.json with your settings\n"+
				"   3. Get a free RPC endpoint from:\n"+
				"      • Helius: https://helius.dev\n"+
				"      • QuickNode: https://quicknode.com\n"+
				"      • Triton: https://triton.one", err)
		}
		logger.Fatal("Failed to load config: %v\n\n"+
			"💡 Check that your config.json file has valid JSON syntax.\n"+
			"   You can validate it at https://jsonlint.com/", err)
	}

	if err := cfg.Validate(); err != nil {
		logger.Fatal("Configuration validation failed:\n%v", err)
	}

	// 初始化扫描器
	scanner, err := monitor.NewWalletMonitor(cfg.NetworkURL, cfg.Wallets, &cfg.Scan)
	if err != nil {
		logger.Fatal("Failed to create wallet monitor: %v\n\n"+
			"💡 This usually means:\n"+
			"   • Invalid wallet address format in config.json\n"+
			"   • Network connectivity issues\n"+
			"   • RPC endpoint problems\n\n"+
			"   Verify your wallet addresses are valid Solana addresses.", err)
	}

	// 初始化告警器
	var alerter alerts.Alerter
	if cfg.Discord.Enabled {
		alerter = alerts.NewDiscordAlerter(cfg.Discord.WebhookURL, cfg.Discord.ChannelID)
		logger.Config("Discord alerts enabled")
	} else {
		alerter = &alerts.ConsoleAlerter{}
		logger.Config("Console alerts enabled")
	}

	// 解析扫描间隔
	scanInterval, err := time.ParseDuration(cfg.ScanInterval)
	if err != nil {
		logger.Warning("Invalid scan interval '%s', using default of 1 minute", cfg.ScanInterval)
		scanInterval = time.Minute
	}

	runMonitor(scanner, alerter, cfg, scanInterval, logger)
}

func runMonitor(scanner WalletScanner, alerter alerts.Alerter, cfg *config.Config, scanInterval time.Duration, logger *utils.Logger) {
	storage := storage.New("./data")

	// 创建缓冲通道以便优雅关闭
	interrupt := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// 跟踪连接状态
	var lastSuccessfulScan time.Time
	var connectionLost bool

	// 定义成功扫描之间允许的最大时间
	maxTimeBetweenScans := scanInterval * 3

	// 在启动时从存储初始化 previousData
	var previousData map[string]*monitor.WalletData
	if savedData, err := storage.LoadWalletData(); err == nil {
		previousData = savedData
		logger.Storage("Loaded previous wallet data from storage")
	} else {
		logger.Warning("Could not load previous data: %v. Will initialize after first scan.", err)
		previousData = make(map[string]*monitor.WalletData)
	}

	// 立即执行初始扫描
	logger.Scan("Performing initial wallet scan...")
	initialResults, err := scanner.ScanAllWallets()
	if err != nil {
		logger.Error("Initial scan failed: %v", err)
		logger.Error("\n💡 Common solutions:")
		logger.Error("   • Check your internet connection")
		logger.Error("   • Verify your RPC endpoint is working")
		logger.Error("   • Ensure wallet addresses in config.json are valid")
		logger.Error("   • Try a different RPC provider if rate limited")
		logger.Error("\nThe monitor will continue trying in the background...")
	} else {
		if err := storage.SaveWalletData(initialResults); err != nil {
			logger.Error("Error saving initial data: %v", err)
		}
		lastSuccessfulScan = time.Now()
		logger.Success("Initial scan complete. Found data for %d wallets", len(initialResults))
		scanner.DisplayWalletOverview(initialResults)
	}

	// 在单独的 goroutine 中开始监控
	go func() {
		ticker := time.NewTicker(scanInterval)
		defer ticker.Stop()

		logger.Info("Starting monitoring loop with %v interval...", scanInterval)

		for {
			select {
			case <-ticker.C:
				// 检查是否超过了扫描之间的最大允许时间
				if time.Since(lastSuccessfulScan) > maxTimeBetweenScans && !connectionLost {
					connectionLost = true
					logger.Warning("No successful scan in %v, marking connection as lost", maxTimeBetweenScans)
					continue
				}

				newResults, err := scanner.ScanAllWallets()
				if err != nil {
					logger.Error("Error scanning wallets: %v", err)
					if !connectionLost {
						connectionLost = true
						logger.Network("Connection appears to be lost, will suppress alerts until restored")
					}
					continue
				}

				// 检查连接是否已恢复
				if connectionLost {
					connectionLost = false
					logger.Network("Connection restored, loading previous data to prevent false alerts")
					if savedData, err := storage.LoadWalletData(); err == nil {
						previousData = savedData
					}
					lastSuccessfulScan = time.Now()
					continue
				}

				// 更新上次成功扫描时间
				lastSuccessfulScan = time.Now()

				// 仅在存在历史数据时处理变化
				if len(previousData) > 0 {
					changes := monitor.DetectChanges(previousData, newResults, cfg.Alerts.SignificantChange)
					processChanges(changes, alerter, cfg.Alerts, logger)
				} else {
					// 第一次扫描，仅存储数据而不生成告警
					logger.Info("Initial scan completed, storing baseline data")
				}

				// 保存新的结果
				if err := storage.SaveWalletData(newResults); err != nil {
					logger.Error("Error saving data: %v", err)
				}
				previousData = newResults

				// 展示钱包概览
				scanner.DisplayWalletOverview(newResults)

			case <-done:
				logger.Info("Monitoring loop stopped")
				return
			}
		}
	}()

	// 等待中断信号
	<-interrupt
	logger.Info("Shutting down gracefully...")
	if err := monitor.LogToFile("./data", "Monitor shutting down gracefully"); err != nil {
		logger.Error("Failed to write shutdown log: %v", err)
	}
	done <- true
	time.Sleep(time.Second) // 留出一点时间用于最后清理
}

func processChanges(changes []monitor.Change, alerter alerts.Alerter, alertCfg config.AlertConfig, logger *utils.Logger) {
	for _, change := range changes {
		var msg string
		var level alerts.AlertLevel
		var alertData map[string]interface{}

		switch change.ChangeType {
		case "new_wallet":
			// 为所有代币创建汇总消息
			var tokenDetails []string
			tokenData := make(map[string]uint64)
			tokenDecimals := make(map[string]uint8)
			for mint, balance := range change.TokenBalances {
				tokenDetails = append(tokenDetails, fmt.Sprintf("%s: %d", mint, balance))
				tokenData[mint] = balance
				tokenDecimals[mint] = 9 // 默认小数位；如有实际小数位请调整
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
				Level:         level,
				Data:          alertData,
			}

			if err := alerter.SendAlert(alert); err != nil {
				logger.Error("Failed to send alert: %v", err)
			}
		} else {
			logger.Info(msg)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
