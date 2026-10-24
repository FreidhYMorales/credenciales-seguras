# AGENTS.md

Instructions for AI agents working in this repository.

## Read first

1. `docs/architecture.md` — stack, module boundaries, quality gate.
2. `docs/development.md` — commands and conventions.
3. `docs/plans/app-credenciales/plan.md` — scope, contracts, definition of done.
4. `docs/api.md` — endpoint payloads (frozen contract).

## Hard rules

- `make check` must pass before any commit. It is Go-only and non-interactive.
- No external Go dependencies. A new one needs an ADR under `docs/adr/`.
- Rule ids are frozen: `P0..P8`, `C0..C6`, `C*`. Do not invent or renumber.
- JSON keys and UI strings in Spanish; code, comments and commits in English.
- Determinism: sorted slices, no map iteration in construction or in the walk, every draw
  through `rand.New(rand.NewSource(semilla))`.
- `internal/automata` stays free of project imports and of UI concerns.
- Never weaken a check (skip a test, loosen a comparison) to make the gate pass.
- Conventional commits. No AI attribution, no `Co-Authored-By`.

## Where things go

| Change | Location |
|---|---|
| New rule | `internal/rules` (ER, RE2, labels, metadata) |
| New endpoint | contract in `docs/api.md`, then `internal/api`, then `web/app.js` |
| New figure | `cmd/figgen`, output to `assets/figures/` |
| Test case | next to the package, table-driven |

## Definition of done

See the `Definition of done` section of the plan. In short: `make check` green, 15
automata matching the document's state table, 0 discrepancies against the RE2 oracle over
1500 strings per rule, 20-pair regeneration byte for byte, both binaries built.
