GO ?= go
BIN := bin/credenciales
PYTHON ?= python3

.PHONY: check check-oracle fmt test build build-windows run figures pairs pares-tabla

# Single quality gate: format check (no rewrite), lint, tests. Read-only,
# non-interactive, Go-only. Never weaken a step to make it pass.
check:
	@test -z "$$(gofmt -l .)" || { echo "gofmt needed:"; gofmt -l .; exit 1; }
	$(GO) vet ./...
	$(GO) test ./...

# Optional port-fidelity check against the Python prototype in codigo_avance/.
# Needs Python 3 and is never required to build or run the application.
check-oracle:
	$(PYTHON) tools/oracle/check.py

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

# Regenerate the embedded SVG figures. Needs Graphviz `dot`; the runtime does not.
figures:
	$(GO) run ./cmd/figgen

# Regenerate the 20 graded pairs. Deterministic: the result is committed and a
# test fails if it does not regenerate byte for byte.
pairs:
	$(GO) run ./cmd/pares

# Render the same 20 pairs as the Markdown tables the final document includes.
# A test fails if the committed file drifts from pares.json.
pares-tabla:
	@mkdir -p docs/entregas/final
	$(GO) run ./cmd/pares -md > docs/entregas/final/pares.md
