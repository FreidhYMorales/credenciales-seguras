GO ?= go
BIN := bin/credenciales

.PHONY: check fmt test build build-windows run

# Single quality gate: format check (no rewrite), lint, tests. Read-only, non-interactive.
check:
	@test -z "$$(gofmt -l .)" || { echo "gofmt needed:"; gofmt -l .; exit 1; }
	$(GO) vet ./...
	$(GO) test ./...

fmt:
	gofmt -w .

test:
	$(GO) test ./...

build:
	$(GO) build -o $(BIN) ./cmd/app

build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BIN).exe ./cmd/app

run: build
	./$(BIN)
