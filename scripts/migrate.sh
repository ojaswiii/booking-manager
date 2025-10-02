#!/bin/bash

# Database Migration Script for Ticket Booking System
# This script runs database migrations in the correct order

echo "🗄️  Running Database Migrations"
echo "==============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Database configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}
DB_NAME=${DB_NAME:-ticket_booking}

# Function to run a migration
run_migration() {
    local migration_dir=$1
    local direction=${2:-up}
    
    echo -e "${BLUE}Running migration: $migration_dir ($direction)${NC}"
    
    # Set password for psql
    export PGPASSWORD=$DB_PASSWORD
    
    # Run migration
    if [ "$direction" = "up" ]; then
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "database/migrations/$migration_dir/up.sql"
    else
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "database/migrations/$migration_dir/down.sql"
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Migration $migration_dir ($direction) completed${NC}"
    else
        echo -e "${RED}❌ Migration $migration_dir ($direction) failed${NC}"
        return 1
    fi
}

# Function to run all migrations up
migrate_up() {
    echo -e "${YELLOW}Running all migrations up...${NC}"
    
    run_migration "001_initial" "up" || return 1
    run_migration "002_users" "up" || return 1
    run_migration "003_events" "up" || return 1
    run_migration "004_tickets" "up" || return 1
    run_migration "005_bookings" "up" || return 1
    
    echo -e "${GREEN}✅ All migrations completed successfully${NC}"
}

# Function to run all migrations down
migrate_down() {
    echo -e "${YELLOW}Running all migrations down...${NC}"
    
    run_migration "005_bookings" "down" || return 1
    run_migration "004_tickets" "down" || return 1
    run_migration "003_events" "down" || return 1
    run_migration "002_users" "down" || return 1
    run_migration "001_initial" "down" || return 1
    
    echo -e "${GREEN}✅ All migrations rolled back successfully${NC}"
}

# Function to show migration status
show_status() {
    echo -e "${BLUE}Migration Status:${NC}"
    
    # Set password for psql
    export PGPASSWORD=$DB_PASSWORD
    
    # Show table counts
    echo -e "${CYAN}Table counts:${NC}"
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        SELECT 
            'users' as table_name, COUNT(*) as count FROM users
        UNION ALL
        SELECT 
            'events' as table_name, COUNT(*) as count FROM events
        UNION ALL
        SELECT 
            'tickets' as table_name, COUNT(*) as count FROM tickets
        UNION ALL
        SELECT 
            'bookings' as table_name, COUNT(*) as count FROM bookings;
    " 2>/dev/null || echo "No tables found or database not accessible"
}

# Main execution
case ${1:-up} in
    "up")
        migrate_up
        ;;
    "down")
        migrate_down
        ;;
    "status")
        show_status
        ;;
    "reset")
        migrate_down
        migrate_up
        ;;
    *)
        echo "Usage: $0 [up|down|status|reset]"
        echo "  up     - Run all migrations up (default)"
        echo "  down   - Run all migrations down"
        echo "  status - Show migration status"
        echo "  reset  - Rollback all and run up again"
        exit 1
        ;;
esac
