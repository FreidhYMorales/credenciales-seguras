GO ?= go
BIN := bin/credenciales
PYTHON ?= python3

.PHONY: check check-oracle fmt test build build-windows run figures pairs pares-tabla documento smoke-windows

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

# Render the final document to .docx and .pdf. Optional tooling: needs
# python-docx and LibreOffice, and the outputs are committed anyway.
documento:
	$(PYTHON) tools/documento/exportar.py

# Smoke test of the Windows binary under Wine: start it, wait for /api/salud,
# validate one credential, stop it. Kept out of `check`, which must stay
# Go-only and non-interactive; this needs Wine, curl and a free port.
smoke-windows: build-windows
	@command -v wine >/dev/null 2>&1 || { echo "falta wine"; exit 1; }
	@command -v curl >/dev/null 2>&1 || { echo "falta curl"; exit 1; }
	@puerto=8099; \
	WINEDEBUG=-all wine $(BIN).exe -addr 127.0.0.1:$$puerto >/tmp/credenciales-wine.log 2>&1 & \
	pid=$$!; \
	trap 'kill $$pid 2>/dev/null' EXIT; \
	for i in $$(seq 1 60); do \
		curl -sf http://127.0.0.1:$$puerto/api/salud >/dev/null 2>&1 && break; \
		sleep 0.5; \
	done; \
	curl -sf http://127.0.0.1:$$puerto/api/salud | grep -q '"estado":"ok"' \
		|| { echo "el .exe no responde en /api/salud"; exit 1; }; \
	curl -sf -X POST -H 'Content-Type: application/json' \
		-d '{"tipo":"password","valor":"Xk9$$mTq2#vLw8@password"}' \
		http://127.0.0.1:$$puerto/api/validar | grep -q '"valido":false' \
		|| { echo "el .exe no valida"; exit 1; }; \
	echo "ok: el .exe responde y valida bajo Wine (puerto $$puerto)"
