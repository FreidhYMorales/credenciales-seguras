# Development

## Prerequisites

| Tool | Version | Needed for |
|---|---|---|
| Go | 1.27+ | build, test, run |
| Graphviz `dot` | 16+ | regenerating `assets/figures/*.svg` (build time only) |
| Python 3 | 3.11+ | `make check-oracle` only (optional) |

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
make check-oracle   # Python differential, optional
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
- No external Go dependencies without an ADR.

## Seeded demo values

| Seed | Kind | Result |
|---|---|---|
| `2026` | password, length 20 | reproduced in the demo and in the video |
| `4490` | email | reproduced in the demo |

Both must reproduce exactly on the presentation machine; the regeneration test fails the
build if they do not.
