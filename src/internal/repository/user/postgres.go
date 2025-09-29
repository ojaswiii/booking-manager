package repository_user

import (
	"context"
	"database/sql"

	"github.com/ojaswiii/booking-manager/src/internal/domain"
	domain_user "github.com/ojaswiii/booking-manager/src/internal/domain/user"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sqlx.DB) *postgresUserRepository {
	return &postgresUserRepository{db: db}
}

// Create stores a new user
func (r *postgresUserRepository) Create(ctx context.Context, usr *domain_user.User) error {
	query := `
		INSERT INTO users (id, email, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query, usr.ID, usr.Email, usr.Name, usr.CreatedAt, usr.UpdatedAt)
	return err
}

// GetByID retrieves a user by ID
func (r *postgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_user.User, error) {
	query := `
		SELECT id, email, name, created_at, updated_at
		FROM users
		WHERE id = $1`

	var usr domain_user.User
	err := r.db.GetContext(ctx, &usr, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &usr, nil
}

// GetByEmail retrieves a user by email
func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain_user.User, error) {
	query := `
		SELECT id, email, name, created_at, updated_at
		FROM users
		WHERE email = $1`

	var usr domain_user.User
	err := r.db.GetContext(ctx, &usr, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &usr, nil
}

// Update updates an existing user
func (r *postgresUserRepository) Update(ctx context.Context, usr *domain_user.User) error {
	query := `
		UPDATE users
		SET email = $2, name = $3, updated_at = $4
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, usr.ID, usr.Email, usr.Name, usr.UpdatedAt)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete removes a user
func (r *postgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
