GO ?= go
BIN := bin/credenciales
GUI := bin/credenciales-gui
PYTHON ?= python3
# Local pkg-config directory for the WebKitGTK alias (see gui-pc).
GUI_PC := build/pkgconfig
MINGW ?= x86_64-w64-mingw32-gcc
MINGW_INCLUDE := build/mingw-include

.PHONY: check check-oracle fmt test build build-windows run app app-windows gui-pc mingw-include figures pairs pares-tabla documento entrega smoke-windows smoke-windows-ventana

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

# Native window. Linux builds WebKitGTK, Windows builds WebView2; both come from
# webview_go and are behind the `gui` tag, so the plain binaries above stay
# cgo-free and still run where no webview runtime is installed.
app: gui-pc
	PKG_CONFIG_PATH='$(abspath $(GUI_PC))' $(GO) build -tags gui -o $(GUI) ./cmd/app

# Cross build needs mingw-w64 (extra/mingw-w64-gcc) and cgo. -H=windowsgui
# hides the console: the window is the whole interface.
app-windows: gui-pc mingw-include
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=$(MINGW) CXX=x86_64-w64-mingw32-g++ \
		CPATH='$(abspath $(MINGW_INCLUDE))' \
		$(GO) build -tags gui -ldflags '-H=windowsgui -s -w' -o $(GUI).exe ./cmd/app

# webview.cc includes "EventToken.h" in Windows SDK spelling, while mingw ships
# it as eventtoken.h. CPATH is used instead of -I because the repository path
# may contain spaces and CPATH is ':'-separated.
mingw-include:
	@mkdir -p $(MINGW_INCLUDE)
	@printf '%s\n' '#pragma once' '#include <eventtoken.h>' > $(MINGW_INCLUDE)/EventToken.h

# webview_go asks pkg-config for webkit2gtk-4.0, but current distributions ship
# WebKitGTK 4.1 only. 4.1 is the same API with libsoup3, so a local alias .pc
# that requires 4.1 is enough. Written only when 4.0 is genuinely missing.
gui-pc:
	@mkdir -p $(GUI_PC)
	@if pkg-config --exists webkit2gtk-4.0; then \
		rm -f $(GUI_PC)/webkit2gtk-4.0.pc; \
	elif pkg-config --exists webkit2gtk-4.1; then \
		printf '%s\n' \
			'Name: webkit2gtk-4.0' \
			'Description: alias of webkit2gtk-4.1 for webview_go' \
			'Version: 4.1.0' \
			'Requires: webkit2gtk-4.1' \
			> $(GUI_PC)/webkit2gtk-4.0.pc; \
	else \
		echo "falta webkit2gtk: instalá webkit2gtk-4.1 (Arch) o libwebkit2gtk-4.1-dev (Debian/Ubuntu)"; \
		exit 1; \
	fi

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

# Pack what gets handed in: both binaries, the document and a readme. The
# document is the committed one, so this does not need python-docx.
entrega: build build-windows app app-windows
	@mkdir -p dist
	rm -rf dist/entrega && mkdir -p dist/entrega
	cp $(GUI) dist/entrega/credenciales-gui-linux
	cp $(GUI).exe dist/entrega/credenciales-gui.exe
	cp $(BIN) dist/entrega/credenciales-linux
	cp $(BIN).exe dist/entrega/credenciales.exe
	cp docs/entregas/final/Credenciales_seguras.pdf dist/entrega/
	cp docs/entregas/final/Credenciales_seguras.docx dist/entrega/
	cp docs/entregas/final/LEEME.txt dist/entrega/
	rm -f dist/credenciales-seguras.zip
	cd dist/entrega && zip -q ../credenciales-seguras.zip *
	@ls -lh dist/credenciales-seguras.zip

# Smoke test of the Windows binaries under Wine: start one, wait for /api/salud,
# validate one credential, stop it. Kept out of `check`, which must stay
# Go-only and non-interactive; these need Wine, curl and a free port.
smoke-windows: build-windows
	@sh tools/smoke/wine.sh $(BIN).exe 8099

# The window binary needs no display to be checked this way. Wine has no
# WebView2 runtime, so this covers the HTTP layer and the browser fallback.
smoke-windows-ventana: app-windows
	@sh tools/smoke/wine.sh $(GUI).exe 8098
