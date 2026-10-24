import random, subprocess, sys
from motor_er import Compiled, leaves_of, Star, cat
from reglas import SIGMA_P, L, C_COMPLETA
D = "/home/deadlock/Documents/AUTOMATAS/JFLAP/target/release/automata-cli"
random.seed(11)
def run(n, c):
    names = {lf.chars: lf.name for lf in leaves_of(c.root)}
    rep = {names[a]: next(iter(a)) for a in c.atoms}
    d = c.dfa_tree; an = [names[a] for a in c.atoms]
    words = set()
    while len(words) < 400:            # mitad al azar, mitad caminatas guiadas por el AFD hacia aceptacion
        if random.random() < .5:
            w = tuple(random.choice(list(rep)) for _ in range(random.randint(1, 12)))
        else:
            q, w = 0, []
            for _ in range(random.randint(1, 14)):
                opts = [(a, r) for (s, a), r in d.delta.items() if s == q]
                if not opts: break
                a, q = random.choice(opts); w.append(an[a])
            w = tuple(w)
        if w: words.add(w)
    words = sorted(words)
    open(f"jff/{n}-w2.txt", "w").write("\n".join(" ".join(w) for w in words) + "\n")
    exp = [c.match("".join(rep[t] for t in w)) for w in words]
    for k in ("nfa", "dfa"):
        out = subprocess.run([D, "sim", "--file", f"jff/{n}-{k}.jff", "--words-file", f"jff/{n}-w2.txt"], capture_output=True, text=True).stdout
        got = [l.startswith("Accepted") for l in out.splitlines() if l.split(" ")[0] in ("Accepted", "Rejected", "Stuck")]
        diff = sum(a != b for a, b in zip(got, exp))
        print(f"{n}-{k}: {len(got)}/{len(exp)} palabras, aceptadas={sum(exp)}, diferencias={diff}")
run("P2", Compiled(cat(Star(SIGMA_P), L, Star(SIGMA_P))))
run("correo", Compiled(C_COMPLETA))
