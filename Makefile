# Makefile

# Variables
BINARY_NAME=httpchat
DOCKER_COMPOSE=docker-compose.yml

# Build binary
build:
	go build -o ${BINARY_NAME} cmd/server/main.go

# Generate Swagger documentation
swagger:
	swag init -g cmd/server/main.go -o docs/swagger

# Run application
run: build
	./${BINARY_NAME}

# Run in Docker
docker-run:
	docker-compose -f ${DOCKER_COMPOSE} up --build

# Stop Docker containers
docker-stop:
	docker-compose -f ${DOCKER_COMPOSE} down

# Run only database and Kafka in Docker
docker-db-kafka:
	docker-compose -f ${DOCKER_COMPOSE} up -d postgres kafka

# Stop database and Kafka
docker-db-kafka-stop:
	docker-compose -f ${DOCKER_COMPOSE} stop postgres kafka

# Install dependencies
deps:
	go mod tidy

# Unit testing
test:
	go test -v ./...

# Integration testing
test-integration:
	docker-compose -f ${DOCKER_COMPOSE} up -d postgres kafka
	timeout /t 5 >nul
	set TEST_DATABASE_URL=postgres://user:password@localhost:5432/messages_db?sslmode=disable && go test -v ./internal/repository
	docker-compose -f ${DOCKER_COMPOSE} stop postgres kafka

# Full testing he-he
test-full:
	-@docker-compose -f ${DOCKER_COMPOSE} down >nul 2>&1
	docker-compose -f ${DOCKER_COMPOSE} up -d
	timeout /t 10 >nul
	go test -v ./...
	set TEST_DATABASE_URL=postgres://user:password@localhost:5432/messages_db?sslmode=disable && go test -v ./internal/repository
	go test -v ./cmd/server -run "TestEndToEnd"
	-@docker-compose -f ${DOCKER_COMPOSE} down >nul 2>&1

# Run linter
lint:
	golangci-lint run --no-config --enable=govet --enable=errcheck --enable=staticcheck --enable=unused --enable=ineffassign --enable=bodyclose --enable=gosec --enable=misspell --enable=unparam --enable=revive

# Cleanup
clean:
	go clean
	if exist ${BINARY_NAME} del ${BINARY_NAME}

.PHONY: build swagger run docker-run docker-stop docker-db-kafka docker-db-kafka-stop deps test test-integration test-full lint clean