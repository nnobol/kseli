package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"kseli-server/models"
)

var CreateTokenFunc = createToken

type Claims struct {
	UserID string      `json:"userId"`
	RoomID string      `json:"roomId"`
	Role   models.Role `json:"role"`
	Exp    int64       `json:"exp"`
}

var header = base64URLEncode([]byte(`{"alg":"HS256","typ":"JWT"}`))

// createToken creates a JWT string signed with HS256.
func createToken(claims Claims, secretKey string) (string, error) {
	// 1. Marshal the claims to JSON
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// 2. Base64URL-encode the payload
	payload := base64URLEncode(payloadBytes)

	// 3. Concatenate header + "." + payload
	unsignedToken := header + "." + payload

	// 4. Sign with HMAC-SHA256
	signatureBytes := signHMACSHA256(unsignedToken, secretKey)

	// 5. Base64URL-encode the signature
	signature := base64URLEncode(signatureBytes)

	// 6. Return "header.payload.signature" as a token string
	return unsignedToken + "." + signature, nil
}

func ValidateToken(token string, secretKey string) (Claims, error) {
	var claims Claims

	// 1. Split the token into 3 parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims, errors.New("invalid token format")
	}

	headerB64 := parts[0]
	payloadB64 := parts[1]
	signatureB64 := parts[2]

	// 2. Recompute the signature from header + payload
	unsignedToken := headerB64 + "." + payloadB64
	expectedSig := signHMACSHA256(unsignedToken, secretKey)

	// 3. Base64URL-decode the signature from the token
	signatureBytes, err := base64URLDecode(signatureB64)
	if err != nil {
		return claims, errors.New("invalid signature encoding")
	}

	// 4. Compare the recomputed signature with the token’s signature
	if !hmac.Equal(expectedSig, signatureBytes) {
		return claims, errors.New("signature mismatch")
	}

	// 5. Decode the payload into the Claims struct
	payloadBytes, err := base64URLDecode(payloadB64)
	if err != nil {
		return claims, errors.New("invalid payload encoding")
	}

	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return claims, errors.New("failed to parse claims")
	}

	// 6. Check the expiration time
	if time.Now().Unix() > claims.Exp {
		return claims, errors.New("token has expired")
	}

	return claims, nil
}

// signHMACSHA256 returns the HMAC-SHA256 signature of a message.
func signHMACSHA256(message, secretKey string) []byte {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

// base64URLEncode encodes a byte slice to a Base64 URL-encoded string without padding.
func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

// base64URLDecode decodes a Base64 URL-encoded string (no padding).
func base64URLDecode(s string) ([]byte, error) {
	// Re-pad the string if necessary
	padding := (4 - len(s)%4) % 4
	s += strings.Repeat("=", padding)

	return base64.URLEncoding.DecodeString(s)
}
