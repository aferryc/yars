package utils

import (
	"fmt"
	"math"
)

// ConvertCurrency converts an amount from one currency to another based on the given exchange rate.
func ConvertCurrency(amount float64, exchangeRate float64) float64 {
	return math.Round(amount*exchangeRate*100) / 100
}

// FormatCurrency formats a float64 amount into a string representation with two decimal places.
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}