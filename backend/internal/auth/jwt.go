package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService generates and parses signed access tokens.
type TokenService interface {
	Generate(user *User) (string, error)
	Parse(token string) (*Claims, error)
}

// JWTConfig holds the parameters needed to sign and verify tokens.
type JWTConfig struct {
	Secret []byte
	Issuer string
	TTL    time.Duration
}

// JWTService implements TokenService using HMAC-SHA256 signed JWTs.
type JWTService struct {
	cfg JWTConfig
}

// NewJWTService constructs a JWTService from the given config.
func NewJWTService(cfg JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

// Generate signs a new access token for user.
func (s *JWTService) Generate(user *User) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: user.ID,
		Login:  user.Username,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.TTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.cfg.Secret)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

// Parse validates the token signature and expiry, then returns its claims.
// Returns ErrInvalidCredentials for any validation failure so callers need not
// inspect jwt-library errors directly.
func (s *JWTService) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.cfg.Secret, nil
	}, jwt.WithIssuedAt(), jwt.WithExpirationRequired())

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("%w: token expired", ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalidCredentials, err)
	}

	if !token.Valid {
		return nil, ErrInvalidCredentials
	}
	return claims, nil
}
