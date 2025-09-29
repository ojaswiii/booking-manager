package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/ojaswiii/booking-manager/src/internal/usecase"
	"github.com/ojaswiii/booking-manager/src/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserController struct {
	userUsecase *usecase.UserUsecase
	logger      *utils.Logger
}

// NewUserController creates a new user controller
func NewUserController(userUsecase *usecase.UserUsecase, logger *utils.Logger) *UserController {
	return &UserController{
		userUsecase: userUsecase,
		logger:      logger,
	}
}

// CreateUser handles POST /api/users
func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := c.userUsecase.CreateUser(r.Context(), req)
	if err != nil {
		c.logger.Error("Failed to create user", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	c.respondWithJSON(w, http.StatusCreated, response)
}

// GetUser handles GET /api/users/{id}
func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := c.userUsecase.GetUser(r.Context(), userID)
	if err != nil {
		if err.Error() == "resource not found" {
			c.respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		c.logger.Error("Failed to get user", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	c.respondWithJSON(w, http.StatusOK, user)
}

// UpdateUser handles PUT /api/users/{id}
func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing user
	user, err := c.userUsecase.GetUser(r.Context(), userID)
	if err != nil {
		if err.Error() == "resource not found" {
			c.respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		c.logger.Error("Failed to get user", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Update user fields
	user.Email = req.Email
	user.Name = req.Name

	if err := c.userUsecase.UpdateUser(r.Context(), user); err != nil {
		c.logger.Error("Failed to update user", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	c.respondWithJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/users/{id}
func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := c.userUsecase.DeleteUser(r.Context(), userID); err != nil {
		if err.Error() == "resource not found" {
			c.respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		c.logger.Error("Failed to delete user", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	c.respondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// Helper methods

func (c *UserController) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (c *UserController) respondWithError(w http.ResponseWriter, code int, message string) {
	c.respondWithJSON(w, code, map[string]string{"error": message})
}
