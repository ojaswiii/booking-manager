package rest

import (
	"github.com/ojaswiii/booking-manager/src/delivery/rest/controllers"
	"github.com/ojaswiii/booking-manager/src/delivery/rest/routers"
	"github.com/ojaswiii/booking-manager/src/internal/usecase"
	"github.com/ojaswiii/booking-manager/src/utils"
)

// RestContainer holds all REST delivery instances
type RestContainer struct {
	Router *routers.Router
}

// NewRestContainer creates a new REST container
func NewRestContainer(usecases *usecase.UsecaseContainer, logger *utils.Logger) *RestContainer {
	// Create controllers
	userController := controllers.NewUserController(usecases.User, logger)
	eventController := controllers.NewEventController(usecases.Event, logger)
	bookingController := controllers.NewBookingController(usecases.Booking, logger)

	// Create router
	router := routers.NewRouter(userController, eventController, bookingController, logger)

	return &RestContainer{
		Router: router,
	}
}
