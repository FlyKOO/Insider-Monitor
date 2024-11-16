.PHONY: setup test lint build clean install update restart

# Configuration
BINARY_NAME := solana-monitor
INSTALL_PATH := /usr/local/bin
SERVICE_NAME := solana-monitor.service
SERVICE_PATH := /etc/systemd/system

setup:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	pip install pre-commit
	pre-commit install

test:
	go test ./... -coverprofile=coverage.out

lint:
	golangci-lint run

build:
	go build -v -o $(BINARY_NAME) ./cmd/monitor

clean:
	go clean
	rm -f coverage.out
	rm -f $(BINARY_NAME)

# Install the binary and service
install: build
	@echo "Installing $(BINARY_NAME)..."
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/
	sudo chown root:root $(INSTALL_PATH)/$(BINARY_NAME)
	sudo chmod 755 $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete!"

# Update the binary and restart the service
update: build
	@echo "Stopping service..."
	sudo systemctl stop $(SERVICE_NAME)
	@echo "Installing new binary..."
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/
	sudo chown root:root $(INSTALL_PATH)/$(BINARY_NAME)
	sudo chmod 755 $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Starting service..."
	sudo systemctl start $(SERVICE_NAME)
	@echo "Update complete!"
	@echo "Service status:"
	sudo systemctl status $(SERVICE_NAME) --no-pager

# Restart the service
restart:
	@echo "Restarting service..."
	sudo systemctl restart $(SERVICE_NAME)
	@echo "Service status:"
	sudo systemctl status $(SERVICE_NAME) --no-pager

all: setup lint test build
