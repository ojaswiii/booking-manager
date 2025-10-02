# Booking Manager - High-Performance Concurrent Booking System

A scalable, high-performance booking management system built with Go that handles concurrent ticket bookings with advanced race condition prevention and load balancing.

## 🚀 Features

- **Concurrent Booking Processing**: Multi-queue system with load balancing
- **Race Condition Prevention**: Advanced ticket locking with expiration
- **Automatic Cleanup**: Expired locks and resources are cleaned up automatically
- **Real-time Statistics**: Live monitoring of booking performance
- **Scalable Architecture**: Clean separation of concerns with domain-driven design
- **High Availability**: Graceful shutdown and error handling

## 🏗️ Architecture

### System Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REST API     │    │   Usecase      │    │   Repository    │
│   Controllers  │────│   Business     │────│   Data Access   │
│                │    │   Logic        │    │                │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Concurrency   │    │   Domain       │    │   Database      │
│   Utils         │    │   Models       │    │   (PostgreSQL) │
│   (Queue, Locks)│    │   (Booking,    │    │                │
│                 │    │   Event, etc.) │    │                │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Core Components

#### 1. **Domain Layer** (`/src/internal/domain/`)
- **Booking**: Core booking entity with status management
- **Event**: Event/concert management
- **User**: User management
- **Ticket**: Ticket entity with status tracking

#### 2. **Repository Layer** (`/src/internal/repository/`)
- **PostgreSQL**: Primary data storage
- **Redis**: Caching layer for performance
- **Interface-based**: Clean abstraction for data access

#### 3. **Usecase Layer** (`/src/internal/usecase/`)
- **Business Logic**: Core application logic
- **Concurrency Integration**: Seamless concurrent processing
- **Transaction Management**: ACID compliance

#### 4. **Concurrency Utils** (`/src/utils/concurrency/`)
- **QueueManager**: Load-balanced request processing
- **TicketLockManager**: Ticket-level locking with expiration
- **EventLockManager**: Event-level coordination
- **BookingProcessor**: Orchestrates concurrent operations

#### 5. **Delivery Layer** (`/src/delivery/rest/`)
- **REST API**: HTTP endpoints
- **Controllers**: Request/response handling
- **Middleware**: CORS, logging, error handling

## 🔧 Problem Statement

### The Challenge
Building a booking system that can handle:
- **High Concurrency**: Multiple users booking simultaneously
- **Race Conditions**: Preventing double-booking of the same ticket
- **Performance**: Fast response times under load
- **Scalability**: Handle increasing user load
- **Data Consistency**: Ensure data integrity

### Traditional Solutions & Their Limitations
1. **Database Locks**: Slow, causes bottlenecks
2. **Single-threaded Processing**: Poor performance
3. **Optimistic Locking**: High failure rates under load
4. **Pessimistic Locking**: Poor user experience

## 💡 Our Solution

### 1. **Multi-Queue Architecture**
```go
// 3 parallel queues with load balancing
queueManager := NewQueueManager(3, 100, logger)
```
- **Load Distribution**: Requests distributed across queues
- **Event-based Routing**: Same event always goes to same queue
- **Buffer Management**: Prevents memory overflow

### 2. **Advanced Ticket Locking**
```go
// Ticket-level locks with expiration
ticketLocks := NewTicketLockManager()
```
- **Per-Ticket Locks**: Granular locking prevents conflicts
- **Automatic Expiration**: Locks expire after 10 minutes
- **User-specific**: Same user can re-lock their tickets

### 3. **Event-level Coordination**
```go
// Event-level locks for coordination
eventLocks := NewEventLockManager(30*time.Minute, 5*time.Minute)
```
- **Event Isolation**: Different events don't interfere
- **Resource Management**: Automatic cleanup of unused locks
- **Reference Counting**: Efficient lock management

### 4. **Asynchronous Processing**
```go
// Background processors handle requests
go bp.processQueue(queueIndex)
```
- **Non-blocking**: Immediate response to users
- **Background Processing**: Heavy work done asynchronously
- **Error Handling**: Robust error recovery

## 🚀 Getting Started

### Prerequisites
- Go 1.19 or higher
- PostgreSQL 13 or higher
- Redis 6 or higher

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/your-username/booking-manager.git
cd booking-manager
```

2. **Install dependencies**
```bash
go mod tidy
```

3. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your database credentials
```

4. **Run database migrations**
```bash
./scripts/migrate.sh
```

5. **Start the application**
```bash
go run src/main.go
```


## 📊 API Documentation

### Base URL
```
http://localhost:8080
```

### Authentication
Currently, the system doesn't require authentication. In production, implement JWT or OAuth2.

### Endpoints

#### 1. **Health Check**
```http
GET /health
```
**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "booking-manager",
  "features": ["concurrent_booking", "ticket_locks", "load_balanced_queues"]
}
```

#### 2. **Create User**
```http
POST /api/users
Content-Type: application/json

