package base62

import (
	"strings"
)

const (
	// Base62 alphabet uses 0-9, A-Z, a-z (62 characters total)
	alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base     = int64(len(alphabet))
)

// ToBase62 converts an integer to a base62 string
func ToBase62(num int64) string {
	if num == 0 {
		return "0"
	}

	var result strings.Builder
	for num > 0 {
		result.WriteByte(alphabet[num%base])
		num /= base
	}

	// Reverse the string since we built it backwards
	s := result.String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// FromBase62 converts a base62 string back to an integer
func FromBase62(s string) int64 {
	var result int64
	for _, char := range s {
		result = result*base + int64(strings.IndexRune(alphabet, char))
	}
	return result
}