package repository_event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ojaswiii/booking-manager/src/internal/domain"
	domain_event "github.com/ojaswiii/booking-manager/src/internal/domain/event"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisEventRepository struct {
	client *redis.Client
}

// NewRedisEventRepository creates a new Redis event repository
func NewRedisEventRepository(client *redis.Client) *redisEventRepository {
	return &redisEventRepository{client: client}
}

// Create caches a new event
func (r *redisEventRepository) Create(ctx context.Context, event *domain_event.Event) error {
	key := fmt.Sprintf("event:%s", event.ID.String())
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Cache for 2 hours
	return r.client.Set(ctx, key, eventJSON, 2*time.Hour).Err()
}

// GetByID retrieves an event by ID from cache
func (r *redisEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_event.Event, error) {
	key := fmt.Sprintf("event:%s", id.String())
	eventJSON, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	var event domain_event.Event
	err = json.Unmarshal([]byte(eventJSON), &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetAll retrieves all events from cache
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

// Update updates a cached event
func (r *redisEventRepository) Update(ctx context.Context, event *domain_event.Event) error {
	key := fmt.Sprintf("event:%s", event.ID.String())
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Update cache for 2 hours
	return r.client.Set(ctx, key, eventJSON, 2*time.Hour).Err()
}

// Delete removes an event from cache
func (r *redisEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	key := fmt.Sprintf("event:%s", id.String())
	return r.client.Del(ctx, key).Err()
}

// SetAllEvents caches all events
func (r *redisEventRepository) SetAllEvents(ctx context.Context, events []*domain_event.Event) error {
	key := "events:all"
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return err
	}

	// Cache for 1 hour
	return r.client.Set(ctx, key, eventsJSON, time.Hour).Err()
}
