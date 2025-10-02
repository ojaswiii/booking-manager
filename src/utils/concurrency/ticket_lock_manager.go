package concurrency

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// TicketLock represents a lock on a ticket with timestamp
type TicketLock struct {
	TicketID  uuid.UUID
	UserID    uuid.UUID
	LockedAt  time.Time
	ExpiresAt time.Time
}

// TicketLockManager manages ticket locks with automatic expiration
type TicketLockManager struct {
	locks map[uuid.UUID]*TicketLock
	mu    sync.RWMutex
}

// NewTicketLockManager creates a new ticket lock manager
func NewTicketLockManager() *TicketLockManager {
	return &TicketLockManager{
		locks: make(map[uuid.UUID]*TicketLock),
	}
}

// LockTicket attempts to lock a ticket for a user
func (tlm *TicketLockManager) LockTicket(ticketID, userID uuid.UUID) bool {
	tlm.mu.Lock()
	defer tlm.mu.Unlock()

	now := time.Now()
	lock, exists := tlm.locks[ticketID]

	// If lock exists and is still valid, check if it's the same user
	if exists && now.Before(lock.ExpiresAt) {
		return lock.UserID == userID // Same user can re-lock
	}

	// Create new lock or replace expired lock
	tlm.locks[ticketID] = &TicketLock{
		TicketID:  ticketID,
		UserID:    userID,
		LockedAt:  now,
		ExpiresAt: now.Add(10 * time.Minute), // 10 minutes expiration
	}

	return true
}

// UnlockTicket removes a ticket lock
func (tlm *TicketLockManager) UnlockTicket(ticketID, userID uuid.UUID) bool {
	tlm.mu.Lock()
	defer tlm.mu.Unlock()

	lock, exists := tlm.locks[ticketID]
	if !exists {
		return false
	}

	// Only the user who locked it can unlock it
	if lock.UserID != userID {
		return false
	}

	delete(tlm.locks, ticketID)
	return true
}

// IsTicketLocked checks if a ticket is currently locked
func (tlm *TicketLockManager) IsTicketLocked(ticketID uuid.UUID) bool {
	tlm.mu.RLock()
	defer tlm.mu.RUnlock()

	lock, exists := tlm.locks[ticketID]
	if !exists {
		return false
	}

	// Check if lock has expired
	return time.Now().Before(lock.ExpiresAt)
}

// GetTicketLockInfo returns lock information for a ticket
func (tlm *TicketLockManager) GetTicketLockInfo(ticketID uuid.UUID) (*TicketLock, bool) {
	tlm.mu.RLock()
	defer tlm.mu.RUnlock()

	lock, exists := tlm.locks[ticketID]
	if !exists {
		return nil, false
	}

	// Check if lock has expired
	if time.Now().After(lock.ExpiresAt) {
		return nil, false
	}

	return lock, true
}

// CleanupExpiredLocks removes expired locks
func (tlm *TicketLockManager) CleanupExpiredLocks() int {
	tlm.mu.Lock()
	defer tlm.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for ticketID, lock := range tlm.locks {
		if now.After(lock.ExpiresAt) {
			delete(tlm.locks, ticketID)
			expiredCount++
		}
	}

	return expiredCount
}

// GetLockStats returns lock statistics
func (tlm *TicketLockManager) GetLockStats() map[string]interface{} {
	tlm.mu.RLock()
	defer tlm.mu.RUnlock()

	now := time.Now()
	activeLocks := 0
	expiredLocks := 0

	for _, lock := range tlm.locks {
		if now.Before(lock.ExpiresAt) {
			activeLocks++
		} else {
			expiredLocks++
		}
	}

	return map[string]interface{}{
		"total_locks":   len(tlm.locks),
		"active_locks":  activeLocks,
		"expired_locks": expiredLocks,
	}
}
