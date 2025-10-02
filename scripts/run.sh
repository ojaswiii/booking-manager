#!/bin/bash

# Concurrent Booking Manager - Run Script
# Clean, focused solution for handling concurrent booking requests

echo "🎯 Starting Concurrent Booking System..."
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Add PostgreSQL to PATH
export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"

# Set environment variables
export SERVER_HOST=localhost
export SERVER_PORT=8080
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=ojaswi
export DB_PASSWORD=""
export DB_NAME=ticket_booking
export DB_SSL_MODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=""
export REDIS_DB=0
export ENV=development
export LOG_LEVEL=info
export BOOKING_EXPIRY_MINUTES=15

echo -e "${BLUE}🔍 Checking services...${NC}"

# Check PostgreSQL
if ! pg_isready -h localhost -p 5432 -U ojaswi > /dev/null 2>&1; then
    echo -e "${RED}❌ PostgreSQL is not running. Starting it...${NC}"
    brew services start postgresql@15
    sleep 2
fi

# Check Redis
if ! redis-cli ping > /dev/null 2>&1; then
    echo -e "${RED}❌ Redis is not running. Starting it...${NC}"
    brew services start redis
    sleep 2
fi

echo -e "${GREEN}✅ All services are running!${NC}"

echo -e "${CYAN}🎯 Concurrent Features:${NC}"
echo -e "  • ${GREEN}✅ Ticket-level locks with 10-minute expiration${NC}"
echo -e "  • ${GREEN}✅ 3 load-balanced queues (round-robin)${NC}"
echo -e "  • ${GREEN}✅ Race condition handling with forced ordering${NC}"
echo -e "  • ${GREEN}✅ Automatic lock cleanup${NC}"
echo -e "  • ${GREEN}✅ Clean architecture preserved${NC}"

echo -e "${BLUE}🌐 Starting concurrent application on http://localhost:8080${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
echo ""

# Run the application
go run src/main.go
