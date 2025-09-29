package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"ticket-booking-system/src/internal/domain"
	domain_booking "ticket-booking-system/src/internal/domain/booking"
	domain_event "ticket-booking-system/src/internal/domain/event"
	domain_ticket "ticket-booking-system/src/internal/domain/ticket"
	domain_user "ticket-booking-system/src/internal/domain/user"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// RepositoryContainer holds all repository instances
type RepositoryContainer struct {
	User    UserRepository
	Event   EventRepository
	Ticket  TicketRepository
	Booking BookingRepository

	// Cache repositories
	UserCache  UserCacheRepository
	EventCache EventCacheRepository
}

// Repository interfaces
type UserRepository interface {
	Create(ctx context.Context, usr *domain_user.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain_user.User, error)
	GetByEmail(ctx context.Context, email string) (*domain_user.User, error)
	Update(ctx context.Context, usr *domain_user.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type EventRepository interface {
	Create(ctx context.Context, evt *domain_event.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain_event.Event, error)
	GetAll(ctx context.Context) ([]*domain_event.Event, error)
	Update(ctx context.Context, evt *domain_event.Event) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TicketRepository interface {
	Create(ctx context.Context, tkt *domain_ticket.Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain_ticket.Ticket, error)
	GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain_ticket.Ticket, error)
	GetAvailableByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain_ticket.Ticket, error)
	Update(ctx context.Context, tkt *domain_ticket.Ticket) error
	Delete(ctx context.Context, id uuid.UUID) error
	ReserveTickets(ctx context.Context, ticketIDs []uuid.UUID) error
	ConfirmTickets(ctx context.Context, ticketIDs []uuid.UUID) error
	ReleaseTickets(ctx context.Context, ticketIDs []uuid.UUID) error
}

type BookingRepository interface {
	Create(ctx context.Context, bk *domain_booking.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain_booking.Booking, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain_booking.Booking, error)
	GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain_booking.Booking, error)
	Update(ctx context.Context, bk *domain_booking.Booking) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetExpiredBookings(ctx context.Context, before time.Time) ([]*domain_booking.Booking, error)
}

