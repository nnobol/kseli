package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

var GenerateRandomIDFunc = generateRandomID
var GenerateRoomIDFunc = generateRoomID

func generateRandomID() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}

func generateRoomID() string {
	randomBytes := make([]byte, 6)
	rand.Read(randomBytes)
	return base64.RawURLEncoding.EncodeToString(randomBytes)
}
