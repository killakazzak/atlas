package auth

import (
	"errors"
	"testing"
	"time"
)

func testJWTService(ttl time.Duration) *JWTService {
	return NewJWTService(JWTConfig{
		Secret: []byte("test-secret"),
		Issuer: "atlas-test",
		TTL:    ttl,
	})
}

func testUser() *User {
	return &User{
		ID:       "user-1",
		Username: "alice",
		Role:     RoleOperator,
	}
}

func TestJWT_GenerateAndParse(t *testing.T) {
	t.Parallel()

	svc := testJWTService(time.Hour)
	u := testUser()

	token, err := svc.Generate(u)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if token == "" {
		t.Fatal("Generate() returned empty token")
	}

	claims, err := svc.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if claims.UserID != u.ID {
		t.Errorf("UserID = %q, want %q", claims.UserID, u.ID)
	}
	if claims.Login != u.Username {
		t.Errorf("Login = %q, want %q", claims.Login, u.Username)
	}
	if claims.Role != u.Role {
		t.Errorf("Role = %q, want %q", claims.Role, u.Role)
	}
	if claims.Issuer != "atlas-test" {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, "atlas-test")
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	t.Parallel()

	svc := testJWTService(-time.Second) // TTL in the past
	token, err := svc.Generate(testUser())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = svc.Parse(token)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for expired token, got %v", err)
	}
}

func TestJWT_WrongSignature(t *testing.T) {
	t.Parallel()

	svcA := testJWTService(time.Hour)
	svcB := NewJWTService(JWTConfig{
		Secret: []byte("different-secret"),
		Issuer: "atlas-test",
		TTL:    time.Hour,
	})

	token, err := svcA.Generate(testUser())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = svcB.Parse(token)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong signature, got %v", err)
	}
}

func TestJWT_CorruptedToken(t *testing.T) {
	t.Parallel()

	svc := testJWTService(time.Hour)
	_, err := svc.Parse("this.is.not.a.jwt")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for corrupted token, got %v", err)
	}
}

func TestJWT_WrongAlgorithm(t *testing.T) {
	t.Parallel()

	// A token hand-crafted with "alg":"none" — the library must reject it.
	// Base64url({"alg":"none","typ":"JWT"}).Base64url({"sub":"x"}).
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ4In0."

	svc := testJWTService(time.Hour)
	_, err := svc.Parse(noneToken)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for alg=none token, got %v", err)
	}
}
