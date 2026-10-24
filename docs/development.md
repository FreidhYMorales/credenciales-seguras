# Development

## Prerequisites

| Tool | Version | Needed for |
|---|---|---|
| Go | 1.27+ | build, test, run |
| Graphviz `dot` | 16+ | regenerating `assets/figures/*.svg` (build time only) |
| Python 3 | 3.11+ | `make check-oracle` and `make documento` (optional) |
| python-docx + LibreOffice | any recent | `make documento` only (optional) |
| mingw-w64 (`extra/mingw-w64-gcc`) | 16+ | `make app-windows` only (optional, cgo cross build) |
| GTK3 + WebKitGTK 4.1 | any recent | `make app` only (optional window build) |

The plain binaries need none of them. The window builds need a C toolchain because
`webview_go` is cgo.

## Commands

```sh
make check          # gofmt + vet + tests (the only gate)
make test           # go test ./...
make fmt            # gofmt -w .
make build          # linux binary  -> bin/credenciales
make build-windows  # windows .exe  -> bin/credenciales.exe
make run            # build and serve on http://localhost:8080
make app            # native window binary -> bin/credenciales-gui (WebKitGTK)
make app-windows    # native window .exe    -> bin/credenciales-gui.exe (WebView2)
make figures        # regenerate assets/figures/*.svg with graphviz
make pairs          # regenerate testdata/pares.json (deterministic seeds)
make pares-tabla    # render the 20 pairs as Markdown (docs/entregas/final/pares.md)
make check-oracle   # Python differential, optional
make documento      # render the final document to .docx and .pdf, optional
make entrega        # pack binaries + document into dist/credenciales-seguras.zip
make smoke-windows  # start the .exe under Wine and validate one credential
```

`make check` is the only gate and it stays Go-only. `make smoke-windows` needs Wine and
curl; it builds the `.exe`, waits for `/api/salud` on port 8099, validates one credential
and stops the process. Neither `make check` nor `make check-oracle` starts a server.

## Native window

`make app` builds the same program with `-tags gui`: instead of asking for a browser it
opens a 1280x860 window (ADR 0006). Both modes share `cmd/app/main.go`; `gui.go` and
`nogui.go` only differ in the entry point.

- Without the tag nothing changes: `bin/credenciales` serves on `127.0.0.1:8080` and waits
  for Ctrl+C. The window build listens on `127.0.0.1:0` (kernel-picked free port) and
  prints it; `-addr` still overrides both.
- `webview_go` requests `pkg-config: webkit2gtk-4.0` and current distributions ship 4.1
  only. `make gui-pc` (a prerequisite of both window targets) writes
  `build/pkgconfig/webkit2gtk-4.0.pc` pointing at 4.1, and only when 4.0 is absent.
- Cross-compiling the window for Windows needs `extra/mingw-w64-gcc`; `make app-windows`
  passes `CC`/`CXX` and links with `-H=windowsgui`, so no console flashes behind the
  window. The result needs the WebView2 runtime, which ships with Edge on Windows 10/11.
- `webview.cc` includes `EventToken.h` in Windows SDK spelling, but mingw ships
  `eventtoken.h`; on a case-sensitive filesystem that is a fatal error. `make mingw-include`
  writes `build/mingw-include/EventToken.h` (`#include <eventtoken.h>`) and the build
  exports `CPATH`, not `-I`: the repository path may contain spaces and `CPATH` is
  `:`-separated, so entries with spaces survive.
- If there is no web engine (`DISPLAY`/`WAYLAND_DISPLAY` empty, or no WebView2 runtime)
  the window build opens the default browser and keeps serving instead of dying silently,
  which matters because the Windows binary has no console to print to.

Packages, if missing:

```sh
sudo pacman -S --needed gtk3 webkit2gtk-4.1 mingw-w64-gcc   # Arch
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev gcc-mingw-w64-x86-64  # Debian/Ubuntu
```

## Layout rules

- New rule? Add it to `internal/rules` with its ER, its RE2 translation, its labels and its
  metadata row. Never hardcode rule ids in handlers or in the frontend.
- New endpoint? Contract goes into `docs/api.md` first, then the handler, then the frontend.
- `internal/automata` stays dependency-free and UI-free.

## Conventions

- Package names, identifiers and comments in English. UI strings and JSON keys in Spanish.
- Table-driven tests; every rule has at least one accept and one reject case.
- Sorted slices instead of maps wherever iteration order can reach output.
- No external Go dependencies without an ADR. The one on record is `webview_go`, optional
  and behind the `gui` tag (ADR 0006).

## Seeded demo values

| Seed | Kind | Result |
|---|---|---|
| `2026` | password, length 20 | reproduced in the demo and in the video |
| `4490` | email | reproduced in the demo |

Both must reproduce exactly on the presentation machine; the regeneration test fails the
build if they do not.
