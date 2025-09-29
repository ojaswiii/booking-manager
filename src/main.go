package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ticket-booking-system/src/delivery/rest"
	"ticket-booking-system/src/internal/repository"
	"ticket-booking-system/src/internal/usecase"
	"ticket-booking-system/src/utils"
	"ticket-booking-system/src/utils/database"
)

func main() {
	// Load configuration
	config := utils.LoadConfig()

	// Initialize logger
	logger := utils.NewLogger()
	logger.Info("Starting ticket booking system", "environment", config.Environment)

	// Initialize database connections
	postgresClient, err := database.NewPostgresClient(config)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer postgresClient.Close()

	redisClient, err := database.NewRedisClient(config)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// Initialize repositories
	repos := repository.NewRepositoryContainer(postgresClient.DB, redisClient.Client)
	logger.Info("Repositories initialized")

	// Initialize usecases
	usecases := usecase.NewUsecaseContainer(repos, logger)
	logger.Info("Usecases initialized")

	// Initialize REST delivery
	restContainer := rest.NewRestContainer(usecases, logger)
	router := restContainer.Router.SetupRoutes()
	logger.Info("REST delivery initialized")

	// Create server
	server := &http.Server{
		Addr:         config.ServerHost + ":" + config.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", "host", config.ServerHost, "port", config.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited")
}
