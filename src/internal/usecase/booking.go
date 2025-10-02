package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	domain_booking "github.com/ojaswiii/booking-manager/src/internal/domain/booking"
	domain_ticket "github.com/ojaswiii/booking-manager/src/internal/domain/ticket"
	"github.com/ojaswiii/booking-manager/src/internal/repository"
	"github.com/ojaswiii/booking-manager/src/utils"
	concurrency "github.com/ojaswiii/booking-manager/src/utils/concurrency"

	"github.com/google/uuid"
)

type BookingUsecase struct {
	bookingRepo repository.BookingRepository
	ticketRepo  repository.TicketRepository
	eventRepo   repository.EventRepository
	userRepo    repository.UserRepository
	logger      *utils.Logger

	// Concurrency components
	processor *concurrency.BookingProcessor

	// Legacy concurrency control (for backward compatibility)
	bookingMutex sync.RWMutex
	eventLocks   map[uuid.UUID]*sync.Mutex
	eventMutex   sync.RWMutex
}

// NewBookingUsecase creates a new booking usecase
func NewBookingUsecase(
	bookingRepo repository.BookingRepository,
	ticketRepo repository.TicketRepository,
	eventRepo repository.EventRepository,
	userRepo repository.UserRepository,
	logger *utils.Logger,
) *BookingUsecase {
	// Initialize the concurrent booking processor
	processor := concurrency.NewBookingProcessor(
		bookingRepo,
		ticketRepo,
		eventRepo,
		userRepo,
		logger,
	)

	return &BookingUsecase{
		bookingRepo: bookingRepo,
		ticketRepo:  ticketRepo,
		eventRepo:   eventRepo,
		userRepo:    userRepo,
		logger:      logger,
		processor:   processor,
		eventLocks:  make(map[uuid.UUID]*sync.Mutex),
	}
}

// CreateBookingRequest represents a request to create a booking
type CreateBookingRequest struct {
	UserID    uuid.UUID   `json:"user_id"`
	EventID   uuid.UUID   `json:"event_id"`
	TicketIDs []uuid.UUID `json:"ticket_ids"`
}

// CreateBookingResponse represents the response of creating a booking
type CreateBookingResponse struct {
	BookingID   uuid.UUID `json:"booking_id"`
	TotalAmount float64   `json:"total_amount"`
	ExpiresAt   string    `json:"expires_at"`
	Status      string    `json:"status"`
}

// CreateBooking creates a new booking using the concurrent processor
func (b *BookingUsecase) CreateBooking(ctx context.Context, req CreateBookingRequest) (*CreateBookingResponse, error) {
	// Create booking request for the processor
	bookingReq := concurrency.BookingRequest{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		EventID:   req.EventID,
		TicketIDs: req.TicketIDs,
		Timestamp: time.Now(),
		Priority:  1,
	}

	// Enqueue the request
	if err := b.processor.EnqueueBookingRequest(bookingReq); err != nil {
		return nil, fmt.Errorf("failed to enqueue booking request: %w", err)
	}

	// Return immediate response
	return &CreateBookingResponse{
		BookingID:   uuid.New(), // Temporary, will be updated when processed
		TotalAmount: float64(len(req.TicketIDs)) * 50.0,
		ExpiresAt:   time.Now().Add(15 * time.Minute).Format("2006-01-02T15:04:05Z"),
		Status:      "pending",
	}, nil
}

// CreateBookingLegacy creates a new booking with legacy concurrency control (for comparison)
func (b *BookingUsecase) CreateBookingLegacy(ctx context.Context, req CreateBookingRequest) (*CreateBookingResponse, error) {
	// Validate user exists
	user, err := b.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	_ = user // Use user to avoid unused variable

	// Validate event exists and is valid
	event, err := b.eventRepo.GetByID(ctx, req.EventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}
	if event == nil {
		return nil, fmt.Errorf("event is not valid for booking")
	}

	// Get event-specific lock
	eventLock := b.getEventLock(req.EventID)
	eventLock.Lock()
	defer eventLock.Unlock()

	// Get available tickets
	availableTickets, err := b.ticketRepo.GetAvailableByEventID(ctx, req.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available tickets: %w", err)
	}

	// Check if requested tickets are available
	availableTicketMap := make(map[uuid.UUID]*domain_ticket.Ticket)
	for _, ticket := range availableTickets {
		availableTicketMap[ticket.ID] = ticket
	}

	var selectedTickets []*domain_ticket.Ticket
	var totalAmount float64

	for _, ticketID := range req.TicketIDs {
		ticket, exists := availableTicketMap[ticketID]
		if !exists {
			return nil, fmt.Errorf("ticket %s is not available", ticketID)
		}
		selectedTickets = append(selectedTickets, ticket)
		totalAmount += ticket.Price
	}

	// Reserve tickets atomically
	ticketIDs := make([]uuid.UUID, len(selectedTickets))
	for i, ticket := range selectedTickets {
		ticketIDs[i] = ticket.ID
	}

	if err := b.ticketRepo.ReserveTickets(ctx, ticketIDs); err != nil {
		return nil, fmt.Errorf("failed to reserve tickets: %w", err)
	}

	// Create booking
	booking := &domain_booking.Booking{
		ID:          uuid.New(),
		UserID:      req.UserID,
		EventID:     req.EventID,
		TicketIDs:   ticketIDs,
		Status:      domain_booking.BookingStatusPending,
		TotalAmount: totalAmount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(15 * time.Minute), // 15 minutes expiry
	}

	if err := b.bookingRepo.Create(ctx, booking); err != nil {
		// Release tickets if booking save fails
		b.ticketRepo.ReleaseTickets(ctx, ticketIDs)
		return nil, fmt.Errorf("failed to save booking: %w", err)
	}

	b.logger.Info("Booking created successfully",
		"booking_id", booking.ID,
		"user_id", req.UserID,
		"event_id", req.EventID,
		"tickets", len(ticketIDs))

	return &CreateBookingResponse{
		BookingID:   booking.ID,
		TotalAmount: totalAmount,
		ExpiresAt:   booking.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		Status:      string(booking.Status),
	}, nil
}

