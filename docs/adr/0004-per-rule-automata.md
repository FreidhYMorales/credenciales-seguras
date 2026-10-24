# ADR 0004 — Per-rule automata plus a product check

- Status: accepted
- Date: 2026-10-23

## Context

The assignment requires reporting *which* rule fails, not only that a credential is
invalid. A single ER for the whole policy can only say "no".

## Decision

One automaton per rule (P1-P8, C1-C6, C*) and a product check when the combination
matters. `P0` and `C0` are alphabet set checks and have no automaton. The 24-permutation
composition automaton is documented but not built at runtime (48 137 subset states,
~59 s in the prototype).

## Consequences

- The report names the exact failing rule with a message and, for `P0`/`C0`, a position.
- Rules skipped by a short-circuit are still returned with `evaluado: false` so the table
  is always complete.
- The product automaton is used only by the generator, where it is bounded (1040 states,
  see ADR 0005).
