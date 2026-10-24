# Plan review 3

- Verdict: PASS WITH CHANGES
- Blocker count: 0
- Major count: 5
- Minor count: 11

## Resolution check

| Finding | Status | Evidence in current plan |
|---|---|---|
| R2-B1 | RESOLVED | L103 "the working assumption is 10-12 h/day for the coder on Saturday and Sunday"; L107-111 parallel day-by-day rows; L113 "Checkpoint rule ... drop the Explorer ... Do not wait until Monday." The residual arithmetic is N-M1. |
| R2-M1 | RESOLVED | L99 "`P1_min` 65 × `P2_min..P5_min` 2⁴ = 1040"; L380 "`minimize` returns a DFA (classes + delta), not a count; test asserts the reachable product <= 1040." Verified against the shipped engine: `minimize` gives 65, 2, 2, 2, 2; the reachable product is 944 states. Dead-state wording is N-m6. |
| R2-M2 | RESOLVED | L65 "each embedded tree has its own `embed.go`, because `//go:embed` cannot reference parent directories"; L381. Verified: a minimal module with `web/embed.go` and `testdata/embed.go` imported from `cmd/app` builds with `go build ./cmd/app`. Caveat: `go list ./...` omits `testdata`, so `go vet ./...`/`go test ./...` skip that package. |
| R2-M3 | RESOLVED | L382 "15 runtime automata; DoD compares 13 printed rows + P3/P4/P5 ≡ P2, COMP excluded; Rules shows 17 rows, 15 with ER"; L218, L284. |
| R2-M4 | RESOLVED | L90 envelope; L152, L180, L212-214 and L230 all say `datos: ...`, including `/api/pares`. |
| R2-M5 | PARTIALLY RESOLVED | L384 "Promoted to MUST; moved to the end of the contingency order", but L277 still cuts it and L287 still requires it. See N-M3. |
| R2-M6 | RESOLVED | L326 "a 30-second pass through the Rules screen with the ER and its RE2 translation"; L218 RE2 column; L394. |
| R2-M7 | RESOLVED | L292-294 `seed_email`/`seed_password`, invalid pairs "authored by hand", `cmd/pares` in-process; L386. Single `regla_fallida` is N-m2. |
| R2-M8 | RESOLVED | L205 "<= 40 states and <= 80 edges"; measured P2 NFA 34 states/43 transitions, subset 9/36, tree 2/8, C5 26/32. C\* subset is 33/156, not drawn; N-m4. |
| R2-m1 | RESOLVED | L161 "Xk9$mTq2#vLw8@RzQe4&" is 20 characters, labelled "20 caracteres · 131 bits". |
| R2-m2 | RESOLVED | L206 "the sidebar shows the **minimal** state count with a header ('mín.')"; L189-197. |
| R2-m3 | RESOLVED | Request bodies on L152, L180 and L214. |
| R2-m4 | RESOLVED | L50 "position (only P0/C0)"; L391. |
| R2-m5 | RESOLVED | L91 "`make check` runs the Go-only checks; `make check-oracle` adds the Python differential (optional)"; L392. |
| R2-m6 | RESOLVED | L335 five-block storyboard with "a re-record slot before the presentation"; L393. |
| R2-m7 | RESOLVED | L218 "**RE2 translation**" column; L394. |
| R2-m8 | RESOLVED | L307 "zip + repo link uploaded to the course platform; README lists every file"; L395. |
| R2-Q1 | RESOLVED | L396 "the minimized ones (65 × 2⁴ = 1040); asserted by a test"; reachable product measured at 944. |
| R2-Q2 | RESOLVED | L397 `seed_email` and `seed_password` plus `regla_fallida`. |
| R2-Q3 | RESOLVED | L398 "Sunday-noon checkpoint: if the engine is not green, the Explorer drops to counts + table; no Python fallback, Go stays." |
| R2-Q4 | RESOLVED | L399 "per-view threshold (40 states / 80 edges) keeps it"; measured 43 transitions for P2's NFA. |
| R2-Q5 | RESOLVED | L400 "Assumption: course platform, zip + repo link; README lists files. Open." Answered as an assumption with an owner. |

## New findings

