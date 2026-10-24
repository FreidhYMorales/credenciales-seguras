"""Motor de ER: Thompson (AFND-e), metodo del arbol (followpos) y AFD.

Cada hoja de una ER consume UN caracter que pertenece a un conjunto
(una "etiqueta": L, l, N, S, ...). Todo se lee de uno en uno.
"""
import re
import sys
from collections import defaultdict

sys.setrecursionlimit(10000)


# ---------------------------------------------------------------- AST
class Leaf:
    def __init__(self, chars, name):
        self.chars, self.name = frozenset(chars), name

    def __str__(self):
        return self.name


class Eps:
    def __str__(self):
        return "e"


class Cat:
    def __init__(self, a, b):
        self.a, self.b = a, b

    def __str__(self):
        return f"{self.a}{self.b}"


class Alt:
    def __init__(self, a, b):
        self.a, self.b = a, b

    def __str__(self):
        return f"({self.a}|{self.b})"


class Star:
    def __init__(self, a):
        self.a = a

    def __str__(self):
        return f"({self.a})*"


def _balanced(items, ctor):
    if len(items) == 1:
        return items[0]
    m = len(items) // 2
    return ctor(_balanced(items[:m], ctor), _balanced(items[m:], ctor))


def cat(*xs):
    return _balanced(list(xs), Cat)


def alt(*xs):
    return _balanced(list(xs), Alt)


def plus(x):
    return Cat(x, Star(x))


def opt(x):
    return Alt(x, Eps())


def rep(x, lo, hi):
    """x{lo,hi} expandido: x^lo (x|e)^(hi-lo)."""
    parts = [x] * lo + [opt(x)] * (hi - lo)
    return cat(*parts)


# ------------------------------------------------- Thompson (AFND-e)
class NFA:
    def __init__(self, root):
        self.n, self.trans = 0, []  # (src, Leaf|None, dst)
        self.start, self.accept = self._build(root)

    def _new(self):
        self.n += 1
        return self.n - 1

    def _build(self, x):
        if isinstance(x, Leaf):
            s, a = self._new(), self._new()
            self.trans.append((s, x, a))
        elif isinstance(x, Eps):
            s, a = self._new(), self._new()
            self.trans.append((s, None, a))
        elif isinstance(x, Cat):
            s1, a1 = self._build(x.a)
            s2, a2 = self._build(x.b)
            self.trans.append((a1, None, s2))
            s, a = s1, a2
        elif isinstance(x, Alt):
            s, a = self._new(), self._new()
            s1, a1 = self._build(x.a)
            s2, a2 = self._build(x.b)
            self.trans += [(s, None, s1), (s, None, s2), (a1, None, a), (a2, None, a)]
        else:  # Star
            s, a = self._new(), self._new()
            s1, a1 = self._build(x.a)
            self.trans += [(s, None, s1), (s, None, a), (a1, None, s1), (a1, None, a)]
        return s, a

    def eclosure(self, states):
        if not hasattr(self, "_eps"):
            self._eps = defaultdict(list)
            for s, lab, d in self.trans:
                if lab is None:
                    self._eps[s].append(d)
        seen, stack = set(states), list(states)
        while stack:
            q = stack.pop()
            for d in self._eps[q]:
                if d not in seen:
                    seen.add(d)
                    stack.append(d)
        return frozenset(seen)


# -------------------------------------------------------------- DFA
class DFA:
    def __init__(self, atoms, char2atom, start, delta, accepts, n):
        self.atoms, self.char2atom = atoms, char2atom
        self.start, self.delta, self.accepts, self.n = start, delta, accepts, n

    def accepts_str(self, s):
        q = self.start
        for c in s:
            a = self.char2atom.get(c)
            if a is None:
                return False
            q = self.delta.get((q, a))
            if q is None:
                return False
        return q in self.accepts

    def run_prefix_fail(self, s):
        """Indice (0-based) del primer simbolo tras el cual ya no hay camino, o None."""
        q = self.start
        for i, c in enumerate(s):
            a = self.char2atom.get(c)
            q = None if a is None else self.delta.get((q, a))
            if q is None:
                return i
        return None


def leaves_of(x, out=None):
    out = [] if out is None else out
    if isinstance(x, Leaf):
        out.append(x)
    elif isinstance(x, (Cat, Alt)):
        leaves_of(x.a, out)
        leaves_of(x.b, out)
    elif isinstance(x, Star):
        leaves_of(x.a, out)
    return out


def make_atoms(root):
    sig = defaultdict(list)
    lv = list({id(l): l for l in leaves_of(root)}.values())
    universe = set().union(*[l.chars for l in lv])
    for c in universe:
        sig[tuple(i for i, l in enumerate(lv) if c in l.chars)].append(c)
    atoms = [frozenset(v) for v in sig.values()]
    c2a = {c: i for i, at in enumerate(atoms) for c in at}
    return atoms, c2a


