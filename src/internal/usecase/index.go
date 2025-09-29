package usecase

import (
	"github.com/ojaswiii/booking-manager/src/internal/repository"
	"github.com/ojaswiii/booking-manager/src/utils"
)

// UsecaseContainer holds all usecase instances
type UsecaseContainer struct {
	User    *UserUsecase
	Event   *EventUsecase
	Booking *BookingUsecase
}

// NewUsecaseContainer creates a new usecase container
func NewUsecaseContainer(repos *repository.RepositoryContainer, logger *utils.Logger) *UsecaseContainer {
	return &UsecaseContainer{
		User:    NewUserUsecase(repos.User, repos.UserCache, logger),
		Event:   NewEventUsecase(repos.Event, repos.EventCache, repos.Ticket, logger),
		Booking: NewBookingUsecase(repos.Booking, repos.Ticket, repos.Event, repos.User, logger),
	}
}
