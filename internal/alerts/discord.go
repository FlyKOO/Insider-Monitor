package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/accursedgalaxy/insider-monitor/internal/utils"
)

type DiscordAlerter struct {
	WebhookURL string
	ChannelID  string
}

type discordMessage struct {
	Content   string  `json:"content"`
	Username  string  `json:"username,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	Embeds    []embed `json:"embeds,omitempty"`
}

type embed struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Color       int     `json:"color"` // 颜色代码
	Fields      []field `json:"fields,omitempty"`
}

type field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

func NewDiscordAlerter(webhookURL, channelID string) *DiscordAlerter {
	return &DiscordAlerter{
		WebhookURL: webhookURL,
		ChannelID:  channelID,
	}
}

func (d *DiscordAlerter) SendAlert(alert Alert) error {
	color := 0x7289DA // Discord 默认蓝色
	switch alert.Level {
	case Critical:
		color = 0xFF0000 // 红色
	case Warning:
		color = 0xFFA500 // 橙色
	}

	// 从 Data 映射中安全地获取值
	safeGet := func(key string) interface{} {
		if alert.Data == nil {
			return nil
		}
		return alert.Data[key]
	}

	var description string
	var fields []field

	switch alert.AlertType {
	case "balance_change":
		if oldBal, ok := safeGet("old_balance").(uint64); ok {
			if newBal, ok := safeGet("new_balance").(uint64); ok {
				if decimals, ok := safeGet("decimals").(uint8); ok {
					oldFormatted := utils.FormatTokenAmount(oldBal, decimals)
					newFormatted := utils.FormatTokenAmount(newBal, decimals)
					symbol := safeGet("symbol").(string)
					changePercent := safeGet("change_percent").(float64)

					description = fmt.Sprintf("```diff\n- Old: %s\n+ New: %s\nChange: %+.2f%%```",
						oldFormatted,
						newFormatted,
						changePercent)

					// 作为字段添加代币的详细信息
					fields = append(fields, field{
						Name: "Token",
						Value: fmt.Sprintf("%s\n`%s`",
							symbol,
							alert.TokenMint),
						Inline: false,
					})
				}
			}
		}

	case "new_token":
		if balance, ok := safeGet("balance").(uint64); ok {
			if decimals, ok := safeGet("decimals").(uint8); ok {
				formatted := utils.FormatTokenAmount(balance, decimals)
				symbol := safeGet("symbol").(string)
				description = fmt.Sprintf("```ini\n[Initial Balance]\n%s```",
					formatted)

				// 作为字段添加代币的详细信息
				fields = append(fields, field{
					Name: "Token",
					Value: fmt.Sprintf("%s\n`%s`",
						symbol,
						alert.TokenMint),
					Inline: false,
				})
			}
		}
	}

	// 若生成描述失败，则使用备用内容
	if description == "" {
		description = fmt.Sprintf("```%s```", alert.Message)
	}

	// 将钱包地址作为一个字段
	fields = append(fields, field{
		Name:   "Wallet",
		Value:  fmt.Sprintf("`%s`", alert.WalletAddress),
		Inline: false,
	})

	// 添加时间戳
	fields = append(fields, field{
		Name:   "Time",
		Value:  alert.Timestamp.Format("2006-01-02 15:04:05 MST"),
		Inline: true,
	})

	msg := discordMessage{
		Username: "Solana Wallet Monitor",
		Embeds: []embed{{
			Title:       fmt.Sprintf("%s Alert", strings.ToUpper(alert.AlertType)),
			Description: description,
			Color:       color,
			Fields:      fields,
		}},
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal discord message: %w", err)
	}

	log.Printf("Sending Discord alert: %s", string(payload))

	resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send discord message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API returned error status: %d, body: %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully sent Discord alert (status: %d)", resp.StatusCode)
	return nil
}
