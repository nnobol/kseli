package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"kseli-server/common"
	"kseli-server/config"
)

type contextKey string

const (
	ParticipantClaimsKey    contextKey = "claims"
	ParticipantSessionIDKey contextKey = "sessionId"
)

type Claims struct {
	UserID   uint8       `json:"userId"`
	Username string      `json:"username"`
	Role     common.Role `json:"role"`
	RoomID   string      `json:"roomId"`
	Exp      int64       `json:"exp"`
}

// Precomputed Base64-encoded JWT header: {"alg":"HS256","typ":"JWT"}
const header = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"

func CreateToken(claims Claims) (string, error) {
	secretKey := config.SecretKey

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", errors.New("failed to marshal claims")
	}

	payload := base64URLEncode(payloadBytes)

	unsignedToken := header + "." + payload

	signatureBytes := signHMACSHA256(unsignedToken, secretKey)

	signature := base64URLEncode(signatureBytes)

	return unsignedToken + "." + signature, nil
}

func ValidateToken(token string) (Claims, error) {
	var claims Claims
	secretKey := config.SecretKey

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims, errors.New("invalid token format")
	}

	unsignedToken, signatureB64 := parts[0]+"."+parts[1], parts[2]

	expectedSignature := signHMACSHA256(unsignedToken, secretKey)

	signatureBytes, err := base64URLDecode(signatureB64)
	if err != nil || !hmac.Equal(expectedSignature, signatureBytes) {
		return claims, errors.New("signature mismatch")
	}

	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return claims, errors.New("invalid payload encoding")
	}

	// Unmarshal the payload into a temporary struct to handle float64
	var temp struct {
		UserID   float64 `json:"userId"`
		Username string  `json:"username"`
		Role     float64 `json:"role"`
		RoomID   string  `json:"roomId"`
		Exp      int64   `json:"exp"`
	}

	if err := json.Unmarshal(payloadBytes, &temp); err != nil {
		return claims, errors.New("failed to parse claims")
	}

	// Convert UserID & Role from float64 to uint8
	claims = Claims{
		UserID:   uint8(temp.UserID),
		Username: temp.Username,
		Role:     common.Role(uint8(temp.Role)),
		RoomID:   temp.RoomID,
		Exp:      temp.Exp,
	}

	if time.Now().Unix() > claims.Exp {
		return claims, errors.New("token has expired")
	}

	return claims, nil
}

func signHMACSHA256(message, secretKey string) []byte {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
