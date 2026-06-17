.PHONY: build test lint clean cross install

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.Version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/neocode ./cmd/neocode

run:
	go run ./cmd/neocode

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/ dist/

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/neocode

# Cross-compilation for 6 platforms
cross:
	@for target in \
		darwin/amd64 darwin/arm64 \
		linux/amd64 linux/arm64 \
		windows/amd64 windows/arm64; do \
		os=$$(echo $$target | cut -d/ -f1); \
		arch=$$(echo $$target | cut -d/ -f2); \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "Building $$os/$$arch..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -ldflags "$(LDFLAGS)" -o dist/neocode-$$os-$$arch$$ext ./cmd/neocode; \
	done
	@echo "Cross-compilation complete. Binaries in dist/"

# Format code
fmt:
	go fmt ./...

# Check dependencies
tidy:
	go mod tidy
