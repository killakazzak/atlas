package auth

import "github.com/golang-jwt/jwt/v5"

// Claims carries the application-level fields embedded in a JWT.
type Claims struct {
	UserID string `json:"uid"`
	Login  string `json:"login"`
	Role   Role   `json:"role"`

	jwt.RegisteredClaims
}
