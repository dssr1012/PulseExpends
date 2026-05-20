#!/bin/bash

# PulseExpends Test Runner
# This script runs all tests for the PulseExpends project

set -e  # Exit on error

echo "========================================="
echo "PulseExpends Test Suite"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local status=$2
    local message=$3
    echo -e "${color}[${status}]${NC} ${message}"
}

# Function to run Go tests
run_go_tests() {
    echo -e "\n${BLUE}Running Go tests...${NC}"
    
    cd /root/PulseExpends
    
    # Install test dependencies if not already installed
    if ! go list -m github.com/stretchr/testify 2>/dev/null | grep -q testify; then
        print_status $YELLOW "INFO" "Installing testify..."
        go get github.com/stretchr/testify
    fi
    
    # Run tests
    if go test ./internal/mcp/test -v; then
        print_status $GREEN "PASS" "Go tests passed"
        return 0
    else
        print_status $RED "FAIL" "Go tests failed"
        return 1
    fi
}

# Function to run Python tests
run_python_tests() {
    echo -e "\n${BLUE}Running Python tests...${NC}"
    
    cd /root/PulseExpends/python/pdf-parser
    
    # Install dependencies if needed
    if [ ! -d "venv" ]; then
        print_status $YELLOW "INFO" "Creating virtual environment..."
        python3 -m venv venv
        source venv/bin/activate
        pip install -r requirements.txt
        pip install pytest pytest-asyncio pytest-cov
    else
        source venv/bin/activate
    fi
    
    # Run tests
    if python -m pytest tests/ -v --cov=src --cov-report=term-missing; then
        print_status $GREEN "PASS" "Python tests passed"
        return 0
    else
        print_status $RED "FAIL" "Python tests failed"
        return 1
    fi
}

# Function to run security tests
run_security_tests() {
    echo -e "\n${BLUE}Running security tests...${NC}"
    
    cd /root/PulseExpends
    
    # Install Python dependencies for security tests
    if [ ! -d "venv" ]; then
        print_status $YELLOW "INFO" "Creating virtual environment for security tests..."
        python3 -m venv venv
        source venv/bin/activate
        pip install requests
    else
        source venv/bin/activate
    fi
    
    # Check if server is running
    print_status $YELLOW "INFO" "Checking if server is running..."
    
    # Try to start the server if not running
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        print_status $YELLOW "INFO" "Starting MCP server for testing..."
        
        # Start server in background
        cd /root/PulseExpends
        go run cmd/mcp-server/main.go &
        SERVER_PID=$!
        
        # Wait for server to start
        sleep 5
        
        # Check if server started successfully
        if ! kill -0 $SERVER_PID 2>/dev/null; then
            print_status $RED "FAIL" "Failed to start server"
            return 1
        fi
        
        SERVER_STARTED=true
    else
        SERVER_STARTED=false
        print_status $GREEN "INFO" "Server is already running"
    fi
    
    # Run security tests
    print_status $YELLOW "INFO" "Running Red Team security tests..."
    
    if python tests/security/redteam_tests.py --url http://localhost:8080 --verbose; then
        print_status $GREEN "PASS" "Security tests passed"
        SECURITY_RESULT=0
    else
        print_status $RED "FAIL" "Security tests failed"
        SECURITY_RESULT=1
    fi
    
    # Stop server if we started it
    if [ "$SERVER_STARTED" = true ]; then
        print_status $YELLOW "INFO" "Stopping test server..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi
    
    return $SECURITY_RESULT
}

# Function to run integration tests
run_integration_tests() {
    echo -e "\n${BLUE}Running integration tests...${NC}"
    
    cd /root/PulseExpends/python/pdf-parser
    source venv/bin/activate 2>/dev/null || true
    
    # Run integration tests
    if python -m pytest tests/test_integration.py -v -k "integration"; then
        print_status $GREEN "PASS" "Integration tests passed"
        return 0
    else
        print_status $RED "FAIL" "Integration tests failed"
        return 1
    fi
}

# Function to run performance tests
run_performance_tests() {
    echo -e "\n${BLUE}Running performance tests...${NC}"
    
    cd /root/PulseExpends/python/pdf-parser
    source venv/bin/activate 2>/dev/null || true
    
    # Run performance tests
    if python -m pytest tests/test_integration.py::TestIntegration::test_performance_metrics -v; then
        print_status $GREEN "PASS" "Performance tests passed"
        return 0
    else
        print_status $RED "FAIL" "Performance tests failed"
        return 1
    fi
}

