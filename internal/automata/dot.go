package automata

import (
	"fmt"
	"strings"
)

// AtomLabel returns a short display label for an atom: a class letter when the
// atom is a whole class, the symbol itself when it holds a single character.
func AtomLabel(a Atom) string {
	if len(a.Chars) == 1 {
		return string(a.Chars[0])
	}
	upper, lower, digit := true, true, true
	for _, r := range a.Chars {
		if r < 'A' || r > 'Z' {
			upper = false
		}
		if r < 'a' || r > 'z' {
			lower = false
		}
		if r < '0' || r > '9' {
			digit = false
		}
	}
	switch {
	case upper:
		return "L"
	case lower:
		return "l"
	case digit:
		return "N"
	case len(a.Chars) == 52:
		return "A"
	default:
		return "S"
	}
}

// DOTNFA renders the Thompson automaton as Graphviz DOT.
func DOTNFA(n *NFA) string {
	var b strings.Builder
	b.WriteString("digraph G {\n  rankdir=LR;\n  node [shape=circle, fontsize=11];\n")
	fmt.Fprintf(&b, "  __start [shape=point, width=0.1];\n  __start -> q%d;\n", n.Start)
	for q := 0; q < n.N; q++ {
		if q == n.Accept {
			fmt.Fprintf(&b, "  q%d [shape=doublecircle];\n", q)
		} else {
			fmt.Fprintf(&b, "  q%d;\n", q)
		}
	}
	for _, t := range n.Trans {
		label := t.Name
		if t.Chars == nil {
			label = "ε"
		}
		fmt.Fprintf(&b, "  q%d -> q%d [label=%q];\n", t.From, t.To, label)
	}
	b.WriteString("}\n")
	return b.String()
}

// DOTDFA renders a DFA as Graphviz DOT, grouping the transitions that share
// source, target and label so the graph stays readable.
func DOTDFA(d *DFA) string {
	var b strings.Builder
	b.WriteString("digraph G {\n  rankdir=LR;\n  node [shape=circle, fontsize=11];\n")
	fmt.Fprintf(&b, "  __start [shape=point, width=0.1];\n  __start -> q%d;\n", d.Start)
	for q := 0; q < d.N; q++ {
		if d.Accept[q] {
			fmt.Fprintf(&b, "  q%d [shape=doublecircle];\n", q)
		} else {
			fmt.Fprintf(&b, "  q%d;\n", q)
		}
	}
	for q := 0; q < d.N; q++ {
		grouped := make(map[int][]string)
		var order []int
		for a := range d.Atoms {
			t := d.Delta[q][a]
			if t < 0 {
				continue
			}
			if _, ok := grouped[t]; !ok {
				order = append(order, t)
			}
			grouped[t] = append(grouped[t], AtomLabel(d.Atoms[a]))
		}
		for _, t := range order {
			fmt.Fprintf(&b, "  q%d -> q%d [label=%q];\n", q, t, strings.Join(grouped[t], ","))
		}
	}
	b.WriteString("}\n")
	return b.String()
}

// Edges returns the number of drawn transitions of a DFA.
func (d *DFA) Edges() int {
	n := 0
	for q := 0; q < d.N; q++ {
		for a := range d.Atoms {
			if d.Delta[q][a] >= 0 {
				n++
			}
		}
	}
	return n
}
