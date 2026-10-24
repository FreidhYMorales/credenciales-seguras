# ADR 0003 — Two independent oracles: RE2 and the Python prototype

- Status: accepted
- Date: 2026-10-23

## Context

A port that is tested against itself proves nothing. The assignment also asks (steps 5-6)
for a translation of the ERs to the regex tool of a chosen language.

## Decision

- **RE2** (`regexp`, Go) is the independent oracle. The ER of every rule is translated
  from the AST by `toRE2()` and compiled with `regexp`; both implementations must agree
  on 1500 random strings per rule.
- **Python** (`tools/oracle/`) is kept only to check port fidelity: same input, same
  verdict, same state counts as the advance document.
- `min = tree` is proven by the equivalence check between the two DFAs, not by comparing
  state counts alone.

## Consequences

- `make check` needs only Go; the Python differential is behind `make check-oracle`.
- The ER -> RE2 table doubles as deliverable evidence for steps 5-6.
- RE2 has no lookaheads, so the per-rule + product design cannot be shortcut with
  `(?=.*[A-Z])` even accidentally.
