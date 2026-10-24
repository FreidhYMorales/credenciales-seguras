# Plan: lexical validator and generator of secure credentials

- Status: v4 (reviews 1-3 applied; implementation complete and pushed)
- Deadline: 2026-10-28 (presentation window 2026-10-26 to 2026-10-28)
- Mode: collaborative (3 people: 1 coder, 2 document/video)
- Supersedes: the Python prototype in `codigo_avance/` as runtime; it is kept as test oracle and figure source only.

## Platforms

| Target | In v1? | Min version / form factor | Distribution |
|---|---|---|---|
| Web app served from localhost | yes | Chrome/Firefox/Edge last 2 versions; desktop width >= 1024 px | local binary, browser at `http://localhost:8080` |
| Linux amd64 binary | yes | glibc distro, no runtime deps | `go build` |
| Windows amd64 binary | yes | Windows 10+ | cross-compiled `.exe`, to be tested on a real machine during rehearsal |
| Mobile / touch | no | n/a | n/a |
| Remote hosting | no | n/a | n/a |

Out of scope: any deployment that requires a server on the internet, databases, user accounts, TLS.

Demo machine: the development laptop (Linux) by default; the Windows `.exe` is the tested backup.

## Research

Source material read in full:

- `TAREA FINAL AUTÓMATAS PD.pdf` — assignment: NIST SP 800-63B + CIS policy, RFC 5322 email syntax, ER design, Thompson NFA, AFD, application with per-rule report, automaton-guided generation and security level. Grading: the final is **10 pts = 5 presentation + 5 deliverables** (document, 10 valid pairs, 10 invalid pairs, code, 1-minute video). The 5-point advance is a separate grade already earned, not part of the final 10.
- `AYLF Final Avance corregido.pdf` (35 pp.) — already covers steps 1-4: alphabets, rules P0-P8 and C0-C6, ER per rule, regular grammars and derivations, Thompson construction with state tables, tree method, minimal DFA, test annex and security-level criterion (Annex C).
- `codigo_avance/` — working Python engine: AST, Thompson NFA, subset construction, tree/followpos DFA, Moore minimization, equivalence check, DOT export, per-rule validator. Verified: state counts match the document exactly (P2 34/9/2, email 136/33/12/8, etc.).

Findings that shape v1:

- v1: the engine must be reimplemented in Go, not shelled out to Python, because the demo machine may not have Python and the Python build takes 9 s (70 s with the 24-permutation rule).
- v1: steps 5 and 6 require choosing a language and translating the ERs to its regex tool. Go's `regexp` (RE2) is the target syntax; the translation is generated from the AST (`toRE2`) and RE2 doubles as the independent test oracle.
- v1: per-rule automata plus a product check, not one giant ER. This is what enables "which rule failed".
- v1: generation must be a reachability-guided random walk over a DFA, not random strings (assignment 7.a.iii).
- v1: the security level is already defined (Annex C): Baja if any acceptance rule fails or any rejection filter matches; Media 15-19 chars; Fuerte 20+. It applies to passwords only.
- v1: the 10 valid pairs must be produced by the tool with recorded seeds, not copied from the annex.
- v2: personalized email suggestions from name/date of birth. Rejected for v1: it is template-based, does not satisfy 7.a.iii, and costs 5-7 h.

## Product vision

- **v1 (deliverable by deadline):** a local web app that validates any email/password against the formal rules, names the exact failing rule, generates valid credentials by walking the automaton and shows the walk, exposes the NFA -> DFA -> minimal DFA construction for every rule, and renders the 20 test pairs as executable cases.
- **v2 (later):** registration assistant (name/DOB -> email suggestions validated by the automaton, with a guessability warning), animated walk over the rendered graph, export of the audit report as PDF, password strength feedback while typing.

## Core model

```
Rule          id, name, kind (accept|reject), ER text, re2 text, automata (nfa, subset, tree, min), has_er
Credential    kind (password|email), raw value, normalized value (spaces stripped)
RuleResult    rule id, ok, evaluado, message, position (only P0/C0)
Report        verdict, results[], level, length, entropy
Generation    value, path[{step, state, symbol, class}], seed, target length
Simulation    value, accepted, trace[{step, state, symbol}]
Pair          id, email, password, email_expected, password_expected, seed_email, seed_password, failure_rule
```

Key events: validate, generate, simulate, inspect automaton, load pair.

## Foundation

Stack decided with the user: **Go 1.27** backend (stdlib only: `net/http`, `embed`, `html/template`, `encoding/json`, `math/rand`) and **HTML/CSS/JS vanilla** frontend, no build step, no framework.

Rationale: single static binary for the demo (no venv, no `pip install`), the 15 automata build once in ~970 ms at startup instead of 9 s in Python, and Go's `regexp` is RE2 without lookaheads, so the app cannot cheat with `(?=.*[A-Z])` and must use the product automaton the assignment asks for.

