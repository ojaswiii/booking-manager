#!/bin/bash

# Load Testing Script for Ticket Booking System
# This script demonstrates various load testing scenarios

echo "🎫 Ticket Booking System - Load Testing Script"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if hey is installed
if ! command -v hey &> /dev/null; then
    echo -e "${RED}Error: 'hey' is not installed. Please install it first:${NC}"
    echo "go install github.com/rakyll/hey@latest"
    exit 1
fi

# Server URL
SERVER_URL="http://localhost:8080"

# Function to check if server is running
check_server() {
    echo -e "${BLUE}Checking if server is running...${NC}"
    if curl -s "$SERVER_URL/health" > /dev/null; then
        echo -e "${GREEN}✅ Server is running${NC}"
        return 0
    else
        echo -e "${RED}❌ Server is not running. Please start the server first:${NC}"
        echo "go run src/main.go"
        return 1
    fi
}

# Function to create test data
setup_test_data() {
    echo -e "${BLUE}Setting up test data...${NC}"
    
    # Create a test user
    USER_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/users" \
        -H "Content-Type: application/json" \
        -d '{"email": "test@example.com", "name": "Test User"}')
    
    USER_ID=$(echo $USER_RESPONSE | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
    echo "Created user: $USER_ID"
    
    # Create a test event
    EVENT_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/events" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Concert Test",
            "artist": "Test Artist",
            "venue": "Test Venue",
            "date": "2024-12-31T20:00:00Z",
            "total_seats": 1000,
            "price": 50.0
        }')
    
    EVENT_ID=$(echo $EVENT_RESPONSE | grep -o '"event_id":"[^"]*"' | cut -d'"' -f4)
    echo "Created event: $EVENT_ID"
    
    # Get available tickets
    TICKETS_RESPONSE=$(curl -s "$SERVER_URL/api/events/$EVENT_ID/tickets/available")
    echo "Available tickets count: $(echo $TICKETS_RESPONSE | grep -o '"id"' | wc -l)"
    
    # Export variables for other functions
    export USER_ID
    export EVENT_ID
}

# Function to run basic load test
basic_load_test() {
    echo -e "${YELLOW}🚀 Running Basic Load Test${NC}"
    echo "Testing: GET /api/events"
    echo "Requests: 1000, Concurrency: 10"
    
    hey -n 1000 -c 10 "$SERVER_URL/api/events"
}

# Function to run high concurrency test
high_concurrency_test() {
    echo -e "${YELLOW}🚀 Running High Concurrency Test${NC}"
    echo "Testing: GET /api/events"
    echo "Requests: 10000, Concurrency: 100"
    
    hey -n 10000 -c 100 "$SERVER_URL/api/events"
}

# Function to run ticket booking stress test
booking_stress_test() {
    echo -e "${YELLOW}🚀 Running Ticket Booking Stress Test${NC}"
    echo "This test will attempt to book tickets concurrently"
    
    # Create a temporary script for booking requests
    cat > /tmp/booking_test.sh << EOF
#!/bin/bash
for i in {1..100}; do
    curl -s -X POST "$SERVER_URL/api/bookings" \
        -H "Content-Type: application/json" \
        -d "{
            \"user_id\": \"$USER_ID\",
            \"event_id\": \"$EVENT_ID\",
            \"ticket_ids\": [\"$(uuidgen)\", \"$(uuidgen)\"]
        }" &
done
wait
EOF
    
    chmod +x /tmp/booking_test.sh
    echo "Running 100 concurrent booking attempts..."
    time /tmp/booking_test.sh
    rm /tmp/booking_test.sh
}

# Function to run memory usage test
memory_test() {
    echo -e "${YELLOW}🚀 Running Memory Usage Test${NC}"
    echo "This test will create many events and tickets to test memory usage"
    
    # Create multiple events
    for i in {1..10}; do
        curl -s -X POST "$SERVER_URL/api/events" \
            -H "Content-Type: application/json" \
            -d "{
                \"name\": \"Event $i\",
                \"artist\": \"Artist $i\",
                \"venue\": \"Venue $i\",
                \"date\": \"2024-12-31T20:00:00Z\",
                \"total_seats\": 1000,
                \"price\": 50.0
            }" > /dev/null &
    done
    wait
    
    echo "Created 10 events with 1000 tickets each (10,000 total tickets)"
    echo "Memory usage should be monitored during this test"
}

# Function to run benchmark comparison
benchmark_comparison() {
    echo -e "${YELLOW}🚀 Running Benchmark Comparison${NC}"
    echo "Comparing different concurrency levels..."
    
    echo -e "${BLUE}Test 1: Low Concurrency (c=5)${NC}"
    hey -n 1000 -c 5 "$SERVER_URL/api/events" | grep -E "(Requests/sec|Average|Fastest|Slowest)"
    
    echo -e "${BLUE}Test 2: Medium Concurrency (c=20)${NC}"
    hey -n 1000 -c 20 "$SERVER_URL/api/events" | grep -E "(Requests/sec|Average|Fastest|Slowest)"
    
    echo -e "${BLUE}Test 3: High Concurrency (c=50)${NC}"
    hey -n 1000 -c 50 "$SERVER_URL/api/events" | grep -E "(Requests/sec|Average|Fastest|Slowest)"
}

# Function to run all tests
run_all_tests() {
    echo -e "${GREEN}🎯 Running All Tests${NC}"
    
    check_server || exit 1
    setup_test_data
    
    echo -e "\n${BLUE}=== Test 1: Basic Load Test ===${NC}"
    basic_load_test
    
    echo -e "\n${BLUE}=== Test 2: High Concurrency Test ===${NC}"
    high_concurrency_test
    
    echo -e "\n${BLUE}=== Test 3: Booking Stress Test ===${NC}"
    booking_stress_test
    
    echo -e "\n${BLUE}=== Test 4: Memory Usage Test ===${NC}"
    memory_test
    
    echo -e "\n${BLUE}=== Test 5: Benchmark Comparison ===${NC}"
    benchmark_comparison
    
    echo -e "\n${GREEN}✅ All tests completed!${NC}"
}

# Main menu
show_menu() {
    echo -e "\n${BLUE}Select a test to run:${NC}"
    echo "1. Check server status"
    echo "2. Setup test data"
    echo "3. Basic load test"
    echo "4. High concurrency test"
    echo "5. Booking stress test"
    echo "6. Memory usage test"
    echo "7. Benchmark comparison"
    echo "8. Run all tests"
    echo "9. Exit"
    echo -n "Enter your choice (1-9): "
}

# Main execution
if [ $# -eq 0 ]; then
    while true; do
        show_menu
        read choice
        case $choice in
            1) check_server ;;
            2) setup_test_data ;;
            3) basic_load_test ;;
            4) high_concurrency_test ;;
            5) booking_stress_test ;;
            6) memory_test ;;
            7) benchmark_comparison ;;
            8) run_all_tests ;;
            9) echo "Goodbye!"; exit 0 ;;
            *) echo -e "${RED}Invalid choice. Please try again.${NC}" ;;
        esac
        echo -e "\nPress Enter to continue..."
        read
    done
else
    # Run specific test based on argument
    case $1 in
        "basic") basic_load_test ;;
        "concurrency") high_concurrency_test ;;
        "booking") booking_stress_test ;;
        "memory") memory_test ;;
        "benchmark") benchmark_comparison ;;
        "all") run_all_tests ;;
        *) echo "Usage: $0 [basic|concurrency|booking|memory|benchmark|all]"; exit 1 ;;
    esac
fi
