package string

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func RandomHexString(length int) (string, error) {
	if length%2 != 0 {
		return "", fmt.Errorf("length must be an even number")
	}

	bytes := make([]byte, length/2)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	return hex.EncodeToString(bytes), nil
}
