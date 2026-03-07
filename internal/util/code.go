package util

import (
	"crypto/rand"
	"fmt"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	out := make([]byte, length)
	for i, b := range buf {
		out[i] = base62Chars[int(b)%len(base62Chars)]
	}
	return string(out), nil
}
