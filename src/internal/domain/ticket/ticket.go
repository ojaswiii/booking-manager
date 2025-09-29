package domain_ticket

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TicketStatus represents the status of a ticket
type TicketStatus string

const (
	TicketStatusAvailable TicketStatus = "available"
	TicketStatusReserved  TicketStatus = "reserved"
	TicketStatusSold      TicketStatus = "sold"
	TicketStatusCancelled TicketStatus = "cancelled"
)

// Ticket represents a single ticket for an event
type Ticket struct {
	ID         uuid.UUID    `json:"id" db:"id"`
	EventID    uuid.UUID    `json:"event_id" db:"event_id"`
	SeatNumber int          `json:"seat_number" db:"seat_number"`
	Status     TicketStatus `json:"status" db:"status"`
	Price      float64      `json:"price" db:"price"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at" db:"updated_at"`
}

// TicketRepository defines the interface for ticket data operations
type TicketRepository interface {
	Create(ctx context.Context, ticket *Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (*Ticket, error)
	GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*Ticket, error)
	GetAvailableByEventID(ctx context.Context, eventID uuid.UUID) ([]*Ticket, error)
	Update(ctx context.Context, ticket *Ticket) error
	Delete(ctx context.Context, id uuid.UUID) error
	ReserveTickets(ctx context.Context, ticketIDs []uuid.UUID) error
	ConfirmTickets(ctx context.Context, ticketIDs []uuid.UUID) error
	ReleaseTickets(ctx context.Context, ticketIDs []uuid.UUID) error
}

// TicketUsecase defines the interface for ticket business logic
type TicketUsecase interface {
	CreateTicket(ctx context.Context, req CreateTicketRequest) (*CreateTicketResponse, error)
	GetTicket(ctx context.Context, ticketID uuid.UUID) (*Ticket, error)
	GetTicketsByEvent(ctx context.Context, eventID uuid.UUID) ([]*Ticket, error)
	GetAvailableTickets(ctx context.Context, eventID uuid.UUID) ([]*Ticket, error)
	ReserveTickets(ctx context.Context, ticketIDs []uuid.UUID) error
	ConfirmTickets(ctx context.Context, ticketIDs []uuid.UUID) error
	ReleaseTickets(ctx context.Context, ticketIDs []uuid.UUID) error
}

// CreateTicketRequest represents a request to create a ticket
type CreateTicketRequest struct {
	EventID    uuid.UUID `json:"event_id"`
	SeatNumber int       `json:"seat_number"`
	Price      float64   `json:"price"`
}

// CreateTicketResponse represents the response of creating a ticket
type CreateTicketResponse struct {
	TicketID   uuid.UUID    `json:"ticket_id"`
	EventID    uuid.UUID    `json:"event_id"`
	SeatNumber int          `json:"seat_number"`
	Status     TicketStatus `json:"status"`
	Price      float64      `json:"price"`
}