# Function to generate test report
generate_report() {
    echo -e "\n${BLUE}Generating test report...${NC}"
    
    REPORT_FILE="/root/PulseExpends/test_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$REPORT_FILE" << EOF
# PulseExpends Test Report
Generated: $(date)

## Test Results

### Go Tests
- Status: $1
- Details: MCP Server unit tests

### Python Tests
- Status: $2
- Details: PDF Parser API tests

### Security Tests
- Status: $3
- Details: Red Team security assessment

### Integration Tests
- Status: $4
- Details: End-to-end integration tests

### Performance Tests
- Status: $5
- Details: Performance and load testing

## Summary
- Total Tests: 5
- Passed: $(($1 == 0 ? 1 : 0 + $2 == 0 ? 1 : 0 + $3 == 0 ? 1 : 0 + $4 == 0 ? 1 : 0 + $5 == 0 ? 1 : 0))
- Failed: $(($1 != 0 ? 1 : 0 + $2 != 0 ? 1 : 0 + $3 != 0 ? 1 : 0 + $4 != 0 ? 1 : 0 + $5 != 0 ? 1 : 0))

## Recommendations
1. Address any failed tests immediately
2. Review security test warnings
3. Ensure all tests pass before deployment
4. Consider adding more edge case tests
5. Monitor performance in production

## Next Steps
1. Fix any failing tests
2. Run tests in CI/CD pipeline
3. Deploy to staging environment
4. Perform user acceptance testing
5. Deploy to production

EOF
    
    print_status $GREEN "INFO" "Test report generated: $REPORT_FILE"
    
    # Print report summary
    echo -e "\n${BLUE}Test Report Summary:${NC}"
    cat "$REPORT_FILE" | grep -A2 "###\|## Summary\|## Recommendations"
}

# Main test execution
main() {
    echo "Starting test suite..."
    
    # Initialize results
    GO_RESULT=0
    PYTHON_RESULT=0
    SECURITY_RESULT=0
    INTEGRATION_RESULT=0
    PERFORMANCE_RESULT=0
    
    # Run Go tests
    if run_go_tests; then
        GO_RESULT=0
    else
        GO_RESULT=1
    fi
    
    # Run Python tests
    if run_python_tests; then
        PYTHON_RESULT=0
    else
        PYTHON_RESULT=1
    fi
    
    # Run security tests
    if run_security_tests; then
        SECURITY_RESULT=0
    else
        SECURITY_RESULT=1
    fi
    
    # Run integration tests
    if run_integration_tests; then
        INTEGRATION_RESULT=0
    else
        INTEGRATION_RESULT=1
    fi
    
    # Run performance tests
    if run_performance_tests; then
        PERFORMANCE_RESULT=0
    else
        PERFORMANCE_RESULT=1
    fi
    
    # Generate report
    generate_report $GO_RESULT $PYTHON_RESULT $SECURITY_RESULT $INTEGRATION_RESULT $PERFORMANCE_RESULT
    
    # Calculate overall result
    TOTAL_FAILED=$((GO_RESULT + PYTHON_RESULT + SECURITY_RESULT + INTEGRATION_RESULT + PERFORMANCE_RESULT))
    
    echo -e "\n${BLUE}=========================================${NC}"
    echo -e "${BLUE}Test Suite Complete${NC}"
    echo -e "${BLUE}=========================================${NC}"
    
    if [ $TOTAL_FAILED -eq 0 ]; then
        echo -e "${GREEN}✅ ALL TESTS PASSED!${NC}"
        echo "The system is ready for deployment."
        exit 0
    else
        echo -e "${RED}❌ $TOTAL_FAILED TEST SUITE(S) FAILED${NC}"
        echo "Please review the test report and fix the issues."
        exit 1
    fi
}

# Parse command line arguments
case "$1" in
    "go")
        run_go_tests
        ;;
    "python")
        run_python_tests
        ;;
    "security")
        run_security_tests
        ;;
    "integration")
        run_integration_tests
        ;;
    "performance")
        run_performance_tests
        ;;
    "all"|"")
        main
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [suite]"
        echo ""
        echo "Test suites:"
        echo "  go          Run Go tests (MCP Server)"
        echo "  python      Run Python tests (PDF Parser)"
        echo "  security    Run security tests"
        echo "  integration Run integration tests"
        echo "  performance Run performance tests"
        echo "  all         Run all tests (default)"
        echo "  help        Show this help"
        ;;
    *)
        echo "Unknown option: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac