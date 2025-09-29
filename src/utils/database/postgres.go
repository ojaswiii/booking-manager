package database

import (
	"context"
	"fmt"
	"time"

	"ticket-booking-system/src/utils"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresClient represents a PostgreSQL client
type PostgresClient struct {
	DB *sqlx.DB
}

// NewPostgresClient creates a new PostgreSQL client
func NewPostgresClient(config *utils.Config) (*PostgresClient, error) {
	// Create connection string
	connStr := config.GetDBConnectionString()

	// Connect to database
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgresClient{DB: db}, nil
}

// Close closes the database connection
func (c *PostgresClient) Close() error {
	return c.DB.Close()
}

// Ping tests the database connection
func (c *PostgresClient) Ping(ctx context.Context) error {
	return c.DB.PingContext(ctx)
}

// Health checks database health
func (c *PostgresClient) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.Ping(ctx)
}
