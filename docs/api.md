# API contract

Frozen 2026-10-24 18:00. Changing a payload requires updating this file, the handler and
the frontend in the same commit.

## Envelope

Every response is one of:

```json
{ "ok": true,  "datos": <payload> }
{ "ok": false, "error": "mensaje en español" }
```

Validation outcomes (`valido: false`, `nivel: "Baja"`, a rule that fails) are **HTTP 200**:
they are answers, not errors. `4xx` is used for a malformed request (unknown `tipo`, missing
`valor`, unknown rule id, out-of-range `longitud`) and `5xx` for a server fault.

## Shared types

```jsonc
// Rule metadata
{
  "id": "P2",            // frozen id: P0..P8, C0..C6, C*
  "nombre": "Mayúscula",
  "tipo": "accept",      // "accept" | "reject"
  "tiene_er": true,      // false only for P0 and C0 (alphabet set checks)
  "er": "(L)(L|l|N|S)*", // null when tiene_er = false
  "re2": "^[A-Z].*$",    // null when tiene_er = false
  "conteos": { "nfa": 34, "sub": 9, "tree": 2, "min": 2 } // null when tiene_er = false
}

// Per-rule result
{
  "id": "P1",
  "nombre": "Longitud",
  "tipo": "accept",
  "ok": false,
  "evaluado": true,      // false when a short-circuit skipped the rule
  "mensaje": "longitud = 12 (faltan 3)",
  "posicion": null       // 0-based index, only for P0 and C0
}
```

`evaluado: false` means the rule was not checked because an earlier rule already decided
the verdict. The array always contains **every** rule of the requested kind: 9 for
`password` (P0-P8) and 7 for `email` (C0-C6).

## POST /api/validar

```json
{ "tipo": "password", "valor": "Xk9$mTq2#vLw8@RzQe4&" }
```

`datos`:

```jsonc
{
  "valido": true,
  "normalizado": "Xk9$mTq2#vLw8@RzQe4&", // spaces stripped
  "longitud": 20,
  "nivel": "Fuerte",     // "Baja" | "Media" | "Fuerte"; null for email
  "entropia": 131.0,     // bits, bruto; null for email
  "reglas": [ /* RuleResult[], 9 or 7 entries */ ]
}
```

Security level (Annex C, passwords only): `Baja` if any acceptance rule fails or any
rejection filter matches; `Media` for 15-19 characters; `Fuerte` for 20 or more.

## POST /api/generar

```json
{ "tipo": "password", "longitud": 20, "semilla": 2026 }
```

`longitud` is 15-64 for passwords and ignored for emails. `semilla` is optional: when
absent a random seed is drawn and returned so the generation can be replayed.

`datos`:

```jsonc
{
  "valor": "Xk9$mTq2#vLw8@RzQe4&",
  "tipo": "password",
  "longitud": 20,
  "nivel": "Fuerte",     // null for email
  "entropia": 131.0,     // null for email
  "semilla": 2026,
  "camino": [            // the walk over the automaton
    { "paso": 0, "estado": "q0", "simbolo": "X", "clase": "L" }
  ]
}
```

Same `semilla` plus same `tipo` and `longitud` returns the same `valor` and the same
`camino`, across process restarts.

## GET /api/reglas

`datos`: `Rule[]` (17 rows, in id order: P0-P8, C0-C6, C*). P0 and C0 have
`tiene_er: false` and null `er`/`re2`/`conteos`.

## GET /api/automatas/{id}

Query: `vista=nfa|sub|tree|min` (default `sub`), `pagina` (1-based, default 1),
`limite` (default 50, max 200).

`datos`:

```jsonc
{
  "id": "P2",
  "vista": "nfa",
  "svg": "<svg…>",        // null when the view exceeds the drawing threshold
  "motivo": null,         // string when svg is null, e.g. "782 estados"
  "transiciones": [       // the requested page
    { "desde": "q0", "simbolo": "L", "hacia": "q1" }
  ],
  "total": 43,            // total transitions, not states
  "pagina": 1,
  "limite": 50
}
```

A view is drawn when it has **<= 40 states and <= 80 edges**; otherwise `svg` is `null`
and the UI shows the counts plus the paginated table.

## POST /api/simular

```json
{ "id": "P2", "valor": "Aa1!", "vista": "sub" }
```

`datos`:

```jsonc
{
  "acepta": true,
  "traza": [
    { "paso": 0, "estado": "q0", "simbolo": "A" },
    { "paso": 1, "estado": "q1", "simbolo": "a" }
  ],
  "fallo": null           // 0-based index of the first symbol that leaves the trace, or null
}
```

## GET /api/pares

`datos`: the 20 entries of `testdata/pares.json`.

```jsonc
{
  "id": 1,
  "email": "ana.perez@correo.com",
  "password": "Xk9$mTq2#vLw8@RzQe4&",
  "email_esperado": true,
  "password_esperado": true,
  "regla_fallida": null,        // only the invalid pairs carry one
  "seed_email": 4490,           // null for authored pairs
  "seed_password": 2026
}
```

The 10 valid pairs are produced by `POST /api/generar` with the recorded seeds; the 10
invalid pairs are authored by hand and each has exactly one invalid credential, named by
`regla_fallida`.

## GET /api/salud

`datos`: `{"estado": "ok"}`. Used by the smoke test and by the rehearsal.
