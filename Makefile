.PHONY: proto build run test lint clean docker docker-down db-up db-down

# Proto generation
proto:
	cd proto && buf generate

# Build
build:
	go build -o bin/server ./cmd/server

# Run locally
run: build
	DATABASE_URL="postgres://todo:todo@localhost:5433/todo?sslmode=disable" \
	GRPC_PORT=50051 \
	LOG_LEVEL=debug \
	./bin/server

# Run tests
test:
	go test -v -race -cover ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

# Clean
clean:
	rm -rf bin/ coverage.out coverage.html

# Docker
docker:
	docker-compose up --build

docker-down:
	docker-compose down -v

# Database only (for local dev)
db-up:
	docker-compose up -d postgres

db-down:
	docker-compose down postgres -v

# Generate mocks (if using mockgen)
mocks:
	go generate ./...

# Tidy dependencies
tidy:
	go mod tidy