| ID | Severity | Area | Finding | Evidence | Suggested fix |
|---|---|---|---|---|---|
| N-M1 | major | B (deadline) | The coder rows sum to **4.55 d, not 4.45** (0.8+0.1+0.3+0.2+0.7+0.3+0.4+0.2+0.5+0.2+0.15+0.3+0.2+0.1+0.1), i.e. 36.4 h at 8 h/day. Sat+Sun+Mon at 10-12 h gives 30-36 h, so the full scope is short even at 12 h/day, and 0.7 d short at 10 h/day. Sunday alone carries 2.1 d (rules 0.3 + validator 0.2 + generator 0.7 + API 0.3 + Validate 0.4 + Generate 0.2) against 1.25-1.5 d available. | L103 "10-12 h/day ... Saturday and Sunday"; L107-111 rows; L267 "~4.45 on the coder"; L279 "No permutation of these five recovers the full 1.45-day gap." | State Monday's hours, move 0.3-0.6 d from Sunday to Saturday, and make the Explorer cut mandatory (not conditional) if the coder works 10 h/day; or move the Rules/Test-pairs rows to Tuesday 27, which is currently unscheduled buffer. |
| N-M2 | major | B/D (checkpoint) | The Sunday-noon checkpoint cannot be evaluated as written. Its pass criterion is "all 15 automata + RE2 oracle at 0 discrepancies", but the 15 automata and the oracle tests are scheduled **after** the checkpoint (Sunday row); the Saturday row only lists the five algorithms and `toRE2`. At noon the criterion is false by construction, so the day-1 cut always fires. | L108 "Checkpoint: engine green (all 15 automata + RE2 oracle at 0 discrepancies)"; L107 "Engine: AST, Thompson NFA, subset, tree, `minimize` ..., equivalence, `toRE2`"; L109 "Rules, validator, generator, API, Validate and Generate screens". | Checkpoint on the five algorithms + P1/P2 counts + the RE2 oracle for P1-P5, or move the checkpoint to Sunday 18:00 and the 15 automata to Saturday. |
| N-M3 | major | D (DoD vs cuts) | The DoD and the contingency order contradict each other. Cuts 1, 2, 4 and 5 remove exactly what the DoD's screen-acceptance line requires: Rules 17 rows, Explorer tree/min views, `simula Aa1!`, Pairs 20 rows. Applying the plan's own fallback makes the DoD red. | L273-277 cuts; L287 "Explorer shows 34/9/2 for P2 and simulates `Aa1!`; Rules lists 17 rows (15 with ER); Pairs lists 20 rows." | Annotate each DoD line with the cut that voids it (or move those lines into a "full scope" subsection) so the post-cut DoD is explicit. |
| N-M4 | major | D (day-1 cut) | The day-1 cut is described two ways: "drop the Explorer as a whole" and "(counts + transition table only)". If the Explorer route is dropped, contingency 1 (fold the Rules table into the explorer sidebar) has no sidebar, so the Rules screen (MUST, 0.2 d) must stay and the 0.5-0.7 d saving shrinks. | L113 "drop the Explorer as a whole (counts + transition table only, ~0.5-0.7 d recovered)"; L273 "fold the table into the explorer sidebar". | Define the cut as "drop the graphs and the `min` view, keep the Explorer route, sidebar, counts, table and trace", or "drop the Explorer route and keep Rules as a route". |
| N-M5 | major | C (email generation) | Email generation has no target length or stop rule. Passwords get a user-set target 15-64; emails "walk `C*`" with "email length is not user-set", so nothing says when the walk stops or how the 10 email pairs get their lengths. The 1000-email property test cannot be written without this. | L174 "email length is not user-set"; L350 "Emails walk `C*` and reject-sample C2-C6"; ADR 0005 same wording. | Define a seeded target length (e.g. 24-40) and the stop rule (accept at the target, else redraw), and add it to the property test. |
| N-m1 | minor | D (contracts) | `docs/api.md` is called the "frozen contract" by `AGENTS.md` and `docs/architecture.md`, but the file does not exist. The only contracts are in the plan's Screens tables, and the plan freezes the file Saturday 18:00; an outsider reading the repo today finds a dangling reference. | `docs/api.md` missing; L341 "Frozen contracts: `docs/api.md` (frozen Saturday 18:00)"; architecture.md "Endpoint payloads live in `docs/api.md`". | Create the file now with the five payloads, or change the references to the plan until Saturday. |
| N-m2 | minor | E (pairs schema) | `pares.json` has one `regla_fallida` for a pair of two credentials, and the DoD says "Validate names the failing rule for each of the 20 pairs". Valid pairs have no failing rule, and an invalid pair with two invalid credentials cannot be described by one field. | L292 schema; L287 "for each of the 20 pairs". | State that each invalid pair has exactly one invalid credential, or split `regla_fallida_email`/`regla_fallida_password`; reword the DoD to "matches the expected verdict for all 20 and names the rule for the 10 invalid". |
| N-m3 | minor | D (API) | `GET /api/reglas` has no `tiene_er`, and `er`/`re2`/`conteos` for P0/C0 are not defined as nullable, although the Rules screen needs "conjunto" for P0/C0 and the Explorer has no automaton for them. | L212 "`{id,nombre,tipo,er,re2,conteos:{nfa,sub,tree,min}}`"; L218 "P0 and C0 show 'conjunto'". | Add `tiene_er` and mark `er`/`re2`/`conteos` nullable. |
| N-m4 | minor | C/D (explorer) | The explorer's C\* is inconsistent: the sidebar mock lists "C1 … C6" with no C\*, and the plan claims C\*'s subset is drawn, but the C\* subset DFA has 33 states and 156 transitions, over the 80-edge limit. | L199 sidebar; L205 "C5 and C\*'s subset/tree"; measured C\* subset 33/156, tree 12/46. | Add C\* to the sidebar and say "C\* tree only; the subset shows a table". |
| N-m5 | minor | A (grading) | The grading summary is wrong: "5 pts advance, 5 pts presentation, 5 pts document..." sums to 15, but the PDF's final is 10 (5 presentation + 5 deliverables) and the 5-point advance is a separate, already-delivered grade. | L26; PDF "VALOR 5 PUNTOS – APLICA A ZONA (AVANCE)" and "VALOR 10 PUNTOS – APLICA A EVALUACIÓN FINAL". | State "final 10 = 5 presentation + 5 deliverables; the advance 5 is already earned". |
| N-m6 | minor | C (minimize) | The 65 and 1040 figures depend on an unstated dead-state convention. The shipped `minimize` returns 65 by subtracting the dead class ("sin la clase del estado muerto"); ADR 0005 says "minimize must return a full DFA". With the dead state P1 is 66 and the Cartesian product is 1056, so the plan's arithmetic and the ADR disagree. | L99 "`P1_min` 65"; L380; ADR 0005 "full DFA"; `motor_er.py` `minimize`. | State "partial DFA, dead class removed; missing transition = dead; P1_min = 65; reachable product = 944". |
| N-m7 | minor | C (`toRE2`) | `toRE2` must escape `\`, `]`, `^` and `-` inside its character classes. The alphabet includes all four, and a naive `[...]` breaks the RE2 oracle or silently changes the language. The 0.1 d row does not mention escaping. | L33 "the translation is generated from the AST (`toRE2`)"; `reglas.py` `PRINT` = 0x21-0x7E. | Add an escaping test over all 94 symbols and a differential check of `toRE2` vs the DFA on one string per atom. |
| N-m8 | minor | E (DoD precision) | The DoD's counts and boundaries are imprecise: "13 printed rows ... COMP excluded by design" — the 13 includes COMP, so the comparable set is 12 + P3/P4/P5; the boundary list omits the 64/65-character local part and a spaces case, and the Annex C edge case (20 chars with a P8 word → Baja) is not tested. | L284 "13 printed rows; ... COMP excluded"; L286 boundaries; document table has 13 rows including COMP; `validador.py` strips spaces. | Say "12 compared rows + P3/P4/P5 ≡ P2", add 64/65 local, a spaces case and the 20-char blocklist case. |
| N-m9 | minor | F (video/demo) | The video is exactly 60 s and the run sheet exactly 10:00, with no buffer; the assignment says "máximo de 1 minuto" and "10 minutos". The run sheet also has no slot for the code/repo, which is a graded deliverable. | L335 "0-10 ... 55-60 s"; L322-331 10:00; PDF "máximo de 1 minuto" / "10 minutos". | Target 55-57 s, 9:30 for the demo, and 10 s for "the repo has the Go port, the tests and the 20 pairs". |
| N-m10 | minor | C (startup) | The "automata built in milliseconds" claim is unverified and the build policy is unspecified: 15 automata × 4 views eagerly is roughly 500k transitions (C4 subset 2616 × 83, C6 5606 × 83, P6/P7 ~50k each), which may be 0.3-1 s and tens of MB. Lazy per-view building would make startup instant. | L32 "the Python build takes 9 s (70 s with the 24-permutation rule)"; L63 "automata built in milliseconds instead of seconds"; measured C4/C6 sizes. | State eager vs lazy and add a startup assertion (e.g. < 500 ms) to `make check`. |
| N-m11 | minor | D (scope split) | The scope split is off by 0.1 d: the coder rows sum to 4.55, not 4.45; the document/video rows are 0.7, not 0.8. The total 5.25 is right, the split is not, and the schedule's feasibility arithmetic inherits the error. | L267 "~4.45 on the coder, ~0.8 on the two document/video people"; rows L249-265. | Write "4.55 coder + 0.7 doc/video = 5.25". |

## Scope cut line

1. **Rules screen as a separate route → fold the table into the Explorer sidebar** (the plan's #1). Agree: 0.2 d, zero content loss (the ER, RE2 and standard columns survive). Apply it today, not at the checkpoint — the gap is already known.
2. **Email generation toggle in the UI** (the plan's #3). Keep `POST /api/generar` for emails and produce the 10 email pairs with `cmd/pares`. 0.1-0.2 d, zero grading loss: the assignment says "una contraseña **o** correo". This should come before the Explorer SVGs.
3. **Explorer `tree`/`min` SVGs and `vista=min`** (the plan's #2 plus part of #4). Keep the P2 NFA/subset graphs (34/9), the counts and the tables; the document already has Figures 4/4a. 0.2-0.3 d.
4. **`/api/simular` trace** (the plan's #4). The Validate report and the counts carry the story, but then amend the DoD's "simulates `Aa1!`" line.
5. **Test pairs screen** (the plan's #5), last as the plan says. It owns the final minute and is the live evidence for the 20 graded pairs; if cut, the document table plus a screenshot.

I agree with the set and with keeping the Test pairs screen last. I disagree with two things: the email toggle should precede the Explorer SVGs (cheaper, zero loss), and item 1 should not wait for the checkpoint. I also note that the 0.1 d Windows cross-compile has no grading value and should go before item 5 if the deadline is tight. No permutation of these five recovers the full gap; the cut line is a polish lever, not the rescue.

## Questions for the plan owner

1. Which day is the group's 10-minute slot (26, 27 or 28)? The plan assumes 28 and schedules 1.55 coder-days plus the video on Monday 26, the first day of the window.
2. Is Monday 26 a full 10-12 h coder day? The stated assumption covers only Saturday and Sunday; the coder rows are 4.55 d and the three-day window at 10-12 h gives 30-36 h (36.4 h needed).
3. For emails, what is the target length and stop rule of the `C*` walk? The plan says the email length is not user-set and never defines when the walk stops or how the 10 email pairs get their lengths.
4. Does `minimize` return the dead state? The 65 and 1040 figures assume it is excluded (the shipped engine subtracts it); a complete DFA gives 66 and 1056.
5. Do the 10 invalid pairs each have exactly one invalid credential? `pares.json` has a single `regla_fallida` for two credentials, and the DoD says "each of the 20 pairs".

## What the plan gets right

- The generator is now concrete and implementable: I rebuilt the minimized DFAs from the shipped engine and the reachable product is 944 states (P1 65, P2-P5 2 each), so the 1040 bound is real, not an arithmetic guess.
- The embed layout is legal: a minimal module with `web/embed.go` and `testdata/embed.go` imported from `cmd/app` builds; the `..` problem is gone.
- The state table is accurate: the shipped engine reproduces the document's P1 1092/257/65, P2 34/9/2, P6 782/565/377/190, P7 630/763/314/158 and C\* 136/33/12/8; the "verified" claim is real.
- The 40-state/80-edge threshold keeps P2's 34-state NFA (43 transitions), so the demo's 34 → 9 → 2 centrepiece survives the cut.
- The deliverables are executable, not prose: 10 valid pairs with recorded seeds and a byte-for-byte regeneration test, 10 authored invalid pairs covering length/special/repetition/blocklist, a 60 s storyboard, a 10-minute run sheet and a fallback.
