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
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
}
