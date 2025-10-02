package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/ojaswiii/booking-manager/src/internal/usecase"
	"github.com/ojaswiii/booking-manager/src/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type BookingController struct {
	bookingUsecase *usecase.BookingUsecase
	logger         *utils.Logger
}

// NewBookingController creates a new booking controller
func NewBookingController(bookingUsecase *usecase.BookingUsecase, logger *utils.Logger) *BookingController {
	return &BookingController{
		bookingUsecase: bookingUsecase,
		logger:         logger,
	}
}

// CreateBooking handles POST /api/bookings
func (c *BookingController) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Use concurrent booking for better performance
	response, err := c.bookingUsecase.CreateBooking(r.Context(), req)
	if err != nil {
		c.logger.Error("Failed to create booking", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to create booking")
		return
	}

	c.respondWithJSON(w, http.StatusCreated, response)
}

// ConfirmBooking handles POST /api/bookings/{id}/confirm
func (c *BookingController) ConfirmBooking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookingID, err := uuid.Parse(vars["id"])
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	var req struct {
		UserID uuid.UUID `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	confirmReq := usecase.ConfirmBookingRequest{
		BookingID: bookingID,
		UserID:    req.UserID,
	}

	if err := c.bookingUsecase.ConfirmBooking(r.Context(), confirmReq); err != nil {
		c.logger.Error("Failed to confirm booking", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to confirm booking")
		return
	}

	c.respondWithJSON(w, http.StatusOK, map[string]string{"status": "confirmed"})
}

// CancelBooking handles POST /api/bookings/{id}/cancel
func (c *BookingController) CancelBooking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookingID, err := uuid.Parse(vars["id"])
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	var req struct {
		UserID uuid.UUID `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cancelReq := usecase.CancelBookingRequest{
		BookingID: bookingID,
		UserID:    req.UserID,
	}

	if err := c.bookingUsecase.CancelBooking(r.Context(), cancelReq); err != nil {
		c.logger.Error("Failed to cancel booking", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to cancel booking")
		return
	}

	c.respondWithJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// GetUserBookings handles GET /api/users/{id}/bookings
func (c *BookingController) GetUserBookings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	bookings, err := c.bookingUsecase.GetUserBookings(r.Context(), userID)
	if err != nil {
		c.logger.Error("Failed to get user bookings", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to get user bookings")
		return
	}

	c.respondWithJSON(w, http.StatusOK, bookings)
}

// GetStats handles GET /api/bookings/stats
func (c *BookingController) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := c.bookingUsecase.GetConcurrencyStats()
	c.respondWithJSON(w, http.StatusOK, stats)
}

// Helper methods

func (c *BookingController) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (c *BookingController) respondWithError(w http.ResponseWriter, code int, message string) {
	c.respondWithJSON(w, code, map[string]string{"error": message})
}
