package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/terraform-registry/terraform-registry/internal/db/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, email, name, oidc_sub, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.OIDCSub,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	query := `
		SELECT id, email, name, oidc_sub, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.OIDCSub,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, oidc_sub, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.OIDCSub,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByOIDCSub retrieves a user by OIDC subject identifier
func (r *UserRepository) GetUserByOIDCSub(ctx context.Context, oidcSub string) (*models.User, error) {
	query := `
		SELECT id, email, name, oidc_sub, created_at, updated_at
		FROM users
		WHERE oidc_sub = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, oidcSub).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.OIDCSub,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates a user's information
func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET email = $2, name = $3, oidc_sub = $4, updated_at = $5
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.OIDCSub,
		user.UpdatedAt,
	)

	return err
}

// DeleteUser deletes a user (cascades to API keys and memberships)
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// ListUsers retrieves a paginated list of users
func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated users
	query := `
		SELECT id, email, name, oidc_sub, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.OIDCSub,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, rows.Err()
}

// GetOrCreateUserFromOIDC gets or creates a user from OIDC authentication
func (r *UserRepository) GetOrCreateUserFromOIDC(ctx context.Context, oidcSub, email, name string) (*models.User, error) {
	// Try to find existing user by OIDC sub
	user, err := r.GetUserByOIDCSub(ctx, oidcSub)
	if err != nil {
		return nil, err
	}

	if user != nil {
		// User exists, update email and name if changed
		if user.Email != email || user.Name != name {
			user.Email = email
			user.Name = name
			err = r.UpdateUser(ctx, user)
			if err != nil {
				return nil, err
			}
		}
		return user, nil
	}

	// User doesn't exist, create new one
	newUser := &models.User{
		Email:   email,
		Name:    name,
		OIDCSub: &oidcSub,
	}

	err = r.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}
