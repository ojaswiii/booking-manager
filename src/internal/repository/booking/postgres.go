package repository_booking

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ojaswiii/booking-manager/src/internal/domain"
	domain_booking "github.com/ojaswiii/booking-manager/src/internal/domain/booking"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresBookingRepository struct {
	db *sqlx.DB
}

// uuidSliceToString converts []uuid.UUID to PostgreSQL array string format
func uuidSliceToString(uuids []uuid.UUID) string {
	if len(uuids) == 0 {
		return "{}"
	}

	strs := make([]string, len(uuids))
	for i, u := range uuids {
		strs[i] = fmt.Sprintf("\"%s\"", u.String())
	}
	return "{" + strings.Join(strs, ",") + "}"
}

// stringToUUIDSlice converts PostgreSQL array string to []uuid.UUID
func stringToUUIDSlice(s string) ([]uuid.UUID, error) {
	// Remove curly braces
	s = strings.Trim(s, "{}")
	if s == "" {
		return []uuid.UUID{}, nil
	}

	// Split by comma and parse each UUID
	parts := strings.Split(s, ",")
	uuids := make([]uuid.UUID, len(parts))
	for i, part := range parts {
		// Remove quotes if present
		part = strings.Trim(part, "\"")
		u, err := uuid.Parse(part)
		if err != nil {
			return nil, err
		}
		uuids[i] = u
	}
	return uuids, nil
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

	// Convert UUID slice to PostgreSQL array string
	ticketIDsStr := uuidSliceToString(booking.TicketIDs)

	_, err := r.db.ExecContext(ctx, query, booking.ID, booking.UserID, booking.EventID,
		ticketIDsStr, booking.Status, booking.TotalAmount, booking.CreatedAt,
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
	var ticketIDsStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&booking.ID, &booking.UserID, &booking.EventID, &ticketIDsStr,
		&booking.Status, &booking.TotalAmount, &booking.CreatedAt,
		&booking.UpdatedAt, &booking.ExpiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	// Convert PostgreSQL array string back to UUID slice
	ticketIDs, err := stringToUUIDSlice(ticketIDsStr)
	if err != nil {
		return nil, err
	}
	booking.TicketIDs = ticketIDs

	return &booking, nil
}

// GetByUserID retrieves all bookings for a user
func (r *postgresBookingRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain_booking.Booking, error) {
	query := `
		SELECT id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*domain_booking.Booking
	for rows.Next() {
		var booking domain_booking.Booking
		var ticketIDsStr string

		err := rows.Scan(
			&booking.ID, &booking.UserID, &booking.EventID, &ticketIDsStr,
			&booking.Status, &booking.TotalAmount, &booking.CreatedAt,
			&booking.UpdatedAt, &booking.ExpiresAt)
		if err != nil {
			return nil, err
		}

		// Convert PostgreSQL array string back to UUID slice
		ticketIDs, err := stringToUUIDSlice(ticketIDsStr)
		if err != nil {
			return nil, err
		}
		booking.TicketIDs = ticketIDs

		bookings = append(bookings, &booking)
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

	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*domain_booking.Booking
	for rows.Next() {
		var booking domain_booking.Booking
		var ticketIDsStr string

		err := rows.Scan(
			&booking.ID, &booking.UserID, &booking.EventID, &ticketIDsStr,
			&booking.Status, &booking.TotalAmount, &booking.CreatedAt,
			&booking.UpdatedAt, &booking.ExpiresAt)
		if err != nil {
			return nil, err
		}

		// Convert PostgreSQL array string back to UUID slice
		ticketIDs, err := stringToUUIDSlice(ticketIDsStr)
		if err != nil {
			return nil, err
		}
		booking.TicketIDs = ticketIDs

		bookings = append(bookings, &booking)
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

	rows, err := r.db.QueryContext(ctx, query, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*domain_booking.Booking
	for rows.Next() {
		var booking domain_booking.Booking
		var ticketIDsStr string

		err := rows.Scan(
			&booking.ID, &booking.UserID, &booking.EventID, &ticketIDsStr,
			&booking.Status, &booking.TotalAmount, &booking.CreatedAt,
			&booking.UpdatedAt, &booking.ExpiresAt)
		if err != nil {
			return nil, err
		}

		// Convert PostgreSQL array string back to UUID slice
		ticketIDs, err := stringToUUIDSlice(ticketIDsStr)
		if err != nil {
			return nil, err
		}
		booking.TicketIDs = ticketIDs

		bookings = append(bookings, &booking)
	}

	return bookings, nil
}
