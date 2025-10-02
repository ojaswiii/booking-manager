package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ojaswiii/booking-manager/src/utils"

	"github.com/google/uuid"
)

// BookingRequest represents a booking request in the queue
type BookingRequest struct {
	ID        string
	UserID    uuid.UUID
	EventID   uuid.UUID
	TicketIDs []uuid.UUID
	Timestamp time.Time
	Priority  int // Higher number = higher priority
}

// QueueManager manages booking requests with load balancing
type QueueManager struct {
	Queues     []chan BookingRequest
	queueCount int
	mu         sync.RWMutex
	logger     *utils.Logger
}

// NewQueueManager creates a new queue manager with load balancing
func NewQueueManager(queueCount int, bufferSize int, logger *utils.Logger) *QueueManager {
	queues := make([]chan BookingRequest, queueCount)
	for i := 0; i < queueCount; i++ {
		queues[i] = make(chan BookingRequest, bufferSize)
	}

	return &QueueManager{
		Queues:     queues,
		queueCount: queueCount,
		logger:     logger,
	}
}

// GetQueue returns the appropriate queue for an event (round-robin)
func (qm *QueueManager) GetQueue(eventID uuid.UUID) chan BookingRequest {
	// Use event ID hash for consistent queue assignment
	hash := eventID.String()
	queueIndex := 0
	for _, char := range hash {
		queueIndex = (queueIndex + int(char)) % qm.queueCount
	}
	return qm.Queues[queueIndex]
}

// Enqueue adds a booking request to the appropriate queue
func (qm *QueueManager) Enqueue(req BookingRequest) error {
	queue := qm.GetQueue(req.EventID)

	select {
	case queue <- req:
		qm.logger.Debug("Booking request enqueued",
			"request_id", req.ID,
			"event_id", req.EventID,
			"queue_index", qm.getQueueIndex(req.EventID))
		return nil
	default:
		return context.DeadlineExceeded // Queue is full
	}
}

// getQueueIndex returns the queue index for an event
func (qm *QueueManager) getQueueIndex(eventID uuid.UUID) int {
	hash := eventID.String()
	queueIndex := 0
	for _, char := range hash {
		queueIndex = (queueIndex + int(char)) % qm.queueCount
	}
	return queueIndex
}

// GetQueueStats returns statistics for all queues
func (qm *QueueManager) GetQueueStats() map[string]interface{} {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	stats := make(map[string]interface{})
	totalPending := 0

	for i, queue := range qm.Queues {
		queueName := fmt.Sprintf("queue_%d", i)
		queueLength := len(queue)
		totalPending += queueLength

		stats[queueName] = map[string]interface{}{
			"length":   queueLength,
			"capacity": cap(queue),
		}
	}

	stats["total_queues"] = qm.queueCount
	stats["total_pending"] = totalPending
	return stats
}
