package models

import "time"

// Module represents a Terraform module in the registry
type Module struct {
	ID             string
	OrganizationID string
	Namespace      string
	Name           string
	System         string
	Description    *string
	Source         *string
	CreatedBy      *string // User ID who created this module
	CreatedAt      time.Time
	UpdatedAt      time.Time
	// Joined fields (not stored in modules table)
	CreatedByName *string // User name who created this module (joined from users table)
}

// ModuleVersion represents a specific version of a module
type ModuleVersion struct {
	ID                 string
	ModuleID           string
	Version            string
	StoragePath        string
	StorageBackend     string
	SizeBytes          int64
	Checksum           string
	Readme             *string
	PublishedBy        *string
	DownloadCount      int64
	Deprecated         bool       // Whether this version is deprecated
	DeprecatedAt       *time.Time // When the version was deprecated
	DeprecationMessage *string    // Optional message explaining deprecation
	CreatedAt          time.Time
	// Joined fields (not stored in module_versions table)
	PublishedByName *string // User name who published this version (joined from users table)
}