Proposed layout (each embedded tree has its own `embed.go`, because `//go:embed` cannot reference parent directories):

```
cmd/app/            HTTP server, flags; imports web, assets, pairs
cmd/figgen/         build-time SVG generation with graphviz
cmd/pares/          generates the 10 valid pairs from seeds through the API
internal/automata/  AST, Thompson NFA, subset DFA, tree DFA, minimize (returns DFA), equivalence, DOT, simulate, toRE2
internal/rules/     alphabets, labels, 15 automata, registry with metadata
internal/validate/  per-rule report, normalization, security level
internal/generate/  reachability-guided random walk
internal/api/       handlers and JSON contracts
web/                index.html, app.js, styles.css, embed.go (package web)
assets/             figures/*.svg, embed.go (package assets)
testdata/           pares.json, oracle/, embed.go (package testdata)
tools/oracle/       Python scripts, not part of the runtime
docs/               architecture.md, adr/, api.md, plans/
```

Conventions:

- Package names in English, UI strings and JSON keys in Spanish (the document and the audience are Spanish).
- Rule ids are stable: `P0..P8`, `C0..C6`, `C*`.
- No external Go dependencies. If one becomes necessary, it needs an ADR.
- Table-driven tests; every rule has at least one accept and one reject case.
- Determinism is a design constraint, not a hope: atoms and transitions are sorted slices, no map iteration in construction or walk, `rand.New(rand.NewSource(seed))` for every draw including retries.
- Single response envelope: `{"ok": true, "datos": ...}` or `{"ok": false, "error": "..."}`; HTTP 200 for validation outcomes, 4xx/5xx only for real errors. Every endpoint contract below describes the `datos` payload.
- `gofmt` clean; `go vet` clean; `make check` runs the Go-only checks; `make check-oracle` adds the Python differential (optional, not needed by the professor).

ADRs to write during bootstrap:

1. Go + vanilla over Python + Streamlit (decided, recorded).
2. Pre-rendered SVG embedded instead of runtime graphviz (Windows demo).
3. Python engine kept as port-fidelity oracle; RE2 is the independent oracle.
4. Per-rule automata + product check instead of one ER with lookaheads.
5. Product automaton for generation built over the **minimized** DFAs (`P1_min` 65 × `P2_min..P5_min` 2⁴ = 1040), with rejection sampling for P6-P8.

## Schedule

Presentation window 2026-10-26 to 2026-10-28; the exact slot is still unknown, so the
plan targets the worst case (Monday 26).

The original estimate was 5.25 person-days, **4.55 of them on the coder** (the earlier
4.45 was an arithmetic error: the README row was left out). That was the critical path
and it did not fit the window at 10 h/day: Saturday, Sunday and Monday morning give at
most 30-36 h against 36.4 h needed. The implementation is therefore no longer scheduled
as human typing time: it was written by the coding agent on 24 Oct, is green and is
pushed. What remains is the human work that cannot be delegated.

| Date | State |
|---|---|
| Sat 24 | **Implementation complete.** Engine, 15 automata, validator, generator, API, five screens, 20 pairs, 26 figures. `make check` and `make check-oracle` green; `main` pushed. |
| Sun 25 | Document: test-case tables, the 20 pairs, steps 5-6, screenshots. Video recorded against the frozen build. |
| Mon 26 | Rehearsal of the run sheet, Windows cross-compile smoke test, re-record slot. |
| 26-28 | Presentation. |

**Checkpoint, now measurable:** `make check && make check-oracle` must be green. That is
the only pass criterion, it runs in seconds, and it does not depend on work scheduled
later in the day — the earlier wording ("all 15 automata + RE2 oracle by Sunday noon")
was impossible to satisfy because those rows were scheduled after the checkpoint. The
day-1 cut no longer applies to the engine; if the human work slips, the contingency order
below still holds.

## Screens

### Validate (primary)

```
┌─ Credenciales seguras ───────────────────────────────────────────┐
│  ( Correo )  ( Contraseña )                                      │
├──────────────────────────────────────────────────────────────────┤
│  Contraseña                                                      │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ ••••••••••••••••••••••                            [ ojo ]  │  │
│  └────────────────────────────────────────────────────────────┘  │
│  [ Validar ]                                                     │
│                                                                  │
│  ✘ INVÁLIDA — 1 regla incumplida                    Nivel: Baja │
│  ──────────────────────────────────────────────────────────────  │
│  P0 alfabeto           ✔                                        │
│  P1 longitud 15-64     ✘  longitud = 12 (faltan 3)              │
│  P2 mayúscula          ✔                                        │
│  P3 minúscula          ✔                                        │
│  P4 dígito             ✔                                        │
│  P5 especial           ✔                                        │
│  P6 repetición         ✔                                        │
│  P7 secuencia          ✔                                        │
│  P8 palabra prohibida  ✔                                        │
└──────────────────────────────────────────────────────────────────┘
```

