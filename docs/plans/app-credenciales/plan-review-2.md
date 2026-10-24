# Plan review 2

- Verdict: FAIL
- Blocker count: 1
- Major count: 8
- Minor count: 8

## Resolution check

| Finding | Status | Evidence in current plan |
|---|---|---|
| B1 | RESOLVED | L330: "Passwords walk the product `P1×P2×P3×P4×P5` (<=1040 states) with reachability; P6-P8 by rejection sampling, 200 retries, same seeded PRNG"; L269 property test "1000 generated values per kind are 100 % accepted". The 1040 bound is only reachable with minimized inputs (see R2-M1). |
| B2 | RESOLVED | L331: "10 valid pairs produced by `POST /api/generar` with recorded seeds ... regeneration test compares byte for byte"; L276 `semilla` in `pares.json`. Schema gap in R2-M7. |
| B3 | RESOLVED | L33: "the translation is generated from the AST (`toRE2`) and RE2 doubles as the independent test oracle"; L332: document section "Pasos 5 y 6" with the ER->RE2 table. |
| M1 | PARTIALLY RESOLVED | Roles assigned (L333) but no names ("names filled by `/flow:split`", L320) and the order is a serial 4-step list, not a day-by-day schedule; 4.45 coder-days against 3 days (see R2-B1). |
| M2 | RESOLVED | L334: "RE2 is the independent oracle; Python is kept for port fidelity only; `min ≡ tree` checked by equivalence and min counts against the document table". |
| M3 | RESOLVED | L89: "atoms and transitions are sorted slices, no map iteration in construction or walk, `rand.New(rand.NewSource(seed))` for every draw including retries"; L269: 100 runs and two restarts. |
| M4 | PARTIALLY RESOLVED | L336 enumerates `P1-P8, C1-C6, C*` (15 ids) and calls them "16 runtime automata"; no inventory table in the body; no startup assertion (R2-M3, R2-m-omitted). |
| M5 | RESOLVED | L212-217 Test pairs screen; L353 "two independent validations, shown side by side in the Test pairs table". |
| M6 | RESOLVED | L132 "every rule of the kind (9 for password, 7 for email) ... rendered as `—` with `evaluado: false`"; L338 `posicion` only P0/C0/C1. Envelope contradiction in R2-M4. |
| M7 | RESOLVED | L339: "`nivel` and `entropia` are `null` for emails and the badge is hidden"; L134 "for emails the level and entropy are hidden". |
| M8 | RESOLVED | L265-272 DoD with `make check` "single PASS/FAIL line with discrepancy counts"; two of its criteria are miscounted (R2-M3, R2-M5). |
| M9 | RESOLVED | L304-315 run sheet with minute allocations and a fallback. Rules screen missing from it (R2-M6). |
| M10 | RESOLVED | L204: "Promoted from SHOULD to MUST because it is demo priority 5 and the only screen that shows the ER and its standard basis". |
| m1 | RESOLVED | L253 "~5.3 person-days"; rows sum to 5.25 (0.8+0.1+0.3+0.2+0.7+0.3+0.4+0.2+0.5+0.2+0.15+0.3+0.2+0.5+0.2+0.1+0.1). |
| m2 | RESOLVED | L175: "AFND-ε 34 · Subconjuntos 9 · Árbol 2". |
| m3 | RESOLVED | L131: "States: empty (form + hint) \| result (verdict + table) \| error"; no spinner, no offline. |
| m4 | RESOLVED | L199: "`{svg\|null, transiciones[], total, pagina}`" with `pagina`/`limite` parameters. |
| m5 | PARTIALLY RESOLVED | L90 picks the envelope, but L138, L166, L198, L208 and L216 still show bare payloads (R2-M4). |
| m6 | RESOLVED | L134: "Entropy is displayed next to the level for passwords". |
| m7 | RESOLVED | L14: "to be tested on a real machine during rehearsal"; L20: ".exe is the tested backup". |
| m8 | REJECTED WITH EVIDENCE (claim verified) | L350. I reproduced it against the shipped engine: `NFA(P_MAYUS)` gives `n=34, start=0, accept=19`; `Star` allocates 18/19 before its child 20-33. The document's "final 19" is correct. |
| m9 | RESOLVED | L251 "Repo + README + run instructions \| MUST \| 0.1"; L70 layout; L272 DoD. |
| m10 | RESOLVED | L130: "'faltan N' below 15 and 'sobran N' above 64". |
| Q1 | RESOLVED | L353. |
| Q2 | RESOLVED | L354. |
| Q3 | RESOLVED | L355 + L336 + L299 (COMP document-only). |
| Q4 | RESOLVED | L356. |
| Q5 | RESOLVED | L357. |
| Q6 | RESOLVED | L358. |

