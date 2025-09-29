package booking

import (
	"ticket-booking-system/src/delivery/rest/controllers"
	"ticket-booking-system/src/utils"

	"github.com/gorilla/mux"
)

// RegisterBookingRoutes registers all booking-related routes
func RegisterBookingRoutes(router *mux.Router, bookingController *controllers.BookingController, logger *utils.Logger) {
	// Booking routes
	router.HandleFunc("/api/bookings", bookingController.CreateBooking).Methods("POST")
	router.HandleFunc("/api/bookings/{id}/confirm", bookingController.ConfirmBooking).Methods("POST")
	router.HandleFunc("/api/bookings/{id}/cancel", bookingController.CancelBooking).Methods("POST")
	router.HandleFunc("/api/users/{id}/bookings", bookingController.GetUserBookings).Methods("GET")
}
