package automata

import "encoding/binary"

// Minimize runs Moore refinement over the DFA completed with a dead state and
// returns a new DFA with the dead class removed.
//
// The result is a partial DFA: a missing transition means "reject from here".
// This is the convention of the prototype, which is why P1 minimizes to 65
// states rather than 66.
func Minimize(d *DFA) *DFA {
	dead := d.N
	total := d.N + 1

	part := make([]int, total)
	for q := 0; q < d.N; q++ {
		if d.Accept[q] {
			part[q] = 1
		}
	}

	for {
		next := make([]int, total)
		ids := make(map[string]int, total)
		var buf []byte
		for q := 0; q < total; q++ {
			buf = buf[:0]
			buf = binary.AppendUvarint(buf, uint64(part[q]))
			for a := range d.Atoms {
				t := dead
				if q < d.N {
					if v := d.Delta[q][a]; v >= 0 {
						t = v
					}
				}
				buf = binary.AppendUvarint(buf, uint64(part[t]))
			}
			k := string(buf)
			id, ok := ids[k]
			if !ok {
				id = len(ids)
				ids[k] = id
			}
			next[q] = id
		}
		refined := len(ids) != countDistinct(part)
		part = next
		if !refined {
			break
		}
	}

	// Classes in order of first appearance, with a representative state each.
	classIndex := make(map[int]int, total)
	var classes []int
	rep := make(map[int]int, total)
	for q := 0; q < total; q++ {
		c := part[q]
		if _, ok := classIndex[c]; !ok {
			classIndex[c] = len(classes)
			classes = append(classes, c)
		}
		if _, ok := rep[c]; !ok && q < d.N {
			rep[c] = q
		}
	}

	deadClass := part[dead]
	newIndex := make(map[int]int, len(classes))
	n := 0
	for _, c := range classes {
		if c == deadClass {
			continue
		}
		newIndex[c] = n
		n++
	}
	if n == 0 {
		return &DFA{Atoms: d.Atoms, CharAtom: d.CharAtom, Start: 0, N: 0}
	}

	out := &DFA{
		Atoms:    d.Atoms,
		CharAtom: d.CharAtom,
		Start:    newIndex[part[d.Start]],
		Delta:    make([][]int, n),
		Accept:   make([]bool, n),
		N:        n,
	}
	for i := range out.Delta {
		out.Delta[i] = make([]int, len(d.Atoms))
		for a := range out.Delta[i] {
			out.Delta[i][a] = -1
		}
	}
	for _, c := range classes {
		ni, ok := newIndex[c]
		if !ok {
			continue
		}
		r := rep[c]
		out.Accept[ni] = d.Accept[r]
		for a := range d.Atoms {
			t := dead
			if v := d.Delta[r][a]; v >= 0 {
				t = v
			}
			if nj, ok := newIndex[part[t]]; ok {
				out.Delta[ni][a] = nj
			}
		}
	}
	return out
}

func countDistinct(xs []int) int {
	seen := make(map[int]struct{}, len(xs))
	for _, x := range xs {
		seen[x] = struct{}{}
	}
	return len(seen)
}

// Equivalent reports whether two DFAs over the same atom alphabet accept the
// same language. State -1 stands for the dead state.
func Equivalent(a, b *DFA) bool {
	if len(a.Atoms) != len(b.Atoms) {
		return false
	}
	type pair struct{ p, q int }
	seen := map[pair]bool{{a.Start, b.Start}: true}
	work := []pair{{a.Start, b.Start}}
	for len(work) > 0 {
		pr := work[len(work)-1]
		work = work[:len(work)-1]

		accA := pr.p >= 0 && a.Accept[pr.p]
		accB := pr.q >= 0 && b.Accept[pr.q]
		if accA != accB {
			return false
		}
		for i := range a.Atoms {
			na, nb := -1, -1
			if pr.p >= 0 {
				na = a.Delta[pr.p][i]
			}
			if pr.q >= 0 {
				nb = b.Delta[pr.q][i]
			}
			np := pair{na, nb}
			if !seen[np] {
				seen[np] = true
				work = append(work, np)
			}
		}
	}
	return true
}
