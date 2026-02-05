package models

import "time"

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	OrganizationID string
	UserID         string
	Role           string    // "owner", "admin", "member", "viewer"
	CreatedAt      time.Time
}
