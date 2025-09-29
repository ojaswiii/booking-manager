package event

import (
	"ticket-booking-system/src/delivery/rest/controllers"
	"ticket-booking-system/src/utils"

	"github.com/gorilla/mux"
)

// RegisterEventRoutes registers all event-related routes
func RegisterEventRoutes(router *mux.Router, eventController *controllers.EventController, logger *utils.Logger) {
	// Event routes
	router.HandleFunc("/api/events", eventController.CreateEvent).Methods("POST")
	router.HandleFunc("/api/events", eventController.GetAllEvents).Methods("GET")
	router.HandleFunc("/api/events/{id}", eventController.GetEvent).Methods("GET")
	router.HandleFunc("/api/events/{id}/tickets", eventController.GetEventTickets).Methods("GET")
	router.HandleFunc("/api/events/{id}/tickets/available", eventController.GetAvailableTickets).Methods("GET")
}