- Chosen: B (tabs, one credential at a time) with A's report panel.
- Rules: spaces are stripped before validation and the normalized value is shown next to the length; the tab remembers its last input; Enter triggers validation; empty input shows the empty state, not an error; the length line says "faltan N" below 15 and "sobran N" above 64.
- States: empty (form + hint) | result (verdict + table) | error (API/network message). No `loading` spinner and no `offline` state.
- The table always shows every rule of the kind (9 for password, 7 for email); rules not evaluated because of a short-circuit are rendered as `—` with `evaluado: false`.
- Verdict colors: green valid, red invalid, amber when invalid only because of rejection filters.
- Entropy is displayed next to the level for passwords ("131 bits (bruto)"); for emails the level and entropy are hidden.

| UI element | Backend demand |
|---|---|
| Verdict + rule table | `POST /api/validar` body `{tipo:"password"\|"email", valor}` -> `datos: {valido, normalizado, longitud, nivel\|null, entropia\|null, reglas:[{id,nombre,tipo,ok,evaluado,mensaje,posicion}]}` |
| Length message | backend returns `longitud`; frontend computes "faltan/sobran" against 15/64 |
| Security level badge | backend computes level per Annex C, only for passwords (`null` for emails) |

### Generate

```
┌─ Generar credencial ──────────────────────────────────────────────────┐
│ ┌─ Parámetros ─────────────┐  ┌─ Resultado ─────────────────────────┐ │
│ │ Tipo                     │  │ Xk9$mTq2#vLw8@RzQe4&                │ │
│ │  (•) Contraseña          │  │ FUERTE · 20 caracteres · 131 bits   │ │
│ │  ( ) Correo              │  │                                     │ │
│ │                          │  │ [ Copiar ]  [ Validar esta ]        │ │
│ │ Longitud  [====•=====] 20│  └─────────────────────────────────────┘ │
│ │ Semilla   [ 2026       ] │  ┌─ Camino sobre el AFD ───────────────┐ │
│ │                          │  │ q0 ─L→ q1 ─l→ q2 ─N→ q3 ─S→ q4     │ │
│ │ [ Generar ]              │  │ ─L→ q5 ─l→ q6 ─N→ q7 ─S→ q8 …      │ │
│ └──────────────────────────┘  └─────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────────┘
```

- Chosen: B (two columns).
- Rules: seed is optional (empty = random seed, shown after generation so it can be replayed); password length slider 15-64, default 20 (Fuerte); email length is not user-set and the level badge is hidden for emails.
- States: empty | result | error.
- "Validar esta" prefills the Validate tab and switches to it.

| UI element | Backend demand |
|---|---|
| Generated value + level + entropy | `POST /api/generar` body `{tipo, longitud?, semilla?}` -> `datos: {valor, nivel\|null, longitud, entropia\|null, semilla, camino[]}` |
| Path chain | `camino:[{paso, estado, simbolo, clase}]` |
| Replay | same seed returns the same value and the same path (deterministic PRNG) |

### Automata explorer

```
┌─ Explorador de autómatas ─────────────────────────────────────────┐
│ ┌─ Reglas ──────────┐ ┌─ P2 mayúscula ─────────────────────────┐  │
│ │ Regla      mín.   │ │ AFND-ε 34 · Subconjuntos 9 · Árbol 2   │  │
│ │  P1 long.   65    │ │ ┌────────────────────────────────────┐ │  │
│ │  P2 mayús.   2    │ │ │                                    │ │  │
│ │  P3 minús.   2    │ │ │       [ grafo Graphviz SVG ]       │ │  │
│ │  P4 dígito   2    │ │ │                                    │ │  │
│ │  P5 espec.   2    │ │ └────────────────────────────────────┘ │  │
│ │  P6 repet. 190    │ │ ┌─ Transiciones ─────────────────────┐ │  │
│ │  P7 secue. 158    │ │ │ q0  L→q1  l→q0  N→q0  S→q0        │ │  │
│ │  P8 palab.  44    │ │ │ q1  L→q1  l→q1  N→q1  S→q1        │ │  │
│ │ Correos           │ │ └────────────────────────────────────┘ │  │
│ │  C1 … C6          │ │ Cadena [ Aa1! ]   q0─L→q1─l→q1─N→q1 ✔  │  │
│ └───────────────────┘ └────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
```

