package repository_user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ticket-booking-system/src/internal/domain"
	domain_user "ticket-booking-system/src/internal/domain/user"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisUserRepository struct {
	client *redis.Client
}

// NewRedisUserRepository creates a new Redis user repository
func NewRedisUserRepository(client *redis.Client) *redisUserRepository {
	return &redisUserRepository{client: client}
}

// Create caches a new user
func (r *redisUserRepository) Create(ctx context.Context, user *domain_user.User) error {
	key := fmt.Sprintf("user:%s", user.ID.String())
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// Cache for 1 hour
	return r.client.Set(ctx, key, userJSON, time.Hour).Err()
}

// GetByID retrieves a user by ID from cache
func (r *redisUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain_user.User, error) {
	key := fmt.Sprintf("user:%s", id.String())
	userJSON, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	var user domain_user.User
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email from cache
func (r *redisUserRepository) GetByEmail(ctx context.Context, email string) (*domain_user.User, error) {
	key := fmt.Sprintf("user:email:%s", email)
	userID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, userUUID)
}

// Update updates a cached user
func (r *redisUserRepository) Update(ctx context.Context, user *domain_user.User) error {
	key := fmt.Sprintf("user:%s", user.ID.String())
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// Update cache for 1 hour
	return r.client.Set(ctx, key, userJSON, time.Hour).Err()
}

// Delete removes a user from cache
func (r *redisUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	key := fmt.Sprintf("user:%s", id.String())
	return r.client.Del(ctx, key).Err()
}

// SetEmailIndex sets email to user ID mapping
func (r *redisUserRepository) SetEmailIndex(ctx context.Context, email string, userID uuid.UUID) error {
	key := fmt.Sprintf("user:email:%s", email)
	return r.client.Set(ctx, key, userID.String(), time.Hour).Err()
}
