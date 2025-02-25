package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"kseli-server/config"
	"kseli-server/models"
)

// Precomputed Base64-encoded JWT header: {"alg":"HS256","typ":"JWT"}
const header = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"

func CreateToken(claims models.Claims) (string, error) {
	secretKey := config.GlobalConfig.SecretKey

	// Step 1: Serialize claims to JSON
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", errors.New("failed to marshal claims")
	}

	// Step 2: Encode the payload to Base64 URL format
	payload := base64URLEncode(payloadBytes)

	// Step 3: Construct the unsigned token (header + payload)
	unsignedToken := header + "." + payload

	// Step 4: Compute the HMAC-SHA256 signature of the token
	signatureBytes := signHMACSHA256(unsignedToken, secretKey)

	// Step 5: Encode the signature in Base64 URL format
	signature := base64URLEncode(signatureBytes)

	// Step 6: Return the final JWT token (header.payload.signature)
	return unsignedToken + "." + signature, nil
}

func ValidateToken(token string) (models.Claims, error) {
	var claims models.Claims
	secretKey := config.GlobalConfig.SecretKey

	// Step 1: Split the token into header, payload, and signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims, errors.New("invalid token format")
	}

	unsignedToken, signatureB64 := parts[0]+"."+parts[1], parts[2]

	// Step 2: Compute the expected signature for verification
	expectedSignature := signHMACSHA256(unsignedToken, secretKey)

	// Step 3: Decode and compare the signature
	signatureBytes, err := base64URLDecode(signatureB64)
	if err != nil || !hmac.Equal(expectedSignature, signatureBytes) {
		return claims, errors.New("signature mismatch")
	}

	// Step 4: Decode the Base64-encoded payload
	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return claims, errors.New("invalid payload encoding")
	}

	// Step 5: Unmarshal the payload into a temporary struct to handle float64 issue
	var temp struct {
		UserID float64 `json:"userId"`
		RoomID string  `json:"roomId"`
		Role   float64 `json:"role"`
		Exp    int64   `json:"exp"`
	}

	if err := json.Unmarshal(payloadBytes, &temp); err != nil {
		return claims, errors.New("failed to parse claims")
	}

	// Step 6: Convert UserID & Role from float64 to uint8
	claims = models.Claims{
		UserID: uint8(temp.UserID),
		RoomID: temp.RoomID,
		Role:   models.Role(uint8(temp.Role)),
		Exp:    temp.Exp,
	}

	// Step 7: Check if the token has expired
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
