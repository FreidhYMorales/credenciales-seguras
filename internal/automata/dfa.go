package automata

import "encoding/binary"

// DFA is a deterministic automaton over the atom alphabet. Delta[q][a] is -1
// when the transition is absent: the automata are partial, exactly like the
// prototype, so a missing edge means "reject from here".
type DFA struct {
	Atoms    []Atom
	CharAtom map[rune]int
	Start    int
	Delta    [][]int
	Accept   []bool
	N        int
}

// Step returns the state reached from q with r, or -1.
func (d *DFA) Step(q int, r rune) int {
	a, ok := d.CharAtom[r]
	if !ok {
		return -1
	}
	return d.Delta[q][a]
}

// Accepts reports whether s is in the language.
func (d *DFA) Accepts(s []rune) bool {
	q := d.Start
	for _, r := range s {
		q = d.Step(q, r)
		if q < 0 {
			return false
		}
	}
	return d.Accept[q]
}

// RunFail returns the 0-based index of the first symbol after which there is no
// path, or -1 when the whole string has a path (whether or not it is accepted).
func (d *DFA) RunFail(s []rune) int {
	q := d.Start
	for i, r := range s {
		q = d.Step(q, r)
		if q < 0 {
			return i
		}
	}
	return -1
}

// Step is one move of a simulation trace.
type Step struct {
	Paso    int
	Desde   int
	Hacia   int
	Simbolo rune
}

// Simulate runs s over d and returns the verdict, the trace and the failure
// index (-1 when no symbol got stuck).
func Simulate(d *DFA, s []rune) (acepta bool, traza []Step, fallo int) {
	q := d.Start
	traza = make([]Step, 0, len(s))
	for i, r := range s {
		next := d.Step(q, r)
		traza = append(traza, Step{Paso: i, Desde: q, Hacia: next, Simbolo: r})
		if next < 0 {
			return false, traza, i
		}
		q = next
	}
	return d.Accept[q], traza, -1
}

// keyOf encodes a sorted state set as a map key.
func keyOf(st []int) string {
	b := make([]byte, 4*len(st))
	for i, v := range st {
		binary.BigEndian.PutUint32(b[4*i:], uint32(v))
	}
	return string(b)
}

// dfaBuilder interns state sets and grows the transition table as it goes.
type dfaBuilder struct {
	atoms []Atom
	c2a   map[rune]int
	ids   map[string]int
	sets  [][]int
	delta [][]int
}

func newDFABuilder(atoms []Atom, c2a map[rune]int) *dfaBuilder {
	return &dfaBuilder{
		atoms: atoms,
		c2a:   c2a,
		ids:   make(map[string]int),
	}
}

// intern returns the id of the set, and whether it was new.
func (b *dfaBuilder) intern(st []int) (int, bool) {
	k := keyOf(st)
	if id, ok := b.ids[k]; ok {
		return id, false
	}
	id := len(b.sets)
	b.ids[k] = id
	b.sets = append(b.sets, st)
	row := make([]int, len(b.atoms))
	for i := range row {
		row[i] = -1
	}
	b.delta = append(b.delta, row)
	return id, true
}

func (b *dfaBuilder) finish(start int, accepting func([]int) bool) *DFA {
	accept := make([]bool, len(b.sets))
	for i, st := range b.sets {
		accept[i] = accepting(st)
	}
	return &DFA{
		Atoms:    b.atoms,
		CharAtom: b.c2a,
		Start:    start,
		Delta:    b.delta,
		Accept:   accept,
		N:        len(b.sets),
	}
}

type nfaEdge struct {
	atom int
	to   int
}

// Subset builds the DFA of the epsilon-NFA by subset construction.
func Subset(n *NFA, atoms []Atom, c2a map[rune]int) *DFA {
	edges := make([][]nfaEdge, n.N)
	for _, t := range n.Trans {
		if t.Chars == nil {
			continue
		}
		for ai, at := range atoms {
			if contains(t.Chars, at.Rep()) {
				edges[t.From] = append(edges[t.From], nfaEdge{atom: ai, to: t.To})
			}
		}
	}

	b := newDFABuilder(atoms, c2a)
	start, _ := b.intern(n.EClose([]int{n.Start}))
	queue := []int{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		set := b.sets[cur]
		for ai := range atoms {
			var next []int
			for _, q := range set {
				for _, e := range edges[q] {
					if e.atom == ai {
						next = append(next, e.to)
					}
				}
			}
			if len(next) == 0 {
				continue
			}
			target, isNew := b.intern(n.EClose(next))
			b.delta[cur][ai] = target
			if isNew {
				queue = append(queue, target)
			}
		}
	}

	return b.finish(start, func(st []int) bool {
		for _, q := range st {
			if q == n.Accept {
				return true
			}
		}
		return false
	})
}
