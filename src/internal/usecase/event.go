package usecase

import (
	"context"
	"fmt"
	"time"

	domain_event "github.com/ojaswiii/booking-manager/src/internal/domain/event"
	domain_ticket "github.com/ojaswiii/booking-manager/src/internal/domain/ticket"
	"github.com/ojaswiii/booking-manager/src/internal/repository"
	"github.com/ojaswiii/booking-manager/src/utils"

	"github.com/google/uuid"
)

type EventUsecase struct {
	eventRepo  repository.EventRepository
	cacheRepo  repository.EventCacheRepository
	ticketRepo repository.TicketRepository
	logger     *utils.Logger
}

// NewEventUsecase creates a new event usecase
func NewEventUsecase(eventRepo repository.EventRepository, cacheRepo repository.EventCacheRepository, ticketRepo repository.TicketRepository, logger *utils.Logger) *EventUsecase {
	return &EventUsecase{
		eventRepo:  eventRepo,
		cacheRepo:  cacheRepo,
		ticketRepo: ticketRepo,
		logger:     logger,
	}
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

// CreateEvent creates a new event with tickets
func (e *EventUsecase) CreateEvent(ctx context.Context, req CreateEventRequest) (*CreateEventResponse, error) {
	// Parse date
	date, err := utils.ParseTime(req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Create event
	event := &domain_event.Event{
		ID:         uuid.New(),
		Name:       req.Name,
		Artist:     req.Artist,
		Venue:      req.Venue,
		Date:       date,
		TotalSeats: req.TotalSeats,
		Price:      req.Price,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save event to database
	if err := e.eventRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	// Cache event
	if err := e.cacheRepo.Create(ctx, event); err != nil {
		e.logger.Warn("Failed to cache event", "event_id", event.ID, "error", err)
	}

	// Create tickets for the event
	for i := 1; i <= req.TotalSeats; i++ {
		ticket := &domain_ticket.Ticket{
			ID:         uuid.New(),
			EventID:    event.ID,
			SeatNumber: i,
			Status:     domain_ticket.TicketStatusAvailable,
			Price:      req.Price,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := e.ticketRepo.Create(ctx, ticket); err != nil {
			return nil, fmt.Errorf("failed to save ticket %d: %w", i, err)
		}
	}

	e.logger.Info("Event created successfully", "event_id", event.ID, "name", event.Name, "total_seats", event.TotalSeats)

	return &CreateEventResponse{
		EventID:    event.ID,
		Name:       event.Name,
		Artist:     event.Artist,
		Venue:      event.Venue,
		Date:       event.Date.Format("2006-01-02T15:04:05Z"),
		TotalSeats: event.TotalSeats,
		Price:      event.Price,
	}, nil
}

// GetEvent retrieves an event by ID
func (e *EventUsecase) GetEvent(ctx context.Context, eventID uuid.UUID) (*domain_event.Event, error) {
	// Try cache first
	event, err := e.cacheRepo.GetByID(ctx, eventID)
	if err == nil && event != nil {
		return event, nil
	}

	// Fallback to database
	event, err = e.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := e.cacheRepo.Create(ctx, event); err != nil {
		e.logger.Warn("Failed to cache event", "event_id", eventID, "error", err)
	}

	return event, nil
}

// GetAllEvents retrieves all events
func (e *EventUsecase) GetAllEvents(ctx context.Context) ([]*domain_event.Event, error) {
	// Try cache first
	events, err := e.cacheRepo.GetAll(ctx)
	if err == nil && events != nil {
		return events, nil
	}

	// Fallback to database
	events, err = e.eventRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := e.cacheRepo.SetAllEvents(ctx, events); err != nil {
		e.logger.Warn("Failed to cache all events", "error", err)
	}

	return events, nil
}

// GetEventTickets retrieves all tickets for an event
func (e *EventUsecase) GetEventTickets(ctx context.Context, eventID uuid.UUID) ([]*domain_ticket.Ticket, error) {
	return e.ticketRepo.GetByEventID(ctx, eventID)
}

// GetAvailableTickets retrieves available tickets for an event
func (e *EventUsecase) GetAvailableTickets(ctx context.Context, eventID uuid.UUID) ([]*domain_ticket.Ticket, error) {
	return e.ticketRepo.GetAvailableByEventID(ctx, eventID)
}