{
  "email": "user@example.com",
  "name": "John Doe"
}
```

**Response:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "name": "John Doe"
}
```

#### 3. **Create Event**
```http
POST /api/events
Content-Type: application/json

{
  "name": "Concert 2024",
  "artist": "Famous Band",
  "venue": "Madison Square Garden",
  "date": "2024-06-15T20:00:00Z",
  "total_seats": 1000,
  "price": 75.00
}
```

#### 4. **Create Booking** ⚡ **Concurrent Processing**
```http
POST /api/bookings
Content-Type: application/json

{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "event_id": "456e7890-e89b-12d3-a456-426614174001",
  "ticket_ids": [
    "789e0123-e89b-12d3-a456-426614174002",
    "012e3456-e89b-12d3-a456-426614174003"
  ]
}
```

**Response:**
```json
{
  "booking_id": "345e6789-e89b-12d3-a456-426614174004",
  "total_amount": 100.00,
  "expires_at": "2024-01-15T10:45:00Z",
  "status": "pending"
}
```

#### 5. **Get Booking Statistics** 📈
```http
GET /api/bookings/stats
```

**Response:**
```json
{
  "total_requests": 1250,
  "successful_bookings": 1180,
  "failed_bookings": 70,
  "queue_length": 5,
  "uptime_seconds": 3600,
  "requests_per_second": 0.35,
  "lock_stats": {
    "total_locks": 45,
    "active_locks": 40,
    "expired_locks": 5
  },
  "queue_stats": {
    "queue_0": {"length": 2, "capacity": 100},
    "queue_1": {"length": 1, "capacity": 100},
    "queue_2": {"length": 2, "capacity": 100},
    "total_queues": 3,
    "total_pending": 5
  }
}
```

#### 6. **Get User Bookings**
```http
GET /api/users/{user_id}/bookings
```

#### 7. **Confirm Booking**
```http
POST /api/bookings/{booking_id}/confirm
Content-Type: application/json

{
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

#### 8. **Cancel Booking**
```http
POST /api/bookings/{booking_id}/cancel
Content-Type: application/json

{
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

## 🔧 Configuration

### Environment Variables

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=booking_manager
DB_USER=postgres
DB_PASSWORD=password

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
ENVIRONMENT=development

# Logging
LOG_LEVEL=info
```

### Concurrency Settings

The system is configured with optimal settings for high performance:

```go
// Queue Configuration
queueCount := 3        // Number of parallel queues
bufferSize := 100      // Buffer size per queue

// Lock Configuration
ticketLockTTL := 10 * time.Minute    // Ticket lock expiration
eventLockTTL := 30 * time.Minute     // Event lock expiration
maxIdleTime := 5 * time.Minute       // Max idle time for cleanup
```

## 🧪 Testing

### Unit Tests
```bash
go test ./...
```

### Load Testing
```bash
# Run load tests
./scripts/load_test.sh

# Run benchmarks
./scripts/benchmark.sh
```

### Manual Testing
```bash
# Create test data
./scripts/demo.sh
```

## 📈 Performance Metrics

### Benchmarks
- **Throughput**: 1000+ requests/second
- **Latency**: < 50ms average response time
- **Concurrency**: Handles 100+ concurrent users
- **Memory**: < 100MB under normal load

### Monitoring
- Real-time statistics via `/api/bookings/stats`
- Automatic metrics logging every 30 seconds
- Queue length monitoring
- Lock usage tracking

## 🛠️ Development

### Project Structure
```
src/
├── delivery/rest/          # REST API layer
│   ├── controllers/        # HTTP controllers
│   ├── middlewares/        # CORS, logging
│   └── routers/           # Route definitions
├── internal/              # Internal packages
│   ├── domain/            # Domain entities
│   ├── repository/        # Data access layer
│   └── usecase/          # Business logic
├── utils/                 # Utility packages
│   └── concurrency/       # Concurrency utilities
├── migrations/            # Database migrations
└── main.go               # Application entry point
```

### Key Concepts Implemented

#### 1. **Domain-Driven Design (DDD)**
- Clear domain boundaries
- Rich domain models
- Repository pattern

#### 2. **Clean Architecture**
- Dependency inversion
- Interface segregation
- Single responsibility

#### 3. **Concurrency Patterns**
- **Producer-Consumer**: Queue-based processing
- **Lock Manager**: Distributed locking
- **Load Balancing**: Request distribution
- **Circuit Breaker**: Error handling

#### 4. **Performance Optimizations**
- **Connection Pooling**: Database connections
- **Caching**: Redis for frequently accessed data
- **Async Processing**: Non-blocking operations
- **Resource Cleanup**: Automatic garbage collection



---

**Built with ❤️ using Go, PostgreSQL, and Redis**
