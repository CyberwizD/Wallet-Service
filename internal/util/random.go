package util

import (
	"crypto/rand"
	"fmt"
)

// RandomDigits returns a zero-padded string of n random digits.
func RandomDigits(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i := 0; i < n; i++ {
		bytes[i] = (bytes[i] % 10) + '0'
	}
	return string(bytes), nil
}

// GenerateAPIKey produces a Paystack-like service key.
func GenerateAPIKey() (string, error) {
	body, err := RandomDigits(24)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sk_live_%s", body), nil
}
