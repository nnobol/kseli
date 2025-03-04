package models

type claimsContextKey string

const UserClaimsKey claimsContextKey = "claims"

type Claims struct {
	UserID   uint8  `json:"userId"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
	RoomID   string `json:"roomId"`
	Exp      int64  `json:"exp"`
}
