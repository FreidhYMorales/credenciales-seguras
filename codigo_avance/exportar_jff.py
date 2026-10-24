"""Exporta AFND-e (Thompson) y AFD (arbol) a .jff y genera lotes de palabras (tags) con veredicto propio."""
import os, random
from motor_er import Compiled, Leaf, Star, cat, plus, leaves_of
from reglas import SIGMA_P, L, C_COMPLETA, P_MAYUS

random.seed(3)
OUT = "jff"; os.makedirs(OUT, exist_ok=True)

def jff(states, start, accepts, trans):
    x = ['<?xml version="1.0" encoding="UTF-8" standalone="no"?>', "<structure>", "\t<type>fa</type>", "\t<automaton>"]
    for q in states:
        x.append(f'\t\t<state id="{q}" name="q{q}"><x>{60*q}</x><y>0</y>' + ("<initial/>" if q == start else "") + ("<final/>" if q in accepts else "") + "</state>")
    for s, a, d in trans:
        x.append(f"\t\t<transition><from>{s}</from><to>{d}</to>" + (f"<read>{a}</read>" if a else "<read/>") + "</transition>")
    x += ["\t</automaton>", "</structure>"]
    return "\n".join(x)

def export(name, c):
    n = c.nfa
    open(f"{OUT}/{name}-nfa.jff", "w").write(jff(range(n.n), n.start, {n.accept}, [(s, l.name if l else "", d) for s, l, d in n.trans]))
    names = {lf.chars: lf.name for lf in leaves_of(c.root)}
    an = [names[a] for a in c.atoms]
    d = c.dfa_tree
    open(f"{OUT}/{name}-dfa.jff", "w").write(jff(range(d.n), d.start, d.accepts, [(q, an[a], r) for (q, a), r in d.delta.items()]))
    # palabras: secuencias de tags, veredicto del motor
    rep = {names[a]: next(iter(a)) for a in c.atoms}
    words = []
    for _ in range(300):
        w = [random.choice(list(rep)) for _ in range(random.randint(0, 9))]
        words.append(w)
    with open(f"{OUT}/{name}-words.txt", "w") as f, open(f"{OUT}/{name}-expected.txt", "w") as g:
        for w in words:
            f.write(" ".join(w) + "\n")
            g.write(("1" if c.match("".join(rep[t] for t in w)) else "0") + "\n")

export("P2", Compiled(cat(Star(SIGMA_P), L, Star(SIGMA_P))))
export("correo", Compiled(C_COMPLETA))
print("ok")
