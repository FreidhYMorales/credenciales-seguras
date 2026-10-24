// Package automata implements the regular-expression engine: an AST, the
// Thompson construction with epsilon transitions, the subset construction, the
// syntax-tree (followpos) method, Moore minimization, equivalence checking,
// simulation with a trace, and translation to RE2 syntax.
//
// The package has no dependencies on the rest of the project and no UI
// concerns. Every leaf consumes exactly one symbol from a character set (a
// "label": L, l, N, S, ...), as in the advance document.
package automata

import (
	"encoding/binary"
	"sort"
)

// Node is a node of the regular-expression AST.
type Node interface {
	isNode()
	String() string
}

type leafNode struct {
	chars []rune // sorted, deduplicated, never empty
	name  string
}

type epsNode struct{}

type catNode struct{ a, b Node }

type altNode struct{ a, b Node }

type starNode struct{ a Node }

func (*leafNode) isNode() {}
func (*epsNode) isNode()  {}
func (*catNode) isNode()  {}
func (*altNode) isNode()  {}
func (*starNode) isNode() {}

// Leaf returns a node that consumes one symbol from chars.
func Leaf(chars []rune, name string) Node {
	c := make([]rune, len(chars))
	copy(c, chars)
	sort.Slice(c, func(i, j int) bool { return c[i] < c[j] })
	out := c[:0]
	for i, r := range c {
		if i == 0 || r != c[i-1] {
			out = append(out, r)
		}
	}
	return &leafNode{chars: out, name: name}
}

// LeafRange returns a leaf over every rune in [lo, hi].
func LeafRange(lo, hi rune, name string) Node {
	chars := make([]rune, 0, hi-lo+1)
	for r := lo; r <= hi; r++ {
		chars = append(chars, r)
	}
	return Leaf(chars, name)
}

// Eps returns the node that consumes nothing.
func Eps() Node { return &epsNode{} }

// balanced builds a left-to-right balanced tree, so deep repetitions do not
// produce deep recursion (the advance document uses the same trick).
func balanced(ns []Node, f func(a, b Node) Node) Node {
	switch len(ns) {
	case 0:
		return &epsNode{}
	case 1:
		return ns[0]
	}
	m := len(ns) / 2
	return f(balanced(ns[:m], f), balanced(ns[m:], f))
}

// Cat returns the concatenation of ns.
func Cat(ns ...Node) Node {
	return balanced(ns, func(a, b Node) Node { return &catNode{a, b} })
}

// Alt returns the alternation of ns.
func Alt(ns ...Node) Node {
	return balanced(ns, func(a, b Node) Node { return &altNode{a, b} })
}

// Star returns n*.
func Star(n Node) Node { return &starNode{n} }

// Plus returns n n*, i.e. one or more repetitions.
func Plus(n Node) Node { return Cat(n, Star(n)) }

// Opt returns n|epsilon.
func Opt(n Node) Node { return Alt(n, Eps()) }

// Rep returns n{lo,hi} expanded as n^lo (n|epsilon)^(hi-lo), exactly as the
// prototype does; it is what makes the length-bounded rules finite automata.
func Rep(n Node, lo, hi int) Node {
	parts := make([]Node, 0, hi)
	for i := 0; i < lo; i++ {
		parts = append(parts, n)
	}
	for i := 0; i < hi-lo; i++ {
		parts = append(parts, Opt(n))
	}
	return Cat(parts...)
}

func (n *leafNode) String() string { return n.name }
func (n *epsNode) String() string  { return "e" }
func (n *catNode) String() string  { return n.a.String() + n.b.String() }
func (n *altNode) String() string  { return "(" + n.a.String() + "|" + n.b.String() + ")" }
func (n *starNode) String() string { return "(" + n.a.String() + ")*" }

// leavesOf collects the leaves in left-to-right order, the order the atom
// partition and the tree positions depend on.
func leavesOf(n Node, out *[]*leafNode) {
	switch t := n.(type) {
	case *leafNode:
		*out = append(*out, t)
	case *catNode:
		leavesOf(t.a, out)
		leavesOf(t.b, out)
	case *altNode:
		leavesOf(t.a, out)
		leavesOf(t.b, out)
	case *starNode:
		leavesOf(t.a, out)
	}
}

// Atoms returns the partition of the alphabet induced by the leaves and the
// char-to-atom map. Two symbols are in the same atom when they belong to
// exactly the same leaves, so the automaton only ever moves per atom.
func Atoms(root Node) ([]Atom, map[rune]int) {
	var leaves []*leafNode
	leavesOf(root, &leaves)

	seen := make(map[rune]bool)
	var universe []rune
	for _, l := range leaves {
		for _, r := range l.chars {
			if !seen[r] {
				seen[r] = true
				universe = append(universe, r)
			}
		}
	}
	sort.Slice(universe, func(i, j int) bool { return universe[i] < universe[j] })

	type group struct {
		sig  string
		atom Atom
	}
	var groups []*group
	index := make(map[string]*group)
	for _, r := range universe {
		// 4 bytes per matching leaf index: a plain byte would collide past 255
		// leaves, and the repetition rule alone has 282.
		var sig []byte
		for i, l := range leaves {
			if contains(l.chars, r) {
				sig = binary.BigEndian.AppendUint32(sig, uint32(i))
			}
		}
		k := string(sig)
		g, ok := index[k]
		if !ok {
			g = &group{sig: k}
			index[k] = g
			groups = append(groups, g)
		}
		g.atom.Chars = append(g.atom.Chars, r)
	}

	atoms := make([]Atom, len(groups))
	charAtom := make(map[rune]int, len(universe))
	for i, g := range groups {
		atoms[i] = g.atom
		for _, r := range g.atom.Chars {
			charAtom[r] = i
		}
	}
	return atoms, charAtom
}

// Atom is one class of the alphabet partition.
type Atom struct {
	Chars []rune // sorted
}

// Rep returns any member of the atom; valid because an atom is non-empty.
func (a Atom) Rep() rune { return a.Chars[0] }

func contains(xs []rune, r rune) bool {
	i := sort.Search(len(xs), func(i int) bool { return xs[i] >= r })
	return i < len(xs) && xs[i] == r
}