## New findings

| ID | Severity | Area | Finding | Evidence | Suggested fix |
|---|---|---|---|---|---|
| R2-B1 | blocker | B (deadline) | The stated plan order is serial and 5.15 days long against a 3-day window. Capacity is not the constraint (5.25 of 9 available person-days); the dependency chain is, and the plan never states a parallel schedule or an hours/day assumption. | L253 "Total ~5.3 person-days, of which ~4.5 fall on the coder"; L333 "day-by-day order: engine + `docs/api.md` first, then rules/validator/generator, then screens, then pairs/doc/video"; L315 "The video is recorded on day 3 with the final build"; coder rows sum to 4.45 d, doc+video 0.7 d. | Rewrite M1 as a real schedule: D1 coder engine + `api.md` frozen by noon while the doc people write "Pasos 5 y 6" and the test-case tables from the existing 35-page document (no app needed); D2 screens + figures; D3 pairs from the frozen build, screenshots, video, then the Windows rehearsal. State ~10-12 h/day for the coder and pick the day-1 cut instead of "see how far we get". |
| R2-M1 | major | C (generator) | "<=1040 states" holds only if the product is built from minimized (or tree) DFAs. The port's `minimize` is the Python one, which returns a *count*, so the natural reading — product of subset DFAs — is 257 × 9⁴ = 1 686 177 states, not 1040. | L330 "(<=1040 states)"; L235 "minimize" in the port scope; `motor_er.py` `minimize()`: "return len({part[q] for q in range(d.n + 1)}) - 1  # sin la clase del estado muerto"; measured P1 sub=257, P2-P5 sub=9. | State that the product is built over the *minimized* DFAs (P1 65 × 2⁴ = 1040), which requires `minimize` to return a DFA (classes + delta), not a count; add a test asserting the reachable product ≤ 1040. |
| R2-M2 | major | C (build) | `cmd/app` cannot embed `web/`, `assets/` or `testdata/`: `//go:embed` patterns may not contain `..` and cannot reach outside the package directory, so the layout as drawn does not compile. | L68 "cmd/app/ HTTP server, flags, embed of web/ and assets/"; L77-78 "assets/figures/ pre-rendered SVG (embedded)", "testdata/ pares.json"; L216 "`GET /api/pares` -> embedded `pares.json`". | Put the directives next to the files (`web/embed.go` package web, `assets/embed.go` package assets) and import them from `cmd/app`, or move `web/`, `assets/` and `testdata/` under `cmd/app/`. |
| R2-M3 | major | D (inventory) | The runtime has 15 automata, not 16: P1-P8 (8) + C1-C6 (6) + C* (1). The 16th row of the document table is COMP, which the plan declares document-only, yet the DoD demands the COMP sizes. The Rules screen lists 17 ids, of which 15 have an ER, so "16 rows with ER" is wrong on both counts. | L336 "16 runtime automata (P1-P8, C1-C6, C*)"; L237 "16 automata in Go"; L268 "the AFD sizes match the document table for all 16 automata"; L271 "Rules lists 16 rows with ER"; document table has 13 printed rows (P1, P2, P6, P7, P8, COMP, C1-C6, CORREO) plus "P3, P4 y P5 ... igual que P2". | Use 15 everywhere; make the DoD "13 table rows + P3/P4/P5 ≡ P2, COMP excluded by design"; Rules = 17 rows, 15 with ER (P0/C0 show "conjunto"). |
| R2-M4 | major | D (contract) | The new envelope contradicts all five endpoint contracts in the Screens section, and `/api/pares` returns a bare embedded array. | L90 "`{"ok": true, "datos": ...}`"; L138 "`POST /api/validar` -> `{valido, normalizado, ...}`"; L166 "`{valor, nivel\|null, ...}`"; L198 "`[{id,nombre,tipo,er,conteos}]`"; L208 same payload; L216 "embedded `pares.json`". | Write each contract as the `datos` payload ("`datos: {valido, ...}`") or drop the envelope; fix `/api/pares` explicitly. |
| R2-M5 | major | D (consistency) | The Test pairs screen is SHOULD in the scope table, required by the DoD, and cut by contingency item 2 — applying the plan's own fallback makes the DoD unreachable. | L245 "Test pairs screen \| SHOULD \| 0.15"; L271 "Screen acceptance: ... Pairs lists 20 rows"; L260 "2. Test pairs screen (the document table still covers the deliverable)". | Promote it to MUST (0.15 d, and it is the live evidence for the graded 20 pairs) or delete it from the DoD. |
| R2-M6 | major | F (demo) | The Rules screen was promoted to MUST as "demo priority 5" but has no slot in the run sheet; neither has the ER→RE2 evidence for steps 5-6. | L204 "Promoted from SHOULD to MUST because it is demo priority 5 and the only screen that shows the ER and its standard basis"; L244 MUST; L306-313 has no Rules or RE2 row. | Give it 30 s inside 0:00-1:00 or 9:00-10:00, and name the screen that shows the RE2 translation (see R2-m7). |
| R2-M7 | major | E (deliverable testability) | `pares.json` carries one `semilla` per pair, but a valid pair is two generated credentials with two seeds, so the "byte for byte" regeneration test cannot know which seed produced which value. Separately, `cmd/pares` "generates the 20 pairs from seeds" while the 10 invalid pairs must be authored, and B2 claims the API produced the valid ones. | L276 "`id`, `email`, `password`, `email_esperado`, `password_esperado`, `regla_fallida`, `semilla`"; L331 "stores `semilla`, `valor`, `regla_esperada`"; L277 "`2026` (password, length 20), `4490` (email)"; L70 "cmd/pares/ generates the 20 pairs from seeds". | Schema: `semilla_password`, `semilla_email`, `regla_esperada` per credential; state that the 10 invalid pairs are authored and only re-validated; state that `cmd/pares` calls the running API (or shares the handler and asserts equality). |
| R2-M8 | major | C/D (explorer) | The "~30 states" threshold excludes P2's 34-state NFA — i.e. every password NFA (34, 146, 630, 782) — contradicting the mock and making contingency item 1 delete the demo's centrepiece. | L191 "graphs only for automata up to ~30 states; above that the endpoint returns `svg: null`"; L175 mock draws a graph for P2 while listing "AFND-ε 34"; L259 "Explorer `tree`/`min` SVGs and figures for automata >30 states"; the document already prints a 34-state Figure 3. | Base the threshold on drawn edges (or 40 states) so P2-P5 NFA, C5 and C* keep a graph, and say per view which automata are drawn. |
| R2-m1 | minor | D | The Generate mock is 22 characters but is labelled "20 caracteres · 131 bits" (131 bits is the 20-char figure). | L147 "Xk9$mTq2#vLw8@RzQe4&Nb"; L148 "FUERTE · 20 caracteres · 131 bits"; the document's own example is "Xk9$mTq2#vLw8@Rz (16) es Media". | Shorten the sample to 20 chars or change the labels. |
| R2-m2 | minor | D | The explorer sidebar numbers are unlabeled: 65/2/190/158/44 are minimal-DFA counts, while the panel shows NFA/subset/tree. | L176-183 "P1 longitud 65", "P2 mayúscula 2", "P6 repetición 190" vs L175 "AFND-ε 34 · Subconjuntos 9 · Árbol 2". | Add a column header ("estados mín.") or render "34/9/2". |
| R2-m3 | minor | D | No POST request body is specified anywhere, and `/api/simular` does not say which automaton the trace comes from, while the mock shows the 2-state tree trace. | L200 "`POST /api/simular` -> `{acepta, traza:[{paso,estado,simbolo}], fallo}`"; L185 "q0─L→q1─l→q1─N→q1 ✔". | Document request shapes (`{valor,tipo}`, `{tipo,longitud,semilla}`, `{id,valor,vista}`). |
| R2-m4 | minor | D | `posicion` has no meaning for C1: the prototype's C1 failure is an "@ count", not a position. | L338 "`posicion` only for P0/C0/C1"; `validador.py`: "f\"{s.count('@')} arrobas\""; L50 "position (only P0/C0/C1)". | Restrict `posicion` to P0/C0, or define C1's position as the offending '@'. |
| R2-m5 | minor | E | `make check` mixes the pure-Go RE2 oracle with the Python port-fidelity oracle without saying whether Python is required, so an outsider may not be able to run it. | L267 `make check` list; L334 "Python is kept for port fidelity only"; ADR 3 (L97). | State "`make check` needs only Go; the Python differential is behind `make check-oracle`". |
| R2-m6 | minor | F | The video has no storyboard, and it is recorded "day 3 with the final build" — the same day the build must become final. | L249 "1-minute video \| MUST \| 0.2"; L315 "recorded on day 3 with the final build". | 5-line script (Validate a pair, Generate seed 2026, Explorer P2) and a re-record slot on day 3. |
| R2-m7 | minor | A/G | The Rules screen shows the formal ER but not the RE2 translation, so the step-6 artefact exists only in the document even though `toRE2` is already in the code. | L204 "columns id, rule, accept/reject, formal ER, foundation"; L33 "the translation is generated from the AST (`toRE2`)". | Add an `re2` field to `/api/reglas` and a column or copy button. |
| R2-m8 | minor | A | The delivery channel is unspecified: "repo + README" says nothing about how the professor receives the document, the video and the source. | L251 "Repo + README + run instructions"; L272 "document updated ... video <= 1 minute; both binaries built". | Name the channel (repo URL / zip / LMS upload) and the exact file list in the README. |

