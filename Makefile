BIN_PATH := /usr/local/bin/scythix
LOG_PATH := $(HOME)/.cache/scythix.log
CONF_DIR := $(HOME)/.config/scythix
LOCK_FILE := /var/lock/scythix.lock

install:
	@echo "Building scythix..."
	@go build -o scythix . || { echo "Build failed"; exit 1; }
	
	@echo "Installing to $(BIN_PATH)..."
	@sudo install -Dm 0755 ./scythix $(BIN_PATH) || { \
		echo "Installation failed - check permissions"; \
		exit 1; \
	}
	
	@if [ -f "$(BIN_PATH)" ]; then \
		echo "Successfully installed scythix to $(BIN_PATH)"; \
		echo "Version: $$(scythix --version 2>/dev/null || echo 'unknown')"; \
	else \
		echo "Installation verification failed"; \
		exit 1; \
	fi

	@if rm -f ./scythix; then \
		echo "Cleanup complete"; \
	else \
		echo "Cleanup failed"; \
	fi

uninstall:
	@# Gracefully attempt to stop player if exists
	@if command -v scythix >/dev/null 2>&1; then \
		echo "Attempting to stop player..."; \
		scythix -stop || true; \
		sleep 1; \
	else \
		echo "Unable to stop playback, skipping"; \
	fi
    
	@# Remove installation files
	@if [ -f "$(BIN_PATH)" ]; then \
		sudo rm -f "$(BIN_PATH)"; \
		echo "Removed: $(BIN_PATH)"; \
	else \
		echo "Player executable not found: $(BIN_PATH), skipping deleting"; \
	fi
    
	@# Clean up other files
	@rm -f "$(LOG_PATH)"
	@rm -rf "$(CONF_DIR)"
	@sudo rm -f "$(LOCK_FILE)" 2>/dev/null || true
	@echo "Uninstall complete"
