"""Genera diagramas (Graphviz) de AFND-e y AFD para el documento."""
import subprocess
from collections import defaultdict
from motor_er import Compiled, Leaf, Star, alt, cat, plus, opt, leaves_of
from reglas import SIGMA_P, L, l, N, S, A, N_e, H, E, P, AT, at, label

OUT = "figuras"


def render(name, dot):
    import os
    os.makedirs(OUT, exist_ok=True)
    open(f"{OUT}/{name}.dot", "w").write(dot)
    subprocess.run(["dot", "-Tpng", "-Gdpi=130", f"{OUT}/{name}.dot", "-o", f"{OUT}/{name}.png"], check=True)


def nfa_dot(c, title=""):
    n = c.nfa
    d = ["digraph G{rankdir=LR;nodesep=.25;ranksep=.35;node[shape=circle,fontsize=11,width=.4,fixedsize=true];edge[fontsize=11];"]
    d.append(f'labelloc=t;label="{title}";' if title else "")
    d.append('_s[shape=point,width=.05];')
    for q in range(n.n):
        d.append(f'q{q}[label="{q}"{",shape=doublecircle" if q == n.accept else ""}];')
    d.append(f"_s->q{n.start};")
    for s, lab, t in n.trans:
        d.append(f'q{s}->q{t}[label="{"ε" if lab is None else lab.name}"];')
    d.append("}")
    return "\n".join(d)


def atom_names(c):
    names = {lf.chars: lf.name for lf in leaves_of(c.root)}
    return [names.get(a, "".join(sorted(a))[:6]) for a in c.atoms]


def dfa_dot(c, sets=True, title=""):
    d0, ids = c.dfa_tree, c.tree_ids
    nm = atom_names(c)
    lab = {i: chr(65 + i) if i < 26 else f"Q{i}" for i in range(d0.n)}
    end = max(c.tree.pos)
    g = ["digraph G{rankdir=LR;nodesep=.35;ranksep=.5;node[shape=circle,fontsize=11];edge[fontsize=11];"]
    g.append(f'labelloc=t;label="{title}";' if title else "")
    g.append('_s[shape=point,width=.05];')
    for S_, i in ids.items():
        txt = lab[i] + (("\\n{" + ",".join(map(lambda p: "#" if p == end else str(p), sorted(S_))) + "}") if sets else "")
        g.append(f'{lab[i]}[label="{txt}"{",shape=doublecircle" if i in d0.accepts else ""}];')
    g.append(f"_s->{lab[0]};")
    m = defaultdict(list)
    for (q, a), r in d0.delta.items():
        m[(q, r)].append(nm[a])
    for (q, r), ls in m.items():
        g.append(f'{lab[q]}->{lab[r]}[label="{",".join(sorted(ls))}"];')
    g.append("}")
    return "\n".join(g)


def tree_table(c):
    t = c.tree
    rows = [(p, t.pos[p].name, sorted(t.follow[p])) for p in sorted(t.pos)]
    return rows


if __name__ == "__main__":
    # 1) fragmentos de Thompson (a y b genericos)
    a_, b_ = Leaf({"a"}, "a"), Leaf({"b"}, "b")
    for nm_, rx, tt in (("th_simbolo", a_, "a"), ("th_union", alt(a_, b_), "a | b"),
                        ("th_concat", cat(a_, b_), "a · b"), ("th_estrella", Star(a_), "a*")):
        render(nm_, nfa_dot(Compiled(rx), tt))
    # 2) Sigma* L Sigma*
    c = Compiled(cat(Star(SIGMA_P), L, Star(SIGMA_P)))
    render("nfa_P2", nfa_dot(c))
    render("dfa_P2", dfa_dot(c))
    print(tree_table(c))
    # 3) parte local con atext como una sola etiqueta 'a'
    a = Leaf({"a"}, "a"); dot = Leaf({"."}, ".")
    c2 = Compiled(cat(plus(a), Star(cat(dot, plus(a)))))
    render("nfa_local", nfa_dot(c2))
    render("dfa_local", dfa_dot(c2))
    print(tree_table(c2))
    # 4) correo completo (etiquetas A,N,H,E,P,@)
    from reglas import C_COMPLETA
    c3 = Compiled(C_COMPLETA)
    render("dfa_correo", dfa_dot(c3, sets=False))
    print(c3.stats())
