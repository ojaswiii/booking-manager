package concurrency

import (
	"context"
	"sync"
	"time"

	domain_booking "github.com/ojaswiii/booking-manager/src/internal/domain/booking"
	"github.com/ojaswiii/booking-manager/src/internal/repository"
	"github.com/ojaswiii/booking-manager/src/utils"

	"github.com/google/uuid"
)

// BookingProcessor handles concurrent booking processing
type BookingProcessor struct {
	bookingRepo repository.BookingRepository
	ticketRepo  repository.TicketRepository
	eventRepo   repository.EventRepository
	userRepo    repository.UserRepository
	logger      *utils.Logger

	// Concurrency components
	queueManager *QueueManager
	ticketLocks  *TicketLockManager
	eventLocks   *EventLockManager

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
	stats  BookingStats
}

// BookingStats holds booking statistics
type BookingStats struct {
	TotalRequests      int64
	SuccessfulBookings int64
	FailedBookings     int64
	QueueLength        int
	ActiveLocks        int
	StartTime          time.Time
}

// NewBookingProcessor creates a new booking processor
func NewBookingProcessor(
	bookingRepo repository.BookingRepository,
	ticketRepo repository.TicketRepository,
	eventRepo repository.EventRepository,
	userRepo repository.UserRepository,
	logger *utils.Logger,
) *BookingProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize concurrency components
	queueManager := NewQueueManager(3, 100, logger) // 3 queues, 100 buffer each
	ticketLocks := NewTicketLockManager()
	eventLocks := NewEventLockManager(30*time.Minute, 5*time.Minute) // 30min TTL, 5min max idle

	bp := &BookingProcessor{
		bookingRepo:  bookingRepo,
		ticketRepo:   ticketRepo,
		eventRepo:    eventRepo,
		userRepo:     userRepo,
		logger:       logger,
		queueManager: queueManager,
		ticketLocks:  ticketLocks,
		eventLocks:   eventLocks,
		ctx:          ctx,
		cancel:       cancel,
		stats: BookingStats{
			StartTime: time.Now(),
		},
	}

	// Start background processors
	bp.startProcessors()

	return bp
}

// startProcessors starts background processors for each queue
func (bp *BookingProcessor) startProcessors() {
	// Start processors for each queue
	for i := 0; i < 3; i++ {
		bp.wg.Add(1)
		go bp.processQueue(i)
	}

	// Start cleanup routine
	bp.wg.Add(1)
	go bp.cleanupExpiredLocks()

	bp.logger.Info("Booking processor started with 3 queue processors")
}

// processQueue processes requests from a specific queue
func (bp *BookingProcessor) processQueue(queueIndex int) {
	defer bp.wg.Done()

	queue := bp.queueManager.Queues[queueIndex]

	for {
		select {
		case req := <-queue:
			bp.processBookingRequest(req)
		case <-bp.ctx.Done():
			return
		}
	}
}

