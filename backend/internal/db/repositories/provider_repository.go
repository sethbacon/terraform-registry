package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/terraform-registry/terraform-registry/internal/db/models"
)

// ProviderRepository handles database operations for providers
type ProviderRepository struct {
	db *sql.DB
}

// NewProviderRepository creates a new provider repository
func NewProviderRepository(db *sql.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

// CreateProvider inserts a new provider record
func (r *ProviderRepository) CreateProvider(ctx context.Context, provider *models.Provider) error {
	query := `
		INSERT INTO providers (organization_id, namespace, type, description, source)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		provider.OrganizationID,
		provider.Namespace,
		provider.Type,
		provider.Description,
		provider.Source,
	).Scan(&provider.ID, &provider.CreatedAt, &provider.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	return nil
}

// GetProvider retrieves a provider by organization, namespace, and type
func (r *ProviderRepository) GetProvider(ctx context.Context, orgID, namespace, providerType string) (*models.Provider, error) {
	query := `
		SELECT id, organization_id, namespace, type, description, source, created_at, updated_at
		FROM providers
		WHERE organization_id = $1 AND namespace = $2 AND type = $3
	`

	provider := &models.Provider{}
	err := r.db.QueryRowContext(ctx, query, orgID, namespace, providerType).Scan(
		&provider.ID,
		&provider.OrganizationID,
		&provider.Namespace,
		&provider.Type,
		&provider.Description,
		&provider.Source,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	return provider, nil
}

// UpdateProvider updates an existing provider's metadata
func (r *ProviderRepository) UpdateProvider(ctx context.Context, provider *models.Provider) error {
	query := `
		UPDATE providers
		SET description = $1, source = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		provider.Description,
		provider.Source,
		provider.ID,
	).Scan(&provider.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}

	return nil
}

// DeleteProvider deletes a provider and all its versions/platforms (cascade)
func (r *ProviderRepository) DeleteProvider(ctx context.Context, providerID string) error {
	query := `DELETE FROM providers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, providerID)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("provider not found")
	}

	return nil
}

// CreateVersion inserts a new provider version
func (r *ProviderRepository) CreateVersion(ctx context.Context, version *models.ProviderVersion) error {
	// Convert protocols slice to JSON
	protocolsJSON, err := json.Marshal(version.Protocols)
	if err != nil {
		return fmt.Errorf("failed to marshal protocols: %w", err)
	}

	query := `
		INSERT INTO provider_versions (provider_id, version, protocols, gpg_public_key, shasums_url, shasums_signature_url, published_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err = r.db.QueryRowContext(ctx, query,
		version.ProviderID,
		version.Version,
		protocolsJSON,
		version.GPGPublicKey,
		version.ShasumURL,
		version.ShasumSignatureURL,
		version.PublishedBy,
	).Scan(&version.ID, &version.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create provider version: %w", err)
	}

	return nil
}

// GetVersion retrieves a specific provider version
func (r *ProviderRepository) GetVersion(ctx context.Context, providerID, version string) (*models.ProviderVersion, error) {
	query := `
		SELECT id, provider_id, version, protocols, gpg_public_key, shasums_url, shasums_signature_url, published_by, created_at
		FROM provider_versions
		WHERE provider_id = $1 AND version = $2
	`

	v := &models.ProviderVersion{}
	var protocolsJSON []byte

	err := r.db.QueryRowContext(ctx, query, providerID, version).Scan(
		&v.ID,
		&v.ProviderID,
		&v.Version,
		&protocolsJSON,
		&v.GPGPublicKey,
		&v.ShasumURL,
		&v.ShasumSignatureURL,
		&v.PublishedBy,
		&v.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get provider version: %w", err)
	}

	// Unmarshal protocols
	if err := json.Unmarshal(protocolsJSON, &v.Protocols); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protocols: %w", err)
	}

	return v, nil
}

// ListVersions retrieves all versions for a provider, ordered by created_at DESC
func (r *ProviderRepository) ListVersions(ctx context.Context, providerID string) ([]*models.ProviderVersion, error) {
	query := `
		SELECT id, provider_id, version, protocols, gpg_public_key, shasums_url, shasums_signature_url, published_by, created_at
		FROM provider_versions
		WHERE provider_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list provider versions: %w", err)
	}
	defer rows.Close()

	var versions []*models.ProviderVersion
	for rows.Next() {
		v := &models.ProviderVersion{}
		var protocolsJSON []byte

		err := rows.Scan(
			&v.ID,
			&v.ProviderID,
			&v.Version,
			&protocolsJSON,
			&v.GPGPublicKey,
			&v.ShasumURL,
			&v.ShasumSignatureURL,
			&v.PublishedBy,
			&v.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider version: %w", err)
		}

		// Unmarshal protocols
		if err := json.Unmarshal(protocolsJSON, &v.Protocols); err != nil {
			return nil, fmt.Errorf("failed to unmarshal protocols: %w", err)
		}

		versions = append(versions, v)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating provider versions: %w", err)
	}

	return versions, nil
}

// DeleteVersion deletes a specific provider version and all its platforms (cascade)
func (r *ProviderRepository) DeleteVersion(ctx context.Context, versionID string) error {
	query := `DELETE FROM provider_versions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, versionID)
	if err != nil {
		return fmt.Errorf("failed to delete provider version: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("provider version not found")
	}

	return nil
}

