# Makefile for YARS (Yet Another Reconciliation System)

# Define the binary names
SERVER_BINARY=yars-server
COMPILER_BINARY=yars-compiler
RECON_BINARY=yars-reconciliation

# Define Docker image names
SERVER_IMAGE=yars-server-image
COMPILER_IMAGE=yars-compiler-image
RECON_IMAGE=yars-recon-image

# Environment variables with defaults
export POSTGRES_USER ?= postgres
export POSTGRES_PASSWORD ?= password
export POSTGRES_DB ?= yars

# Default target
all: build

# Build all applications locally
build: build-server build-compiler build-reconciliation

# Build the server application locally
build-server:
	go build -o $(SERVER_BINARY) ./cmd/server

# Build the compiler consumer locally
build-compiler:
	go build -o $(COMPILER_BINARY) ./cmd/consumer/compiler

# Build the reconciliation consumer locally
build-reconciliation:
	go build -o $(RECON_BINARY) ./cmd/consumer/reconciliation

# Run the server application locally
run-server: build-server
	./$(SERVER_BINARY)

# Run the compiler consumer locally
run-compiler: build-compiler
	./$(COMPILER_BINARY)

# Run the reconciliation consumer locally
run-reconciliation: build-reconciliation
	./$(RECON_BINARY)

# Clean the build artifacts
clean:
	rm -f $(SERVER_BINARY) $(COMPILER_BINARY) $(RECON_BINARY)

# Format the code
fmt:
	go fmt ./...

# Run tests
test:
	go test ./...

# Build Docker images individually
docker-build-server:
	docker build -t $(SERVER_IMAGE) -f Dockerfile.server .

docker-build-compiler:
	docker build -t $(COMPILER_IMAGE) -f Dockerfile.compiler .

docker-build-reconciliation:
	docker build -t $(RECON_IMAGE) -f Dockerfile.reconciliation .

# Build all Docker images
docker-build: docker-build-server docker-build-compiler docker-build-reconciliation


# Start Docker services
docker-up:
	docker-compose up -d

# Start Docker services with logs displayed
docker-up-logs:
	docker-compose up

# Stop all Docker services
docker-down:
	docker-compose down

# Stop all Docker services and remove volumes
docker-clean:
	docker-compose down -v

# Start only infrastructure (postgres, kafka, etc.)
infra-up:
	docker-compose up -d postgres bucket zookeeper kafka

# Force clean database and create fresh
db-clean:
	docker-compose down -v postgres
	docker-compose up -d postgres

# Start specific services
start-server:
	docker-compose up -d app-server

start-compiler:
	docker-compose up -d app-consumer-compiler

start-reconciliation:
	docker-compose up -d app-consumer-reconciliation

# View logs for specific services
logs-server:
	docker-compose logs -f app-server

logs-compiler:
	docker-compose logs -f app-consumer-compiler

logs-reconciliation:
	docker-compose logs -f app-consumer-reconciliation

logs-postgres:
	docker-compose logs -f postgres

logs-all:
	docker-compose logs -f

# Complete end-to-end setup (replaced "dockerfile" with "docker-build")
setup-fresh: docker-build docker-down db-clean docker-up

clean-yars-image: 
	docker rmi yars-compiler-image yars-server-image yars-app-consumer-compiler yars-app-server yars-app-consumer-reconciliation yars-recon-image

# Create a new migration file
migration-create:
	@read -p "Enter migration name: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	echo "-- Migration: $$name\n-- Created at: $$(date)\n\n" > ./scripts/db/$$timestamp\_$$name.sql; \
	echo "Migration file created: ./scripts/db/$$timestamp\_$$name.sql"

# Help command
help:
	@echo "YARS Makefile Commands:"
	@echo "  make build			   - Build all local binaries"
	@echo "  make docker-build		- Build all Docker images"
	@echo "  make docker-up		   - Start all Docker services in background"
	@echo "  make docker-up-logs	  - Start all Docker services with logs"
	@echo "  make docker-down		 - Stop all Docker services"
	@echo "  make docker-clean		- Stop all services and remove volumes"
	@echo "  make setup-fresh		 - Full clean setup from scratch"
	@echo "  make logs-all			- View logs from all services"
	@echo "  make start-server		- Start just the server service"
	@echo "  make infra-up			- Start just infrastructure services"
	@echo "  make migration-create	- Create a new migration file"

.PHONY: all build build-server build-compiler build-reconciliation run-server run-compiler run-reconciliation clean fmt test \
	docker-build docker-build-server docker-build-compiler docker-build-reconciliation \
	docker-up docker-up-logs docker-down docker-clean \
	infra-up db-clean \
	start-server start-compiler start-reconciliation \
	logs-server logs-compiler logs-reconciliation logs-postgres logs-all \
	setup-fresh migration-create help prepare-migrations clean-yars-image
