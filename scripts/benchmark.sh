#!/bin/bash

# Benchmarking Script for Ticket Booking System
# This script runs various benchmarks and compares performance

echo "🎫 Ticket Booking System - Benchmarking Script"
echo "==============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to run Go benchmarks
run_go_benchmarks() {
    echo -e "${BLUE}Running Go benchmarks...${NC}"
    
    echo -e "${YELLOW}1. Concurrency Benchmarks${NC}"
    go test -bench=BenchmarkMutexApproach -benchmem ./internal/benchmarks/
    go test -bench=BenchmarkChannelApproach -benchmem ./internal/benchmarks/
    go test -bench=BenchmarkAtomicApproach -benchmem ./internal/benchmarks/
    go test -bench=BenchmarkWorkerPoolApproach -benchmem ./internal/benchmarks/
    
    echo -e "\n${YELLOW}2. Memory Usage Benchmarks${NC}"
    go test -bench=BenchmarkMemoryUsage -benchmem ./internal/benchmarks/
    
    echo -e "\n${YELLOW}3. Read/Write Benchmarks${NC}"
    go test -bench=BenchmarkConcurrentReads -benchmem ./internal/benchmarks/
    go test -bench=BenchmarkConcurrentWrites -benchmem ./internal/benchmarks/
    go test -bench=BenchmarkMixedOperations -benchmem ./internal/benchmarks/
}

# Function to run HTTP load tests
run_http_benchmarks() {
    echo -e "${BLUE}Running HTTP load tests...${NC}"
    
    # Check if server is running
    if ! curl -s http://localhost:8080/health > /dev/null; then
        echo -e "${RED}Server is not running. Please start it first:${NC}"
        echo "go run src/main.go"
        return 1
    fi
    
    echo -e "${YELLOW}1. Basic HTTP Load Test${NC}"
    hey -n 1000 -c 10 http://localhost:8080/api/events | grep -E "(Requests/sec|Average|Fastest|Slowest)"
    
    echo -e "\n${YELLOW}2. High Concurrency HTTP Test${NC}"
    hey -n 5000 -c 50 http://localhost:8080/api/events | grep -E "(Requests/sec|Average|Fastest|Slowest)"
    
    echo -e "\n${YELLOW}3. Memory Usage During HTTP Load${NC}"
    hey -n 10000 -c 100 http://localhost:8080/api/events | grep -E "(Requests/sec|Average|Fastest|Slowest)"
}

# Function to run comparison with other languages
run_language_comparison() {
    echo -e "${BLUE}Language Comparison Analysis${NC}"
    echo -e "${YELLOW}Go vs Other Languages for High-Concurrency Applications${NC}"
    
    cat << EOF

${BLUE}Go Advantages:${NC}
✅ Goroutines: Lightweight (2KB stack vs 2MB for OS threads)
✅ Channels: Built-in communication primitives
✅ Garbage Collector: Low-latency GC (sub-millisecond pauses)
✅ Compiler: Fast compilation and static typing
✅ Runtime: Built-in scheduler (M:N threading model)

${BLUE}Performance Characteristics:${NC}
- Memory Usage: ~8MB baseline + ~2KB per goroutine
- Context Switching: ~100ns (vs ~1μs for OS threads)
- Concurrency: Can handle 100K+ goroutines efficiently
- Throughput: 10K+ requests/second on single machine

${BLUE}Comparison with Other Languages:${NC}

${YELLOW}Node.js:${NC}
- Single-threaded event loop
- Good for I/O bound operations
- Limited by CPU cores for CPU-bound tasks
- Memory: ~10MB baseline + ~1KB per connection

${YELLOW}Python (asyncio):${NC}
- GIL (Global Interpreter Lock) limits true parallelism
- Good for I/O bound operations
- Slower than Go for CPU-bound tasks
- Memory: ~20MB baseline + ~8KB per task

${YELLOW}Java:${NC}
- Thread-based concurrency
- Higher memory overhead per thread (~1MB)
- Good performance but more resource intensive
- Memory: ~50MB baseline + ~1MB per thread

${YELLOW}Rust:${NC}
- Zero-cost abstractions
- No garbage collector
- Steeper learning curve
- Memory: ~1MB baseline + ~8KB per task

EOF
}