// processBookingRequest processes a single booking request
func (bp *BookingProcessor) processBookingRequest(req BookingRequest) {
	start := time.Now()

	bp.mu.Lock()
	bp.stats.TotalRequests++
	bp.mu.Unlock()

	// Validate user exists
	user, err := bp.userRepo.GetByID(bp.ctx, req.UserID)
	if err != nil {
		bp.logger.Error("User not found", "user_id", req.UserID, "error", err)
		bp.recordFailure()
		return
	}
	_ = user

	// Validate event exists
	_, err = bp.eventRepo.GetByID(bp.ctx, req.EventID)
	if err != nil {
		bp.logger.Error("Event not found", "event_id", req.EventID, "error", err)
		bp.recordFailure()
		return
	}

	// Try to lock all requested tickets
	lockedTickets := make([]uuid.UUID, 0, len(req.TicketIDs))

	for _, ticketID := range req.TicketIDs {
		if bp.ticketLocks.LockTicket(ticketID, req.UserID) {
			lockedTickets = append(lockedTickets, ticketID)
		} else {
			// Failed to lock ticket, release already locked tickets
			bp.releaseTickets(lockedTickets, req.UserID)
			bp.logger.Warn("Failed to lock ticket", "ticket_id", ticketID, "user_id", req.UserID)
			bp.recordFailure()
			return
		}
	}

	// All tickets locked successfully, create booking
	booking := &domain_booking.Booking{
		ID:          uuid.New(),
		UserID:      req.UserID,
		EventID:     req.EventID,
		TicketIDs:   lockedTickets,
		Status:      domain_booking.BookingStatusPending,
		TotalAmount: bp.calculateTotalAmount(lockedTickets),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(15 * time.Minute),
	}

	// Save booking to database
	if err := bp.bookingRepo.Create(bp.ctx, booking); err != nil {
		// Release tickets if booking save fails
		bp.releaseTickets(lockedTickets, req.UserID)
		bp.logger.Error("Failed to save booking", "error", err)
		bp.recordFailure()
		return
	}

	// Reserve tickets in database
	if err := bp.ticketRepo.ReserveTickets(bp.ctx, lockedTickets); err != nil {
		// Rollback booking and release tickets
		bp.bookingRepo.Delete(bp.ctx, booking.ID)
		bp.releaseTickets(lockedTickets, req.UserID)
		bp.logger.Error("Failed to reserve tickets", "error", err)
		bp.recordFailure()
		return
	}

	duration := time.Since(start)
	bp.logger.Info("Booking created successfully",
		"booking_id", booking.ID,
		"user_id", req.UserID,
		"event_id", req.EventID,
		"tickets", len(lockedTickets),
		"duration", duration)

	bp.recordSuccess()
}

// releaseTickets releases multiple tickets
func (bp *BookingProcessor) releaseTickets(ticketIDs []uuid.UUID, userID uuid.UUID) {
	for _, ticketID := range ticketIDs {
		bp.ticketLocks.UnlockTicket(ticketID, userID)
	}
}

// calculateTotalAmount calculates the total amount for tickets
func (bp *BookingProcessor) calculateTotalAmount(ticketIDs []uuid.UUID) float64 {
	// This would need to be implemented based on your ticket pricing logic
	// For now, return a placeholder
	return float64(len(ticketIDs)) * 50.0 // $50 per ticket
}

// recordSuccess records a successful booking
func (bp *BookingProcessor) recordSuccess() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.stats.SuccessfulBookings++
}

// recordFailure records a failed booking
func (bp *BookingProcessor) recordFailure() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.stats.FailedBookings++
}

// cleanupExpiredLocks periodically cleans up expired locks
func (bp *BookingProcessor) cleanupExpiredLocks() {
	defer bp.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-bp.ctx.Done():
			return
		case <-ticker.C:
			expiredCount := bp.ticketLocks.CleanupExpiredLocks()
			if expiredCount > 0 {
				bp.logger.Debug("Cleaned up expired locks", "count", expiredCount)
			}
		}
	}
}

// EnqueueBookingRequest enqueues a booking request for processing
func (bp *BookingProcessor) EnqueueBookingRequest(req BookingRequest) error {
	return bp.queueManager.Enqueue(req)
}

// GetStats returns current booking statistics
func (bp *BookingProcessor) GetStats() map[string]interface{} {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	uptime := time.Since(bp.stats.StartTime)
	lockStats := bp.ticketLocks.GetLockStats()
	queueStats := bp.queueManager.GetQueueStats()

	return map[string]interface{}{
		"total_requests":      bp.stats.TotalRequests,
		"successful_bookings": bp.stats.SuccessfulBookings,
		"failed_bookings":     bp.stats.FailedBookings,
		"queue_length":        bp.getTotalQueueLength(),
		"uptime_seconds":      uptime.Seconds(),
		"requests_per_second": float64(bp.stats.TotalRequests) / uptime.Seconds(),
		"lock_stats":          lockStats,
		"queue_stats":         queueStats,
	}
}

// getTotalQueueLength returns the total length of all queues
func (bp *BookingProcessor) getTotalQueueLength() int {
	total := 0
	for _, queue := range bp.queueManager.Queues {
		total += len(queue)
	}
	return total
}

// Shutdown gracefully shuts down the booking processor
func (bp *BookingProcessor) Shutdown() {
	bp.logger.Info("Shutting down booking processor")
	bp.cancel()
	bp.wg.Wait()
	bp.eventLocks.Shutdown()
	bp.logger.Info("Booking processor stopped")
}
