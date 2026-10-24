# ADR 0005 — Generation walks the minimized product DFA

- Status: accepted
- Date: 2026-10-23

## Context

Assignment 7.a.iii requires generating credentials **by random walk over the automaton**,
not by sampling random strings and filtering them. A walk needs a DFA small enough to
build at startup, and it must be deterministic given a seed.

## Decision

- Passwords walk the product of the **minimized** DFAs `P1_min x P2_min x P3_min x
  P4_min x P5_min` (65 x 2 x 2 x 2 x 2 = 1040 states), restricted to states that can
  still reach an accepting state.
- `P6` (repetition), `P7` (sequence) and `P8` (forbidden words) are handled by rejection
  sampling: draw a walk, reject, redraw. 200 retries with the same seeded PRNG.
- Emails walk `C*` with the same reachability guidance towards a seeded target length of
  12-32 (accept at exactly that length, else redraw), and reject-sample `C2`-`C6`.
- `minimize` must return a DFA (classes + transition table), not a state count, or the
  product cannot be built. It is a **partial** DFA: the dead class is removed, a missing
  transition means reject, `P1_min` is 65 and the reachable product is 944 states.

## Consequences

- Building the product over subset DFAs instead would be 257 x 9^4 = 1.69 M states; a test
  asserts the reachable product stays at or below 1040.
- Rejection sampling keeps the walk automaton-guided while leaving the hard rules simple.
- Determinism is testable: same seed, same value and same path, across process restarts.
