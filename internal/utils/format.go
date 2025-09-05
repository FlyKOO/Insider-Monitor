package utils

import (
	"fmt"
	"math"
)

// FormatTokenAmount 将代币数量格式化为带适当后缀（K、M）及小数的字符串
func FormatTokenAmount(amount uint64, decimals uint8) string {
	if decimals == 0 {
		return fmt.Sprintf("%d", amount)
	}

	// 转换为 float64 并除以 10^decimals
	divisor := math.Pow(10, float64(decimals))
	value := float64(amount) / divisor

	// 根据数值大小格式化小数位
	if value >= 1000000 {
		// 使用百万格式：1.23M
		return fmt.Sprintf("%.2fM", value/1000000)
	} else if value >= 1000 {
		// 使用千位格式：1.23K
		return fmt.Sprintf("%.2fK", value/1000)
	}

	// 使用标准格式，最多保留 4 位小数
	return fmt.Sprintf("%.4f", value)
}
