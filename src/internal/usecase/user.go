package usecase

import (
	"context"
	"fmt"
	"time"

	domain_user "github.com/ojaswiii/booking-manager/src/internal/domain/user"
	"github.com/ojaswiii/booking-manager/src/internal/repository"
	"github.com/ojaswiii/booking-manager/src/utils"

	"github.com/google/uuid"
)

type UserUsecase struct {
	userRepo  repository.UserRepository
	cacheRepo repository.UserCacheRepository
	logger    *utils.Logger
}

// UserRepository and UserCacheRepository interfaces are defined in repository/index.go

// NewUserUsecase creates a new user usecase
func NewUserUsecase(userRepo repository.UserRepository, cacheRepo repository.UserCacheRepository, logger *utils.Logger) *UserUsecase {
	return &UserUsecase{
		userRepo:  userRepo,
		cacheRepo: cacheRepo,
		logger:    logger,
	}
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// CreateUserResponse represents the response of creating a user
type CreateUserResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
}

// CreateUser creates a new user
func (u *UserUsecase) CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	// Check if user already exists
	existingUser, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Create user
	user := &domain_user.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save user to database
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Cache user
	if err := u.cacheRepo.Create(ctx, user); err != nil {
		u.logger.Warn("Failed to cache user", "user_id", user.ID, "error", err)
	}

	// Set email index in cache
	if err := u.cacheRepo.SetEmailIndex(ctx, user.Email, user.ID); err != nil {
		u.logger.Warn("Failed to set email index", "email", user.Email, "error", err)
	}

	u.logger.Info("User created successfully", "user_id", user.ID, "email", user.Email)

	return &CreateUserResponse{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
	}, nil
}

// GetUser retrieves a user by ID
func (u *UserUsecase) GetUser(ctx context.Context, userID uuid.UUID) (*domain_user.User, error) {
	// Try cache first
	user, err := u.cacheRepo.GetByID(ctx, userID)
	if err == nil && user != nil {
		return user, nil
	}

	// Fallback to database
	user, err = u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := u.cacheRepo.Create(ctx, user); err != nil {
		u.logger.Warn("Failed to cache user", "user_id", userID, "error", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (u *UserUsecase) GetUserByEmail(ctx context.Context, email string) (*domain_user.User, error) {
	// Try cache first
	user, err := u.cacheRepo.GetByEmail(ctx, email)
	if err == nil && user != nil {
		return user, nil
	}

	// Fallback to database
	user, err = u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := u.cacheRepo.Create(ctx, user); err != nil {
		u.logger.Warn("Failed to cache user", "email", email, "error", err)
	}

	return user, nil
}

// UpdateUser updates a user
func (u *UserUsecase) UpdateUser(ctx context.Context, user *domain_user.User) error {
	// Update in database
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Update cache
	if err := u.cacheRepo.Update(ctx, user); err != nil {
		u.logger.Warn("Failed to update user cache", "user_id", user.ID, "error", err)
	}

	u.logger.Info("User updated successfully", "user_id", user.ID)
	return nil
}

// DeleteUser deletes a user
func (u *UserUsecase) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Delete from database
	if err := u.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	// Delete from cache
	if err := u.cacheRepo.Delete(ctx, userID); err != nil {
		u.logger.Warn("Failed to delete user from cache", "user_id", userID, "error", err)
	}

	u.logger.Info("User deleted successfully", "user_id", userID)
	return nil
}
