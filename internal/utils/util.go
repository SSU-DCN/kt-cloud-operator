package utils

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"

// RandomString generates a random string of a given length.
func RandomString(length int) string {
	if length <= 0 {
		return ""
	}
	b := make([]byte, length)
	for i := range b {
		randomByte, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err) // Handle error gracefully in real scenarios
		}
		b[i] = charset[randomByte.Int64()]
	}
	return string(b)
}
