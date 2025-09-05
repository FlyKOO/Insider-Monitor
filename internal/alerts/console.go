package alerts

import (
	"fmt"
	"strings"

	"github.com/accursedgalaxy/insider-monitor/internal/utils"
)

// ConsoleAlerter 实现基础的控制台日志输出
type ConsoleAlerter struct{}

func (a *ConsoleAlerter) SendAlert(alert Alert) error {
	// 根据告警级别使用不同颜色
	var color, symbol string
	switch alert.Level {
	case Critical:
		color = utils.ColorRed
		symbol = "🔴"
	case Warning:
		color = utils.ColorYellow
		symbol = "🟡"
	default:
		color = utils.ColorGreen
		symbol = "🟢"
	}

	// 格式化时间戳
	timestamp := alert.Timestamp.Format("15:04:05")

	// 格式化告警类型
	alertType := strings.ToUpper(alert.AlertType)
	if alertType == "BALANCE_CHANGE" {
		alertType = "BALANCE CHANGE"
	} else if alertType == "NEW_TOKEN" {
		alertType = "NEW TOKEN"
	} else if alertType == "NEW_WALLET" {
		alertType = "NEW WALLET"
	}

	// 为告警绘制框线
	width := 80
	topBorder := fmt.Sprintf("%s%s%s", color, strings.Repeat("━", width), utils.ColorReset)
	bottomBorder := topBorder

	// 输出告警头部
	fmt.Println(topBorder)
	fmt.Printf("%s%s [%s] %s ALERT - %s %s\n",
		color,
		symbol,
		timestamp,
		alertType,
		utils.ColorBold,
		utils.ColorReset)

	// 输出告警详情
	shortWallet := alert.WalletAddress
	if len(shortWallet) > 20 {
		shortWallet = shortWallet[:8] + "..." + shortWallet[len(shortWallet)-8:]
	}

	fmt.Printf("Wallet: %s%s%s\n", utils.ColorBold, shortWallet, utils.ColorReset)

	// 格式化消息内容
	lines := strings.Split(alert.Message, "\n")
	for _, line := range lines {
		fmt.Println(line)
	}

	// 如有相关的附加数据则输出
	if data, ok := alert.Data["change_percent"]; ok {
		if pct, ok := data.(float64); ok {
			direction := "↑"
			valueColor := utils.ColorGreen
			if pct < 0 {
				direction = "↓"
				valueColor = utils.ColorRed
			}
			fmt.Printf("Change: %s%s %.2f%%%s\n",
				valueColor,
				direction,
				pct,
				utils.ColorReset)
		}
	}

	fmt.Println(bottomBorder)

	return nil
}
