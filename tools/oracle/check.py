#!/usr/bin/env python3
"""Differential test between the Go engine and the Python prototype.

This is the port-fidelity check: it does NOT test the Go engine against itself.
It builds the automata with codigo_avance/ (the prototype that produced the
advance document), generates a deterministic corpus, asks the Go binary for its
verdicts, and compares:

  1. the four state counts of every rule (NFA, subset, tree, minimum);
  2. the verdict of both engines on the same strings.

Optional: `make check` never needs Python. Run it with `make check-oracle`.
Exit code 0 means no discrepancy.
"""

from __future__ import annotations

import json
import random
import string
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "codigo_avance"))

from motor_er import Compiled  # noqa: E402
import reglas as R  # noqa: E402

# The 15 rules with an automaton. P* (24 permutations) is document-only and is
# deliberately excluded: building it takes ~59 s and nothing depends on it.
REGLAS = {
    "P1": R.P_LONG,
    "P2": R.P_MAYUS,
    "P3": R.P_MINUS,
    "P4": R.P_DIGITO,
    "P5": R.P_ESPECIAL,
    "P6": R.REPETICION,
    "P7": R.SECUENCIA,
    "P8": R.PROHIBIDAS,
    "C1": R.C_ESTRUCTURA,
    "C2": R.C_LOCAL,
    "C3": R.C_LOCAL_LEN,
    "C4": R.C_DOM_ETIQ,
    "C5": R.C_TLD,
    "C6": R.C_DOM_LEN,
    "C*": R.C_COMPLETA,
}

# Pools and lengths per rule, mirroring the Go oracle test so both corpora
# stress the same boundaries.
POOLS = {
    "P1": ("aA1!", 8, 70),
    "P2": ("aA1!", 0, 12),
    "P3": ("aA1!", 0, 12),
    "P4": ("aA1!", 0, 12),
    "P5": ("aA1!", 0, 12),
    "P6": ("aAb1!", 0, 14),
    "P7": ("abcdefABCD12345!", 0, 14),
    "P8": ("passwordADMINqwerty", 0, 14),
    "C1": ("a.@-", 0, 10),
    "C2": ("a.-_", 0, 10),
    "C3": ("ab.", 55, 70),
    "C4": ("aB1-.", 0, 14),
    "C5": ("aB1-.", 0, 10),
    "C6": ("ab.", 250, 260),
    "C*": ("aB1-._@!", 0, 16),
}

MUESTRAS = {
    "P1": "abcdefghijklmno",
    "P2": "A",
    "P3": "a",
    "P4": "1",
    "P5": "!",
    "P6": "aaa",
    "P7": "abcd",
    "P8": "password",
    "C1": "a@b",
    "C2": "a.b",
    "C3": "a",
    "C4": "a.b",
    "C5": ".co",
    "C6": "a",
    "C*": "a@b.co",
}

POR_REGLA = 400
SEMILLA = 20261023


def mutar(rng: random.Random, s: str, pool: str) -> str:
    b = list(s)
    for _ in range(rng.randint(0, 3)):
        op = rng.randrange(3)
        if op == 0:
            b.insert(rng.randint(0, len(b)), rng.choice(pool))
        elif op == 1 and b:
            del b[rng.randrange(len(b))]
        elif b:
            b[rng.randrange(len(b))] = rng.choice(pool)
    return "".join(b)


def corpus() -> list[dict]:
    rng = random.Random(SEMILLA)
    casos = []
    for rid, (pool, lo, hi) in POOLS.items():
        for i in range(POR_REGLA):
            if i % 2 == 0:
                valor = mutar(rng, MUESTRAS[rid], pool)
            else:
                valor = "".join(rng.choice(pool) for _ in range(rng.randint(lo, hi)))
            casos.append({"id": rid, "valor": valor})
    return casos


def main() -> int:
    print("construyendo los autómatas del prototipo…", flush=True)
    compiladas = {rid: Compiled(root) for rid, root in REGLAS.items()}

    casos = corpus()
    print(f"corpus: {len(casos)} cadenas ({POR_REGLA} por regla)", flush=True)

    proc = subprocess.run(
        ["go", "run", "./cmd/oracle"],
        cwd=ROOT,
        input=json.dumps({"casos": casos}),
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        print("el binario Go falló:", proc.stderr.strip(), file=sys.stderr)
        return 1
    salida = json.loads(proc.stdout)

    fallos = 0

    # 1) Conteos de estados.
    for rid, c in compiladas.items():
        st = c.stats()
        esperado = [st["nfa"], st["dfa_sub"], st["dfa_tree"], st["dfa_min"]]
        obtenido = salida["conteos"].get(rid)
        if obtenido is None:
            print(f"DISCREPANCIA {rid}: el motor Go no reporta conteos", file=sys.stderr)
            fallos += 1
            continue
        obtenido = [obtenido["nfa"], obtenido["sub"], obtenido["tree"], obtenido["min"]]
        if esperado != obtenido:
            print(f"DISCREPANCIA {rid}: conteos Python={esperado} Go={obtenido}", file=sys.stderr)
            fallos += 1

    # 2) Veredictos.
    for caso, go_acepta in zip(casos, salida["veredictos"]):
        py_acepta = bool(compiladas[caso["id"]].match(caso["valor"]))
        if py_acepta != go_acepta:
            print(
                f"DISCREPANCIA {caso['id']}: {caso['valor']!r} Python={py_acepta} Go={go_acepta}",
                file=sys.stderr,
            )
            fallos += 1

    total_conteos = len(compiladas)
    if fallos:
        print(f"\n{fallos} discrepancias ({total_conteos} reglas, {len(casos)} cadenas)", file=sys.stderr)
        return 1

    print(f"OK: {total_conteos} reglas con los mismos conteos y {len(casos)} cadenas con el mismo veredicto")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
