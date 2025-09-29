package repository_event

import (
	"context"
	"database/sql"

	"github.com/ojaswiii/booking-manager/src/internal/domain"
	domain_event "github.com/ojaswiii/booking-manager/src/internal/domain/event"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresEventRepository struct {
	db *sqlx.DB
}

// NewPostgresEventRepository creates a new PostgreSQL event repository
func NewPostgresEventRepository(db *sqlx.DB) *postgresEventRepository {
	return &postgresEventRepository{db: db}
}

// Create stores a new event
func (r *postgresEventRepository) Create(ctx context.Context, event *domain_event.Event) error {
	query := `
		INSERT INTO events (id, name, artist, venue, date, total_seats, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query, event.ID, event.Name, event.Artist, event.Venue,
		event.Date, event.TotalSeats, event.Price, event.CreatedAt, event.UpdatedAt)
	return err
}

// GetByID retrieves an event by ID
func (r *postgresEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_event.Event, error) {
	query := `
		SELECT id, name, artist, venue, date, total_seats, price, created_at, updated_at
		FROM events
		WHERE id = $1`

	var event domain_event.Event
	err := r.db.GetContext(ctx, &event, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &event, nil
}

// GetAll retrieves all events
func (r *postgresEventRepository) GetAll(ctx context.Context) ([]*domain_event.Event, error) {
	query := `
		SELECT id, name, artist, venue, date, total_seats, price, created_at, updated_at
		FROM events
		ORDER BY date ASC`

	var events []*domain_event.Event
	err := r.db.SelectContext(ctx, &events, query)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// Update updates an existing event
func (r *postgresEventRepository) Update(ctx context.Context, event *domain_event.Event) error {
	query := `
		UPDATE events
		SET name = $2, artist = $3, venue = $4, date = $5, total_seats = $6, price = $7, updated_at = $8
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, event.ID, event.Name, event.Artist,
		event.Venue, event.Date, event.TotalSeats, event.Price, event.UpdatedAt)
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

// Delete removes an event
func (r *postgresEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`

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