def subset_dfa(nfa, atoms, c2a):
    """Construccion de subconjuntos sobre el AFND-e."""
    by_src = defaultdict(list)
    for s, lab, d in nfa.trans:
        if lab is not None:
            by_src[s].append((lab, d))
    s0 = nfa.eclosure({nfa.start})
    ids, delta, work = {s0: 0}, {}, [s0]
    while work:
        S = work.pop()
        for ai, at in enumerate(atoms):
            rep_c = next(iter(at))
            T = {d for q in S for lab, d in by_src[q] if rep_c in lab.chars}
            if not T:
                continue
            T = nfa.eclosure(T)
            if T not in ids:
                ids[T] = len(ids)
                work.append(T)
            delta[(ids[S], ai)] = ids[T]
    acc = {i for S, i in ids.items() if nfa.accept in S}
    return DFA(atoms, c2a, 0, delta, acc, len(ids)), ids


# --------------------------------------------- metodo del arbol
class Tree:
    """Arbol sintactico aumentado (r)# con nullable/first/last/follow."""

    END = Leaf({"#"}, "#")

    def __init__(self, root):
        self.root = Cat(root, Tree.END)
        self.pos = {}  # posicion -> Leaf
        self.follow = defaultdict(set)
        self.rows = []  # (descripcion, nullable, first, last) en post-orden
        _, self.first0, _ = self._walk(self.root)

    def _walk(self, x):
        if isinstance(x, Leaf):
            p = len(self.pos) + 1
            self.pos[p] = x
            return False, {p}, {p}
        if isinstance(x, Eps):
            return True, set(), set()
        if isinstance(x, Star):
            n, f, l = self._walk(x.a)
            for p in l:
                self.follow[p] |= f
            return True, f, l
        n1, f1, l1 = self._walk(x.a)
        n2, f2, l2 = self._walk(x.b)
        if isinstance(x, Cat):
            for p in l1:
                self.follow[p] |= f2
            return (n1 and n2), (f1 | f2 if n1 else f1), (l1 | l2 if n2 else l2)
        return (n1 or n2), f1 | f2, l1 | l2

    def dfa(self, atoms, c2a):
        end = max(self.pos)
        s0 = frozenset(self.first0)
        ids, delta, work = {s0: 0}, {}, [s0]
        while work:
            S = work.pop()
            for ai, at in enumerate(atoms):
                c = next(iter(at))
                T = frozenset(
                    q for p in S if p != end and c in self.pos[p].chars for q in self.follow[p]
                )
                if not T:
                    continue
                if T not in ids:
                    ids[T] = len(ids)
                    work.append(T)
                delta[(ids[S], ai)] = ids[T]
        acc = {i for S, i in ids.items() if end in S}
        return DFA(atoms, c2a, 0, delta, acc, len(ids)), ids


# ------------------------------------------------------ utilidades
def minimize(d):
    """Moore sobre el AFD completado con estado muerto. Devuelve #estados utiles."""
    dead = d.n
    A = range(len(d.atoms))
    part = {q: (q in d.accepts) for q in range(d.n + 1)}
    while True:
        sig = {
            q: (part[q], tuple(part[d.delta.get((q, a), dead)] for a in A))
            for q in range(d.n + 1)
        }
        ids = {}
        new = {q: ids.setdefault(s, len(ids)) for q, s in sig.items()}
        if len(set(new.values())) == len(set(part.values())):
            part = new
            break
        part = new
    return len({part[q] for q in range(d.n + 1)}) - 1  # sin la clase del estado muerto


def equivalent(d1, d2):
    """Equivalencia por BFS de pares (mismos atomos)."""
    seen, work = {(d1.start, d2.start)}, [(d1.start, d2.start)]
    while work:
        p, q = work.pop()
        if (p in d1.accepts if p is not None else False) != (q in d2.accepts if q is not None else False):
            return False
        for a in range(len(d1.atoms)):
            n = (d1.delta.get((p, a)) if p is not None else None,
                 d2.delta.get((q, a)) if q is not None else None)
            if n not in seen:
                seen.add(n)
                work.append(n)
    return True


def to_python_re(x):
    if isinstance(x, Leaf):
        return "[" + "".join(re.escape(c) for c in sorted(x.chars)) + "]"
    if isinstance(x, Eps):
        return "(?:)"
    if isinstance(x, Cat):
        return to_python_re(x.a) + to_python_re(x.b)
    if isinstance(x, Alt):
        return "(?:" + to_python_re(x.a) + "|" + to_python_re(x.b) + ")"
    return "(?:" + to_python_re(x.a) + ")*"


class Compiled:
    """ER -> AFND-e (Thompson) -> AFD (subconjuntos) y AFD (arbol)."""

    def __init__(self, root, name=""):
        self.name, self.root = name, root
        self.nfa = NFA(root)
        self.atoms, self.c2a = make_atoms(root)
        self.dfa_sub, _ = subset_dfa(self.nfa, self.atoms, self.c2a)
        self.tree = Tree(root)
        self.dfa_tree, self.tree_ids = self.tree.dfa(self.atoms, self.c2a)
        self._re = None

    @property
    def pyre(self):
        if self._re is None:
            self._re = re.compile(to_python_re(self.root), re.S)
        return self._re

    def match(self, s):
        return self.dfa_tree.accepts_str(s)

    def stats(self):
        return dict(
            nfa=self.nfa.n,
            leaves=len(self.tree.pos) - 1,
            dfa_sub=self.dfa_sub.n,
            dfa_tree=self.dfa_tree.n,
            dfa_min=minimize(self.dfa_tree),
            same=equivalent(self.dfa_sub, self.dfa_tree),
        )
