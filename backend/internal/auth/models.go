// Package auth defines the auth domain model and persistence ports.
// It contains entities and repository interfaces only — no storage or HTTP.
package auth

import "time"

// Role controls what actions a user is permitted to perform.
type Role string

// Known Role values.
const (
	RoleAdministrator Role = "administrator"
	RoleOperator      Role = "operator"
	RoleViewer        Role = "viewer"
)

// User is an authenticated principal that can interact with Atlas.
type User struct {
	ID           string
	Username     string
	Email        string
	Role         Role
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
