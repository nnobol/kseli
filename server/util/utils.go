package util

import (
	"crypto/rand"
	"fmt"
	"io"
)

var GenerateRandomIDFunc = generateRandomID

func generateRandomID() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}
