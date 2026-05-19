# Testing Guide for PulseExpends

This document provides comprehensive testing guidelines for the PulseExpends project, including unit tests, integration tests, security tests, and performance tests.

## Table of Contents

1. [Test Structure](#test-structure)
2. [Running Tests](#running-tests)
3. [Test Types](#test-types)
4. [Security Testing](#security-testing)
5. [Performance Testing](#performance-testing)
6. [Continuous Integration](#continuous-integration)
7. [Test Coverage](#test-coverage)
8. [Troubleshooting](#troubleshooting)

## Test Structure

### Go Tests (MCP Server)
```
/root/PulseExpends/
├── internal/mcp/
│   ├── server.go          # Main server implementation
│   ├── tools.go           # Tool definitions and handlers
│   └── test/
│       └── server_test.go # Unit tests for MCP server
└── cmd/mcp-server/
    └── main.go            # Server entry point
```

### Python Tests (PDF Parser)
```
/root/PulseExpends/python/pdf-parser/
├── src/
│   ├── main.py            # FastAPI application
│   ├── services/
│   │   ├── parser.py      # PDF parsing service
│   │   ├── ocr.py         # OCR service
│   │   ├── ai.py          # AI service
│   │   └── obs.py         # OBS service
│   └── models/
│       └── document.py    # Data models
└── tests/
    ├── test_api.py        # API endpoint tests
    ├── test_integration.py # Integration tests
    └── conftest.py        # Test fixtures
```

### Security Tests
```
/root/PulseExpends/tests/security/
└── redteam_tests.py       # Red Team security tests
```

## Running Tests

### Using Make Commands

```bash
# Run all tests
make test

# Run specific test suites
make test-go          # Go tests only
make test-python      # Python tests only
make test-security    # Security tests only
make test-integration # Integration tests only
make test-performance # Performance tests only

# Generate coverage reports
make test-coverage

# Run linters
make lint

# Format code
make format
```

### Using Test Script

```bash
# Make script executable
chmod +x /root/PulseExpends/run_tests.sh

# Run all tests
./run_tests.sh

# Run specific test suite
./run_tests.sh go
./run_tests.sh python
./run_tests.sh security
./run_tests.sh integration
./run_tests.sh performance
```

### Manual Testing

#### Go Tests
```bash
cd /root/PulseExpends
go test ./internal/mcp/test -v
go test ./internal/mcp/test -cover
```

#### Python Tests
```bash
cd /root/PulseExpends/python/pdf-parser

# Create virtual environment (first time)
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
pip install -r requirements-test.txt

# Run tests
pytest tests/ -v
pytest tests/ --cov=src --cov-report=html
```

#### Security Tests
```bash
cd /root/PulseExpends
python tests/security/redteam_tests.py --url http://localhost:8080 --verbose
```

## Test Types

### 1. Unit Tests

**Purpose**: Test individual components in isolation.

**Go Unit Tests**:
- Test MCP server handlers
- Test tool execution logic
- Test error handling
- Mock repository dependencies

**Python Unit Tests**:
- Test API endpoints
- Test service classes
- Test data models
- Mock external dependencies

### 2. Integration Tests

**Purpose**: Test interactions between components.

**Features**:
- End-to-end API flows
- Service integration
- Database interactions (when configured)
- File upload/download
- Error propagation

### 3. Security Tests

**Purpose**: Identify security vulnerabilities.

**Tests include**:
- Authentication/authorization bypass
- SQL injection
- XSS injection
- Command injection
- Path traversal
- Rate limiting
- File upload security
- Sensitive data exposure
- API endpoint security

### 4. Performance Tests

**Purpose**: Ensure system performance under load.

**Tests include**:
- Concurrent request handling
- Response time benchmarks
- Memory usage
- CPU utilization
- Throughput measurement

## Security Testing

### Red Team Tests

The security tests simulate real-world attack scenarios:

```bash
# Basic security scan
python tests/security/redteam_tests.py --url http://localhost:8080

# Verbose output
python tests/security/redteam_tests.py --url http://localhost:8080 --verbose

# Save report to file
python tests/security/redteam_tests.py --url http://localhost:8080 --output security_report.json
```

### Test Categories

1. **Authentication Bypass**:
   - Test endpoints without authentication
   - Check for proper 401/403 responses

2. **Input Validation**:
   - SQL injection attempts
   - XSS payloads
   - Command injection
   - Path traversal

3. **File Upload Security**:
   - Malicious file types
   - Large files (DoS)
   - ZIP bombs
   - Malformed files

4. **API Security**:
   - HTTP method testing (PUT, DELETE, TRACE)
   - Security headers
   - CORS configuration
   - Error information leakage

### Security Recommendations

Based on test results, the script provides recommendations:
- Implement proper authentication
- Add input validation/sanitization
- Configure security headers
- Implement rate limiting
- Add file type/size validation

## Performance Testing

### Key Metrics

1. **Response Time**: < 2 seconds for API calls
2. **Throughput**: > 50 requests/second
3. **Concurrent Users**: Support 100+ concurrent users
4. **Memory Usage**: < 500MB under load
5. **CPU Usage**: < 70% under load

### Performance Test Commands

```bash
# Run performance tests
make test-performance

# Manual performance testing
cd /root/PulseExpends/python/pdf-parser
source venv/bin/activate
python -m pytest tests/test_integration.py::TestIntegration::test_performance_metrics -v -s

# Load testing (using external tool)
# Install locust: pip install locust
locust -f tests/load_test.py
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
    
    - name: Install Go dependencies
      run: make deps-go
    
    - name: Install Python dependencies
      run: make deps-python
    
    - name: Run Go tests
      run: make test-go
    
    - name: Run Python tests
      run: make test-python
    
    - name: Run security tests
      run: make test-security
    
    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.xml
        flags: unittests
```

### GitLab CI Example

```yaml
stages:
  - test

go-test:
  stage: test
  image: golang:1.21
  script:
    - cd /root/PulseExpends
    - go test ./... -v -coverprofile=coverage.out
    - go tool cover -func=coverage.out
  artifacts:
    paths:
      - coverage.out

python-test:
  stage: test
  image: python:3.11
  script:
    - cd /root/PulseExpends/python/pdf-parser
    - pip install -r requirements-test.txt
    - pytest tests/ --cov=src --cov-report=xml:coverage.xml
  artifacts:
    paths:
      - coverage.xml

security-test:
  stage: test
  image: python:3.11
  script:
    - cd /root/PulseExpends
    - pip install requests
    - python tests/security/redteam_tests.py --url http://localhost:8080
```

## Test Coverage

### Generating Coverage Reports

```bash
# Go coverage
cd /root/PulseExpends
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Python coverage
cd /root/PulseExpends/python/pdf-parser
source venv/bin/activate
pytest tests/ --cov=src --cov-report=html --cov-report=xml
```

### Coverage Goals

- **Unit Tests**: > 80% coverage
- **Integration Tests**: > 70% coverage
- **Critical Paths**: 100% coverage
- **Security Tests**: All vulnerabilities addressed

## Troubleshooting

### Common Issues

1. **Go Tests Failing**:
   ```bash
   # Clean and rebuild
   go clean -testcache
   go test ./... -v
   
   # Check dependencies
   go mod tidy
   go mod download
   ```

2. **Python Tests Failing**:
   ```bash
   # Recreate virtual environment
   rm -rf venv
   python -m venv venv
   source venv/bin/activate
   pip install -r requirements-test.txt
   
   # Run with verbose output
   pytest tests/ -v -s
   ```

3. **Security Tests Failing**:
   - Ensure server is running: `make run-mcp`
   - Check server URL: `http://localhost:8080`
   - Review firewall/network settings

4. **Performance Issues**:
   - Check system resources
   - Review database connections
   - Optimize queries
   - Implement caching

### Debugging Tips

1. **Enable Debug Logging**:
   ```bash
   # Go tests
   go test ./... -v
   
   # Python tests
   pytest tests/ -v -s
   
   # Security tests
   python tests/security/redteam_tests.py --verbose
   ```

2. **Test Specific Components**:
   ```bash
   # Test specific Go package
   go test ./internal/mcp -v
   
   # Test specific Python module
   pytest tests/test_api.py::test_health_check -v
   
   # Test with coverage for specific file
   pytest tests/ --cov=src.services.parser -v
   ```

3. **Profile Performance**:
   ```bash
   # Go profiling
   go test ./... -bench=. -cpuprofile=cpu.out
   go tool pprof cpu.out
   
   # Python profiling
   python -m cProfile -o profile.stats tests/test_integration.py
   snakeviz profile.stats
   ```

## Best Practices

### Writing Tests

1. **Follow AAA Pattern**:
   - Arrange: Set up test data
   - Act: Execute the code
   - Assert: Verify results

2. **Use Descriptive Names**:
   ```python
   # Good
   def test_parse_document_with_valid_pdf():
   
   # Bad
   def test_parser():
   ```

3. **Isolate Tests**:
   - Each test should be independent
   - Use fixtures for common setup
   - Clean up after tests

4. **Test Edge Cases**:
   - Empty inputs
   - Invalid inputs
   - Boundary conditions
   - Error conditions

5. **Mock External Dependencies**:
   - Database calls
   - API calls
   - File system operations
   - Network requests

### Maintaining Tests

1. **Keep Tests Fast**:
   - Avoid sleep/wait calls
   - Use in-memory databases
   - Mock slow operations

2. **Update with Code Changes**:
   - Update tests when APIs change
   - Add tests for new features
   - Remove obsolete tests

3. **Regular Review**:
   - Review test coverage monthly
   - Remove flaky tests
   - Optimize slow tests

4. **Document Test Data**:
   - Document test scenarios
   - Include sample data
   - Explain edge cases

## Next Steps

1. **Add More Test Scenarios**:
   - Database integration tests
   - OBS storage tests
   - OCR service tests
   - AI model tests

2. **Implement Load Testing**:
   - Use Locust or k6
   - Test with realistic load
   - Monitor resource usage

3. **Add Contract Tests**:
   - API contract validation
   - Schema validation
   - Backward compatibility

4. **Implement Chaos Testing**:
   - Network failures
   - Service outages
   - Database failures
   - Memory leaks

5. **Set Up Monitoring**:
   - Test execution metrics
   - Performance baselines
   - Failure rates
   - Coverage trends

## Support

For issues with tests:
1. Check the troubleshooting section
2. Review test logs
3. Check system requirements
4. Contact development team

Remember: Tests are documentation. Keep them clean, readable, and maintainable.