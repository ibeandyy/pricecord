include .env
BINARY=pricecord


# Build the Go application
build:
	@echo "Building..."
	go build -o ./build/$(BINARY) ./cmd
	@echo "Build complete"
# Run the Go application with the TOKEN environment variable
run:
	@echo "Running..."
	./build/$(BINARY) --token=$(TOKEN)

# Clean up any build artifacts
clean:
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY)

.PHONY: build run clean
