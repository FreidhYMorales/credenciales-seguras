package generate

import (
	"math/rand"
	"strconv"

	"github.com/FreidhYMorales/credenciales-seguras/internal/automata"
	"github.com/FreidhYMorales/credenciales-seguras/internal/rules"
)

// product is the product automaton P1 x P2 x P3 x P4 x P5 built over the
// minimized DFAs. The P1 component counts characters, so a state already encodes
// how many symbols have been read.
type product struct {
	atoms  []automata.Atom
	labels []string
	delta  [][]int
	accept []bool
	reach  [][]bool // reach[id][r]: accepting state reachable in exactly r steps
	start  int
}

// componentIDs are the rules whose product the walk follows, in order.
var componentIDs = []string{"P1", "P2", "P3", "P4", "P5"}

// buildProduct builds the product from the minimized DFAs of P1-P5.
func buildProduct() *product {
	comps := make([]*automata.DFA, len(componentIDs))
	for i, id := range componentIDs {
		comps[i] = rules.Get(id).Min
	}

	// Canonical alphabet: the atoms of P1, which are the four classes L, l, N, S.
	atoms := comps[0].Atoms
	labels := make([]string, len(atoms))
	for i, a := range atoms {
		labels[i] = classOf(a.Rep())
	}

	// Atom index of each component for each canonical atom.
	atomMap := make([][]int, len(comps))
	for j, c := range comps {
		atomMap[j] = make([]int, len(atoms))
		for a, at := range atoms {
			idx, ok := c.CharAtom[at.Rep()]
			if !ok {
				idx = -1
			}
			atomMap[j][a] = idx
		}
	}

	type key [5]int
	var (
		index  = make(map[key]int)
		states []key
		delta  [][]int
		accept []bool
	)

	startKey := key{comps[0].Start, comps[1].Start, comps[2].Start, comps[3].Start, comps[4].Start}
	index[startKey] = 0
	states = append(states, startKey)
	delta = append(delta, makeRow(len(atoms)))
	accept = append(accept, isAccepting(comps, startKey))

	for cursor := 0; cursor < len(states); cursor++ {
		cur := states[cursor]
		for a := range atoms {
			var next key
			ok := true
			for j := range comps {
				ai := atomMap[j][a]
				if ai < 0 {
					ok = false
					break
				}
				nj := comps[j].Delta[cur[j]][ai]
				if nj < 0 {
					ok = false
					break
				}
				next[j] = nj
			}
			if !ok {
				continue
			}
			id, seen := index[next]
			if !seen {
				id = len(states)
				index[next] = id
				states = append(states, next)
				delta = append(delta, makeRow(len(atoms)))
				accept = append(accept, isAccepting(comps, next))
			}
			delta[cursor][a] = id
		}
	}

	// reach[id][r]: can an accepting state be reached in exactly r steps.
	reach := make([][]bool, len(states))
	for i := range reach {
		reach[i] = make([]bool, maxPasswordLen+1)
		reach[i][0] = accept[i]
	}
	for r := 1; r <= maxPasswordLen; r++ {
		for i := range states {
			ok := false
			for a := range atoms {
				if t := delta[i][a]; t >= 0 && reach[t][r-1] {
					ok = true
					break
				}
			}
			reach[i][r] = ok
		}
	}

	return &product{atoms: atoms, labels: labels, delta: delta, accept: accept, reach: reach, start: 0}
}

// Size returns the number of reachable product states.
func (p *product) Size() int { return len(p.delta) }

// walk returns a string of exactly length symbols accepted by the product, and
// the recorded path. It only takes transitions from which an accepting state is
// still reachable in the remaining steps, so it cannot dead-end.
func (p *product) walk(rng *rand.Rand, length int) (string, []Step, bool) {
	if length > maxPasswordLen || !p.reach[p.start][length] {
		return "", nil, false
	}
	id := p.start
	out := make([]rune, 0, length)
	path := make([]Step, 0, length)

	for k := 0; k < length; k++ {
		remaining := length - k - 1
		var cands []int
		for a := range p.atoms {
			if t := p.delta[id][a]; t >= 0 && p.reach[t][remaining] {
				cands = append(cands, a)
			}
		}
		if len(cands) == 0 {
			return "", nil, false
		}
		a := cands[rng.Intn(len(cands))]
		atom := p.atoms[a]
		r := atom.Chars[rng.Intn(len(atom.Chars))]
		out = append(out, r)
		path = append(path, Step{
			Paso:    k,
			Estado:  "p" + strconv.Itoa(id),
			Simbolo: string(r),
			Clase:   p.labels[a],
		})
		id = p.delta[id][a]
	}
	if !p.accept[id] {
		return "", nil, false
	}
	return string(out), path, true
}

func isAccepting(comps []*automata.DFA, k [5]int) bool {
	for j, c := range comps {
		if !c.Accept[k[j]] {
			return false
		}
	}
	return true
}

func makeRow(n int) []int {
	row := make([]int, n)
	for i := range row {
		row[i] = -1
	}
	return row
}
