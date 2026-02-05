package models

import "time"

// User represents a user in the system
type User struct {
	ID        string
	Email     string
	Name      string
	OIDCSub   *string   // OIDC subject identifier (unique per provider)
	CreatedAt time.Time
	UpdatedAt time.Time
}
