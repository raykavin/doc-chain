package utils

import (
	"math/big"
)

// NewBigInt creates a new big.Int from an int64
func NewBigInt(value int64) *big.Int {
	return big.NewInt(value)
}

// StringArrayContains checks if a string is in a string array
func StringArrayContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}