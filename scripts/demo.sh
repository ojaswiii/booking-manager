#!/bin/bash

# Demo Script for Ticket Booking System
# This script demonstrates the system capabilities with real examples

echo "🎫 Ticket Booking System - Interactive Demo"
echo "==========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Server URL
SERVER_URL="http://localhost:8080"

# Function to check if server is running
check_server() {
    echo -e "${BLUE}Checking server status...${NC}"
    if curl -s "$SERVER_URL/health" > /dev/null; then
        echo -e "${GREEN}✅ Server is running${NC}"
        return 0
    else
        echo -e "${RED}❌ Server is not running${NC}"
        echo -e "${YELLOW}Please start the server first:${NC}"
        echo "go run src/main.go"
        return 1
    fi
}

# Function to create demo data
setup_demo_data() {
    echo -e "${BLUE}Setting up demo data...${NC}"
    
    # Create users
    echo -e "${CYAN}Creating users...${NC}"
    USER1_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/users" \
        -H "Content-Type: application/json" \
        -d '{"email": "alice@example.com", "name": "Alice Johnson"}')
    USER1_ID=$(echo $USER1_RESPONSE | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
    echo "Created user: Alice Johnson ($USER1_ID)"
    
    USER2_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/users" \
        -H "Content-Type: application/json" \
        -d '{"email": "bob@example.com", "name": "Bob Smith"}')
    USER2_ID=$(echo $USER2_RESPONSE | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
    echo "Created user: Bob Smith ($USER2_ID)"
    
    # Create events
    echo -e "${CYAN}Creating events...${NC}"
    EVENT1_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/events" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Rock Concert 2024",
            "artist": "The Rockers",
            "venue": "Madison Square Garden",
            "date": "2024-12-31T20:00:00Z",
            "total_seats": 100,
            "price": 75.0
        }')
    EVENT1_ID=$(echo $EVENT1_RESPONSE | grep -o '"event_id":"[^"]*"' | cut -d'"' -f4)
    echo "Created event: Rock Concert 2024 ($EVENT1_ID)"
    
    EVENT2_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/events" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Jazz Night",
            "artist": "Smooth Jazz Band",
            "venue": "Blue Note",
            "date": "2024-12-25T19:30:00Z",
            "total_seats": 50,
            "price": 45.0
        }')
    EVENT2_ID=$(echo $EVENT2_RESPONSE | grep -o '"event_id":"[^"]*"' | cut -d'"' -f4)
    echo "Created event: Jazz Night ($EVENT2_ID)"
    
    # Export variables
    export USER1_ID
    export USER2_ID
    export EVENT1_ID
    export EVENT2_ID
    
    echo -e "${GREEN}✅ Demo data setup complete${NC}"
}

