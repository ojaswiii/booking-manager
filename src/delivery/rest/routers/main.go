package routers

import (
	"encoding/json"
	"net/http"
	"time"

	"ticket-booking-system/src/delivery/rest/controllers"
	"ticket-booking-system/src/delivery/rest/middlewares"
	"ticket-booking-system/src/delivery/rest/routers/booking"
	"ticket-booking-system/src/delivery/rest/routers/event"
	"ticket-booking-system/src/delivery/rest/routers/user"
	"ticket-booking-system/src/utils"

	"github.com/gorilla/mux"
)

// Router contains all route handlers
type Router struct {
	userController    *controllers.UserController
	eventController   *controllers.EventController
	bookingController *controllers.BookingController
	logger            *utils.Logger
}

// NewRouter creates a new router
func NewRouter(
	userController *controllers.UserController,
	eventController *controllers.EventController,
	bookingController *controllers.BookingController,
	logger *utils.Logger,
) *Router {
	return &Router{
		userController:    userController,
		eventController:   eventController,
		bookingController: bookingController,
		logger:            logger,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Add middleware
	router.Use(middlewares.CORS)
	router.Use(middlewares.Logging(r.logger))

	// Health check
	router.HandleFunc("/health", r.healthCheck).Methods("GET")

	// Register domain-specific routes
	user.RegisterUserRoutes(router, r.userController, r.logger)
	event.RegisterEventRoutes(router, r.eventController, r.logger)
	booking.RegisterBookingRoutes(router, r.bookingController, r.logger)

	return router
}

// healthCheck handles GET /health
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "ticket-booking-system",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
