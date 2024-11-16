package utils

import (
	"fmt"
	"math"
)

// FormatTokenAmount formats a token amount with appropriate suffixes (K, M) and decimals
func FormatTokenAmount(amount uint64, decimals uint8) string {
	if decimals == 0 {
		return fmt.Sprintf("%d", amount)
	}

	// Convert to float64 and divide by 10^decimals
	divisor := math.Pow(10, float64(decimals))
	value := float64(amount) / divisor

	// Format with appropriate decimal places
	if value >= 1000000 {
		// Use millions format: 1.23M
		return fmt.Sprintf("%.2fM", value/1000000)
	} else if value >= 1000 {
		// Use thousands format: 1.23K
		return fmt.Sprintf("%.2fK", value/1000)
	}

	// Use standard format with max 4 decimal places
	return fmt.Sprintf("%.4f", value)
}