- Chosen: C (explorer with sidebar).
- Graph rule: a view is drawn when it has **<= 40 states and <= 80 edges**; otherwise the endpoint returns `svg: null` and the UI shows the Thompson structural figure, the paginated transition table and the counts, with a note. This keeps P2's 34-state NFA, the 9-state subset, the 2-state tree, C5 and C*'s subset/tree; P1 (1092/257/65), P6 (782), P7 (630), C4 (2616/374) and C6 (5606) show tables.
- Views: `nfa`, `sub`, `tree`, `min`; the sidebar shows the **minimal** state count with a header ("mín.").
- States: empty (no rule selected) | rule selected | error.
- Simulation: input string, step-by-step trace on the selected view, final accept/reject, failure position.

| UI element | Backend demand |
|---|---|
| Rule list with counts | `GET /api/reglas` -> `datos: [{id,nombre,tipo,er,re2,conteos:{nfa,sub,tree,min}}]` |
| Graph per view | `GET /api/automatas/{id}?vista=nfa\|sub\|tree\|min&pagina=1&limite=50` -> `datos: {svg\|null, transiciones[], total, pagina}` |
| Trace | `POST /api/simular` body `{id, valor, vista}` -> `datos: {acepta, traza:[{paso,estado,simbolo}], fallo}` |

### Rules reference

Table with filters by kind (password/email): id, rule, accept/reject, **formal ER**, **RE2 translation**, foundation (NIST/CIS/RFC), state counts. Clicking a rule opens it in the explorer; "probar" loads a sample into Validate. 17 rows: P0-P8, C0-C6, C*; 15 have an ER, P0 and C0 show "conjunto" because they are set checks.

| UI element | Backend demand |
|---|---|
| Static table | `GET /api/reglas` (same payload as the explorer, no extra endpoint) |

### Test pairs

Table of the 20 pairs from `testdata/pares.json`: id, email, password, expected verdict per credential, failing rule, and the two seeds. Each row has "probar", which loads both credentials into Validate (email tab and password tab) without a new API.

| UI element | Backend demand |
|---|---|
| Pair table | `GET /api/pares` -> `datos: [...]` from the embedded `pares.json` |
| "Probar" | frontend only (fills the Validate form) |

## Data captured / data inventory

| Screen | Data shown | Actions (events) | Source |
|---|---|---|---|
| Validate | normalized value, rule results, level, entropy | validate, switch tab, reveal password | `POST /api/validar` |
| Generate | value, level, path, seed | generate, copy, validate this, replay | `POST /api/generar` |
| Explorer | rule metadata, SVG, transition table, trace | select rule, switch view, page, simulate | `GET /api/reglas`, `GET /api/automatas/{id}`, `POST /api/simular` |
| Rules | rule table with ER, RE2 and foundation | filter, open in explorer, load sample | `GET /api/reglas` |
| Test pairs | 20 pairs with expected verdicts and seeds | load pair into Validate | `GET /api/pares` |

No persistence: nothing is stored, no cookies, no accounts.

## Scope

| Item | Priority | Est. dev-days | Demo value |
|---|---|---|---|
| Go port: AST, Thompson NFA, subset DFA, tree DFA, minimize (returns DFA), equivalence, DOT, simulate | MUST | 0.8 | invisible but everything depends on it |
| `toRE2` + RE2 oracle harness | MUST | 0.1 | medium (defense for steps 5-6) |
| 15 automata in Go with labels and metadata | MUST | 0.3 | high |
| Per-rule validator + normalization + security level | MUST | 0.2 | high |
| Reachability-guided generator (minimized product, rejection sampling, seed) | MUST | 0.7 | high |
| HTTP API + JSON contracts | MUST | 0.3 | none |
| Validate screen | MUST | 0.4 | highest |
| Generate screen | MUST | 0.2 | high |
| Explorer screen + pre-rendered figures | MUST | 0.5 | high |
| Rules screen | MUST | 0.2 | medium |
| Test pairs screen | MUST | 0.15 | high (shows the graded deliverable) |
| 20 pairs generated from seeds + regeneration test | MUST | 0.3 | high (deliverable) |
| Oracle/differential tests | MUST | 0.2 | medium (defense) |
| Document completion (test cases, 20 pairs, steps 5-6, screenshots) | MUST | 0.5 | deliverable |
| 1-minute video | MUST | 0.2 | deliverable |
| Cross-compile Windows + rehearsal | MUST | 0.1 | medium |
| Repo + README + run instructions | MUST | 0.1 | deliverable |

Total ~5.25 person-days: **~4.55 on the coder, ~0.7 on the two document/video people**
(document 0.5 + video 0.2). The coder's load was the critical path; see the Schedule
section for how it was actually absorbed.

## Contingency cut order

Applied top-down only if the human work slips, and recorded in the plan when applied.
Each cut names exactly what disappears, so the post-cut definition of done stays true.

