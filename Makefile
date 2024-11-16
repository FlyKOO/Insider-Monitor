.PHONY: setup test lint build clean install update restart

# Configuration
BINARY_NAME := solana-monitor
INSTALL_PATH := /usr/local/bin
SERVICE_NAME := solana-monitor.service
SERVICE_PATH := /etc/systemd/system

# Add backup directory
BACKUP_DIR := /tmp/solana-monitor-backup

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

# Install the binary and service (first time installation)
install: build
	@echo "Installing $(BINARY_NAME)..."
	@if systemctl is-active --quiet $(SERVICE_NAME); then \
		echo "Service is running. Use 'make update' instead for updates."; \
		exit 1; \
	fi
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/
	sudo chown root:root $(INSTALL_PATH)/$(BINARY_NAME)
	sudo chmod 755 $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete!"

# Safe update with backup and rollback
update: build
	@echo "Creating backup directory..."
	@mkdir -p $(BACKUP_DIR)
	@echo "Backing up current binary..."
	@sudo cp $(INSTALL_PATH)/$(BINARY_NAME) $(BACKUP_DIR)/$(BINARY_NAME).backup
	@echo "Stopping service..."
	@sudo systemctl stop $(SERVICE_NAME)
	@echo "Installing new binary..."
	@if ! sudo cp $(BINARY_NAME) $(INSTALL_PATH)/; then \
		echo "Failed to install new binary. Rolling back..."; \
		sudo cp $(BACKUP_DIR)/$(BINARY_NAME).backup $(INSTALL_PATH)/$(BINARY_NAME); \
		sudo systemctl start $(SERVICE_NAME); \
		exit 1; \
	fi
	@sudo chown root:root $(INSTALL_PATH)/$(BINARY_NAME)
	@sudo chmod 755 $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Reloading systemd daemon..."
	@sudo systemctl daemon-reload
	@echo "Starting service..."
	@if ! sudo systemctl start $(SERVICE_NAME); then \
		echo "Failed to start service. Rolling back..."; \
		sudo cp $(BACKUP_DIR)/$(BINARY_NAME).backup $(INSTALL_PATH)/$(BINARY_NAME); \
		sudo systemctl start $(SERVICE_NAME); \
		exit 1; \
	fi
	@echo "Update complete!"
	@echo "Service status:"
	@sudo systemctl status $(SERVICE_NAME) --no-pager

# Restart the service
restart:
	@echo "Restarting service..."
	@sudo systemctl restart $(SERVICE_NAME)
	@echo "Service status:"
	@sudo systemctl status $(SERVICE_NAME) --no-pager

all: setup lint test build
