package automata

// Stats holds the four state counts the advance document prints for a rule.
type Stats struct {
	NFA    int `json:"nfa"`
	Subset int `json:"sub"`
	Tree   int `json:"tree"`
	Min    int `json:"min"`
}

// Compiled bundles every view of one regular expression: the Thompson NFA, the
// subset DFA, the syntax-tree DFA and the minimized DFA.
type Compiled struct {
	Root     Node
	NFA      *NFA
	Subset   *DFA
	Tree     *DFA
	Min      *DFA
	Atoms    []Atom
	CharAtom map[rune]int
	re2      string
}

// Compile builds all four automata of root.
func Compile(root Node) *Compiled {
	atoms, c2a := Atoms(root)
	nfa := Thompson(root)
	sub := Subset(nfa, atoms, c2a)
	tree := TreeDFA(root, atoms, c2a)
	return &Compiled{
		Root:     root,
		NFA:      nfa,
		Subset:   sub,
		Tree:     tree,
		Min:      Minimize(tree),
		Atoms:    atoms,
		CharAtom: c2a,
		re2:      ToRE2(root),
	}
}

// Accepts reports whether s belongs to the language, using the tree DFA (the
// same automaton the prototype used as its reference).
func (c *Compiled) Accepts(s string) bool {
	return c.Tree.Accepts([]rune(s))
}

// RE2 returns the anchored RE2 translation.
func (c *Compiled) RE2() string { return c.re2 }

// Stats returns the four state counts.
func (c *Compiled) Stats() Stats {
	return Stats{
		NFA:    c.NFA.N,
		Subset: c.Subset.N,
		Tree:   c.Tree.N,
		Min:    c.Min.N,
	}
}
