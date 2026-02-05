package models

import "time"

// APIKey represents an API key for authentication
type APIKey struct {
	ID             string
	UserID         *string    // Optional: can be service account key
	OrganizationID string
	Name           string     // Friendly name (e.g., "CI/CD Pipeline Key")
	KeyHash        string     // Bcrypt hash of the full key
	KeyPrefix      string     // First 8-10 chars for display (e.g., "tfr_abc123")
	Scopes         []string   // JSONB array: ["modules:read", "modules:write", "providers:write"]
	ExpiresAt      *time.Time // Optional expiration
	LastUsedAt     *time.Time // Track last usage
	CreatedAt      time.Time
}
