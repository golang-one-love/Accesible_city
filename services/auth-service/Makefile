.PHONY: build test run migrate-up migrate-down lint tidy

BINARY_NAME=server
BUILD_DIR=./bin
CMD_DIR=./cmd/server
MIGRATIONS_DIR=./migrations

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

build-local:
	@echo "Building $(BINARY_NAME) for local..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

run: build-local
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test -race -coverprofile=coverage.out ./...

test-verbose:
	go test -race -v ./...

coverage:
	go tool cover -html=coverage.out

lint:
	golangci-lint run

tidy:
	go mod tidy

migrate-up:
	@echo "Running migrations up..."
	@for f in $(MIGRATIONS_DIR)/*.up.sql; do \
		echo "Applying $$f..."; \
		psql "$(DATABASE_URL)" -f "$$f"; \
	done

migrate-down:
	@echo "Running migrations down..."
	@for f in $(MIGRATIONS_DIR)/*.down.sql; do \
		echo "Reverting $$f..."; \
		psql "$(DATABASE_URL)" -f "$$f"; \
	done

docker-build:
	docker build -t auth-service:latest .

docker-run:
	docker run --rm -p 8081:8081 \
		-e PORT=8081 \
		-e POSTGRES_HOST=host.docker.internal \
		-e POSTGRES_PORT=5432 \
		-e POSTGRES_USER=auth_user \
		-e POSTGRES_PASSWORD=auth_pass \
		-e POSTGRES_DB=auth_db \
		-e JWT_PRIVATE_KEY_PATH=/keys/private.pem \
		-e JWT_PUBLIC_KEY_PATH=/keys/public.pem \
		-v $(PWD)/keys:/keys:ro \
		auth-service:latest

clean:
	rm -rf $(BUILD_DIR) coverage.out

generate-keys:
	@mkdir -p keys
	openssl genrsa -out keys/private.pem 2048
	openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@echo "Keys generated in keys/"

help:
	@echo "Available targets:"
	@echo "  build          - Build binary for Linux"
	@echo "  build-local    - Build binary for current OS"
	@echo "  run            - Build and run locally"
	@echo "  test           - Run tests with race detector"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  coverage       - Show test coverage in browser"
	@echo "  lint           - Run golangci-lint"
	@echo "  tidy           - Run go mod tidy"
	@echo "  migrate-up     - Apply migrations (requires DATABASE_URL)"
	@echo "  migrate-down   - Revert migrations (requires DATABASE_URL)"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container locally"
	@echo "  clean          - Clean build artifacts"
	@echo "  generate-keys  - Generate JWT RSA keys"