type UserCacheRepository interface {
	Create(ctx context.Context, usr *domain_user.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain_user.User, error)
	GetByEmail(ctx context.Context, email string) (*domain_user.User, error)
	Update(ctx context.Context, usr *domain_user.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetEmailIndex(ctx context.Context, email string, userID uuid.UUID) error
}

type EventCacheRepository interface {
	Create(ctx context.Context, evt *domain_event.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain_event.Event, error)
	GetAll(ctx context.Context) ([]*domain_event.Event, error)
	Update(ctx context.Context, evt *domain_event.Event) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetAllEvents(ctx context.Context, events []*domain_event.Event) error
}

// NewRepositoryContainer creates a new repository container
func NewRepositoryContainer(db *sqlx.DB, redisClient *redis.Client) *RepositoryContainer {
	// Create repository implementations directly
	userRepo := &postgresUserRepository{db: db}
	eventRepo := &postgresEventRepository{db: db}
	ticketRepo := &postgresTicketRepository{db: db}
	bookingRepo := &postgresBookingRepository{db: db}

	userCache := &redisUserRepository{client: redisClient}
	eventCache := &redisEventRepository{client: redisClient}

	return &RepositoryContainer{
		User:       userRepo,
		Event:      eventRepo,
		Ticket:     ticketRepo,
		Booking:    bookingRepo,
		UserCache:  userCache,
		EventCache: eventCache,
	}
}

// Repository implementations

// PostgreSQL User Repository
type postgresUserRepository struct {
	db *sqlx.DB
}

func (r *postgresUserRepository) Create(ctx context.Context, usr *domain_user.User) error {
	query := `INSERT INTO users (id, email, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, usr.ID, usr.Email, usr.Name, usr.CreatedAt, usr.UpdatedAt)
	return err
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_user.User, error) {
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`
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

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain_user.User, error) {
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE email = $1`
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

func (r *postgresUserRepository) Update(ctx context.Context, usr *domain_user.User) error {
	query := `UPDATE users SET email = $2, name = $3, updated_at = $4 WHERE id = $1`
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

// Redis User Repository
type redisUserRepository struct {
	client *redis.Client
}

func (r *redisUserRepository) Create(ctx context.Context, usr *domain_user.User) error {
	key := fmt.Sprintf("user:%s", usr.ID.String())
	userJSON, err := json.Marshal(usr)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, userJSON, time.Hour).Err()
}

func (r *redisUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_user.User, error) {
	key := fmt.Sprintf("user:%s", id.String())
	userJSON, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	var usr domain_user.User
	err = json.Unmarshal([]byte(userJSON), &usr)
	if err != nil {
		return nil, err
	}
	return &usr, nil
}

func (r *redisUserRepository) GetByEmail(ctx context.Context, email string) (*domain_user.User, error) {
	key := fmt.Sprintf("user:email:%s", email)
	userID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, userUUID)
}

func (r *redisUserRepository) Update(ctx context.Context, usr *domain_user.User) error {
	key := fmt.Sprintf("user:%s", usr.ID.String())
	userJSON, err := json.Marshal(usr)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, userJSON, time.Hour).Err()
}

func (r *redisUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	key := fmt.Sprintf("user:%s", id.String())
	return r.client.Del(ctx, key).Err()
}

func (r *redisUserRepository) SetEmailIndex(ctx context.Context, email string, userID uuid.UUID) error {
	key := fmt.Sprintf("user:email:%s", email)
	return r.client.Set(ctx, key, userID.String(), time.Hour).Err()
}

// PostgreSQL Event Repository
type postgresEventRepository struct {
	db *sqlx.DB
}

func (r *postgresEventRepository) Create(ctx context.Context, evt *domain_event.Event) error {
	query := `INSERT INTO events (id, name, artist, venue, date, total_seats, price, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query, evt.ID, evt.Name, evt.Artist, evt.Venue, evt.Date, evt.TotalSeats, evt.Price, evt.CreatedAt, evt.UpdatedAt)
	return err
}

func (r *postgresEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_event.Event, error) {
	query := `SELECT id, name, artist, venue, date, total_seats, price, created_at, updated_at FROM events WHERE id = $1`
	var evt domain_event.Event
	err := r.db.GetContext(ctx, &evt, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &evt, nil
}

func (r *postgresEventRepository) GetAll(ctx context.Context) ([]*domain_event.Event, error) {
	query := `SELECT id, name, artist, venue, date, total_seats, price, created_at, updated_at FROM events ORDER BY date ASC`
	var events []*domain_event.Event
	err := r.db.SelectContext(ctx, &events, query)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *postgresEventRepository) Update(ctx context.Context, evt *domain_event.Event) error {
	query := `UPDATE events SET name = $2, artist = $3, venue = $4, date = $5, total_seats = $6, price = $7, updated_at = $8 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, evt.ID, evt.Name, evt.Artist, evt.Venue, evt.Date, evt.TotalSeats, evt.Price, evt.UpdatedAt)
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

// Redis Event Repository
type redisEventRepository struct {
	client *redis.Client
}

func (r *redisEventRepository) Create(ctx context.Context, evt *domain_event.Event) error {
	key := fmt.Sprintf("event:%s", evt.ID.String())
	eventJSON, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, eventJSON, 2*time.Hour).Err()
}

func (r *redisEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_event.Event, error) {
	key := fmt.Sprintf("event:%s", id.String())
	eventJSON, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	var evt domain_event.Event
	err = json.Unmarshal([]byte(eventJSON), &evt)
	if err != nil {
		return nil, err
	}
	return &evt, nil
}

func (r *redisEventRepository) GetAll(ctx context.Context) ([]*domain_event.Event, error) {
	key := "events:all"
	eventsJSON, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	var events []*domain_event.Event
	err = json.Unmarshal([]byte(eventsJSON), &events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *redisEventRepository) Update(ctx context.Context, evt *domain_event.Event) error {
	key := fmt.Sprintf("event:%s", evt.ID.String())
	eventJSON, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, eventJSON, 2*time.Hour).Err()
}

func (r *redisEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	key := fmt.Sprintf("event:%s", id.String())
	return r.client.Del(ctx, key).Err()
}

func (r *redisEventRepository) SetAllEvents(ctx context.Context, events []*domain_event.Event) error {
	key := "events:all"
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, eventsJSON, time.Hour).Err()
}

// PostgreSQL Ticket Repository
type postgresTicketRepository struct {
	db *sqlx.DB
}

func (r *postgresTicketRepository) Create(ctx context.Context, tkt *domain_ticket.Ticket) error {
	query := `INSERT INTO tickets (id, event_id, seat_number, status, price, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, tkt.ID, tkt.EventID, tkt.SeatNumber, tkt.Status, tkt.Price, tkt.CreatedAt, tkt.UpdatedAt)
	return err
}

func (r *postgresTicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_ticket.Ticket, error) {
	query := `SELECT id, event_id, seat_number, status, price, created_at, updated_at FROM tickets WHERE id = $1`
	var tkt domain_ticket.Ticket
	err := r.db.GetContext(ctx, &tkt, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &tkt, nil
}

func (r *postgresTicketRepository) GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain_ticket.Ticket, error) {
	query := `SELECT id, event_id, seat_number, status, price, created_at, updated_at FROM tickets WHERE event_id = $1 ORDER BY seat_number ASC`
	var tickets []*domain_ticket.Ticket
	err := r.db.SelectContext(ctx, &tickets, query, eventID)
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *postgresTicketRepository) GetAvailableByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain_ticket.Ticket, error) {
	query := `SELECT id, event_id, seat_number, status, price, created_at, updated_at FROM tickets WHERE event_id = $1 AND status = 'available' ORDER BY seat_number ASC`
	var tickets []*domain_ticket.Ticket
	err := r.db.SelectContext(ctx, &tickets, query, eventID)
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *postgresTicketRepository) Update(ctx context.Context, tkt *domain_ticket.Ticket) error {
	query := `UPDATE tickets SET status = $2, updated_at = $3 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, tkt.ID, tkt.Status, tkt.UpdatedAt)
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

	query := fmt.Sprintf(`SELECT id, status FROM tickets WHERE id IN (%s)`, fmt.Sprintf("$%d", len(ticketIDs)+1))

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
	updateQuery := fmt.Sprintf(`UPDATE tickets SET status = 'reserved', updated_at = NOW() WHERE id IN (%s)`, fmt.Sprintf("$%d", len(ticketIDs)+1))

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

	query := fmt.Sprintf(`UPDATE tickets SET status = 'sold', updated_at = NOW() WHERE id IN (%s) AND status = 'reserved'`, fmt.Sprintf("$%d", len(ticketIDs)+1))

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

	query := fmt.Sprintf(`UPDATE tickets SET status = 'available', updated_at = NOW() WHERE id IN (%s) AND status IN ('reserved', 'cancelled')`, fmt.Sprintf("$%d", len(ticketIDs)+1))

	// Convert ticketIDs to interface{} slice
	ticketIDsInterface := make([]interface{}, len(ticketIDs))
	for i, id := range ticketIDs {
		ticketIDsInterface[i] = id
	}
	args = append(args, ticketIDsInterface...)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// PostgreSQL Booking Repository
type postgresBookingRepository struct {
	db *sqlx.DB
}

func (r *postgresBookingRepository) Create(ctx context.Context, bk *domain_booking.Booking) error {
	query := `INSERT INTO bookings (id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query, bk.ID, bk.UserID, bk.EventID, bk.TicketIDs, bk.Status, bk.TotalAmount, bk.CreatedAt, bk.UpdatedAt, bk.ExpiresAt)
	return err
}

func (r *postgresBookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_booking.Booking, error) {
	query := `SELECT id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at FROM bookings WHERE id = $1`
	var bk domain_booking.Booking
	err := r.db.GetContext(ctx, &bk, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &bk, nil
}

func (r *postgresBookingRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain_booking.Booking, error) {
	query := `SELECT id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at FROM bookings WHERE user_id = $1 ORDER BY created_at DESC`
	var bookings []*domain_booking.Booking
	err := r.db.SelectContext(ctx, &bookings, query, userID)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *postgresBookingRepository) GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain_booking.Booking, error) {
	query := `SELECT id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at FROM bookings WHERE event_id = $1 ORDER BY created_at DESC`
	var bookings []*domain_booking.Booking
	err := r.db.SelectContext(ctx, &bookings, query, eventID)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *postgresBookingRepository) Update(ctx context.Context, bk *domain_booking.Booking) error {
	query := `UPDATE bookings SET status = $2, total_amount = $3, updated_at = $4, expires_at = $5 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, bk.ID, bk.Status, bk.TotalAmount, bk.UpdatedAt, bk.ExpiresAt)
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

func (r *postgresBookingRepository) GetExpiredBookings(ctx context.Context, before time.Time) ([]*domain_booking.Booking, error) {
	query := `SELECT id, user_id, event_id, ticket_ids, status, total_amount, created_at, updated_at, expires_at FROM bookings WHERE expires_at < $1 AND status = 'pending' ORDER BY expires_at ASC`
	var bookings []*domain_booking.Booking
	err := r.db.SelectContext(ctx, &bookings, query, before)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}