# Function to generate performance report
generate_performance_report() {
    echo -e "${BLUE}Generating Performance Report...${NC}"
    
    REPORT_FILE="performance_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$REPORT_FILE" << EOF
# Ticket Booking System - Performance Report

Generated: $(date)

## System Architecture

- **Language**: Go 1.21
- **Architecture**: Clean Architecture + Domain-Driven Design
- **Concurrency Model**: Goroutines + Channels + Mutexes
- **Storage**: In-memory (for testing purposes)

## Benchmark Results

### Concurrency Approaches

| Approach | Operations/sec | Memory/op | Allocations/op |
|----------|----------------|-----------|---------------|
| Mutex-based | TBD | TBD | TBD |
| Channel-based | TBD | TBD | TBD |
| Atomic operations | TBD | TBD | TBD |
| Worker pool | TBD | TBD | TBD |

### HTTP Performance

| Test | Requests/sec | Avg Response Time | 95th Percentile |
|------|--------------|------------------|-----------------|
| Basic Load | TBD | TBD | TBD |
| High Concurrency | TBD | TBD | TBD |
| Memory Stress | TBD | TBD | TBD |

## Key Findings

1. **Go's Concurrency Model**: Excellent for handling high-concurrency scenarios
2. **Memory Efficiency**: Low memory footprint per goroutine
3. **Scalability**: Can handle thousands of concurrent operations
4. **Clean Architecture**: Maintainable and testable code structure

## Recommendations

1. Use worker pools for CPU-intensive tasks
2. Use channels for coordination between goroutines
3. Use mutexes sparingly, prefer channels when possible
4. Monitor memory usage under load
5. Implement circuit breakers for fault tolerance

## Next Steps

1. Implement database persistence
2. Add Redis for caching
3. Implement rate limiting
4. Add monitoring and metrics
5. Implement distributed locking for multi-instance deployments

EOF

    echo -e "${GREEN}Performance report generated: $REPORT_FILE${NC}"
}

# Function to run all benchmarks
run_all_benchmarks() {
    echo -e "${GREEN}🎯 Running All Benchmarks${NC}"
    
    echo -e "\n${BLUE}=== Go Benchmarks ===${NC}"
    run_go_benchmarks
    
    echo -e "\n${BLUE}=== HTTP Load Tests ===${NC}"
    run_http_benchmarks
    
    echo -e "\n${BLUE}=== Language Comparison ===${NC}"
    run_language_comparison
    
    echo -e "\n${BLUE}=== Generating Report ===${NC}"
    generate_performance_report
    
    echo -e "\n${GREEN}✅ All benchmarks completed!${NC}"
}

# Main menu
show_menu() {
    echo -e "\n${BLUE}Select a benchmark to run:${NC}"
    echo "1. Run Go benchmarks"
    echo "2. Run HTTP load tests"
    echo "3. Show language comparison"
    echo "4. Generate performance report"
    echo "5. Run all benchmarks"
    echo "6. Exit"
    echo -n "Enter your choice (1-6): "
}

# Main execution
if [ $# -eq 0 ]; then
    while true; do
        show_menu
        read choice
        case $choice in
            1) run_go_benchmarks ;;
            2) run_http_benchmarks ;;
            3) run_language_comparison ;;
            4) generate_performance_report ;;
            5) run_all_benchmarks ;;
            6) echo "Goodbye!"; exit 0 ;;
            *) echo -e "${RED}Invalid choice. Please try again.${NC}" ;;
        esac
        echo -e "\nPress Enter to continue..."
        read
    done
else
    # Run specific benchmark based on argument
    case $1 in
        "go") run_go_benchmarks ;;
        "http") run_http_benchmarks ;;
        "comparison") run_language_comparison ;;
        "report") generate_performance_report ;;
        "all") run_all_benchmarks ;;
        *) echo "Usage: $0 [go|http|comparison|report|all]"; exit 1 ;;
    esac
fi
