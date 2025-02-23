package models

type claimsContextKey string

const UserClaimsKey claimsContextKey = "claims"

type Claims struct {
	UserID uint8  `json:"userId"`
	RoomID string `json:"roomId"`
	Role   Role   `json:"role"`
	Exp    int64  `json:"exp"`
}
