package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/terraform-registry/terraform-registry/internal/db/models"
)

// ModuleRepository handles database operations for modules
type ModuleRepository struct {
	db *sql.DB
}

// NewModuleRepository creates a new module repository
func NewModuleRepository(db *sql.DB) *ModuleRepository {
	return &ModuleRepository{db: db}
}

// CreateModule inserts a new module record
func (r *ModuleRepository) CreateModule(ctx context.Context, module *models.Module) error {
	query := `
		INSERT INTO modules (organization_id, namespace, name, system, description, source)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		module.OrganizationID,
		module.Namespace,
		module.Name,
		module.System,
		module.Description,
		module.Source,
	).Scan(&module.ID, &module.CreatedAt, &module.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create module: %w", err)
	}

	return nil
}

// GetModule retrieves a module by organization, namespace, name, and system
func (r *ModuleRepository) GetModule(ctx context.Context, orgID, namespace, name, system string) (*models.Module, error) {
	query := `
		SELECT id, organization_id, namespace, name, system, description, source, created_at, updated_at
		FROM modules
		WHERE organization_id = $1 AND namespace = $2 AND name = $3 AND system = $4
	`

	module := &models.Module{}
	err := r.db.QueryRowContext(ctx, query, orgID, namespace, name, system).Scan(
		&module.ID,
		&module.OrganizationID,
		&module.Namespace,
		&module.Name,
		&module.System,
		&module.Description,
		&module.Source,
		&module.CreatedAt,
		&module.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get module: %w", err)
	}

	return module, nil
}

// UpdateModule updates an existing module's metadata
func (r *ModuleRepository) UpdateModule(ctx context.Context, module *models.Module) error {
	query := `
		UPDATE modules
		SET description = $1, source = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		module.Description,
		module.Source,
		module.ID,
	).Scan(&module.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update module: %w", err)
	}

	return nil
}

// CreateVersion inserts a new module version
func (r *ModuleRepository) CreateVersion(ctx context.Context, version *models.ModuleVersion) error {
	query := `
		INSERT INTO module_versions (module_id, version, storage_path, storage_backend, size_bytes, checksum, published_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		version.ModuleID,
		version.Version,
		version.StoragePath,
		version.StorageBackend,
		version.SizeBytes,
		version.Checksum,
		version.PublishedBy,
	).Scan(&version.ID, &version.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create module version: %w", err)
	}

	return nil
}

// GetVersion retrieves a specific module version
func (r *ModuleRepository) GetVersion(ctx context.Context, moduleID, version string) (*models.ModuleVersion, error) {
	query := `
		SELECT id, module_id, version, storage_path, storage_backend, size_bytes, checksum, published_by, download_count, created_at
		FROM module_versions
		WHERE module_id = $1 AND version = $2
	`

	v := &models.ModuleVersion{}
	err := r.db.QueryRowContext(ctx, query, moduleID, version).Scan(
		&v.ID,
		&v.ModuleID,
		&v.Version,
		&v.StoragePath,
		&v.StorageBackend,
		&v.SizeBytes,
		&v.Checksum,
		&v.PublishedBy,
		&v.DownloadCount,
		&v.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get module version: %w", err)
	}

	return v, nil
}

// ListVersions retrieves all versions for a module, ordered by version DESC
func (r *ModuleRepository) ListVersions(ctx context.Context, moduleID string) ([]*models.ModuleVersion, error) {
	query := `
		SELECT id, module_id, version, storage_path, storage_backend, size_bytes, checksum, published_by, download_count, created_at
		FROM module_versions
		WHERE module_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, moduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list module versions: %w", err)
	}
	defer rows.Close()

	var versions []*models.ModuleVersion
	for rows.Next() {
		v := &models.ModuleVersion{}
		err := rows.Scan(
			&v.ID,
			&v.ModuleID,
			&v.Version,
			&v.StoragePath,
			&v.StorageBackend,
			&v.SizeBytes,
			&v.Checksum,
			&v.PublishedBy,
			&v.DownloadCount,
			&v.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan module version: %w", err)
		}
		versions = append(versions, v)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating module versions: %w", err)
	}

	return versions, nil
}

// IncrementDownloadCount increments the download counter for a version
func (r *ModuleRepository) IncrementDownloadCount(ctx context.Context, versionID string) error {
	query := `
		UPDATE module_versions
		SET download_count = download_count + 1
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, versionID)
	if err != nil {
		return fmt.Errorf("failed to increment download count: %w", err)
	}

	return nil
}

// SearchModules searches for modules matching the query
func (r *ModuleRepository) SearchModules(ctx context.Context, orgID, query, namespace, system string, limit, offset int) ([]*models.Module, int, error) {
	// Build WHERE clause
	whereClause := "WHERE organization_id = $1"
	args := []interface{}{orgID}
	argCount := 1

	if query != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND (namespace ILIKE $%d OR name ILIKE $%d OR description ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, "%"+query+"%")
	}

	if namespace != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND namespace = $%d", argCount)
		args = append(args, namespace)
	}

	if system != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND system = $%d", argCount)
		args = append(args, system)
	}

	// Count total results
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM modules %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count modules: %w", err)
	}

	// Query with pagination
	query = fmt.Sprintf(`
		SELECT id, organization_id, namespace, name, system, description, source, created_at, updated_at
		FROM modules
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount+1, argCount+2)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search modules: %w", err)
	}
	defer rows.Close()

	var modules []*models.Module
	for rows.Next() {
		m := &models.Module{}
		err := rows.Scan(
			&m.ID,
			&m.OrganizationID,
			&m.Namespace,
			&m.Name,
			&m.System,
			&m.Description,
			&m.Source,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan module: %w", err)
		}
		modules = append(modules, m)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating modules: %w", err)
	}

	return modules, total, nil
}

// DeleteModule deletes a module and all its versions (cascade)
func (r *ModuleRepository) DeleteModule(ctx context.Context, moduleID string) error {
	query := `DELETE FROM modules WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, moduleID)
	if err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("module not found")
	}

	return nil
}

// DeleteVersion deletes a specific module version
func (r *ModuleRepository) DeleteVersion(ctx context.Context, versionID string) error {
	query := `DELETE FROM module_versions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, versionID)
	if err != nil {
		return fmt.Errorf("failed to delete module version: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("module version not found")
	}

	return nil
}
