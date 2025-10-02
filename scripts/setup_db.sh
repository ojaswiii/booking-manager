#!/bin/bash

# Database Setup Script for Ticket Booking System
# This script sets up PostgreSQL database and runs migrations

echo "🗄️  Setting up PostgreSQL database for Ticket Booking System"
echo "============================================================="

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

# Function to check if PostgreSQL is running
check_postgres() {
    echo -e "${BLUE}Checking PostgreSQL connection...${NC}"
    if pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; then
        echo -e "${GREEN}✅ PostgreSQL is running${NC}"
        return 0
    else
        echo -e "${RED}❌ PostgreSQL is not running${NC}"
        echo -e "${YELLOW}Please start PostgreSQL first:${NC}"
        echo "brew services start postgresql  # macOS"
        echo "sudo systemctl start postgresql # Linux"
        return 1
    fi
}

# Function to create database
create_database() {
    echo -e "${BLUE}Creating database...${NC}"
    
    # Set password for psql
    export PGPASSWORD=$DB_PASSWORD
    
    # Check if database exists
    if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
        echo -e "${YELLOW}Database $DB_NAME already exists${NC}"
    else
        # Create database
        createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ Database $DB_NAME created successfully${NC}"
        else
            echo -e "${RED}❌ Failed to create database${NC}"
            return 1
        fi
    fi
}

# Function to run migrations
run_migrations() {
    echo -e "${BLUE}Running database migrations...${NC}"
    
    # Set password for psql
    export PGPASSWORD=$DB_PASSWORD
    
    # Run each migration file
    for migration in migrations/*.sql; do
        if [ -f "$migration" ]; then
            echo -e "${CYAN}Running $(basename $migration)...${NC}"
            psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration"
            if [ $? -eq 0 ]; then
                echo -e "${GREEN}✅ $(basename $migration) completed${NC}"
            else
                echo -e "${RED}❌ $(basename $migration) failed${NC}"
                return 1
            fi
        fi
    done
}

# Function to verify tables
verify_tables() {
    echo -e "${BLUE}Verifying database tables...${NC}"
    
    # Set password for psql
    export PGPASSWORD=$DB_PASSWORD
    
    # Check if all tables exist
    tables=("users" "events" "tickets" "bookings")
    for table in "${tables[@]}"; do
        if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\dt $table" | grep -q "$table"; then
            echo -e "${GREEN}✅ Table $table exists${NC}"
        else
            echo -e "${RED}❌ Table $table does not exist${NC}"
            return 1
        fi
    done
}

# Function to create sample data
create_sample_data() {
    echo -e "${BLUE}Creating sample data...${NC}"
    
    # Set password for psql
    export PGPASSWORD=$DB_PASSWORD
    
    # Create sample users
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        INSERT INTO users (id, email, name) VALUES 
        ('550e8400-e29b-41d4-a716-446655440000', 'alice@example.com', 'Alice Johnson'),
        ('550e8400-e29b-41d4-a716-446655440001', 'bob@example.com', 'Bob Smith')
        ON CONFLICT (email) DO NOTHING;
    "
    
    # Create sample events
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        INSERT INTO events (id, name, artist, venue, date, total_seats, price) VALUES 
        ('550e8400-e29b-41d4-a716-446655440010', 'Rock Concert 2024', 'The Rockers', 'Madison Square Garden', '2024-12-31 20:00:00+00', 100, 75.00),
        ('550e8400-e29b-41d4-a716-446655440011', 'Jazz Night', 'Smooth Jazz Band', 'Blue Note', '2024-12-25 19:30:00+00', 50, 45.00)
        ON CONFLICT (id) DO NOTHING;
    "
    
    # Create sample tickets for first event
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        INSERT INTO tickets (event_id, seat_number, price) 
        SELECT '550e8400-e29b-41d4-a716-446655440010', generate_series(1, 100), 75.00
        ON CONFLICT (event_id, seat_number) DO NOTHING;
    "
    
    # Create sample tickets for second event
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        INSERT INTO tickets (event_id, seat_number, price) 
        SELECT '550e8400-e29b-41d4-a716-446655440011', generate_series(1, 50), 45.00
        ON CONFLICT (event_id, seat_number) DO NOTHING;
    "
    
    echo -e "${GREEN}✅ Sample data created successfully${NC}"
}

# Function to show database status
show_status() {
    echo -e "${BLUE}Database Status:${NC}"
    
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
    "
}

# Main execution
main() {
    echo -e "${YELLOW}Starting database setup...${NC}"
    
    # Check PostgreSQL connection
    if ! check_postgres; then
        exit 1
    fi
    
    # Create database
    if ! create_database; then
        exit 1
    fi
    
    # Run migrations
    if ! run_migrations; then
        exit 1
    fi
    
    # Verify tables
    if ! verify_tables; then
        exit 1
    fi
    
    # Create sample data
    create_sample_data
    
    # Show status
    show_status
    
    echo -e "${GREEN}✅ Database setup completed successfully!${NC}"
    echo -e "${BLUE}You can now start the application with:${NC}"
    echo "go run cmd/server/main.go"
}

# Run main function
main
