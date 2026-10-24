# Development

## Prerequisites

| Tool | Version | Needed for |
|---|---|---|
| Go | 1.27+ | build, test, run |
| Graphviz `dot` | 16+ | regenerating `assets/figures/*.svg` (build time only) |
| Python 3 | 3.11+ | `make check-oracle` and `make documento` (optional) |
| python-docx + LibreOffice | any recent | `make documento` only (optional) |

The runtime binary needs none of them.

## Commands

```sh
make check          # gofmt + vet + tests (the only gate)
make test           # go test ./...
make fmt            # gofmt -w .
make build          # linux binary  -> bin/credenciales
make build-windows  # windows .exe  -> bin/credenciales.exe
make run            # build and serve on http://localhost:8080
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

## Layout rules

- New rule? Add it to `internal/rules` with its ER, its RE2 translation, its labels and its
  metadata row. Never hardcode rule ids in handlers or in the frontend.
- New endpoint? Contract goes into `docs/api.md` first, then the handler, then the frontend.
- `internal/automata` stays dependency-free and UI-free.

## Conventions

- Package names, identifiers and comments in English. UI strings and JSON keys in Spanish.
- Table-driven tests; every rule has at least one accept and one reject case.
- Sorted slices instead of maps wherever iteration order can reach output.
- No external Go dependencies without an ADR.

## Seeded demo values

| Seed | Kind | Result |
|---|---|---|
| `2026` | password, length 20 | reproduced in the demo and in the video |
| `4490` | email | reproduced in the demo |

Both must reproduce exactly on the presentation machine; the regeneration test fails the
build if they do not.
