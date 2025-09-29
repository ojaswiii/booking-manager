package user

import (
	"ticket-booking-system/src/delivery/rest/controllers"
	"ticket-booking-system/src/utils"

	"github.com/gorilla/mux"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(router *mux.Router, userController *controllers.UserController, logger *utils.Logger) {
	// User routes
	router.HandleFunc("/api/users", userController.CreateUser).Methods("POST")
	router.HandleFunc("/api/users/{id}", userController.GetUser).Methods("GET")
	router.HandleFunc("/api/users/{id}", userController.UpdateUser).Methods("PUT")
	router.HandleFunc("/api/users/{id}", userController.DeleteUser).Methods("DELETE")
}
