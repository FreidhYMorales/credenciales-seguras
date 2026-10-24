# Architecture

## Purpose

Local web app for the Automata final project: validate emails and passwords against
formal regular expressions, report which rule fails, generate credentials by walking a
finite automaton, and expose the ER -> AFND-e -> AFD -> AFD minimo construction per rule.

Source of truth for scope: `docs/plans/app-credenciales/plan.md`.

## Stack

| Layer | Choice | Why |
|---|---|---|
| Backend | Go 1.27, stdlib only | one static binary, no venv, automata build in ms |
| Frontend | HTML + CSS + JS vanilla, no build step | served by the same binary, zero toolchain |
| Graph rendering | Graphviz `dot` at build time, SVG embedded | Windows demo machine may not have Graphviz |
| Persistence | none | nothing to store, no accounts |
| Tests | `go test` + RE2 differential oracle | Go's `regexp` is an independent implementation |

No external Go module dependencies. A new dependency requires an ADR.

## Module boundaries

```
cmd/app/            HTTP server, flags, wiring
cmd/figgen/         build-time SVG generation (graphviz)
cmd/pares/          regenerates testdata/pares.json through the same handlers
internal/automata/  AST, Thompson, subset, tree/followpos, minimize, equivalence, DOT, simulate, toRE2
internal/rules/     alphabets, labels, 15 automata, metadata registry
internal/validate/  per-rule report, normalization, security level
internal/generate/  reachability-guided random walk over the minimized product
internal/api/       handlers and JSON contracts
web/                index.html, app.js, styles.css, embed.go
assets/             figures/*.svg, embed.go
testdata/           pares.json, oracle/, embed.go
tools/oracle/       Python differential scripts (not runtime)
docs/               architecture.md, development.md, adr/, api.md, plans/
```

Dependency direction is one-way: `cmd` -> `api` -> `validate`/`generate`/`rules` -> `automata`.
`automata` imports nothing from the project.

Each embedded tree has its own `embed.go` because `//go:embed` cannot reference parent
directories: `web/embed.go` (package `web`), `assets/embed.go` (package `assets`),
`testdata/embed.go` (package `testdata`).

## Core model

```
Rule        id, nombre, tipo (accept|reject), er, re2, tieneER, conteos{nfa,sub,tree,min}
Credential  tipo (password|email), valor, normalizado
RuleResult  id, ok, evaluado, mensaje, posicion (only P0/C0)
Report      valido, normalizado, longitud, nivel|null, entropia|null, reglas[]
Generation  valor, nivel|null, longitud, entropia|null, semilla, camino[]
```

Rule ids are frozen: `P0..P8`, `C0..C6`, `C*`. 15 runtime automata; `P0`/`C0` are set
checks; `COMP` (24 permutations) is document-only.

## Contracts

Single envelope: `{"ok": true, "datos": ...}` or `{"ok": false, "error": "..."}`.
HTTP 200 for validation outcomes, 4xx/5xx only for real errors. Endpoint payloads live
in `docs/api.md`.

## Determinism

Determinism is a design constraint, not a hope:

- atoms and transitions are sorted slices; no map iteration in construction or in the walk
- every draw uses `rand.New(rand.NewSource(semilla))`, including rejection retries
- the same seed reproduces the same value and the same path across process restarts

## Quality gate

`make check` is the only gate. It is Go-only and non-interactive:

```
gofmt -l .        (must be empty)
go vet ./...
go test ./...
```

Plus the project tests: port fidelity against the RE2 oracle (1500 strings per rule),
`min = tree` equivalence, 20-pair regeneration, boundary values. `make check-oracle`
adds the Python differential and is never required to run the app.

## Constraints and known risks

- `//go:embed` cannot escape its package (see `embed.go` per tree).
- The generator multiplies **minimized** DFAs (`P1_min` 65 x 2^4 = 1040), not subset DFAs
  (257 x 9^4 = 1.69 M). `minimize` therefore returns a DFA, not a count.
- Explorer graphs are drawn only when a view has <= 40 states and <= 80 edges; above that
  the endpoint returns `svg: null` and the UI shows counts plus a paginated table.
- Windows binary is cross-compiled and tested during rehearsal; the Linux build is the
  primary demo target.
