package concurrency

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventLock represents a lock with expiration
type EventLock struct {
	mutex     *sync.Mutex
	lastUsed  time.Time
	expiresAt time.Time
	refCount  int32
}

// EventLockManager manages event locks with automatic expiration
type EventLockManager struct {
	locks         map[uuid.UUID]*EventLock
	mutex         sync.RWMutex
	cleanupTicker *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	ttl           time.Duration
	maxIdle       time.Duration
}

// NewEventLockManager creates a new event lock manager with automatic cleanup
func NewEventLockManager(ttl, maxIdle time.Duration) *EventLockManager {
	ctx, cancel := context.WithCancel(context.Background())

	elm := &EventLockManager{
		locks:         make(map[uuid.UUID]*EventLock),
		ttl:           ttl,
		maxIdle:       maxIdle,
		ctx:           ctx,
		cancel:        cancel,
		cleanupTicker: time.NewTicker(1 * time.Minute), // Cleanup every minute
	}

	// Start background cleanup
	go elm.cleanupExpiredLocks()

	return elm
}

// GetLock returns a lock for the given event ID
func (elm *EventLockManager) GetLock(eventID uuid.UUID) *sync.Mutex {
	elm.mutex.RLock()
	lock, exists := elm.locks[eventID]
	elm.mutex.RUnlock()

	if !exists {
		elm.mutex.Lock()
		// Double-check after acquiring write lock
		lock, exists = elm.locks[eventID]
		if !exists {
			lock = &EventLock{
				mutex:     &sync.Mutex{},
				lastUsed:  time.Now(),
				expiresAt: time.Now().Add(elm.ttl),
				refCount:  0,
			}
			elm.locks[eventID] = lock
		}
		elm.mutex.Unlock()
	}

	// Update usage time
	lock.lastUsed = time.Now()
	lock.expiresAt = time.Now().Add(elm.ttl)
	lock.refCount++

	return lock.mutex
}

// ReleaseLock decrements the reference count
func (elm *EventLockManager) ReleaseLock(eventID uuid.UUID) {
	elm.mutex.RLock()
	lock, exists := elm.locks[eventID]
	elm.mutex.RUnlock()

	if exists {
		lock.refCount--
		if lock.refCount <= 0 {
			// Mark for cleanup
			lock.expiresAt = time.Now().Add(-time.Second)
		}
	}
}

// cleanupExpiredLocks runs in background to clean up expired locks
func (elm *EventLockManager) cleanupExpiredLocks() {
	for {
		select {
		case <-elm.ctx.Done():
			return
		case <-elm.cleanupTicker.C:
			elm.performCleanup()
		}
	}
}

// performCleanup removes expired and unused locks
func (elm *EventLockManager) performCleanup() {
	elm.mutex.Lock()
	defer elm.mutex.Unlock()

	now := time.Now()
	for eventID, lock := range elm.locks {
		// Remove if expired or idle for too long
		if now.After(lock.expiresAt) ||
			(lock.refCount == 0 && now.After(lock.lastUsed.Add(elm.maxIdle))) {
			delete(elm.locks, eventID)
		}
	}
}

// GetStats returns lock manager statistics
func (elm *EventLockManager) GetStats() map[string]interface{} {
	elm.mutex.RLock()
	defer elm.mutex.RUnlock()

	activeLocks := 0
	totalRefs := int32(0)

	for _, lock := range elm.locks {
		if lock.refCount > 0 {
			activeLocks++
		}
		totalRefs += lock.refCount
	}

	return map[string]interface{}{
		"total_locks":      len(elm.locks),
		"active_locks":     activeLocks,
		"total_refs":       totalRefs,
		"ttl_seconds":      elm.ttl.Seconds(),
		"max_idle_seconds": elm.maxIdle.Seconds(),
	}
}

// Shutdown gracefully shuts down the lock manager
func (elm *EventLockManager) Shutdown() {
	elm.cancel()
	elm.cleanupTicker.Stop()
}
