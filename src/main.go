package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ojaswiii/booking-manager/src/delivery/rest"
	"github.com/ojaswiii/booking-manager/src/internal/repository"
	"github.com/ojaswiii/booking-manager/src/internal/usecase"
	"github.com/ojaswiii/booking-manager/src/utils"
	"github.com/ojaswiii/booking-manager/src/utils/database"
)

func main() {
	// Load configuration
	config := utils.LoadConfig()

	// Initialize logger
	logger := utils.NewLogger()
	logger.Info("Starting booking system with integrated concurrency", "environment", config.Environment)

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
	userUsecase := usecase.NewUserUsecase(repos.User, repos.UserCache, logger)
	eventUsecase := usecase.NewEventUsecase(repos.Event, repos.EventCache, repos.Ticket, logger)
	bookingUsecase := usecase.NewBookingUsecase(repos.Booking, repos.Ticket, repos.Event, repos.User, logger)
	defer bookingUsecase.Shutdown()

	// Create usecase container
	usecases := &usecase.UsecaseContainer{
		User:    userUsecase,
		Event:   eventUsecase,
		Booking: bookingUsecase,
	}

	logger.Info("Usecases initialized with integrated concurrency")

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

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server with integrated concurrency",
			"host", config.ServerHost,
			"port", config.ServerPort,
			"features", []string{
				"integrated_concurrency",
				"ticket_locks_with_expiration",
				"load_balanced_queues",
				"race_condition_handling",
				"automatic_cleanup",
			})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Start metrics reporting goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats := bookingUsecase.GetConcurrencyStats()
				logger.Info("Booking concurrency metrics", "stats", stats)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Cancel context to stop background services
	cancel()

	// Give outstanding requests 30 seconds to complete
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}
