# ADR 0006 — Native window via webview_go

- Status: accepted
- Date: 2026-10-24

## Context

The application ships as an HTTP server read from a browser. Grading asks for a program
that opens its own window: a double-click should show the interface, with no address bar,
no manual `make run`, no port to remember.

The UI is plain HTML/CSS/JS served from `internal/api`; rewriting it in a desktop toolkit
(Qt, GTK, Fyne, Wails) would duplicate the whole frontend and break the "no build step"
rule. What is missing is only a window with a web engine inside.

## Decision

Add one Go dependency, `github.com/webview/webview_go` (a thin cgo binding over WebView2
on Windows and WebKitGTK on Linux), behind the build tag `gui`:

- `cmd/app/main.go` listens on a free port (`127.0.0.1:0` by default in this mode), then
  hands the URL to `openWindow`.
- `cmd/app/gui.go` (`//go:build gui`) opens a 1280x860 window titled
  "Credenciales seguras" and blocks until it is closed; the server is shut down after.
- `cmd/app/nogui.go` (`//go:build !gui`) keeps today's behaviour: print the address and
  wait for Ctrl+C. It leaves cgo out of the default build.

`make app` and `make app-windows` build the windowed binary; `make build` and
`make build-windows` keep producing the browser-only binary.

## Consequences

- The plain binaries stay dependency-free and portable; the windowed one needs the
  platform engine at runtime: WebKitGTK 4.1 on Linux, WebView2 on Windows (preinstalled
  with Edge on Windows 10/11, otherwise a small runtime installer).
- The `gui` binary is not a static Go binary: it links GTK3/WebKitGTK, so the target
  machine needs those shared libraries, and Windows needs mingw-w64 to cross-compile.
- webview_go requests `pkg-config: gtk+-3.0 webkit2gtk-4.0`, but current distributions
  (Arch, Fedora, Debian trixie) ship WebKitGTK 4.1 only — same API, libsoup3 instead of
  libsoup2. `make gui-pc` writes a local alias `.pc` that requires 4.1, and only when 4.0
  is genuinely absent. This is a build-environment shim, not a patch of the dependency:
  no fork, no vendoring, so `go.sum` stays verifiable.
- `webview.cc` includes `EventToken.h` in Windows SDK spelling, but mingw ships
  `eventtoken.h`; on a case-sensitive filesystem that is a fatal error. `make mingw-include`
  writes a one-line shim and the build exports `CPATH` (not `-I`) so a repository path
  with spaces still works. Build-environment shim again, not a patched dependency.
- If no engine is available (no display, or no WebView2 runtime) the window build opens
  the default browser and keeps serving. The Windows binary is linked with
  `-H=windowsgui`, so a silent exit would look like the app doing nothing at all.
- The window itself is only verified on Linux; under Wine the WebView2 runtime does not
  exist, so there the `.exe` exercises the browser fallback and the API, not the window.
- `go vet ./...` does not compile the tagged file by default, so the gate stays Go-only
  and non-interactive; `make app` is what type-checks the window code.
- The dependency is optional at the source level: deleting `gui.go`, `nogui.go` and the
  Makefile targets would leave a working program.