## Scope cut line

1. **Rules screen as a separate route → fold the table into the explorer sidebar** (the plan's #4). Saves the route plus a screen of JS for the least grading loss: the ER + standard table survives, and the screen has no run-sheet slot anyway.
2. **Explorer `tree`/`min` SVGs** (the plan's #1), but keep the P2 NFA and subset graphs and the 34/9/2 counts — that is what the 6:30-9:00 block actually shows.
3. **Email generation toggle in the UI** (the plan's #3). "Generar una contraseña **o** correo" — one interactive path is enough; the API and the email pairs stay.
4. **Test pairs screen** (the plan's #2) — last, not second. It is the only live surface for the graded 20 pairs and it owns the final minute of the run sheet; if cut, 9:00-10:00 must become the document table plus a screenshot.
5. **`/api/simular` trace and the `vista=min` view** (the plan's #5). The Validate report and the counts already carry the story.

I agree with the *set* of items, not with the *order*. The plan puts the explorer SVGs first, which — because P2's NFA is 34 states — deletes the one graph the middle of the demo is built on, while the Rules fold saves more for less. More importantly, no permutation of these five recovers more than ~0.6 d of the ~1.45 d gap on the coder, so the cut order is a polish lever, not a deadline rescue. If the gap is real it must be decided on day 1 (drop the Explorer as a whole, ~0.5-0.7 d, and lean on the document's Figures 3-6), not on day 3.

## Questions for the plan owner

1. Which DFAs does the generator multiply — the minimized ones (65 × 2⁴ = 1040) or the subset ones (257 × 9⁴ = 1.69 M)? Has anyone actually built the 1040-state product, or is it an arithmetic guess?
2. Does `pares.json` get two seeds per pair (email and password)? Which field proves that a valid pair came from `POST /api/generar` rather than from the annex?
3. If the coder's 4.45 days do not fit, which MUST goes on day 1 — the Explorer, or the Go port itself (i.e. shipping the Python prototype behind a thin HTTP server)? The contingency order cannot close the gap.
4. Is the P2 NFA graph (34 states) required in the explorer, or is the 9-state subset graph enough for 4.d.ii? That single answer decides whether the ~30-state threshold survives.
5. Where exactly is the package uploaded (repo URL, zip, LMS) and what is the definitive file list the professor receives?

## What the plan gets right

- The m8 rejection is factually correct, and I reproduced it: the shipped engine gives P2 `n=34, start=0, accept=19`, and the document's "inicio 0, final 19" matches. The full size table also matches the engine digit for digit (P1 1092/257/65/65, P2 34/9/2, P6 782/565/377/190, P7 630/763/314/158, C* 136/33/12/8, COMP 2158/48137/12034/16), so the "verified" claim in the Research section is real and not asserted.
- The generator design is now specific enough to implement and is the correct reading of 7.a.iii: a seeded reachability walk over the product for P1-P5, bounded rejection sampling for P6-P8, plus a 1000-per-kind property test.
- Making the 10 valid pairs seed-generated with a byte-for-byte regeneration test turns a document deliverable into an executable one — exactly what "generadas por la herramienta" asks for, and the one thing a professor can check without reading Go.
- The Definition of done is runnable by an outsider: one `make check`, a PASS/FAIL line, 1500 RE2 strings per rule, `min ≡ tree`, boundary values 14/15/16, 19/20, 64/65, 63/64, 255/256.
- Scope discipline is honest: the contingency order exists, the gaps table names the real risks (basic-intermediate Go, no Windows machine, professor's exclusions) instead of hiding them, and the 24-permutation automaton is correctly kept out of the runtime (measured 59 s / 48 137 subset states in Python).
