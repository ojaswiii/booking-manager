package repository_ticket

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ojaswiii/booking-manager/src/internal/domain"
	domain_ticket "github.com/ojaswiii/booking-manager/src/internal/domain/ticket"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresTicketRepository struct {
	db *sqlx.DB
}

// NewPostgresTicketRepository creates a new PostgreSQL ticket repository
func NewPostgresTicketRepository(db *sqlx.DB) *postgresTicketRepository {
	return &postgresTicketRepository{db: db}
}

// Create stores a new ticket
func (r *postgresTicketRepository) Create(ctx context.Context, ticket *domain_ticket.Ticket) error {
	query := `
		INSERT INTO tickets (id, event_id, seat_number, status, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query, ticket.ID, ticket.EventID, ticket.SeatNumber,
		ticket.Status, ticket.Price, ticket.CreatedAt, ticket.UpdatedAt)
	return err
}

// GetByID retrieves a ticket by ID
func (r *postgresTicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_ticket.Ticket, error) {
	query := `
		SELECT id, event_id, seat_number, status, price, created_at, updated_at
		FROM tickets
		WHERE id = $1`

	var ticket domain_ticket.Ticket
	err := r.db.GetContext(ctx, &ticket, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &ticket, nil
}

// GetByEventID retrieves all tickets for an event
func (r *postgresTicketRepository) GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain_ticket.Ticket, error) {
	query := `
		SELECT id, event_id, seat_number, status, price, created_at, updated_at
		FROM tickets
		WHERE event_id = $1
		ORDER BY seat_number ASC`

	var tickets []*domain_ticket.Ticket
	err := r.db.SelectContext(ctx, &tickets, query, eventID)
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

// GetAvailableByEventID retrieves available tickets for an event
func (r *postgresTicketRepository) GetAvailableByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain_ticket.Ticket, error) {
	query := `
		SELECT id, event_id, seat_number, status, price, created_at, updated_at
		FROM tickets
		WHERE event_id = $1 AND status = 'available'
		ORDER BY seat_number ASC`

	var tickets []*domain_ticket.Ticket
	err := r.db.SelectContext(ctx, &tickets, query, eventID)
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

// Update updates an existing ticket
func (r *postgresTicketRepository) Update(ctx context.Context, ticket *domain_ticket.Ticket) error {
	query := `
		UPDATE tickets
		SET status = $2, updated_at = $3
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, ticket.ID, ticket.Status, ticket.UpdatedAt)
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

// Delete removes a ticket
func (r *postgresTicketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tickets WHERE id = $1`

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

// ReserveTickets reserves multiple tickets atomically
func (r *postgresTicketRepository) ReserveTickets(ctx context.Context, ticketIDs []uuid.UUID) error {
	if len(ticketIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if all tickets are available
	placeholders := make([]string, len(ticketIDs))
	args := make([]interface{}, len(ticketIDs))
	for i, id := range ticketIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, status
		FROM tickets
		WHERE id IN (%s)`,
		fmt.Sprintf("$%d", len(ticketIDs)+1))

	// Convert ticketIDs to interface{} slice
	ticketIDsInterface := make([]interface{}, len(ticketIDs))
	for i, id := range ticketIDs {
		ticketIDsInterface[i] = id
	}
	args = append(args, ticketIDsInterface...)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	availableTickets := make(map[uuid.UUID]bool)
	for rows.Next() {
		var id uuid.UUID
		var status string
		if err := rows.Scan(&id, &status); err != nil {
			return err
		}
		availableTickets[id] = (status == "available")
	}

	// Check if all requested tickets are available
	for _, id := range ticketIDs {
		if !availableTickets[id] {
			return fmt.Errorf("ticket %s is not available", id)
		}
	}

	// Reserve all tickets
	updateQuery := fmt.Sprintf(`
		UPDATE tickets
		SET status = 'reserved', updated_at = NOW()
		WHERE id IN (%s)`,
		fmt.Sprintf("$%d", len(ticketIDs)+1))

	// Convert ticketIDs to interface{} slice
	ticketIDsInterface = make([]interface{}, len(ticketIDs))
	for i, id := range ticketIDs {
		ticketIDsInterface[i] = id
	}
	args = append(args, ticketIDsInterface...)
	_, err = tx.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ConfirmTickets confirms multiple tickets atomically
func (r *postgresTicketRepository) ConfirmTickets(ctx context.Context, ticketIDs []uuid.UUID) error {
	if len(ticketIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(ticketIDs))
	args := make([]interface{}, len(ticketIDs))
	for i, id := range ticketIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		UPDATE tickets
		SET status = 'sold', updated_at = NOW()
		WHERE id IN (%s) AND status = 'reserved'`,
		fmt.Sprintf("$%d", len(ticketIDs)+1))

	// Convert ticketIDs to interface{} slice
	ticketIDsInterface := make([]interface{}, len(ticketIDs))
	for i, id := range ticketIDs {
		ticketIDsInterface[i] = id
	}
	args = append(args, ticketIDsInterface...)
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if int(rowsAffected) != len(ticketIDs) {
		return fmt.Errorf("not all tickets could be confirmed")
	}

	return nil
}

// ReleaseTickets releases multiple tickets atomically
func (r *postgresTicketRepository) ReleaseTickets(ctx context.Context, ticketIDs []uuid.UUID) error {
	if len(ticketIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(ticketIDs))
	args := make([]interface{}, len(ticketIDs))
	for i, id := range ticketIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		UPDATE tickets
		SET status = 'available', updated_at = NOW()
		WHERE id IN (%s) AND status IN ('reserved', 'cancelled')`,
		fmt.Sprintf("$%d", len(ticketIDs)+1))

	// Convert ticketIDs to interface{} slice
	ticketIDsInterface := make([]interface{}, len(ticketIDs))
	for i, id := range ticketIDs {
		ticketIDsInterface[i] = id
	}
	args = append(args, ticketIDsInterface...)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}
