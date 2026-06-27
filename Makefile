.PHONY: all run test lint build clean swagger

# Default target
all: test lint build

# Run the application locally
run:
	docker compose up -d db
	go run main.go

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Build the application
build:
	go build -o server.exe main.go

# Generate Swagger documentation
swagger:
	swag init -g main.go

# Clean up
clean:
	docker compose down
	rm -f server.exe