# Function to demonstrate ticket booking
demonstrate_booking() {
    echo -e "${BLUE}Demonstrating ticket booking...${NC}"
    
    # Get available tickets for event 1
    echo -e "${CYAN}Getting available tickets for Rock Concert...${NC}"
    TICKETS_RESPONSE=$(curl -s "$SERVER_URL/api/events/$EVENT1_ID/tickets/available")
    TICKET_COUNT=$(echo $TICKETS_RESPONSE | grep -o '"id"' | wc -l)
    echo "Available tickets: $TICKET_COUNT"
    
    # Get first two ticket IDs
    TICKET1_ID=$(echo $TICKETS_RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    TICKET2_ID=$(echo $TICKETS_RESPONSE | grep -o '"id":"[^"]*"' | head -2 | tail -1 | cut -d'"' -f4)
    
    echo "Selected tickets: $TICKET1_ID, $TICKET2_ID"
    
    # Create booking for Alice
    echo -e "${CYAN}Alice is booking 2 tickets...${NC}"
    BOOKING_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/bookings" \
        -H "Content-Type: application/json" \
        -d "{
            \"user_id\": \"$USER1_ID\",
            \"event_id\": \"$EVENT1_ID\",
            \"ticket_ids\": [\"$TICKET1_ID\", \"$TICKET2_ID\"]
        }")
    
    BOOKING_ID=$(echo $BOOKING_RESPONSE | grep -o '"booking_id":"[^"]*"' | cut -d'"' -f4)
    TOTAL_AMOUNT=$(echo $BOOKING_RESPONSE | grep -o '"total_amount":[0-9.]*' | cut -d':' -f2)
    EXPIRES_AT=$(echo $BOOKING_RESPONSE | grep -o '"expires_at":"[^"]*"' | cut -d'"' -f4)
    
    echo "Booking created: $BOOKING_ID"
    echo "Total amount: \$$TOTAL_AMOUNT"
    echo "Expires at: $EXPIRES_AT"
    
    # Confirm booking
    echo -e "${CYAN}Confirming booking...${NC}"
    curl -s -X POST "$SERVER_URL/api/bookings/$BOOKING_ID/confirm" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\": \"$USER1_ID\"}" > /dev/null
    echo "Booking confirmed!"
    
    # Show Alice's bookings
    echo -e "${CYAN}Alice's bookings:${NC}"
    curl -s "$SERVER_URL/api/users/$USER1_ID/bookings" | jq '.[] | {booking_id: .id, event_id: .event_id, status: .status, total_amount: .total_amount}'
}

# Function to demonstrate concurrent booking
demonstrate_concurrent_booking() {
    echo -e "${BLUE}Demonstrating concurrent booking (stress test)...${NC}"
    echo -e "${YELLOW}This will show how the system handles multiple users booking simultaneously${NC}"
    
    # Create a script for concurrent bookings
    cat > /tmp/concurrent_booking.sh << EOF
#!/bin/bash
for i in {1..20}; do
    (
        # Get available tickets
        TICKETS=\$(curl -s "$SERVER_URL/api/events/$EVENT2_ID/tickets/available")
        TICKET1=\$(echo \$TICKETS | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        TICKET2=\$(echo \$TICKETS | grep -o '"id":"[^"]*"' | head -2 | tail -1 | cut -d'"' -f4)
        
        if [ ! -z "\$TICKET1" ] && [ ! -z "\$TICKET2" ]; then
            # Create booking
            BOOKING=\$(curl -s -X POST "$SERVER_URL/api/bookings" \
                -H "Content-Type: application/json" \
                -d "{
                    \"user_id\": \"$USER2_ID\",
                    \"event_id\": \"$EVENT2_ID\",
                    \"ticket_ids\": [\"\$TICKET1\", \"\$TICKET2\"]
                }")
            
            BOOKING_ID=\$(echo \$BOOKING | grep -o '"booking_id":"[^"]*"' | cut -d'"' -f4)
            if [ ! -z "\$BOOKING_ID" ]; then
                echo "Booking \$i: \$BOOKING_ID"
                # Confirm booking
                curl -s -X POST "$SERVER_URL/api/bookings/\$BOOKING_ID/confirm" \
                    -H "Content-Type: application/json" \
                    -d "{\"user_id\": \"$USER2_ID\"}" > /dev/null
            else
                echo "Booking \$i: Failed (no tickets available)"
            fi
        else
            echo "Booking \$i: No tickets available"
        fi
    ) &
done
wait
EOF
    
    chmod +x /tmp/concurrent_booking.sh
    echo "Running 20 concurrent booking attempts..."
    time /tmp/concurrent_booking.sh
    rm /tmp/concurrent_booking.sh
    
    # Show final ticket status
    echo -e "${CYAN}Final ticket status for Jazz Night:${NC}"
    curl -s "$SERVER_URL/api/events/$EVENT2_ID/tickets" | jq '.[] | {seat_number: .seat_number, status: .status}' | head -10
}

# Function to demonstrate system monitoring
demonstrate_monitoring() {
    echo -e "${BLUE}System monitoring and statistics...${NC}"
    
    # Show all events
    echo -e "${CYAN}All events:${NC}"
    curl -s "$SERVER_URL/api/events" | jq '.[] | {name: .name, artist: .artist, venue: .venue, total_seats: .total_seats, price: .price}'
    
    # Show booking statistics
    echo -e "\n${CYAN}Booking statistics:${NC}"
    echo "Alice's bookings:"
    curl -s "$SERVER_URL/api/users/$USER1_ID/bookings" | jq 'length'
    echo "Bob's bookings:"
    curl -s "$SERVER_URL/api/users/$USER2_ID/bookings" | jq 'length'
    
    # Show ticket availability
    echo -e "\n${CYAN}Ticket availability:${NC}"
    echo "Rock Concert available tickets:"
    curl -s "$SERVER_URL/api/events/$EVENT1_ID/tickets/available" | jq 'length'
    echo "Jazz Night available tickets:"
    curl -s "$SERVER_URL/api/events/$EVENT2_ID/tickets/available" | jq 'length'
}

# Function to demonstrate error handling
demonstrate_error_handling() {
    echo -e "${BLUE}Demonstrating error handling...${NC}"
    
    # Try to book non-existent tickets
    echo -e "${CYAN}Attempting to book non-existent tickets...${NC}"
    curl -s -X POST "$SERVER_URL/api/bookings" \
        -H "Content-Type: application/json" \
        -d '{
            "user_id": "'$USER1_ID'",
            "event_id": "'$EVENT1_ID'",
            "ticket_ids": ["00000000-0000-0000-0000-000000000000"]
        }' | jq '.error'
    
    # Try to access non-existent user
    echo -e "${CYAN}Attempting to access non-existent user...${NC}"
    curl -s "$SERVER_URL/api/users/00000000-0000-0000-0000-000000000000" | jq '.error'
    
    # Try to confirm non-existent booking
    echo -e "${CYAN}Attempting to confirm non-existent booking...${NC}"
    curl -s -X POST "$SERVER_URL/api/bookings/00000000-0000-0000-0000-000000000000/confirm" \
        -H "Content-Type: application/json" \
        -d '{"user_id": "'$USER1_ID'"}' | jq '.error'
}

# Function to run performance test
demonstrate_performance() {
    echo -e "${BLUE}Running performance test...${NC}"
    echo -e "${YELLOW}This will test the system under load${NC}"
    
    # Check if hey is installed
    if ! command -v hey &> /dev/null; then
        echo -e "${RED}hey is not installed. Installing...${NC}"
        go install github.com/rakyll/hey@latest
    fi
    
    echo -e "${CYAN}Testing GET /api/events with 1000 requests, 10 concurrent...${NC}"
    hey -n 1000 -c 10 "$SERVER_URL/api/events" | grep -E "(Requests/sec|Average|Fastest|Slowest)"
    
    echo -e "\n${CYAN}Testing GET /api/events with 5000 requests, 50 concurrent...${NC}"
    hey -n 5000 -c 50 "$SERVER_URL/api/events" | grep -E "(Requests/sec|Average|Fastest|Slowest)"
}

# Function to show system architecture
show_architecture() {
    echo -e "${BLUE}System Architecture Overview${NC}"
    cat << EOF

${PURPLE}🏗️  Architecture Layers:${NC}

${CYAN}1. HTTP Interface Layer${NC}
   ├── HTTP Handlers (REST API)
   ├── Request/Response mapping
   └── Error handling

${CYAN}2. Application Layer${NC}
   ├── Booking Service (business logic)
   ├── Event Service (event management)
   └── User Service (user management)

${CYAN}3. Domain Layer${NC}
   ├── Entities (Event, Ticket, Booking, User)
   ├── Value Objects (Status, IDs)
   └── Repository Interfaces

${CYAN}4. Infrastructure Layer${NC}
   ├── In-Memory Repositories
   ├── Concurrency Control (Mutex, Channels)
   └── Data Persistence

${PURPLE}🔄 Concurrency Patterns:${NC}

${YELLOW}• Mutex-based Locking${NC}
   - Protects shared resources
   - Simple but can create bottlenecks

${YELLOW}• Channel-based Coordination${NC}
   - Go-idiomatic communication
   - Prevents race conditions

${YELLOW}• Atomic Operations${NC}
   - Lock-free programming
   - Very fast for simple operations

${YELLOW}• Worker Pool Pattern${NC}
   - Controlled concurrency
   - Prevents resource exhaustion

${PURPLE}📊 Performance Characteristics:${NC}

${GREEN}• Memory Efficiency:${NC}
   - Goroutine: ~2KB (vs 2MB for OS threads)
   - Channel: ~96 bytes + buffer
   - Mutex: ~8 bytes

${GREEN}• Concurrency Limits:${NC}
   - Max goroutines: 100K+
   - Context switch: ~100ns
   - Throughput: 10K+ req/sec

EOF
}

# Main demo menu
show_demo_menu() {
    echo -e "\n${BLUE}Demo Menu:${NC}"
    echo "1. Check server status"
    echo "2. Setup demo data"
    echo "3. Demonstrate ticket booking"
    echo "4. Demonstrate concurrent booking"
    echo "5. Show system monitoring"
    echo "6. Demonstrate error handling"
    echo "7. Run performance test"
    echo "8. Show system architecture"
    echo "9. Run full demo"
    echo "0. Exit"
    echo -n "Enter your choice (0-9): "
}

# Function to run full demo
run_full_demo() {
    echo -e "${GREEN}🎯 Running Full Demo${NC}"
    
    check_server || return 1
    setup_demo_data
    echo -e "\n${BLUE}=== Ticket Booking Demo ===${NC}"
    demonstrate_booking
    echo -e "\n${BLUE}=== Concurrent Booking Demo ===${NC}"
    demonstrate_concurrent_booking
    echo -e "\n${BLUE}=== System Monitoring ===${NC}"
    demonstrate_monitoring
    echo -e "\n${BLUE}=== Error Handling Demo ===${NC}"
    demonstrate_error_handling
    echo -e "\n${BLUE}=== Performance Test ===${NC}"
    demonstrate_performance
    echo -e "\n${BLUE}=== Architecture Overview ===${NC}"
    show_architecture
    
    echo -e "\n${GREEN}✅ Full demo completed!${NC}"
}

# Main execution
if [ $# -eq 0 ]; then
    while true; do
        show_demo_menu
        read choice
        case $choice in
            1) check_server ;;
            2) setup_demo_data ;;
            3) demonstrate_booking ;;
            4) demonstrate_concurrent_booking ;;
            5) demonstrate_monitoring ;;
            6) demonstrate_error_handling ;;
            7) demonstrate_performance ;;
            8) show_architecture ;;
            9) run_full_demo ;;
            0) echo "Goodbye!"; exit 0 ;;
            *) echo -e "${RED}Invalid choice. Please try again.${NC}" ;;
        esac
        echo -e "\nPress Enter to continue..."
        read
    done
else
    # Run specific demo based on argument
    case $1 in
        "setup") setup_demo_data ;;
        "booking") demonstrate_booking ;;
        "concurrent") demonstrate_concurrent_booking ;;
        "monitoring") demonstrate_monitoring ;;
        "errors") demonstrate_error_handling ;;
        "performance") demonstrate_performance ;;
        "architecture") show_architecture ;;
        "full") run_full_demo ;;
        *) echo "Usage: $0 [setup|booking|concurrent|monitoring|errors|performance|architecture|full]"; exit 1 ;;
    esac
fi