1. **Explorer `tree`/`min` SVGs and `vista=min`.** The route, the sidebar, the four
   counts, the paginated tables and the trace stay; only the two graphs and the fourth
   tab go. The P2 NFA (34 states) and subset (9 states) graphs are kept: they are the
   demo's centrepiece. Saves ~0.3 d.
2. **Email generation toggle in the UI.** `POST /api/generar` keeps working and the 10
   email pairs are still produced by `cmd/pares`; only the button disappears. Saves
   ~0.15 d and costs nothing that is graded (the assignment says "a password **or** an
   email").
3. **Rules screen as a separate route.** Fold the table into the explorer sidebar; the
   ER, RE2 and standard columns survive. Only valid if cut 1 was *not* applied, because
   the sidebar must still exist. Saves ~0.2 d.
4. **`/api/simular` trace.** The Validate report and the counts carry the story. This
   voids the "simulates `Aa1!`" line of the definition of done.
5. **Test pairs screen, last.** The document table plus a screenshot replace it. This
   voids the "Pairs lists 20 rows" line.

The cut line is a polish lever, not a rescue: no permutation of these five recovers the
1.45-day gap that the original schedule had. That gap was closed by writing the code with
the agent, not by cutting scope.

## Definition of done

Each line states what it measures and, where a cut voids it, which one.

- **Gate.** `make check` is Go-only and non-interactive: `gofmt -l` (must be empty),
  `go vet ./...`, `go test ./...`. It is green from the first commit and never weakened.
  `make check-oracle` adds the Python differential and is not required to run the app.
- **Port correctness.** The four counts of **12 printed rows** match the advance document
  digit for digit: P1, P2, P6, P7, P8, C1, C2, C3, C4, C5, C6, C\*. P3/P4/P5 are
  identical to P2 by construction. COMP is excluded by design (document-only), so the
  comparable set is 12 rows, not 13. `min ≡ tree` and `subset ≡ tree` hold by the
  equivalence check for all 15 automata.
- **Independent oracles.** RE2: 0 discrepancies over 1500 strings per rule (22 500
  comparisons). Python: 15 rules with the same four counts and 6000 strings with the
  same verdict, 0 discrepancies.
- **Generator property.** 1000 passwords (lengths 15-64) and 300 emails are 100 %
  accepted; the reachable product has **944** states, under the 1040 bound; the same
  seed reproduces the same value and the same path, and the test runs it twice in one
  process plus across restarts.
- **Boundaries.** Password lengths 14/15/16 and 64/65, levels 19/20; email label 63/64,
  local part 64/65, domain 255/256; spaces stripped on both kinds; a 20-character
  password containing a blocklisted word is Baja, not Fuerte.
- **Graded pairs.** `testdata/pares.json` regenerates byte for byte from its seeds; all
  20 verdicts match; each of the 10 invalid pairs has exactly one defective credential
  and names the rule that fails.
- **Screen acceptance** (each line is voided by the cut in brackets): Validate names the
  failing rule for the 10 invalid pairs and accepts the 10 valid ones; Generate
  reproduces seed `2026` -> `>nslx613LK2%|#PxFbA5` (Fuerte, 131 bits) and seed `4490` ->
  `IykLbhvynB@F.mT`; Explorer shows 34/9/2 for P2 and simulates `Aa1!` [cut 4]; Rules
  lists 17 rows, 15 with ER [cut 3]; Pairs lists 20 rows [cut 5].
- **Deliverables.** Document updated with test cases, the 20 pairs, steps 5-6 and
  screenshots; video <= 1 minute; Linux and Windows binaries built.

## Demo data

- `testdata/pares.json`: 10 valid pairs produced by `POST /api/generar` with `seed_email` and `seed_password`, and 10 invalid pairs **authored by hand** (they are not generated) covering length, missing special, missing digit, missing upper, missing lower, triple repetition, ascending sequence, forbidden word, invalid email structure, invalid TLD. Fields: `id`, `email`, `password`, `email_esperado`, `password_esperado`, `regla_fallida`, `seed_email`, `seed_password`. Each invalid pair has exactly one defective credential, which is the one `regla_fallida` names.
- `cmd/pares` calls the same generator as the API (in-process) and asserts the generated values match the file, so the regeneration test does not need a running server.
- Seeds for the demo: `2026` (password, length 20), `4490` (email). Both must reproduce exactly on the presentation machine.
- Sample strings from the document annex are reused where they overlap, so the app and the document agree.
- **Email generation rule** (the earlier plan left it undefined): the walk targets a seeded length of **12-32 characters**; a draw stops when the automaton is in an accepting state at exactly that length, otherwise it is redrawn, up to 200 attempts with the same seeded PRNG. Atoms are weighted by the square of their size, so letters dominate and addresses look like addresses, and a quality filter rejects atext specials (``!#$%&'*+/=?^_`{|}~``) in the generated address even though the grammar allows them. The validator still accepts them, so the filter narrows generation without weakening validation.

## Data gaps & assumptions

| Gap | Assumption | Owner | Resolved? |
|---|---|---|---|
| Graphviz available at build time on the dev machine | yes, 16.1 installed; SVG pre-rendered and embedded, never required at runtime | coder | yes |
| Windows machine available for the rehearsal | yes; the `.exe` is tested before the presentation, otherwise the demo runs on Linux | group | no |
| Go experience of the coder | basic to intermediate; the port is the highest-risk item and the first thing to slip | coder | no |
| Professor accepts the documented exclusions (quoted-string, IP literal, Unicode, CFWS) | yes, the advance document already justifies them | group | no |
| Go engine reproduces the Python AFD sizes exactly | verified by differential tests and by the document table | coder | no |
| Browser used in the classroom | Chrome or Firefox last 2 versions; no polyfills | group | no |
| Delivery channel for document, video and code | zip + repo link uploaded to the course platform; README lists every file | group | no |

## Deferred to v2

- Registration wizard with name/surname/date of birth — rejected for v1: 5-7 h, does not satisfy 7.a.iii, and the assignment only requires direct validation.
- Personalized email suggestions — same reason; would be an extra, not the generator.
- Animated walk over the rendered graph — the explorer's trace covers the requirement.
- Report export (PDF/JSON) — nice for the document, not needed for grading.
- Password strength while typing — not required.
- Mobile layout — the demo is on a laptop/projector.
- 24-permutation composition automaton in the runtime — document-only (59 s in Python, 48 137 subset states; the design uses P2-P5 separately).
- Minimized DFA as an independent explorer view — the count is shown, the graph is not.

## Demo priorities

Run sheet (targets 9:30, leaving 30 s of buffer under the 10-minute cap):

| Time | Content |
|---|---|
| 0:00-0:50 | Problem, standards (NIST/CIS, RFC 5322) and a 30-second pass through the Rules screen with the ER and its RE2 translation |
| 0:50-2:20 | Invalid password -> failing rule P1 and level Baja; then a valid one |
| 2:20-3:50 | Invalid email -> C2 (double dot) and C5 (TLD); then a valid one |
| 3:50-6:10 | Generate with seed 2026, show the walk, replay the seed, level Fuerte |
| 6:10-8:20 | Explorer: P2 goes 34 (NFA) -> 9 (subset) -> 2 (tree/min), plus the trace |
| 8:20-9:20 | Pairs table (the graded 20) and the repo: Go port, tests, `make check` |
| 9:20-9:30 | Closing |

Fallback: the 1-minute video plus a screenshot deck, opened before the demo starts.

Video storyboard (target 55-57 s, under the 60 s cap): 0-8 s invalid password and its
failing rule; 8-20 s a valid pair from the Pairs table; 20-35 s generate with seed 2026
and replay; 35-50 s Explorer with 34/9/2 and the trace; 50-56 s the repo and closing.
Recorded against the frozen build, with a re-record slot before the presentation.

## Collaboration

- Mode: collaborative (3 people)
- People: 1 coder, 2 document/video (names filled by `/flow:split`)
- Frozen contracts: `docs/api.md` (frozen Saturday 18:00), `testdata/pares.json`, rule ids `P0..P8` / `C0..C6`
- Work packages: `docs/work-packages/*.md` only if the group splits further; otherwise a single ordered task list from `/flow:handoff`

## Plan-review resolutions

Reference: `plan-review-1.md`, `plan-review-2.md`.

| # | Finding | Resolution | Decided by |
|---|---|---|---|
| B1 | Generation automaton undefined | Passwords walk the product of the **minimized** DFAs (`P1_min` 65 × `P2_min..P5_min` 2⁴ = 1040) with reachability; P6-P8 by rejection sampling, 200 retries, same seeded PRNG. Emails walk `C*` and reject-sample C2-C6. Property test 1000/kind. | planner |
| B2 | Valid pairs not generated by the tool | 10 valid pairs produced by `POST /api/generar` with recorded seeds; regeneration test compares byte for byte. | planner |
| B3 | Steps 5-6 missing | `toRE2()` from the AST; Go `regexp` (RE2) as independent oracle; document section "Pasos 5 y 6" with the ER->RE2 table; RE2 column in the Rules screen. | planner |
| M1 | Estimate exceeds window, no owners | 1 coder + 2 document/video; Schedule section with day-by-day parallel tracks and a Sunday-noon checkpoint. | user |
| M2 | Python oracle is not independent | RE2 is the independent oracle; Python is kept for port fidelity only; `min ≡ tree` checked by equivalence and min counts against the document table. | planner |
| M3 | Seed determinism asserted, not designed | Sorted slices, no map iteration, `rand.New(rand.NewSource(seed))` for all draws including retries; 100-run and two-restart test. | planner |
| M4 | Automata inventory ambiguous | 15 runtime automata (P1-P8, C1-C6, C*); P0/C0 are set checks without automaton; COMP is document-only. | planner |
| M5 | App validates single credentials, deliverable is pairs | Pair = two independent validations; Test pairs screen lists the 20 from `pares.json` with expected verdicts and loads them into Validate. | user |
| M6 | Report contract under-specified | Always return the full rule set with `evaluado: false` for skipped rules; `posicion` only for P0/C0. | planner |
| M7 | Level/entropy undefined for emails | `nivel` and `entropia` are `null` for emails and the badge is hidden; Annex C applies to passwords only. | planner |
| M8 | No definition of done | Added a Definition of done section: `make check`, port correctness criteria, generator property, boundaries, per-screen acceptance. | planner |
| M9 | No run sheet or fallback | 10-minute run sheet with minute allocations; fallback is the 1-minute video plus screenshots; demo on the dev laptop, `.exe` as backup. | planner |
| M10 | Rules screen SHOULD but demo priority 5 | Promoted to MUST. | planner |
| m1 | Scope total wrong (4.4 vs 4.7) | Recomputed: ~5.25 person-days. | planner |
| m2 | Mock says AFD 2 for P2 | Fixed to 34 / 9 / 2 (NFA / subset / tree-min). | planner |
| m3 | Offline/loading states make no sense | States reduced to empty / result / error. | planner |
| m4 | Explorer contract cannot express pagination or a missing graph | `svg` nullable, `pagina` + `limite` parameters, `total` in the response. | planner |
| m5 | Two error conventions | Single envelope `{ok, datos}` / `{ok, error}`, applied to all five contracts. | planner |
| m6 | Entropy returned but not shown | Shown in Validate next to the level for passwords. | planner |
| m7 | Windows claimed as tested while unresolved | Platform table states intent; testing moved to the rehearsal task. | planner |
| m8 | Figure 3 final state 19 vs 33 | Rejected: verified by running the engine, the P2 NFA final state is 19 because `Star` allocates its states before the child. Document is correct. | planner |
| m9 | No repo, README or run instructions | Added item: git repo + README with build/run and seed list. | planner |
| m10 | "faltan N" only covers short inputs | Defined both branches: "faltan N" below 15, "sobran N" above 64. | planner |
| Q1 | Does "par" mean a combined report? | No: two independent validations, shown side by side in the Test pairs table. | user |
| Q2 | Must valid pairs be regenerable from seeds? | Yes, the assignment says generated by the tool; seeds are recorded and tested. | planner |
| Q3 | Is the 24-permutation automaton in runtime/explorer? | Document-only. | planner |
| Q4 | Who owns which row, and which machine for the demo? | 1 coder, 2 document/video; demo on the dev laptop, `.exe` as backup. | user |
| Q5 | Is the explorer expected, or is a figure enough? | Explorer kept: it is the live evidence for 4.d.ii and the document figures complement it. | planner |
| Q6 | Should the level be shown for emails? | No; `null` and hidden badge. | planner |
| R2-B1 | Serial schedule, no parallel plan or day-1 decision | Schedule section with parallel tracks, 10-12 h/day for the coder and a Sunday-noon checkpoint with a pre-decided cut. | user |
| R2-M1 | Product of subset DFAs would be 1.69 M states | Product is built over minimized DFAs (65 × 2⁴ = 1040); `minimize` returns a DFA (classes + delta), not a count; test asserts the reachable product <= 1040. | planner |
| R2-M2 | `//go:embed` cannot leave the package | Each embedded tree has its own `embed.go` (`web`, `assets`, `testdata`); `cmd/app` imports them. | planner |
| R2-M3 | 16 automata counted; it is 15 | 15 runtime automata; DoD compares 13 printed rows + P3/P4/P5 ≡ P2, COMP excluded; Rules shows 17 rows, 15 with ER. | planner |
| R2-M4 | Envelope contradicts the endpoint contracts | Every contract now describes the `datos` payload explicitly, including `/api/pares`. | planner |
| R2-M5 | Test pairs SHOULD but required by the DoD | Promoted to MUST; moved to the end of the contingency order. | planner |
| R2-M6 | Rules screen has no slot in the run sheet | 30 s inside the first minute; RE2 column added to the screen. | planner |
| R2-M7 | One seed per pair cannot regenerate two credentials | Schema uses `seed_email` and `seed_password`; the 10 invalid pairs are authored; `cmd/pares` calls the handler in-process and asserts equality. | planner |
| R2-M8 | 30-state threshold excludes P2's 34-state NFA | Threshold is per view: drawn when <= 40 states and <= 80 edges, so P2's NFA, the 9-state subset, the 2-state tree, C5 and C*'s subset/tree are drawn. | planner |
| R2-m1 | 22-character mock labelled 20 | Sample shortened to 20 characters. | planner |
| R2-m2 | Sidebar counts unlabeled | Header "mín." added; panel shows NFA/subset/tree. | planner |
| R2-m3 | No request bodies specified | Bodies added to all three POST contracts. | planner |
| R2-m4 | `posicion` meaningless for C1 | Restricted to P0/C0; C1 reports the '@' count as a message. | planner |
| R2-m5 | `make check` mixes Go and Python | `make check` is Go-only; Python differential behind `make check-oracle`. | planner |
| R2-m6 | Video has no storyboard and no re-record slot | 5-block storyboard added, recorded Monday with a re-record slot. | planner |
| R2-m7 | Rules screen lacks the RE2 translation | `re2` field added to `/api/reglas` and a column to the screen. | planner |
| R2-m8 | Delivery channel unspecified | Assumption recorded: zip + repo link to the course platform, README with the file list; owner: group. | planner |
| R2-Q1 | Which DFAs does the generator multiply? | The minimized ones (65 × 2⁴ = 1040); asserted by a test. | planner |
| R2-Q2 | Two seeds per pair? | Yes: `seed_email` and `seed_password`, plus `regla_fallida`. | planner |
| R2-Q3 | Which MUST goes on day 1 if the gap is real? | Sunday-noon checkpoint: if the engine is not green, the Explorer drops to counts + table; no Python fallback, Go stays. | user |
| R2-Q4 | Is the 34-state NFA graph required? | Yes: the per-view threshold (40 states / 80 edges) keeps it. | planner |
| R2-Q5 | Where is the package uploaded? | Assumption: course platform, zip + repo link; README lists files. Open. | group |

### Review 3

Reference: `plan-review-3.md`. Verdict: PASS WITH CHANGES, 0 blockers.

| # | Finding | Resolution | Decided by |
|---|---|---|---|
| N-M1 | Coder rows sum to 4.55, not 4.45; Monday had no hours | Arithmetic corrected and the human schedule replaced by the actual execution: the code was written by the agent on 24 Oct; the remaining days are document, video and rehearsal. | planner |
| N-M2 | The Sunday-noon checkpoint could not be satisfied: its criterion depended on work scheduled after it | Checkpoint is now `make check && make check-oracle`, measurable in seconds and independent of later work. | planner |
| N-M3 | The DoD required exactly what the cuts remove | Every DoD line that a cut voids now names that cut. | planner |
| N-M4 | The Explorer cut was described two ways | Cut 1 is "drop the `tree`/`min` SVGs and `vista=min`, keep the route, sidebar, counts, tables and trace"; cut 3 (fold Rules into the sidebar) is only valid if cut 1 was not applied. | planner |
| N-M5 | Email generation had no target length or stop rule | Seeded target 12-32; stop on an accepting state at exactly that length, otherwise redraw (200 attempts); atoms weighted by size squared; documented quality filter. | planner |
| N-m1 | `docs/api.md` was referenced but missing | The file exists with the five contracts and the shared types. | planner |
| N-m2 | One `regla_fallida` for two credentials | Each invalid pair has exactly one defective credential; the DoD and `testdata` say so. | planner |
| N-m3 | `/api/reglas` lacked `tiene_er` and nullability | `tiene_er` plus nullable `er`/`re2`/`conteos` for P0/C0. | planner |
| N-m4 | C\* missing from the sidebar and over the edge limit | The sidebar lists C\*; its subset (33 states / 156 edges) is a table, its tree (12/46) is drawn. | planner |
| N-m5 | Grading summary added up to 15 | Final is 10 = 5 presentation + 5 deliverables; the advance 5 is separate and already earned. | planner |
| N-m6 | Dead-state convention unstated | Partial DFA, dead class removed; `P1_min` = 65; reachable product = 944. | planner |
| N-m7 | `toRE2` must escape class metacharacters | Test over all 94 symbols, plus a negative test proving an unescaped `]` does not compile. | planner |
| N-m8 | DoD counts and boundaries imprecise | 12 compared rows + P3/P4/P5 ≡ P2; 64/65 local, spaces and the 20-character blocklist case added. | planner |
| N-m9 | Video and run sheet had no buffer and no code slot | Run sheet targets 9:30 and includes the repo; video targets 55-57 s. | planner |
| N-m10 | "Automata built in milliseconds" was unverified | Measured: ~970 ms once at startup for all 15. Eager build kept; no lazy loading needed. | planner |
| N-m11 | Scope split off by 0.1 day | 4.55 coder + 0.7 document/video = 5.25. | planner |
