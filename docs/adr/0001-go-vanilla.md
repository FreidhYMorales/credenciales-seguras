# ADR 0001 — Go + vanilla web instead of Python + Streamlit

- Status: accepted
- Date: 2026-10-23

## Context

The advance document shipped a working Python engine (`codigo_avance/`). Its full build
takes ~9 s (~70 s with the 24-permutation automaton). The demo machine may not have
Python, and the presentation is on a machine we do not control.

## Decision

Reimplement the engine in Go 1.27 (stdlib only) and serve a vanilla HTML/CSS/JS frontend
from the same binary. Keep the Python engine as a test oracle only.

## Consequences

- One static binary, no venv, no `pip install`, no build step in the browser.
- All 15 automata build once in ~970 ms at startup; the server pays it once.
- The port is the highest-risk item and is verified against the RE2 oracle and against
  the state counts printed in the document.
- No third-party Go module; the whole dependency surface is the standard library.
