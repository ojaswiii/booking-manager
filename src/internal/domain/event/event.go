package domain_event

import (
	"context"
	"time"

	domain_ticket "github.com/ojaswiii/booking-manager/src/internal/domain/ticket"

	"github.com/google/uuid"
)

// Event represents a show/concert event
type Event struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Artist     string    `json:"artist" db:"artist"`
	Venue      string    `json:"venue" db:"venue"`
	Date       time.Time `json:"date" db:"date"`
	TotalSeats int       `json:"total_seats" db:"total_seats"`
	Price      float64   `json:"price" db:"price"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// EventRepository defines the interface for event data operations
type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)
	GetAll(ctx context.Context) ([]*Event, error)
	Update(ctx context.Context, event *Event) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// EventCacheRepository defines the interface for event cache operations
type EventCacheRepository interface {
	Create(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)
	GetAll(ctx context.Context) ([]*Event, error)
	Update(ctx context.Context, event *Event) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetAllEvents(ctx context.Context, events []*Event) error
}

// EventUsecase defines the interface for event business logic
type EventUsecase interface {
	CreateEvent(ctx context.Context, req CreateEventRequest) (*CreateEventResponse, error)
	GetEvent(ctx context.Context, eventID uuid.UUID) (*Event, error)
	GetAllEvents(ctx context.Context) ([]*Event, error)
	GetEventTickets(ctx context.Context, eventID uuid.UUID) ([]*domain_ticket.Ticket, error)
	GetAvailableTickets(ctx context.Context, eventID uuid.UUID) ([]*domain_ticket.Ticket, error)
}

// CreateEventRequest represents a request to create an event
type CreateEventRequest struct {
	Name       string  `json:"name"`
	Artist     string  `json:"artist"`
	Venue      string  `json:"venue"`
	Date       string  `json:"date"` // ISO 8601 format
	TotalSeats int     `json:"total_seats"`
	Price      float64 `json:"price"`
}

// CreateEventResponse represents the response of creating an event
type CreateEventResponse struct {
	EventID    uuid.UUID `json:"event_id"`
	Name       string    `json:"name"`
	Artist     string    `json:"artist"`
	Venue      string    `json:"venue"`
	Date       string    `json:"date"`
	TotalSeats int       `json:"total_seats"`
	Price      float64   `json:"price"`
}
