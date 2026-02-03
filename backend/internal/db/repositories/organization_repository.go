package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/terraform-registry/terraform-registry/internal/db/models"
)

// OrganizationRepository handles database operations for organizations
type OrganizationRepository struct {
	db *sql.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// GetDefaultOrganization retrieves the default organization for single-tenant mode
func (r *OrganizationRepository) GetDefaultOrganization(ctx context.Context) (*models.Organization, error) {
	return r.GetByName(ctx, "default")
}

// GetByName retrieves an organization by its name
func (r *OrganizationRepository) GetByName(ctx context.Context, name string) (*models.Organization, error) {
	query := `
		SELECT id, name, display_name, created_at, updated_at
		FROM organizations
		WHERE name = $1
	`

	org := &models.Organization{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&org.ID,
		&org.Name,
		&org.DisplayName,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	query := `
		SELECT id, name, display_name, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`

	org := &models.Organization{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.DisplayName,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}

// CreateOrganization creates a new organization
func (r *OrganizationRepository) CreateOrganization(ctx context.Context, org *models.Organization) error {
	query := `
		INSERT INTO organizations (name, display_name)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, org.Name, org.DisplayName).Scan(
		&org.ID,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	return nil
}

// === Organization Membership Operations ===

// AddMemberWithParams adds a user to an organization with the specified role (individual parameters)
func (r *OrganizationRepository) AddMemberWithParams(ctx context.Context, orgID, userID, role string) error {
	query := `
		INSERT INTO organization_members (organization_id, user_id, role, created_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err := r.db.ExecContext(ctx, query, orgID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

// RemoveMember removes a user from an organization
func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	query := `DELETE FROM organization_members WHERE organization_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

// UpdateMemberRole changes a user's role in an organization
func (r *OrganizationRepository) UpdateMemberRole(ctx context.Context, orgID, userID, role string) error {
	query := `
		UPDATE organization_members
		SET role = $3
		WHERE organization_id = $1 AND user_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, orgID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	return nil
}

// GetMember retrieves a user's membership in an organization
func (r *OrganizationRepository) GetMember(ctx context.Context, orgID, userID string) (*models.OrganizationMember, error) {
	query := `
		SELECT organization_id, user_id, role, created_at
		FROM organization_members
		WHERE organization_id = $1 AND user_id = $2
	`

	member := &models.OrganizationMember{}
	err := r.db.QueryRowContext(ctx, query, orgID, userID).Scan(
		&member.OrganizationID,
		&member.UserID,
		&member.Role,
		&member.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return member, nil
}

// ListMembers retrieves all members of an organization
func (r *OrganizationRepository) ListMembers(ctx context.Context, orgID string) ([]*models.OrganizationMember, error) {
	query := `
		SELECT organization_id, user_id, role, created_at
		FROM organization_members
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	defer rows.Close()

	members := make([]*models.OrganizationMember, 0)
	for rows.Next() {
		member := &models.OrganizationMember{}
		err := rows.Scan(
			&member.OrganizationID,
			&member.UserID,
			&member.Role,
			&member.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

// GetUserOrganizations retrieves all organizations a user belongs to
func (r *OrganizationRepository) GetUserOrganizations(ctx context.Context, userID string) ([]*models.Organization, error) {
	query := `
		SELECT o.id, o.name, o.display_name, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = $1
		ORDER BY o.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}
	defer rows.Close()

	organizations := make([]*models.Organization, 0)
	for rows.Next() {
		org := &models.Organization{}
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.DisplayName,
			&org.CreatedAt,
			&org.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, org)
	}

	return organizations, rows.Err()
}

// CheckMembership checks if a user is a member of an organization and returns their role
func (r *OrganizationRepository) CheckMembership(ctx context.Context, orgID, userID string) (bool, string, error) {
	member, err := r.GetMember(ctx, orgID, userID)
	if err != nil {
		return false, "", err
	}

	if member == nil {
		return false, "", nil
	}

	return true, member.Role, nil
}

// Create is an alias for CreateOrganization to match admin handlers
func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	return r.CreateOrganization(ctx, org)
}

// Update updates an organization
func (r *OrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	query := `
		UPDATE organizations
		SET display_name = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, org.ID, org.DisplayName)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	return nil
}

// Delete deletes an organization
func (r *OrganizationRepository) Delete(ctx context.Context, orgID string) error {
	query := `DELETE FROM organizations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

// List retrieves a paginated list of organizations
func (r *OrganizationRepository) List(ctx context.Context, limit, offset int) ([]*models.Organization, error) {
	query := `
		SELECT id, name, display_name, created_at, updated_at
		FROM organizations
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer rows.Close()

	orgs := make([]*models.Organization, 0)
	for rows.Next() {
		org := &models.Organization{}
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.DisplayName,
			&org.CreatedAt,
			&org.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, org)
	}

	return orgs, rows.Err()
}

// Count returns the total number of organizations
func (r *OrganizationRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM organizations`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count organizations: %w", err)
	}

	return count, nil
}

// Search searches for organizations by name or display name
func (r *OrganizationRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Organization, error) {
	searchQuery := `
		SELECT id, name, display_name, created_at, updated_at
		FROM organizations
		WHERE name ILIKE $1 OR display_name ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search organizations: %w", err)
	}
	defer rows.Close()

	orgs := make([]*models.Organization, 0)
	for rows.Next() {
		org := &models.Organization{}
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.DisplayName,
			&org.CreatedAt,
			&org.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, org)
	}

	return orgs, rows.Err()
}

// ListUserOrganizations is an alias for GetUserOrganizations
func (r *OrganizationRepository) ListUserOrganizations(ctx context.Context, userID string) ([]*models.Organization, error) {
	return r.GetUserOrganizations(ctx, userID)
}

// AddMember with models.OrganizationMember parameter
func (r *OrganizationRepository) AddMember(ctx context.Context, member *models.OrganizationMember) error {
	query := `
		INSERT INTO organization_members (organization_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query,
		member.OrganizationID,
		member.UserID,
		member.Role,
		member.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

// UpdateMember updates a member's information
func (r *OrganizationRepository) UpdateMember(ctx context.Context, member *models.OrganizationMember) error {
	return r.UpdateMemberRole(ctx, member.OrganizationID, member.UserID, member.Role)
}
