package domain_booking

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// BookingStatus represents the status of a booking
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusExpired   BookingStatus = "expired"
)

// Booking represents a ticket booking
type Booking struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	UserID      uuid.UUID     `json:"user_id" db:"user_id"`
	EventID     uuid.UUID     `json:"event_id" db:"event_id"`
	TicketIDs   []uuid.UUID   `json:"ticket_ids" db:"ticket_ids"`
	Status      BookingStatus `json:"status" db:"status"`
	TotalAmount float64       `json:"total_amount" db:"total_amount"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
	ExpiresAt   time.Time     `json:"expires_at" db:"expires_at"`
}

// BookingRepository defines the interface for booking data operations
type BookingRepository interface {
	Create(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*Booking, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Booking, error)
	GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*Booking, error)
	Update(ctx context.Context, booking *Booking) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetExpiredBookings(ctx context.Context, before time.Time) ([]*Booking, error)
}

// BookingUsecase defines the interface for booking business logic
type BookingUsecase interface {
	CreateBooking(ctx context.Context, req CreateBookingRequest) (*CreateBookingResponse, error)
	ConfirmBooking(ctx context.Context, req ConfirmBookingRequest) error
	CancelBooking(ctx context.Context, req CancelBookingRequest) error
	GetUserBookings(ctx context.Context, userID uuid.UUID) ([]*Booking, error)
	GetBooking(ctx context.Context, bookingID uuid.UUID) (*Booking, error)
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

// ConfirmBookingRequest represents a request to confirm a booking
type ConfirmBookingRequest struct {
	BookingID uuid.UUID `json:"booking_id"`
	UserID    uuid.UUID `json:"user_id"`
}

// CancelBookingRequest represents a request to cancel a booking
type CancelBookingRequest struct {
	BookingID uuid.UUID `json:"booking_id"`
	UserID    uuid.UUID `json:"user_id"`
}