// ConfirmBookingRequest represents a request to confirm a booking
type ConfirmBookingRequest struct {
	BookingID uuid.UUID `json:"booking_id"`
	UserID    uuid.UUID `json:"user_id"`
}

// ConfirmBooking confirms a booking and marks tickets as sold
func (b *BookingUsecase) ConfirmBooking(ctx context.Context, req ConfirmBookingRequest) error {
	booking, err := b.bookingRepo.GetByID(ctx, req.BookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}

	if booking.UserID != req.UserID {
		return fmt.Errorf("unauthorized: booking does not belong to user")
	}

	if booking.Status != domain_booking.BookingStatusPending {
		return fmt.Errorf("booking is not valid (expired or cancelled)")
	}

	// Confirm booking
	booking.Status = domain_booking.BookingStatusConfirmed
	booking.UpdatedAt = time.Now()

	// Confirm tickets
	if err := b.ticketRepo.ConfirmTickets(ctx, booking.TicketIDs); err != nil {
		return fmt.Errorf("failed to confirm tickets: %w", err)
	}

	// Update booking in repository
	if err := b.bookingRepo.Update(ctx, booking); err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	b.logger.Info("Booking confirmed successfully",
		"booking_id", booking.ID,
		"user_id", req.UserID)

	return nil
}

// CancelBookingRequest represents a request to cancel a booking
type CancelBookingRequest struct {
	BookingID uuid.UUID `json:"booking_id"`
	UserID    uuid.UUID `json:"user_id"`
}

// CancelBooking cancels a booking and releases tickets
func (b *BookingUsecase) CancelBooking(ctx context.Context, req CancelBookingRequest) error {
	booking, err := b.bookingRepo.GetByID(ctx, req.BookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}

	if booking.UserID != req.UserID {
		return fmt.Errorf("unauthorized: booking does not belong to user")
	}

	if booking.Status == domain_booking.BookingStatusConfirmed {
		return fmt.Errorf("confirmed bookings cannot be cancelled")
	}

	// Cancel booking
	booking.Status = domain_booking.BookingStatusCancelled
	booking.UpdatedAt = time.Now()

	// Release tickets
	if err := b.ticketRepo.ReleaseTickets(ctx, booking.TicketIDs); err != nil {
		return fmt.Errorf("failed to release tickets: %w", err)
	}

	// Update booking in repository
	if err := b.bookingRepo.Update(ctx, booking); err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	b.logger.Info("Booking cancelled successfully",
		"booking_id", booking.ID,
		"user_id", req.UserID)

	return nil
}

// GetUserBookings retrieves all bookings for a user
func (b *BookingUsecase) GetUserBookings(ctx context.Context, userID uuid.UUID) ([]*domain_booking.Booking, error) {
	return b.bookingRepo.GetByUserID(ctx, userID)
}

// getEventLock returns a mutex for the specific event
func (b *BookingUsecase) getEventLock(eventID uuid.UUID) *sync.Mutex {
	b.eventMutex.RLock()
	lock, exists := b.eventLocks[eventID]
	b.eventMutex.RUnlock()

	if !exists {
		b.eventMutex.Lock()
		lock, exists = b.eventLocks[eventID]
		if !exists {
			lock = &sync.Mutex{}
			b.eventLocks[eventID] = lock
		}
		b.eventMutex.Unlock()
	}

	return lock
}

// GetConcurrencyStats returns current booking statistics from the processor
func (b *BookingUsecase) GetConcurrencyStats() map[string]interface{} {
	return b.processor.GetStats()
}

// Shutdown gracefully shuts down the booking usecase and its processor
func (b *BookingUsecase) Shutdown() {
	b.logger.Info("Shutting down booking usecase")
	b.processor.Shutdown()
	b.logger.Info("Booking usecase stopped")
}
