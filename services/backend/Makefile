# PulseExpends Makefile
# Build, test, and deployment commands

.PHONY: help build test test-go test-python test-security test-integration test-performance clean deps deps-go deps-python run run-mcp run-python deploy

# Default target
help:
	@echo "PulseExpends Development Commands"
	@echo ""
	@echo "Build:"
	@echo "  build           Build all components"
	@echo "  build-go        Build Go MCP server"
	@echo "  build-python    Build Python PDF parser"
	@echo ""
	@echo "Testing:"
	@echo "  test            Run all tests"
	@echo "  test-go         Run Go tests"
	@echo "  test-python     Run Python tests"
	@echo "  test-security   Run security tests"
	@echo "  test-integration Run integration tests"
	@echo "  test-performance Run performance tests"
	@echo "  test-coverage   Generate test coverage reports"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps            Install all dependencies"
	@echo "  deps-go         Install Go dependencies"
	@echo "  deps-python     Install Python dependencies"
	@echo ""
	@echo "Development:"
	@echo "  run             Run all services"
	@echo "  run-mcp         Run MCP server"
	@echo "  run-python      Run PDF parser service"
	@echo "  lint            Run linters"
	@echo "  format          Format code"
	@echo ""
	@echo "Deployment:"
	@echo "  deploy-dev      Deploy to development"
	@echo "  deploy-staging  Deploy to staging"
	@echo "  deploy-prod     Deploy to production"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean           Clean build artifacts"
	@echo "  clean-deps      Clean dependency caches"

# Build targets
build: build-go build-python

build-go:
	@echo "Building Go MCP server..."
	cd /root/PulseExpends && go build -o bin/mcp-server ./cmd/mcp-server

build-python:
	@echo "Building Python PDF parser..."
	cd /root/PulseExpends/python/pdf-parser && \
	python -m pip install -r requirements.txt && \
	python -m pip install -e .

# Test targets
test: test-go test-python test-security test-integration

test-go:
	@echo "Running Go tests..."
	cd /root/PulseExpends && \
	go test ./internal/mcp/test -v -coverprofile=coverage-go.out

test-python:
	@echo "Running Python tests..."
	cd /root/PulseExpends/python/pdf-parser && \
	source venv/bin/activate && \
	python -m pytest tests/ -v --cov=src --cov-report=term-missing --cov-report=html:coverage_html

test-security:
	@echo "Running security tests..."
	chmod +x /root/PulseExpends/run_tests.sh && \
	cd /root/PulseExpends && \
	./run_tests.sh security

test-integration:
	@echo "Running integration tests..."
	cd /root/PulseExpends/python/pdf-parser && \
	source venv/bin/activate && \
	python -m pytest tests/test_integration.py -v -m integration

test-performance:
	@echo "Running performance tests..."
	cd /root/PulseExpends/python/pdf-parser && \
	source venv/bin/activate && \
	python -m pytest tests/test_integration.py::TestIntegration::test_performance_metrics -v

test-coverage:
	@echo "Generating test coverage reports..."
	@echo "Go coverage:"
	cd /root/PulseExpends && go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
	@echo "Python coverage:"
	cd /root/PulseExpends/python/pdf-parser && \
	source venv/bin/activate && \
	python -m pytest tests/ --cov=src --cov-report=html:coverage_html --cov-report=xml:coverage.xml

# Dependency targets
deps: deps-go deps-python

deps-go:
	@echo "Installing Go dependencies..."
	cd /root/PulseExpends && \
	go mod download && \
	go get github.com/stretchr/testify

deps-python:
	@echo "Installing Python dependencies..."
	cd /root/PulseExpends/python/pdf-parser && \
	python -m venv venv && \
	source venv/bin/activate && \
	pip install -r requirements.txt && \
	pip install -r requirements-test.txt

# Development targets
run: run-mcp run-python

run-mcp: build-go
	@echo "Starting MCP server..."
	cd /root/PulseExpends && \
	./bin/mcp-server

run-python:
	@echo "Starting PDF parser service..."
	cd /root/PulseExpends/python/pdf-parser && \
	source venv/bin/activate && \
	uvicorn src.main:app --host 0.0.0.0 --port 8000 --reload

lint:
	@echo "Running linters..."
	@echo "Go linting..."
	cd /root/PulseExpends && go vet ./...
	@echo "Python linting..."
	cd /root/PulseExpends/python/pdf-parser && \
	source venv/bin/activate && \
	flake8 src/ tests/ && \
	black --check src/ tests/ && \
	mypy src/

format:
	@echo "Formatting code..."
	@echo "Go formatting..."
	cd /root/PulseExpends && gofmt -w .
	@echo "Python formatting..."
	cd /root/PulseExpends/python/pdf-parser && \
	source venv/bin/activate && \
	black src/ tests/ && \
	isort src/ tests/

# Deployment targets (placeholder - implement based on your deployment process)
deploy-dev:
	@echo "Deploying to development..."
	@echo "TODO: Implement development deployment"

deploy-staging:
	@echo "Deploying to staging..."
	@echo "TODO: Implement staging deployment"

deploy-prod:
	@echo "Deploying to production..."
	@echo "TODO: Implement production deployment"

# Cleanup targets
clean:
	@echo "Cleaning build artifacts..."
	rm -rf /root/PulseExpends/bin
	rm -rf /root/PulseExpends/coverage*
	rm -rf /root/PulseExpends/python/pdf-parser/__pycache__
	rm -rf /root/PulseExpends/python/pdf-parser/src/__pycache__
	rm -rf /root/PulseExpends/python/pdf-parser/tests/__pycache__
	rm -rf /root/PulseExpends/python/pdf-parser/.coverage
	rm -rf /root/PulseExpends/python/pdf-parser/coverage_html
	rm -rf /root/PulseExpends/python/pdf-parser/coverage.xml
	rm -rf /root/PulseExpends/python/pdf-parser/.pytest_cache
	find /root/PulseExpends -name "*.pyc" -delete
	find /root/PulseExpends -name "__pycache__" -type d -exec rm -rf {} +
	find /root/PulseExpends -name ".pytest_cache" -type d -exec rm -rf {} +

clean-deps:
	@echo "Cleaning dependency caches..."
	rm -rf /root/PulseExpends/python/pdf-parser/venv
	go clean -modcache