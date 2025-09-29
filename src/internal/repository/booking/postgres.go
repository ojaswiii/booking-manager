package repository_booking

import (
	"context"
	"database/sql"
	"time"

	"ticket-booking-system/internal/domain"
	domain_booking "ticket-booking-system/internal/domain/booking"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresBookingRepository struct {
	db *sqlx.DB
}

// NewPostgresBookingRepository creates a new PostgreSQL booking repository
func NewPostgresBookingRepository(db *sqlx.DB) *postgresBookingRepository {
	return &postgresBookingRepository{db: db}
}

// Create stores a new booking
func (r *postgresBookingRepository) Create(ctx context.Context, booking *domain_booking.Booking) error {
	query := `
		INSERT INTO bookings (id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query, booking.ID, booking.UserID, booking.EventID,
		booking.TicketIDs, booking.Status, booking.TotalAmount, booking.CreatedAt,
		booking.UpdatedAt, booking.ExpiresAt)
	return err
}

// GetByID retrieves a booking by ID
func (r *postgresBookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_booking.Booking, error) {
	query := `
		SELECT id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at
		FROM bookings
		WHERE id = $1`

	var booking domain_booking.Booking
	err := r.db.GetContext(ctx, &booking, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &booking, nil
}

// GetByUserID retrieves all bookings for a user
func (r *postgresBookingRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain_booking.Booking, error) {
	query := `
		SELECT id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC`

	var bookings []*domain_booking.Booking
	err := r.db.SelectContext(ctx, &bookings, query, userID)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

// GetByEventID retrieves all bookings for an event
func (r *postgresBookingRepository) GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain_booking.Booking, error) {
	query := `
		SELECT id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at
		FROM bookings
		WHERE event_id = $1
		ORDER BY created_at DESC`

	var bookings []*domain_booking.Booking
	err := r.db.SelectContext(ctx, &bookings, query, eventID)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

// Update updates an existing booking
func (r *postgresBookingRepository) Update(ctx context.Context, booking *domain_booking.Booking) error {
	query := `
		UPDATE bookings
		SET status = $2, total_amount = $3, updated_at = $4, expires_at = $5
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, booking.ID, booking.Status,
		booking.TotalAmount, booking.UpdatedAt, booking.ExpiresAt)
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

// Delete removes a booking
func (r *postgresBookingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM bookings WHERE id = $1`

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

// GetExpiredBookings retrieves bookings that have expired
func (r *postgresBookingRepository) GetExpiredBookings(ctx context.Context, before time.Time) ([]*domain_booking.Booking, error) {
	query := `
		SELECT id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at
		FROM bookings
		WHERE expires_at < $1 AND status = 'pending'
		ORDER BY expires_at ASC`

	var bookings []*domain_booking.Booking
	err := r.db.SelectContext(ctx, &bookings, query, before)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}
