# ADR 0002 — Pre-rendered SVG embedded in the binary

- Status: accepted
- Date: 2026-10-23

## Context

The explorer draws automata. Rendering them at runtime would need Graphviz on the demo
machine, and the Windows `.exe` would need it too.

## Decision

Generate the figures at build time with `dot` (`cmd/figgen`), store them as SVG under
`assets/figures/`, and embed them with `//go:embed`. At runtime the explorer serves those
files plus a paginated transition table generated in Go.

## Consequences

- The demo works offline and identically on Linux and Windows.
- `make figures` must be re-run when a rule or an automaton changes; the build does not
  detect it automatically.
- Automata above the drawing threshold (40 states / 80 edges) have no SVG; the UI falls
  back to counts plus the transition table.
