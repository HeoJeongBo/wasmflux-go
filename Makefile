GOROOT_WASM_EXEC := $(shell go env GOROOT)/lib/wasm/wasm_exec.js
WASM_OUT         := web/app.wasm
GO_BUILD_FLAGS   := -ldflags="-s -w"

.PHONY: build build-tiny build-debug serve test test-wasm setup clean watch size \
        fmt vet lint tidy coverage bench check \
        build-example setup-example example

# === Build ===

build:
	GOOS=js GOARCH=wasm go build $(GO_BUILD_FLAGS) -o $(WASM_OUT) ./example/

build-debug:
	GOOS=js GOARCH=wasm go build -tags debug -o $(WASM_OUT) ./example/

build-tiny: build
	@command -v wasm-opt > /dev/null 2>&1 && wasm-opt -Oz -o $(WASM_OUT) $(WASM_OUT) || echo "wasm-opt not found, skipping optimization"

# === Development ===

serve: build
	go run ./cmd/devserver/

watch:
	@echo "Watching for changes..."
	@while true; do \
		find . -name '*.go' -newer $(WASM_OUT) 2>/dev/null | head -1 | \
		while read f; do \
			echo "Changed: $$f — rebuilding..."; \
			$(MAKE) build; \
		done; \
		sleep 1; \
	done

setup:
	cp "$(GOROOT_WASM_EXEC)" web/wasm_exec.js

# === Example (Vite + React) ===

build-example:
	GOOS=js GOARCH=wasm go build $(GO_BUILD_FLAGS) -o example/public/app.wasm ./example/

setup-example:
	cp "$(GOROOT_WASM_EXEC)" example/public/wasm_exec.js
	cp web/glue.js example/public/glue.js

example: build-example
	cd example && npx vite

# === Quality ===

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	@command -v golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not found, run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

check: fmt vet lint test

# === Testing ===

test:
	go test -race ./...

test-wasm:
	GOOS=js GOARCH=wasm go test ./...

bench:
	go test -bench=. -benchmem ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "---"
	@echo "For HTML report: go tool cover -html=coverage.out"

# === Maintenance ===

tidy:
	go mod tidy

clean:
	rm -f $(WASM_OUT) coverage.out

size: build
	@ls -lh $(WASM_OUT) | awk '{print "WASM binary size:", $$5}'