// CreatePlatform inserts a new platform binary record
func (r *ProviderRepository) CreatePlatform(ctx context.Context, platform *models.ProviderPlatform) error {
	query := `
		INSERT INTO provider_platforms (provider_version_id, os, arch, filename, storage_path, storage_backend, size_bytes, shasum)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		platform.ProviderVersionID,
		platform.OS,
		platform.Arch,
		platform.Filename,
		platform.StoragePath,
		platform.StorageBackend,
		platform.SizeBytes,
		platform.Shasum,
	).Scan(&platform.ID)

	if err != nil {
		return fmt.Errorf("failed to create provider platform: %w", err)
	}

	return nil
}

// GetPlatform retrieves a specific platform binary by version ID, OS, and arch
func (r *ProviderRepository) GetPlatform(ctx context.Context, versionID, os, arch string) (*models.ProviderPlatform, error) {
	query := `
		SELECT id, provider_version_id, os, arch, filename, storage_path, storage_backend, size_bytes, shasum, download_count
		FROM provider_platforms
		WHERE provider_version_id = $1 AND os = $2 AND arch = $3
	`

	platform := &models.ProviderPlatform{}
	err := r.db.QueryRowContext(ctx, query, versionID, os, arch).Scan(
		&platform.ID,
		&platform.ProviderVersionID,
		&platform.OS,
		&platform.Arch,
		&platform.Filename,
		&platform.StoragePath,
		&platform.StorageBackend,
		&platform.SizeBytes,
		&platform.Shasum,
		&platform.DownloadCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get provider platform: %w", err)
	}

	return platform, nil
}

// ListPlatforms retrieves all platform binaries for a provider version
func (r *ProviderRepository) ListPlatforms(ctx context.Context, versionID string) ([]*models.ProviderPlatform, error) {
	query := `
		SELECT id, provider_version_id, os, arch, filename, storage_path, storage_backend, size_bytes, shasum, download_count
		FROM provider_platforms
		WHERE provider_version_id = $1
		ORDER BY os, arch
	`

	rows, err := r.db.QueryContext(ctx, query, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list provider platforms: %w", err)
	}
	defer rows.Close()

	var platforms []*models.ProviderPlatform
	for rows.Next() {
		p := &models.ProviderPlatform{}
		err := rows.Scan(
			&p.ID,
			&p.ProviderVersionID,
			&p.OS,
			&p.Arch,
			&p.Filename,
			&p.StoragePath,
			&p.StorageBackend,
			&p.SizeBytes,
			&p.Shasum,
			&p.DownloadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider platform: %w", err)
		}
		platforms = append(platforms, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating provider platforms: %w", err)
	}

	return platforms, nil
}

// IncrementDownloadCount increments the download counter for a platform
func (r *ProviderRepository) IncrementDownloadCount(ctx context.Context, platformID string) error {
	query := `
		UPDATE provider_platforms
		SET download_count = download_count + 1
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, platformID)
	if err != nil {
		return fmt.Errorf("failed to increment download count: %w", err)
	}

	return nil
}

// DeletePlatform deletes a specific platform binary
func (r *ProviderRepository) DeletePlatform(ctx context.Context, platformID string) error {
	query := `DELETE FROM provider_platforms WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, platformID)
	if err != nil {
		return fmt.Errorf("failed to delete provider platform: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("provider platform not found")
	}

	return nil
}

// SearchProviders searches for providers matching the query
func (r *ProviderRepository) SearchProviders(ctx context.Context, orgID, query, namespace string, limit, offset int) ([]*models.Provider, int, error) {
	// Build WHERE clause
	var whereClause string
	var args []interface{}
	argCount := 0

	// Only filter by organization if orgID is provided (multi-tenant mode)
	if orgID != "" {
		argCount++
		whereClause = fmt.Sprintf("WHERE organization_id = $%d", argCount)
		args = append(args, orgID)
	} else {
		whereClause = "WHERE 1=1" // No org filter in single-tenant mode
	}

	if query != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND (namespace ILIKE $%d OR type ILIKE $%d OR description ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, "%"+query+"%")
	}

	if namespace != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND namespace = $%d", argCount)
		args = append(args, namespace)
	}

	// Count total results
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM providers %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count providers: %w", err)
	}

	// Query with pagination
	searchQuery := fmt.Sprintf(`
		SELECT id, organization_id, namespace, type, description, source, created_at, updated_at
		FROM providers
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount+1, argCount+2)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, searchQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search providers: %w", err)
	}
	defer rows.Close()

	var providers []*models.Provider
	for rows.Next() {
		p := &models.Provider{}
		err := rows.Scan(
			&p.ID,
			&p.OrganizationID,
			&p.Namespace,
			&p.Type,
			&p.Description,
			&p.Source,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, p)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating providers: %w", err)
	}

	return providers, total, nil
}
