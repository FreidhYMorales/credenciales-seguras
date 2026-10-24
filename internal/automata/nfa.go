package automata

// NFATrans is one transition of the Thompson automaton. A nil Chars slice is an
// epsilon transition.
type NFATrans struct {
	From  int
	To    int
	Chars []rune
	Name  string // label of the leaf, empty for epsilon transitions
}

// NFA is the Thompson automaton of an expression.
type NFA struct {
	N      int
	Start  int
	Accept int
	Trans  []NFATrans

	eps [][]int
}

// Thompson builds the epsilon-NFA of root. The allocation order matches the
// prototype exactly (Alt and Star reserve their own two states before building
// the child), which is why the state numbering agrees with the figures in the
// advance document.
func Thompson(root Node) *NFA {
	b := &nfaBuilder{}
	s, a := b.build(root)
	return &NFA{N: b.n, Start: s, Accept: a, Trans: b.trans}
}

type nfaBuilder struct {
	n     int
	trans []NFATrans
}

func (b *nfaBuilder) newState() int {
	b.n++
	return b.n - 1
}

func (b *nfaBuilder) eps(from, to int) {
	b.trans = append(b.trans, NFATrans{From: from, To: to})
}

func (b *nfaBuilder) build(n Node) (start, accept int) {
	switch t := n.(type) {
	case *leafNode:
		s, a := b.newState(), b.newState()
		b.trans = append(b.trans, NFATrans{From: s, To: a, Chars: t.chars, Name: t.name})
		return s, a
	case *epsNode:
		s, a := b.newState(), b.newState()
		b.eps(s, a)
		return s, a
	case *catNode:
		s1, a1 := b.build(t.a)
		s2, a2 := b.build(t.b)
		b.eps(a1, s2)
		return s1, a2
	case *altNode:
		s, a := b.newState(), b.newState()
		s1, a1 := b.build(t.a)
		s2, a2 := b.build(t.b)
		b.eps(s, s1)
		b.eps(s, s2)
		b.eps(a1, a)
		b.eps(a2, a)
		return s, a
	case *starNode:
		s, a := b.newState(), b.newState()
		s1, a1 := b.build(t.a)
		b.eps(s, s1)
		b.eps(s, a)
		b.eps(a1, s1)
		b.eps(a1, a)
		return s, a
	}
	panic("automata: unknown node type")
}

func (n *NFA) buildEps() {
	if n.eps != nil {
		return
	}
	n.eps = make([][]int, n.N)
	for _, t := range n.Trans {
		if t.Chars == nil {
			n.eps[t.From] = append(n.eps[t.From], t.To)
		}
	}
}

// EClose returns the epsilon closure of seed, sorted ascending.
func (n *NFA) EClose(seed []int) []int {
	n.buildEps()
	seen := make([]bool, n.N)
	stack := make([]int, 0, len(seed))
	for _, s := range seed {
		if !seen[s] {
			seen[s] = true
			stack = append(stack, s)
		}
	}
	for len(stack) > 0 {
		q := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, d := range n.eps[q] {
			if !seen[d] {
				seen[d] = true
				stack = append(stack, d)
			}
		}
	}
	out := make([]int, 0, len(seed))
	for q, v := range seen {
		if v {
			out = append(out, q)
		}
	}
	return out
}
